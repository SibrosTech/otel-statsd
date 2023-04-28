package statsd

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type int64Inst struct {
	embedded.Int64Counter
	embedded.Int64UpDownCounter
	embedded.Int64Histogram

	provider   *MeterProvider
	instrument sdkmetric.Instrument
}

var _ metric.Int64Counter = (*int64Inst)(nil)
var _ metric.Int64UpDownCounter = (*int64Inst)(nil)
var _ metric.Int64Histogram = (*int64Inst)(nil)

func (i *int64Inst) Add(ctx context.Context, val int64, opts ...metric.AddOption) {
	c := metric.NewAddConfig(opts)
	_ = i.provider.statsdClient.Inc(i.instrument.Name, val, 1.0, collectTags(i.provider, c.Attributes())...)
}

func (i *int64Inst) Record(ctx context.Context, val int64, opts ...metric.RecordOption) {
	c := metric.NewRecordConfig(opts)
	_ = i.provider.statsdClient.Timing(i.instrument.Name, int64(val), 1.0, collectTags(i.provider, c.Attributes())...)
}

type float64Inst struct {
	embedded.Float64Counter
	embedded.Float64UpDownCounter
	embedded.Float64Histogram

	provider   *MeterProvider
	instrument sdkmetric.Instrument
}

var _ metric.Float64Counter = (*float64Inst)(nil)
var _ metric.Float64UpDownCounter = (*float64Inst)(nil)
var _ metric.Float64Histogram = (*float64Inst)(nil)

func (i *float64Inst) Add(ctx context.Context, val float64, opts ...metric.AddOption) {
	c := metric.NewAddConfig(opts)
	_ = i.provider.statsdClient.Inc(i.instrument.Name, int64(val), 1.0, collectTags(i.provider, c.Attributes())...)
}

func (i *float64Inst) Record(ctx context.Context, val float64, opts ...metric.RecordOption) {
	c := metric.NewRecordConfig(opts)
	_ = i.provider.statsdClient.Timing(i.instrument.Name, int64(val), 1.0, collectTags(i.provider, c.Attributes())...)
}

// observablID is a comparable unique identifier of an observable.
type observablID[N int64 | float64] struct {
	name        string
	description string
	kind        sdkmetric.InstrumentKind
	unit        string
	scope       instrumentation.Scope
}

type float64Observable struct {
	metric.Float64Observable
	embedded.Float64ObservableCounter
	embedded.Float64ObservableUpDownCounter
	embedded.Float64ObservableGauge

	*observable[float64]
}

var _ metric.Float64ObservableCounter = float64Observable{}
var _ metric.Float64ObservableUpDownCounter = float64Observable{}
var _ metric.Float64ObservableGauge = float64Observable{}

func newFloat64Observable(provider *MeterProvider, scope instrumentation.Scope, kind sdkmetric.InstrumentKind, name, desc string, u string) float64Observable {
	return float64Observable{
		observable: newObservable[float64](provider, scope, kind, name, desc, u),
	}
}

type int64Observable struct {
	metric.Int64Observable
	embedded.Int64ObservableCounter
	embedded.Int64ObservableUpDownCounter
	embedded.Int64ObservableGauge

	*observable[int64]
}

var _ metric.Int64ObservableCounter = int64Observable{}
var _ metric.Int64ObservableUpDownCounter = int64Observable{}
var _ metric.Int64ObservableGauge = int64Observable{}

func newInt64Observable(provider *MeterProvider, scope instrumentation.Scope, kind sdkmetric.InstrumentKind, name, desc string, u string) int64Observable {
	return int64Observable{
		observable: newObservable[int64](provider, scope, kind, name, desc, u),
	}
}

type observable[N int64 | float64] struct {
	metric.Observable
	observablID[N]

	provider *MeterProvider
}

func newObservable[N int64 | float64](provider *MeterProvider, scope instrumentation.Scope, kind sdkmetric.InstrumentKind, name, desc string, u string) *observable[N] {
	return &observable[N]{
		observablID: observablID[N]{
			name:        name,
			description: desc,
			kind:        kind,
			unit:        u,
			scope:       scope,
		},
		provider: provider,
	}
}

// observe records the val for the set of attrs.
func (o *observable[N]) observe(val N, opts ...metric.ObserveOption) {
	c := metric.NewObserveConfig(opts)
	_ = o.provider.statsdClient.Inc(o.name, int64(val), 1.0, collectTags(o.provider, c.Attributes())...)
}

var errEmptyAgg = errors.New("no aggregators for observable instrument")

// registerable returns an error if the observable o should not be registered,
// and nil if it should. An errEmptyAgg error is returned if o is effecively a
// no-op because it does not have any aggregators. Also, an error is returned
// if scope defines a Meter other than the one o was created by.
func (o *observable[N]) registerable(scope instrumentation.Scope) error {
	if scope != o.scope {
		return fmt.Errorf(
			"invalid registration: observable %q from Meter %q, registered with Meter %q",
			o.name,
			o.scope.Name,
			scope.Name,
		)
	}
	return nil
}
