package statsd

import (
	"context"
	"sync"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/registry"
	"go.opentelemetry.io/otel/sdk/metric/sdkapi"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Controller struct {
	scopes sync.Map

	statsdClient statsd.StatSender
	resource     *resource.Resource
}

var _ metric.MeterProvider = &Controller{}

func New(opts ...Option) *Controller {
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

	ret := &Controller{
		resource: c.Resource,
	}

	if c.Workers > 0 {
		workers := newWorkerStatSender(c.Workers, c.WorkerChanBufferSize, statsdClient)
		statsdClient = workers
	}

	ret.statsdClient = statsdClient
	return ret
}

func (c *Controller) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	cfg := metric.NewMeterConfig(opts...)
	scope := instrumentation.Scope{
		Name:      instrumentationName,
		Version:   cfg.InstrumentationVersion(),
		SchemaURL: cfg.SchemaURL(),
	}

	m, ok := c.scopes.Load(scope)
	if !ok {
		m, _ = c.scopes.LoadOrStore(
			scope,
			registry.NewUniqueInstrumentMeterImpl(&meterImpl{
				controller: c,
			}))
	}
	return sdkapi.WrapMeterImpl(m.(*registry.UniqueInstrumentMeterImpl))
}

func (c *Controller) Start(_ context.Context) error {
	if w, ok := c.statsdClient.(*workerStatSender); ok {
		err := w.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) Stop(_ context.Context) error {
	if w, ok := c.statsdClient.(*workerStatSender); ok {
		return w.Stop()
	}
	return nil
}
