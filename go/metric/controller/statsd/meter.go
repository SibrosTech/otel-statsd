package statsd

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var _ metric.Meter = &meterImpl{}

type meterImpl struct {
	provider *MeterProvider
	scope    instrumentation.Scope
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
	return &syncInt64Provider{m.provider, m.scope}
}

func (m *meterImpl) SyncFloat64() syncfloat64.InstrumentProvider {
	return &syncFloat64Provider{m.provider, m.scope}
}
