package statsd

import (
	"context"
	"errors"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var _ metric.Meter = &meterImpl{}

type meterImpl struct {
	controller *Controller
	scope      instrumentation.Scope
}

func (m *meterImpl) AsyncInt64() asyncint64.InstrumentProvider {
	return &asyncInt64Provider{}
}

func (m *meterImpl) AsyncFloat64() asyncfloat64.InstrumentProvider {
	return &asyncFloat64Provider{}
}

func (m *meterImpl) RegisterCallback(insts []instrument.Asynchronous, function func(context.Context)) error {
	return errors.New("async metrics not supported")
}

func (m *meterImpl) SyncInt64() syncint64.InstrumentProvider {
	return &syncInt64Provider{m.controller, m.scope}
}

func (m *meterImpl) SyncFloat64() syncfloat64.InstrumentProvider {
	// TODO implement me
	panic("implement me")
}

var _ asyncint64.InstrumentProvider = &asyncInt64Provider{}

type asyncInt64Provider struct {
}

func (a *asyncInt64Provider) Counter(name string, opts ...instrument.Option) (asyncint64.Counter, error) {
	return nil, errors.New("async metrics not supported")
}

func (a *asyncInt64Provider) UpDownCounter(name string, opts ...instrument.Option) (asyncint64.UpDownCounter, error) {
	return nil, errors.New("async metrics not supported")
}

func (a *asyncInt64Provider) Gauge(name string, opts ...instrument.Option) (asyncint64.Gauge, error) {
	return nil, errors.New("async metrics not supported")
}

var _ asyncfloat64.InstrumentProvider = &asyncFloat64Provider{}

type asyncFloat64Provider struct {
}

func (a *asyncFloat64Provider) Counter(name string, opts ...instrument.Option) (asyncfloat64.Counter, error) {
	return nil, errors.New("async metrics not supported")
}

func (a *asyncFloat64Provider) UpDownCounter(name string, opts ...instrument.Option) (asyncfloat64.UpDownCounter, error) {
	return nil, errors.New("async metrics not supported")
}

func (a *asyncFloat64Provider) Gauge(name string, opts ...instrument.Option) (asyncfloat64.Gauge, error) {
	return nil, errors.New("async metrics not supported")
}

var _ syncint64.InstrumentProvider = &syncInt64Provider{}

type syncInt64Provider struct {
	controller *Controller
	scope      instrumentation.Scope
}

func (s *syncInt64Provider) Counter(name string, opts ...instrument.Option) (syncint64.Counter, error) {
	cfg := instrument.NewConfig(opts...)
	i := sdkmetric.Instrument{
		Name:        name,
		Description: cfg.Description(),
		Unit:        cfg.Unit(),
		Kind:        sdkmetric.InstrumentKindSyncCounter,
		Scope:       s.scope,
	}
	return &counterImpl{
		controller: s.controller,
		instrument: i,
	}, nil
}

func (s *syncInt64Provider) UpDownCounter(name string, opts ...instrument.Option) (syncint64.UpDownCounter, error) {
	cfg := instrument.NewConfig(opts...)
	i := sdkmetric.Instrument{
		Name:        name,
		Description: cfg.Description(),
		Unit:        cfg.Unit(),
		Kind:        sdkmetric.InstrumentKindSyncUpDownCounter,
		Scope:       s.scope,
	}
	return &updownImpl{
		controller: s.controller,
		instrument: i,
	}, nil
}

func (s *syncInt64Provider) Histogram(name string, opts ...instrument.Option) (syncint64.Histogram, error) {
	cfg := instrument.NewConfig(opts...)
	i := sdkmetric.Instrument{
		Name:        name,
		Description: cfg.Description(),
		Unit:        cfg.Unit(),
		Kind:        sdkmetric.InstrumentKindSyncHistogram,
		Scope:       s.scope,
	}
	return &histogramImpl{
		controller: s.controller,
		instrument: i,
	}, nil
}

// counterImpl
type counterImpl struct {
	instrument.Synchronous
	controller *Controller
	instrument sdkmetric.Instrument
}

var _ syncint64.Counter = &counterImpl{}

func (c *counterImpl) Add(ctx context.Context, incr int64, attrs ...attribute.KeyValue) {
	_ = c.controller.statsdClient.Inc(c.instrument.Name, incr, 1.0, collectTags(c.controller, attrs)...)
}

// updownImpl
type updownImpl struct {
	instrument.Synchronous
	controller *Controller
	instrument sdkmetric.Instrument
}

var _ syncint64.UpDownCounter = &updownImpl{}

func (c *updownImpl) Add(ctx context.Context, incr int64, attrs ...attribute.KeyValue) {
	_ = c.controller.statsdClient.Inc(c.instrument.Name, incr, 1.0, collectTags(c.controller, attrs)...)
}

// histogramImpl
type histogramImpl struct {
	instrument.Synchronous
	controller *Controller
	instrument sdkmetric.Instrument
}

var _ syncint64.Histogram = &histogramImpl{}

func (c *histogramImpl) Record(ctx context.Context, incr int64, attrs ...attribute.KeyValue) {
	_ = c.controller.statsdClient.Timing(c.instrument.Name, incr, 1.0, collectTags(c.controller, attrs)...)
}

var _ syncfloat64.InstrumentProvider = &syncFloat64Provider{}

type syncFloat64Provider struct {
	controller *Controller
	scope      instrumentation.Scope
}

func (s *syncFloat64Provider) Counter(name string, opts ...instrument.Option) (syncfloat64.Counter, error) {
	cfg := instrument.NewConfig(opts...)
	i := sdkmetric.Instrument{
		Name:        name,
		Description: cfg.Description(),
		Unit:        cfg.Unit(),
		Kind:        sdkmetric.InstrumentKindSyncCounter,
		Scope:       s.scope,
	}
	return &counterFloatImpl{
		controller: s.controller,
		instrument: i,
	}, nil
}

func (s *syncFloat64Provider) UpDownCounter(name string, opts ...instrument.Option) (syncfloat64.UpDownCounter, error) {
	cfg := instrument.NewConfig(opts...)
	i := sdkmetric.Instrument{
		Name:        name,
		Description: cfg.Description(),
		Unit:        cfg.Unit(),
		Kind:        sdkmetric.InstrumentKindSyncUpDownCounter,
		Scope:       s.scope,
	}
	return &updownFloatImpl{
		controller: s.controller,
		instrument: i,
	}, nil
}

func (s *syncFloat64Provider) Histogram(name string, opts ...instrument.Option) (syncfloat64.Histogram, error) {
	cfg := instrument.NewConfig(opts...)
	i := sdkmetric.Instrument{
		Name:        name,
		Description: cfg.Description(),
		Unit:        cfg.Unit(),
		Kind:        sdkmetric.InstrumentKindSyncHistogram,
		Scope:       s.scope,
	}
	return &histogramFloatImpl{
		controller: s.controller,
		instrument: i,
	}, nil
}

// counterFloatImpl
type counterFloatImpl struct {
	instrument.Synchronous
	controller *Controller
	instrument sdkmetric.Instrument
}

var _ syncfloat64.Counter = &counterFloatImpl{}

func (c *counterFloatImpl) Add(ctx context.Context, incr float64, attrs ...attribute.KeyValue) {
	_ = c.controller.statsdClient.Inc(c.instrument.Name, int64(incr), 1.0, collectTags(c.controller, attrs)...)
}

// updownFloatImpl
type updownFloatImpl struct {
	instrument.Synchronous
	controller *Controller
	instrument sdkmetric.Instrument
}

var _ syncfloat64.UpDownCounter = &updownFloatImpl{}

func (c *updownFloatImpl) Add(ctx context.Context, incr float64, attrs ...attribute.KeyValue) {
	_ = c.controller.statsdClient.Inc(c.instrument.Name, int64(incr), 1.0, collectTags(c.controller, attrs)...)
}

// histogramFloatImpl
type histogramFloatImpl struct {
	instrument.Synchronous
	controller *Controller
	instrument sdkmetric.Instrument
}

var _ syncfloat64.Histogram = &histogramFloatImpl{}

func (c *histogramFloatImpl) Record(ctx context.Context, incr float64, attrs ...attribute.KeyValue) {
	_ = c.controller.statsdClient.Timing(c.instrument.Name, int64(incr), 1.0, collectTags(c.controller, attrs)...)
}

// helper

func collectTags(controller *Controller, attrs []attribute.KeyValue) []statsd.Tag {
	var ret []statsd.Tag

	for _, attr := range controller.resource.Attributes() {
		ret = append(ret, statsd.Tag{string(attr.Key), attr.Value.Emit()})
	}
	for _, attr := range attrs {
		ret = append(ret, statsd.Tag{string(attr.Key), attr.Value.Emit()})
	}

	return ret
}
