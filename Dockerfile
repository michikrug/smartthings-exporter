# 🏗 Stage 1: Build the Go binary
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o exporter main.go

# 🏗 Stage 2: Create a minimal runtime environment
FROM alpine:latest

LABEL org.opencontainers.image.authors="Michael Krug <michi.krug@gmail.com>"
LABEL org.opencontainers.image.description="An implementation of a Prometheus exporter for SmartThings Devices"
LABEL org.opencontainers.image.source=https://github.com/michikrug/smartthings_exporter
LABEL org.opencontainers.image.licenses=GPL-3.0

# Install CA certificates (needed for HTTPS requests)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/exporter .

USER nobody

# Run the bot
ENTRYPOINT ["/app/exporter"]
