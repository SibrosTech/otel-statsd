package statsd

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type instrumentImpl[N int64 | float64] struct {
	instrument.Asynchronous
	instrument.Synchronous

	provider   *MeterProvider
	instrument sdkmetric.Instrument
}

var _ asyncfloat64.Counter = &instrumentImpl[float64]{}
var _ asyncfloat64.UpDownCounter = &instrumentImpl[float64]{}
var _ asyncfloat64.Gauge = &instrumentImpl[float64]{}
var _ asyncint64.Counter = &instrumentImpl[int64]{}
var _ asyncint64.UpDownCounter = &instrumentImpl[int64]{}
var _ asyncint64.Gauge = &instrumentImpl[int64]{}
var _ syncfloat64.Counter = &instrumentImpl[float64]{}
var _ syncfloat64.UpDownCounter = &instrumentImpl[float64]{}
var _ syncfloat64.Histogram = &instrumentImpl[float64]{}
var _ syncint64.Counter = &instrumentImpl[int64]{}
var _ syncint64.UpDownCounter = &instrumentImpl[int64]{}
var _ syncint64.Histogram = &instrumentImpl[int64]{}

func (i *instrumentImpl[N]) Observe(ctx context.Context, val N, attrs ...attribute.KeyValue) {
	panic("async metrics not supported")
}

func (i *instrumentImpl[N]) Add(ctx context.Context, val N, attrs ...attribute.KeyValue) {
	_ = i.provider.statsdClient.Inc(i.instrument.Name, int64(val), 1.0, collectTags(i.provider, attrs)...)
}

func (i *instrumentImpl[N]) Record(ctx context.Context, val N, attrs ...attribute.KeyValue) {
	_ = i.provider.statsdClient.Timing(i.instrument.Name, int64(val), 1.0, collectTags(i.provider, attrs)...)
}
