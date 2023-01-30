package statsd

import (
	"context"
	"sync"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
)

type MeterProvider struct {
	scopes sync.Map

	statsdClient statsd.StatSender
	resource     *resource.Resource
}

var _ metric.MeterProvider = &MeterProvider{}

func New(opts ...Option) *MeterProvider {
	c := config{}
	for _, opt := range opts {
		c = opt.apply(c)
	}
	statsdClient := c.StatsdClient
	if statsdClient == nil {
		var err error
		statsdClient, err = statsd.NewClientWithConfig(&statsd.ClientConfig{
			Address:     "127.0.0.1:8125",
			UseBuffered: false,
		})
		if err != nil {
			otel.Handle(err)
		}
	}

	if c.Resource == nil {
		c.Resource = resource.Default()
	} else {
		var err error
		c.Resource, err = resource.Merge(resource.Environment(), c.Resource)
		if err != nil {
			otel.Handle(err)
		}
	}

	ret := &MeterProvider{
		resource: c.Resource,
	}

	if c.Workers > 0 {
		workers := newWorkerStatSender(c.Workers, c.WorkerChanBufferSize, statsdClient)
		statsdClient = workers
	}

	ret.statsdClient = statsdClient
	return ret
}

func (c *MeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	cfg := metric.NewMeterConfig(opts...)
	scope := instrumentation.Scope{
		Name:      instrumentationName,
		Version:   cfg.InstrumentationVersion(),
		SchemaURL: cfg.SchemaURL(),
	}

	return newMeter(c, scope)
}

func (c *MeterProvider) Start(_ context.Context) error {
	if w, ok := c.statsdClient.(*workerStatSender); ok {
		err := w.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *MeterProvider) Stop(_ context.Context) error {
	if w, ok := c.statsdClient.(*workerStatSender); ok {
		return w.Stop()
	}
	return nil
}
