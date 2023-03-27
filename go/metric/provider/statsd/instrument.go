package statsd

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
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

// observablID is a comparable unique identifier of an observable.
type observablID[N int64 | float64] struct {
	name        string
	description string
	kind        sdkmetric.InstrumentKind
	unit        string
	scope       instrumentation.Scope
}

type float64Observable struct {
	instrument.Float64Observable
	*observable[float64]
}

var _ instrument.Float64ObservableCounter = float64Observable{}
var _ instrument.Float64ObservableUpDownCounter = float64Observable{}
var _ instrument.Float64ObservableGauge = float64Observable{}

func newFloat64Observable(provider *MeterProvider, scope instrumentation.Scope, kind sdkmetric.InstrumentKind, name, desc string, u string) float64Observable {
	return float64Observable{
		observable: newObservable[float64](provider, scope, kind, name, desc, u),
	}
}

type int64Observable struct {
	instrument.Int64Observable
	*observable[int64]
}

var _ instrument.Int64ObservableCounter = int64Observable{}
var _ instrument.Int64ObservableUpDownCounter = int64Observable{}
var _ instrument.Int64ObservableGauge = int64Observable{}

func newInt64Observable(provider *MeterProvider, scope instrumentation.Scope, kind sdkmetric.InstrumentKind, name, desc string, u string) int64Observable {
	return int64Observable{
		observable: newObservable[int64](provider, scope, kind, name, desc, u),
	}
}

type observable[N int64 | float64] struct {
	instrument.Asynchronous
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
func (o *observable[N]) observe(val N, attrs []attribute.KeyValue) {
	_ = o.provider.statsdClient.Inc(o.name, int64(val), 1.0, collectTags(o.provider, attrs)...)
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
