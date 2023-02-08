package statsd

import (
	"errors"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var _ metric.Meter = &meterImpl{}

type meterImpl struct {
	provider *MeterProvider
	scope    instrumentation.Scope

	int64IP   *instProvider[int64]
	float64IP *instProvider[float64]
}

func newMeter(provider *MeterProvider, scope instrumentation.Scope) *meterImpl {
	return &meterImpl{
		provider:  provider,
		scope:     scope,
		int64IP:   newInstProvider[int64](provider, scope),
		float64IP: newInstProvider[float64](provider, scope),
	}
}

func (m *meterImpl) Int64Counter(name string, options ...instrument.Int64Option) (instrument.Int64Counter, error) {
	cfg := instrument.NewInt64Config(options...)
	const kind = sdkmetric.InstrumentKindCounter
	return m.int64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Int64UpDownCounter(name string, options ...instrument.Int64Option) (instrument.Int64UpDownCounter, error) {
	cfg := instrument.NewInt64Config(options...)
	const kind = sdkmetric.InstrumentKindUpDownCounter
	return m.int64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Int64Histogram(name string, options ...instrument.Int64Option) (instrument.Int64Histogram, error) {
	cfg := instrument.NewInt64Config(options...)
	const kind = sdkmetric.InstrumentKindHistogram
	return m.int64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Int64ObservableCounter(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableCounter, error) {
	return nil, errors.New("async metrics not supported")
}

func (m *meterImpl) Int64ObservableUpDownCounter(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableUpDownCounter, error) {
	return nil, errors.New("async metrics not supported")
}

func (m *meterImpl) Int64ObservableGauge(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableGauge, error) {
	return nil, errors.New("async metrics not supported")
}

func (m *meterImpl) Float64Counter(name string, options ...instrument.Float64Option) (instrument.Float64Counter, error) {
	cfg := instrument.NewFloat64Config(options...)
	const kind = sdkmetric.InstrumentKindCounter
	return m.float64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Float64UpDownCounter(name string, options ...instrument.Float64Option) (instrument.Float64UpDownCounter, error) {
	cfg := instrument.NewFloat64Config(options...)
	const kind = sdkmetric.InstrumentKindUpDownCounter
	return m.float64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Float64Histogram(name string, options ...instrument.Float64Option) (instrument.Float64Histogram, error) {
	cfg := instrument.NewFloat64Config(options...)
	const kind = sdkmetric.InstrumentKindHistogram
	return m.float64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Float64ObservableCounter(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableCounter, error) {
	return nil, errors.New("async metrics not supported")
}

func (m *meterImpl) Float64ObservableUpDownCounter(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableUpDownCounter, error) {
	return nil, errors.New("async metrics not supported")
}

func (m *meterImpl) Float64ObservableGauge(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableGauge, error) {
	return nil, errors.New("async metrics not supported")
}

func (m *meterImpl) RegisterCallback(f metric.Callback, instruments ...instrument.Asynchronous) (metric.Registration, error) {
	return nil, errors.New("async metrics not supported")
}

// instProvider provides all OpenTelemetry instruments.
type instProvider[N int64 | float64] struct {
	provider *MeterProvider
	scope    instrumentation.Scope
}

func newInstProvider[N int64 | float64](p *MeterProvider, s instrumentation.Scope) *instProvider[N] {
	return &instProvider[N]{provider: p, scope: s}
}

// lookup returns the resolved instrumentImpl.
func (p *instProvider[N]) lookup(kind sdkmetric.InstrumentKind, name, desc string, u unit.Unit) (*instrumentImpl[N], error) {
	i := sdkmetric.Instrument{
		Name:        name,
		Description: "", // TODO
		Unit:        u,
		Kind:        kind,
		Scope:       p.scope,
	}
	return &instrumentImpl[N]{provider: p.provider, instrument: i}, nil
}
