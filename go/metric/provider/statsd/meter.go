package statsd

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	pipes    *pipeline

	int64IP   *instProvider[int64]
	float64IP *instProvider[float64]
}

func newMeter(provider *MeterProvider, scope instrumentation.Scope, p *pipeline) *meterImpl {
	return &meterImpl{
		provider:  provider,
		scope:     scope,
		pipes:     p,
		int64IP:   newInstProvider[int64](provider, p, scope),
		float64IP: newInstProvider[float64](provider, p, scope),
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
	cfg := instrument.NewInt64ObserverConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableCounter
	p := int64ObservProvider{m.int64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil
}

func (m *meterImpl) Int64ObservableUpDownCounter(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableUpDownCounter, error) {
	cfg := instrument.NewInt64ObserverConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableUpDownCounter
	p := int64ObservProvider{m.int64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil

}

func (m *meterImpl) Int64ObservableGauge(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableGauge, error) {
	cfg := instrument.NewInt64ObserverConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableGauge
	p := int64ObservProvider{m.int64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil

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
	cfg := instrument.NewFloat64ObserverConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableCounter
	p := float64ObservProvider{m.float64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil

}

func (m *meterImpl) Float64ObservableUpDownCounter(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableUpDownCounter, error) {
	cfg := instrument.NewFloat64ObserverConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableUpDownCounter
	p := float64ObservProvider{m.float64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil
}

func (m *meterImpl) Float64ObservableGauge(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableGauge, error) {
	cfg := instrument.NewFloat64ObserverConfig(options...)
	const kind = sdkmetric.InstrumentKindObservableGauge
	p := float64ObservProvider{m.float64IP}
	inst, err := p.lookup(kind, name, cfg.Description(), cfg.Unit())
	if err != nil {
		return nil, err
	}
	p.registerCallbacks(inst, cfg.Callbacks())
	return inst, nil
}

func (m *meterImpl) RegisterCallback(f metric.Callback, instruments ...instrument.Asynchronous) (metric.Registration, error) {
	if len(instruments) == 0 {
		// Don't allocate a observer if not needed.
		return noopRegister{}, nil
	}

	reg := newObserver()
	var errs multierror
	for _, inst := range instruments {
		// Unwrap any global.
		if u, ok := inst.(interface {
			Unwrap() instrument.Asynchronous
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

func (r observer) ObserveFloat64(o instrument.Float64Observable, v float64, a ...attribute.KeyValue) {
	var oImpl float64Observable
	switch conv := o.(type) {
	case float64Observable:
		oImpl = conv
	case interface {
		Unwrap() instrument.Asynchronous
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
	oImpl.observe(v, a)
}

func (r observer) ObserveInt64(o instrument.Int64Observable, v int64, a ...attribute.KeyValue) {
	var oImpl int64Observable
	switch conv := o.(type) {
	case int64Observable:
		oImpl = conv
	case interface {
		Unwrap() instrument.Asynchronous
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
	oImpl.observe(v, a)
}

type noopRegister struct{}

func (noopRegister) Unregister() error {
	return nil
}

// instProvider provides all OpenTelemetry instruments.
type instProvider[N int64 | float64] struct {
	provider *MeterProvider
	pipes    *pipeline
	scope    instrumentation.Scope
}

func newInstProvider[N int64 | float64](p *MeterProvider, pipes *pipeline, s instrumentation.Scope) *instProvider[N] {
	return &instProvider[N]{provider: p, pipes: pipes, scope: s}
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

type int64ObservProvider struct{ *instProvider[int64] }

func (p int64ObservProvider) lookup(kind sdkmetric.InstrumentKind, name, desc string, u unit.Unit) (int64Observable, error) {
	return newInt64Observable(p.provider, p.scope, kind, name, desc, u), nil
}

func (p int64ObservProvider) registerCallbacks(inst int64Observable, cBacks []instrument.Int64Callback) {
	if inst.observable == nil {
		// Drop.
		return
	}

	for _, cBack := range cBacks {
		p.pipes.addCallback(p.callback(inst, cBack))
	}
}

func (p int64ObservProvider) callback(i int64Observable, f instrument.Int64Callback) func(context.Context) error {
	inst := int64Observer{i}
	return func(ctx context.Context) error { return f(ctx, inst) }
}

type int64Observer struct {
	int64Observable
}

func (o int64Observer) Observe(val int64, attrs ...attribute.KeyValue) {
	o.observe(val, attrs)
}

type float64ObservProvider struct{ *instProvider[float64] }

func (p float64ObservProvider) lookup(kind sdkmetric.InstrumentKind, name, desc string, u unit.Unit) (float64Observable, error) {
	return newFloat64Observable(p.provider, p.scope, kind, name, desc, u), nil
}

func (p float64ObservProvider) registerCallbacks(inst float64Observable, cBacks []instrument.Float64Callback) {
	if inst.observable == nil {
		// Drop aggregator.
		return
	}

	for _, cBack := range cBacks {
		p.pipes.addCallback(p.callback(inst, cBack))
	}
}

func (p float64ObservProvider) callback(i float64Observable, f instrument.Float64Callback) func(context.Context) error {
	inst := float64Observer{i}
	return func(ctx context.Context) error { return f(ctx, inst) }
}

type float64Observer struct {
	float64Observable
}

func (o float64Observer) Observe(val float64, attrs ...attribute.KeyValue) {
	o.observe(val, attrs)
}
