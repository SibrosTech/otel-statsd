package statsd

import (
	"context"
	"errors"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric/number"
	"go.opentelemetry.io/otel/sdk/metric/sdkapi"
)

type meterImpl struct {
	controller *Controller
}

func (m *meterImpl) NewSyncInstrument(descriptor sdkapi.Descriptor) (sdkapi.SyncImpl, error) {
	switch descriptor.InstrumentKind() {
	case sdkapi.CounterInstrumentKind:
		return &counterImpl{
			descriptor: descriptor,
			controller: m.controller,
		}, nil
	case sdkapi.UpDownCounterInstrumentKind:
		return &updownImpl{
			descriptor: descriptor,
			controller: m.controller,
		}, nil
	case sdkapi.HistogramInstrumentKind:
		return &histogramImpl{
			descriptor: descriptor,
			controller: m.controller,
		}, nil
	}
	return nil, errors.New("unsupported metric instrument")
}

func (m *meterImpl) NewAsyncInstrument(descriptor sdkapi.Descriptor) (sdkapi.AsyncImpl, error) {
	return nil, errors.New("async metrics not supported")
}

func (m *meterImpl) RegisterCallback(insts []instrument.Asynchronous, callback func(context.Context)) error {
	return errors.New("async metrics not supported")
}

var _ sdkapi.MeterImpl = &meterImpl{}

// counterImpl
type counterImpl struct {
	instrument.Synchronous
	descriptor sdkapi.Descriptor
	controller *Controller
}

var _ sdkapi.SyncImpl = &counterImpl{}

func (c *counterImpl) Implementation() interface{} {
	return c
}

func (c *counterImpl) Descriptor() sdkapi.Descriptor {
	return c.descriptor
}

func (c *counterImpl) RecordOne(_ context.Context, n number.Number, attrs []attribute.KeyValue) {
	_ = c.controller.statsdClient.Inc(c.descriptor.Name(), n.AsInt64(), 1.0, collectTags(c.controller, attrs)...)
}

// updownImpl
type updownImpl struct {
	instrument.Synchronous
	descriptor sdkapi.Descriptor
	controller *Controller
}

var _ sdkapi.SyncImpl = &updownImpl{}

func (c *updownImpl) Implementation() interface{} {
	return c
}

func (c *updownImpl) Descriptor() sdkapi.Descriptor {
	return c.descriptor
}

func (c *updownImpl) RecordOne(ctx context.Context, n number.Number, attrs []attribute.KeyValue) {
	_ = c.controller.statsdClient.Inc(c.descriptor.Name(), n.AsInt64(), 1.0, collectTags(c.controller, attrs)...)
}

// histogramImpl
type histogramImpl struct {
	instrument.Synchronous
	descriptor sdkapi.Descriptor
	controller *Controller
}

var _ sdkapi.SyncImpl = &histogramImpl{}

func (c *histogramImpl) Implementation() interface{} {
	return c
}

func (c *histogramImpl) Descriptor() sdkapi.Descriptor {
	return c.descriptor
}

func (c *histogramImpl) RecordOne(_ context.Context, n number.Number, attrs []attribute.KeyValue) {
	_ = c.controller.statsdClient.Timing(c.descriptor.Name(), int64(n.AsFloat64()), 1.0, collectTags(c.controller, attrs)...)
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
