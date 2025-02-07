package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SmartThingsResponse struct {
	Components struct {
		Main struct {
			Status map[string]interface{} `json:"status"`
		} `json:"main"`
	} `json:"components"`
}

type Metric struct {
	gauge      prometheus.Gauge
	lastUpdate time.Time
}

type Worker struct {
	client              *http.Client
	smartthingsToken    string
	deviceID            string
	deviceName          string
	deviceMetrics       []string
	metricsRegistry     *prometheus.Registry
	metricsCollector    map[string]*Metric
	collectingInterval  int
	expirationThreshold int
}

func NewWorker(token, deviceID, deviceName string, metrics []string, interval, expiration int, registry *prometheus.Registry) *Worker {
	return &Worker{
		client:              &http.Client{},
		smartthingsToken:    token,
		deviceID:            deviceID,
		deviceName:          deviceName,
		deviceMetrics:       metrics,
		metricsRegistry:     registry,
		metricsCollector:    make(map[string]*Metric),
		collectingInterval:  interval,
		expirationThreshold: expiration,
	}
}

func (w *Worker) fetchData() (*SmartThingsResponse, error) {
	url := fmt.Sprintf("https://api.smartthings.com/v1/devices/%s/status", w.deviceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+w.smartthingsToken)
	req.Header.Set("Accept", "application/json")
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data, status code: %d", resp.StatusCode)
	}

	var data SmartThingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (w *Worker) updateMetrics() {
	for {
		data, err := w.fetchData()
		if err != nil {
			log.Println("Error fetching data:", err)
		} else {
			for _, key := range w.deviceMetrics {
				if value, ok := data.Components.Main.Status[key]; ok {
					if numValue, ok := value.(float64); ok {
						w.setMetric(key, numValue)
					}
				}
			}
		}
		w.clearExpiredMetrics()
		time.Sleep(time.Duration(w.collectingInterval) * time.Second)
	}
}

func (w *Worker) setMetric(key string, value float64) {
	if metric, exists := w.metricsCollector[key]; exists {
		metric.gauge.Set(value)
		metric.lastUpdate = time.Now()
	} else {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("smartthings_%s", key),
			Help: fmt.Sprintf("Metric from SmartThings API: %s", key),
		})
		w.metricsRegistry.MustRegister(gauge)
		gauge.Set(value)
		w.metricsCollector[key] = &Metric{gauge: gauge, lastUpdate: time.Now()}
	}
}

func (w *Worker) clearExpiredMetrics() {
	now := time.Now()
	for key, metric := range w.metricsCollector {
		if now.Sub(metric.lastUpdate).Seconds() > float64(w.expirationThreshold) {
			w.metricsRegistry.Unregister(metric.gauge)
			delete(w.metricsCollector, key)
			log.Printf("Cleared expired metric: %s", key)
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	deviceID := os.Getenv("DEVICE_ID")
	if deviceID == "" {
		log.Fatal("DEVICE_ID is required")
	}

	smartthingsToken := os.Getenv("SMARTTHINGS_TOKEN")
	if smartthingsToken == "" {
		log.Fatal("SMARTTHINGS_TOKEN is required")
	}

	deviceMetricsStr := os.Getenv("DEVICE_METRICS")
	if deviceMetricsStr == "" {
		log.Fatal("DEVICE_METRICS is required")
	}
	deviceMetrics := strings.Split(deviceMetricsStr, ",")

	deviceName := os.Getenv("DEVICE_NAME")
	if deviceName == "" {
		deviceName = deviceID
	}

	collectingInterval := 30
	if interval, exists := os.LookupEnv("COLLECTING_INTERVAL"); exists {
		if i, err := strconv.Atoi(interval); err == nil {
			collectingInterval = i
		} else {
			log.Printf("Invalid COLLECTING_INTERVAL: %s", interval)
		}
	}

	expirationThreshold := 900
	if threshold, exists := os.LookupEnv("EXPIRATION_THRESHOLD"); exists {
		if t, err := strconv.Atoi(threshold); err == nil {
			expirationThreshold = t
		} else {
			log.Printf("Invalid EXPIRATION_THRESHOLD: %s", threshold)
		}
	}

	exporterPort := "9090"
	if port, exists := os.LookupEnv("EXPORTER_PORT"); exists {
		exporterPort = port
	}

	metricsRegistry := prometheus.NewRegistry()
	worker := NewWorker(smartthingsToken, deviceID, deviceName, deviceMetrics, collectingInterval, expirationThreshold, metricsRegistry)
	go worker.updateMetrics()

	http.Handle("/metrics", promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{}))
	server := &http.Server{Addr: ":" + exporterPort}
	go func() {
		log.Printf("Starting HTTP server on port %s", exporterPort)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server failed: %s", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %s", err)
	}
}
