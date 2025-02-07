# SmartThings Exporter

SmartThings Exporter is a Prometheus exporter for SmartThings devices. It fetches device status from the SmartThings API and exposes relevant metrics for monitoring and analysis.

## Features

- Fetches real-time data from SmartThings devices
- Exposes Prometheus-compatible metrics
- Configurable via environment variables
- Supports multiple metrics per device
- Automatic expiration of stale metrics

## Requirements

- Go 1.18+
- SmartThings API token
- Prometheus for scraping metrics

## Installation

```sh
git clone https://github.com/michikrug/smartthings-exporter.git
cd smartthings-exporter
go build -o smartthings-exporter
```

## Configuration

Create a `.env` file or use environment variables:

```ini
SMARTTHINGS_TOKEN=your_smartthings_api_token
DEVICE_ID=your_device_id
DEVICE_NAME=your_device_name
DEVICE_METRICS=temperature,humidity,power
EXPORTER_PORT=9090
COLLECTING_INTERVAL=30
EXPIRATION_THRESHOLD=300
```

### Running the Exporter

```sh
./smartthings-exporter
```

### Running with Docker

```sh
docker build -t smartthings-exporter .
docker run --env-file .env -p 9090:9090 smartthings-exporter
```

## Metrics

The exporter exposes metrics at:

```sh
http://localhost:9090/metrics
```

Example metrics:

```sh
# HELP smartthings_temperature Temperature from SmartThings device
# TYPE smartthings_temperature gauge
smartthings_temperature{device="LivingRoom"} 23.5
```

## License

This project is licensed under the GNU General Public License (GPL). See the LICENSE file for details.

## Contributions

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Contact

For questions, open an issue on [GitHub](https://github.com/michikrug/smartthings-exporter).