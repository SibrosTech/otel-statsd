package statsd

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var _ metric.Meter = &meterImpl{}

type meterImpl struct {
	embedded.Meter

	provider *MeterProvider
	scope    instrumentation.Scope
	pipes    *pipeline

	int64IP   *int64InstProvider
	float64IP *float64InstProvider
}

func newMeter(provider *MeterProvider, scope instrumentation.Scope, p *pipeline) *meterImpl {
	return &meterImpl{
		provider:  provider,
		scope:     scope,
		pipes:     p,
		int64IP:   newInt64InstProvider(provider, p, scope),
		float64IP: newFloat64InstProvider(provider, p, scope),
	}
}

func (m *meterImpl) Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	cfg := metric.NewInt64CounterConfig(options...)
	const kind = sdkmetric.InstrumentKindCounter
	return m.int64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Int64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	cfg := metric.NewInt64UpDownCounterConfig(options...)
	const kind = sdkmetric.InstrumentKindUpDownCounter
	return m.int64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Int64Histogram(name string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	cfg := metric.NewInt64HistogramConfig(options...)
	const kind = sdkmetric.InstrumentKindHistogram
	return m.int64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Int64ObservableCounter(name string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	cfg := metric.NewInt64ObservableCounterConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableCounter
	p := int64ObservProvider{m.int64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil
}

func (m *meterImpl) Int64ObservableUpDownCounter(name string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	cfg := metric.NewInt64ObservableUpDownCounterConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableUpDownCounter
	p := int64ObservProvider{m.int64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil

}

func (m *meterImpl) Int64ObservableGauge(name string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	cfg := metric.NewInt64ObservableGaugeConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableGauge
	p := int64ObservProvider{m.int64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil

}

func (m *meterImpl) Float64Counter(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	cfg := metric.NewFloat64CounterConfig(options...)
	const kind = sdkmetric.InstrumentKindCounter
	return m.float64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Float64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	cfg := metric.NewFloat64UpDownCounterConfig(options...)
	const kind = sdkmetric.InstrumentKindUpDownCounter
	return m.float64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Float64Histogram(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	cfg := metric.NewFloat64HistogramConfig(options...)
	const kind = sdkmetric.InstrumentKindHistogram
	return m.float64IP.lookup(kind, name, cfg.Description(), cfg.Unit())
}

func (m *meterImpl) Float64ObservableCounter(name string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	cfg := metric.NewFloat64ObservableCounterConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableCounter
	p := float64ObservProvider{m.float64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil

}

func (m *meterImpl) Float64ObservableUpDownCounter(name string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	cfg := metric.NewFloat64ObservableUpDownCounterConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableUpDownCounter
	p := float64ObservProvider{m.float64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil
}

func (m *meterImpl) Float64ObservableGauge(name string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	cfg := metric.NewFloat64ObservableGaugeConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableGauge
	p := float64ObservProvider{m.float64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil
}

func (m *meterImpl) RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error) {
	if len(instruments) == 0 {
		// Don't allocate a observer if not needed.
		return noopRegister{}, nil
	}

	reg := newObserver()
	var errs multierror
	for _, inst := range instruments {
		// Unwrap any global.
		if u, ok := inst.(interface {
			Unwrap() metric.Observable
		}); ok {
			inst = u.Unwrap()
		}

		switch o := inst.(type) {
		case int64Observable:
			if err := o.registerable(m.scope); err != nil {
				if !errors.Is(err, errEmptyAgg) {
					errs.append(err)
				}
				continue
			}
			reg.registerInt64(o.observablID)
		case float64Observable:
			if err := o.registerable(m.scope); err != nil {
				if !errors.Is(err, errEmptyAgg) {
					errs.append(err)
				}
				continue
			}
			reg.registerFloat64(o.observablID)
		default:
			// Instrument external to the SDK.
			return nil, fmt.Errorf("invalid observable: from different implementation")
		}
	}

	if err := errs.errorOrNil(); err != nil {
		return nil, err
	}

	if reg.len() == 0 {
		// All insts use drop aggregation.
		return noopRegister{}, nil
	}

	cback := func(ctx context.Context) error {
		return f(ctx, reg)
	}
	return m.pipes.registerMultiCallback(cback), nil
}

type observer struct {
	embedded.Observer

	float64 map[observablID[float64]]struct{}
	int64   map[observablID[int64]]struct{}
}

func newObserver() observer {
	return observer{
		float64: make(map[observablID[float64]]struct{}),
		int64:   make(map[observablID[int64]]struct{}),
	}
}

func (r observer) len() int {
	return len(r.float64) + len(r.int64)
}

func (r observer) registerFloat64(id observablID[float64]) {
	r.float64[id] = struct{}{}
}

func (r observer) registerInt64(id observablID[int64]) {
	r.int64[id] = struct{}{}
}

var (
	errUnknownObserver = errors.New("unknown observable instrument")
	errUnregObserver   = errors.New("observable instrument not registered for callback")
)

func (r observer) ObserveFloat64(o metric.Float64Observable, v float64, opts ...metric.ObserveOption) {
	var oImpl float64Observable
	switch conv := o.(type) {
	case float64Observable:
		oImpl = conv
	case interface {
		Unwrap() metric.Observable
	}:
		// Unwrap any global.
		async := conv.Unwrap()
		var ok bool
		if oImpl, ok = async.(float64Observable); !ok {
			otel.Handle(errUnknownObserver)
			// global.Error(errUnknownObserver, "failed to record asynchronous")
			return
		}
	default:
		otel.Handle(errUnknownObserver)
		// global.Error(errUnknownObserver, "failed to record")
		return
	}

	if _, registered := r.float64[oImpl.observablID]; !registered {
		otel.Handle(errUnregObserver)
		// global.Error(errUnregObserver, "failed to record",
		// 	"name", oImpl.name,
		// 	"description", oImpl.description,
		// 	"unit", oImpl.unit,
		// 	"number", fmt.Sprintf("%T", float64(0)),
		// )
		return
	}
	oImpl.observe(v, opts...)
}

func (r observer) ObserveInt64(o metric.Int64Observable, v int64, opts ...metric.ObserveOption) {
	var oImpl int64Observable
	switch conv := o.(type) {
	case int64Observable:
		oImpl = conv
	case interface {
		Unwrap() metric.Observable
	}:
		// Unwrap any global.
		async := conv.Unwrap()
		var ok bool
		if oImpl, ok = async.(int64Observable); !ok {
			otel.Handle(errUnknownObserver)
			// global.Error(errUnknownObserver, "failed to record asynchronous")
			return
		}
	default:
		otel.Handle(errUnknownObserver)
		// global.Error(errUnknownObserver, "failed to record")
		return
	}

	if _, registered := r.int64[oImpl.observablID]; !registered {
		otel.Handle(errUnregObserver)
		// global.Error(errUnregObserver, "failed to record",
		// 	"name", oImpl.name,
		// 	"description", oImpl.description,
		// 	"unit", oImpl.unit,
		// 	"number", fmt.Sprintf("%T", int64(0)),
		// )
		return
	}
	oImpl.observe(v, opts...)
}

type noopRegister struct {
	embedded.Registration
}

func (noopRegister) Unregister() error {
	return nil
}

// instProvider provides all OpenTelemetry instruments.
type int64InstProvider struct {
	provider *MeterProvider
	pipes    *pipeline
	scope    instrumentation.Scope
}

func newInt64InstProvider(p *MeterProvider, pipes *pipeline, s instrumentation.Scope) *int64InstProvider {
	return &int64InstProvider{provider: p, pipes: pipes, scope: s}
}

// lookup returns the resolved instrumentImpl.
func (p *int64InstProvider) lookup(kind sdkmetric.InstrumentKind, name, desc string, u string) (*int64Inst, error) {
	i := sdkmetric.Instrument{
		Name:        name,
		Description: "", // TODO
		Unit:        u,
		Kind:        kind,
		Scope:       p.scope,
	}
	return &int64Inst{provider: p.provider, instrument: i}, nil
}

// float64InstProvider provides all OpenTelemetry instruments.
type float64InstProvider struct {
	provider *MeterProvider
	pipes    *pipeline
	scope    instrumentation.Scope
}

func newFloat64InstProvider(p *MeterProvider, pipes *pipeline, s instrumentation.Scope) *float64InstProvider {
	return &float64InstProvider{provider: p, pipes: pipes, scope: s}
}

// lookup returns the resolved instrumentImpl.
func (p *float64InstProvider) lookup(kind sdkmetric.InstrumentKind, name, desc string, u string) (*float64Inst, error) {
	i := sdkmetric.Instrument{
		Name:        name,
		Description: "", // TODO
		Unit:        u,
		Kind:        kind,
		Scope:       p.scope,
	}
	return &float64Inst{provider: p.provider, instrument: i}, nil
}

type int64ObservProvider struct{ *int64InstProvider }

func (p int64ObservProvider) lookup(kind sdkmetric.InstrumentKind, name, desc string, u string) (int64Observable, error) {
	return newInt64Observable(p.provider, p.scope, kind, name, desc, u), nil
}

func (p int64ObservProvider) registerCallbacks(inst int64Observable, cBacks []metric.Int64Callback) {
	if inst.observable == nil {
		// Drop.
		return
	}

	for _, cBack := range cBacks {
		p.pipes.addCallback(p.callback(inst, cBack))
	}
}

func (p int64ObservProvider) callback(i int64Observable, f metric.Int64Callback) func(context.Context) error {
	inst := int64Observer{int64Observable: i}
	return func(ctx context.Context) error { return f(ctx, inst) }
}

type int64Observer struct {
	embedded.Int64Observer
	int64Observable
}

func (o int64Observer) Observe(val int64, opts ...metric.ObserveOption) {
	o.observe(val, opts...)
}

type float64ObservProvider struct{ *float64InstProvider }

func (p float64ObservProvider) lookup(kind sdkmetric.InstrumentKind, name, desc string, u string) (float64Observable, error) {
	return newFloat64Observable(p.provider, p.scope, kind, name, desc, u), nil
}

func (p float64ObservProvider) registerCallbacks(inst float64Observable, cBacks []metric.Float64Callback) {
	if inst.observable == nil {
		// Drop aggregator.
		return
	}

	for _, cBack := range cBacks {
		p.pipes.addCallback(p.callback(inst, cBack))
	}
}

func (p float64ObservProvider) callback(i float64Observable, f metric.Float64Callback) func(context.Context) error {
	inst := float64Observer{float64Observable: i}
	return func(ctx context.Context) error { return f(ctx, inst) }
}

type float64Observer struct {
	embedded.Float64Observer
	float64Observable
}

func (o float64Observer) Observe(val float64, opts ...metric.ObserveOption) {
	o.observe(val, opts...)
}
