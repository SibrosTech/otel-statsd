package statsd

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type instrumentImpl[N int64 | float64] struct {
	instrument.Synchronous

	provider   *MeterProvider
	instrument sdkmetric.Instrument
}

var _ instrument.Float64Counter = (*instrumentImpl[float64])(nil)
var _ instrument.Float64UpDownCounter = (*instrumentImpl[float64])(nil)
var _ instrument.Float64Histogram = (*instrumentImpl[float64])(nil)
var _ instrument.Int64Counter = (*instrumentImpl[int64])(nil)
var _ instrument.Int64UpDownCounter = (*instrumentImpl[int64])(nil)
var _ instrument.Int64Histogram = (*instrumentImpl[int64])(nil)

func (i *instrumentImpl[N]) Add(ctx context.Context, val N, attrs ...attribute.KeyValue) {
	_ = i.provider.statsdClient.Inc(i.instrument.Name, int64(val), 1.0, collectTags(i.provider, attrs)...)
}

func (i *instrumentImpl[N]) Record(ctx context.Context, val N, attrs ...attribute.KeyValue) {
	_ = i.provider.statsdClient.Timing(i.instrument.Name, int64(val), 1.0, collectTags(i.provider, attrs)...)
}
