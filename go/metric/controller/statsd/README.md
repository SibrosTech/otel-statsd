# Go OpenTelemetry StatsD Meter Provider

## Usage

```shell
go get github.com/RangelReale/otel-statsd/go/metric/controller/statsd@latest
```

```go
import (
    "go.opentelemetry.io/otel/metric/global"
    "go.opentelemetry.io/otel/sdk/resource"
    "github.com/cactus/go-statsd-client/v5/statsd"
    otel_statsd "github.com/RangelReale/otel-statsd/go/metric/controller/statsd"
)

func main() {
    statsdClient, err := statsd.NewClientWithConfig(&statsd.ClientConfig{
        Address:     "127.0.0.1:8125",
        UseBuffered: false, // for AWS lambda
        TagFormat:   statsd.SuffixOctothorpe, // for OpenTelemetry-Collector's statsd receiver
    })
    if err != nil {
        panic(err)
    }

	r := resource.Default()
	
    pusher := otel_statsd.New(
        otel_statsd.WithStatsdClient(statsdClient),
        otel_statsd.WithResource(r),
    )
    
    global.SetMeterProvider(pusher)
    if err := pusher.Start(ctx); err != nil {
        panic(err)
    }
}

```
