package statsd

import (
	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
)

// config contains configuration for the MeterProvider.
type config struct {
	// Resource is the OpenTelemetry resource associated with all Meters
	// created by the MeterProvider.
	Resource *resource.Resource

	// Statsd client to use
	StatsdClient statsd.StatSender

	// Number of Workers. If <= 0, send synchronously
	Workers int

	// Size of the worker chan buffer. Default is workers * 10
	WorkerChanBufferSize int
}

// Option is the interface that applies the value to a configuration option.
type Option interface {
	// apply sets the Option value of a Config.
	apply(config) config
}

// WithResource sets the Resource configuration option of a Config by merging it
// with the Resource configuration in the environment.
func WithResource(r *resource.Resource) Option {
	return resourceOption{r}
}

type resourceOption struct{ *resource.Resource }

func (o resourceOption) apply(cfg config) config {
	res, err := resource.Merge(cfg.Resource, o.Resource)
	if err != nil {
		otel.Handle(err)
	}
	cfg.Resource = res
	return cfg
}

// WithStatsdClient sets the StatsD client. If not set, a default one will be created.
func WithStatsdClient(s statsd.StatSender) Option {
	return statsdclientOption{s}
}

type statsdclientOption struct{ statsd.StatSender }

func (o statsdclientOption) apply(cfg config) config {
	cfg.StatsdClient = o.StatSender
	return cfg
}

// WithWorkers sets the number of Workers. If <= 0, send synchronously
func WithWorkers(workers int) Option {
	return workersOption{workers}
}

type workersOption struct{ workers int }

func (o workersOption) apply(cfg config) config {
	cfg.Workers = o.workers
	return cfg
}

// WithWorkerChanBufferSize sets the size of the worker chan buffer. Default is workers * 10
func WithWorkerChanBufferSize(workers int) Option {
	return workersOption{workers}
}

type workerChanBufferSizeOption struct{ size int }

func (o workerChanBufferSizeOption) apply(cfg config) config {
	cfg.WorkerChanBufferSize = o.size
	return cfg
}
