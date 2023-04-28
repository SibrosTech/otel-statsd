package statsd

import (
	"context"
	"sync"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Default periodic reader timing.
const (
	defaultInterval = time.Millisecond * 60000
)

type MeterProvider struct {
	embedded.MeterProvider

	scopes sync.Map
	pipes  *pipeline

	statsdClient statsd.StatSender
	resource     *resource.Resource

	interval time.Duration

	done         chan struct{}
	cancel       context.CancelFunc
	shutdownOnce sync.Once
}

var _ metric.MeterProvider = &MeterProvider{}

func NewMeterProvider(opts ...Option) *MeterProvider {
	c := config{
		Interval: defaultInterval,
	}
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
		pipes:    newPipeline(c.Resource),
		resource: c.Resource,
		interval: c.Interval,
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

	return newMeter(c, scope, c.pipes)
}

func (c *MeterProvider) Start(_ context.Context) error {
	if w, ok := c.statsdClient.(*workerStatSender); ok {
		err := w.Start()
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	c.cancel = cancel
	c.done = make(chan struct{})

	go func() {
		defer func() { close(c.done) }()
		c.run(ctx, c.interval)
	}()

	return nil
}

func (c *MeterProvider) Stop(_ context.Context) error {
	var err error = nil

	c.shutdownOnce.Do(func() {
		// Stop the run loop.
		c.cancel()
		<-c.done

		if w, ok := c.statsdClient.(*workerStatSender); ok {
			err = w.Stop()
		}
	})

	return err
}

func (c *MeterProvider) produce(ctx context.Context) error {
	return c.pipes.produce(ctx)
}

// newTicker allows testing override.
var newTicker = time.NewTicker

// run continuously collects and exports metric data at the specified
// interval. This will run until ctx is canceled or times out.
func (c *MeterProvider) run(ctx context.Context, interval time.Duration) {
	ticker := newTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.produce(ctx)
			if err != nil {
				otel.Handle(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
