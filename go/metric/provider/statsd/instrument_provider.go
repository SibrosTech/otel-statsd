package statsd

import (
	"errors"

	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// instProvider provides all OpenTelemetry instruments.
type instProvider[N int64 | float64] struct {
	provider *MeterProvider
	scope    instrumentation.Scope
}

func newInstProvider[N int64 | float64](p *MeterProvider, s instrumentation.Scope) *instProvider[N] {
	return &instProvider[N]{provider: p, scope: s}
}

// lookup returns the resolved instrumentImpl.
func (p *instProvider[N]) lookup(kind sdkmetric.InstrumentKind, name string, opts []instrument.Option) (*instrumentImpl[N], error) {
	cfg := instrument.NewConfig(opts...)
	i := sdkmetric.Instrument{
		Name:        name,
		Description: cfg.Description(),
		Unit:        cfg.Unit(),
		Kind:        kind,
		Scope:       p.scope,
	}
	return &instrumentImpl[N]{provider: p.provider, instrument: i}, nil
}

type asyncInt64Provider struct {
	*instProvider[int64]
}

var _ asyncint64.InstrumentProvider = asyncInt64Provider{}

// Counter creates an instrument for recording increasing values.
func (p asyncInt64Provider) Counter(name string, opts ...instrument.Option) (asyncint64.Counter, error) {
	// return p.lookup(sdkmetric.InstrumentKindAsyncCounter, name, opts)
	return nil, errors.New("async metrics not supported")
}

// UpDownCounter creates an instrument for recording changes of a value.
func (p asyncInt64Provider) UpDownCounter(name string, opts ...instrument.Option) (asyncint64.UpDownCounter, error) {
	// return p.lookup(sdkmetric.InstrumentKindAsyncUpDownCounter, name, opts)
	return nil, errors.New("async metrics not supported")
}

// Gauge creates an instrument for recording the current value.
func (p asyncInt64Provider) Gauge(name string, opts ...instrument.Option) (asyncint64.Gauge, error) {
	// return p.lookup(sdkmetric.InstrumentKindAsyncGauge, name, opts)
	return nil, errors.New("async metrics not supported")
}

type asyncFloat64Provider struct {
	*instProvider[float64]
}

var _ asyncfloat64.InstrumentProvider = asyncFloat64Provider{}

// Counter creates an instrument for recording increasing values.
func (p asyncFloat64Provider) Counter(name string, opts ...instrument.Option) (asyncfloat64.Counter, error) {
	// return p.lookup(sdkmetric.InstrumentKindAsyncCounter, name, opts)
	return nil, errors.New("async metrics not supported")
}

// UpDownCounter creates an instrument for recording changes of a value.
func (p asyncFloat64Provider) UpDownCounter(name string, opts ...instrument.Option) (asyncfloat64.UpDownCounter, error) {
	// return p.lookup(sdkmetric.InstrumentKindAsyncUpDownCounter, name, opts)
	return nil, errors.New("async metrics not supported")
}

// Gauge creates an instrument for recording the current value.
func (p asyncFloat64Provider) Gauge(name string, opts ...instrument.Option) (asyncfloat64.Gauge, error) {
	// return p.lookup(sdkmetric.InstrumentKindAsyncGauge, name, opts)
	return nil, errors.New("async metrics not supported")
}

type syncInt64Provider struct {
	*instProvider[int64]
}

var _ syncint64.InstrumentProvider = syncInt64Provider{}

// Counter creates an instrument for recording increasing values.
func (p syncInt64Provider) Counter(name string, opts ...instrument.Option) (syncint64.Counter, error) {
	return p.lookup(sdkmetric.InstrumentKindSyncCounter, name, opts)
}

// UpDownCounter creates an instrument for recording changes of a value.
func (p syncInt64Provider) UpDownCounter(name string, opts ...instrument.Option) (syncint64.UpDownCounter, error) {
	return p.lookup(sdkmetric.InstrumentKindSyncUpDownCounter, name, opts)
}

// Histogram creates an instrument for recording the current value.
func (p syncInt64Provider) Histogram(name string, opts ...instrument.Option) (syncint64.Histogram, error) {
	return p.lookup(sdkmetric.InstrumentKindSyncHistogram, name, opts)
}

type syncFloat64Provider struct {
	*instProvider[float64]
}

var _ syncfloat64.InstrumentProvider = syncFloat64Provider{}

// Counter creates an instrument for recording increasing values.
func (p syncFloat64Provider) Counter(name string, opts ...instrument.Option) (syncfloat64.Counter, error) {
	return p.lookup(sdkmetric.InstrumentKindSyncCounter, name, opts)
}

// UpDownCounter creates an instrument for recording changes of a value.
func (p syncFloat64Provider) UpDownCounter(name string, opts ...instrument.Option) (syncfloat64.UpDownCounter, error) {
	return p.lookup(sdkmetric.InstrumentKindSyncUpDownCounter, name, opts)
}

// Histogram creates an instrument for recording the current value.
func (p syncFloat64Provider) Histogram(name string, opts ...instrument.Option) (syncfloat64.Histogram, error) {
	return p.lookup(sdkmetric.InstrumentKindSyncHistogram, name, opts)
}
