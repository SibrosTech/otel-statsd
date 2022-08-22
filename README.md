# OpenTelemetry StatsD Meter Provider

This repository contains [Python](python/opentelemetry-sdk-extension-statsd) and [Go](go/metric/controller/statsd) implementations of an OpenTelemetry meter provider
that sends metrics to a StatsD client. Usually metrics are sent synchronously using UDP,
which makes it a good option for AWS Lambda metrics.

## Author

Rangel Reale (rangelreale@gmail.com)
