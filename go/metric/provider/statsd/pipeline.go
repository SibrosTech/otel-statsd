package statsd

import (
	"container/list"
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
	"go.opentelemetry.io/otel/sdk/resource"
)

func newPipeline(res *resource.Resource) *pipeline {
	if res == nil {
		res = resource.Empty()
	}
	return &pipeline{
		resource: res,
	}
}

// pipeline connects all of the instruments created by a meter provider to a Reader.
// This is the object that will be `Reader.register()` when a meter provider is created.
//
// As instruments are created the instrument should be checked if it exists in the
// views of a the Reader, and if so each aggregator should be added to the pipeline.
type pipeline struct {
	resource *resource.Resource

	sync.Mutex
	callbacks      []func(context.Context) error
	multiCallbacks list.List
}

// addCallback registers a single instrument callback to be run when
// `produce()` is called.
func (p *pipeline) addCallback(cback func(context.Context) error) {
	p.Lock()
	defer p.Unlock()
	p.callbacks = append(p.callbacks, cback)
}

type multiCallback func(context.Context) error

// addMultiCallback registers a multi-instrument callback to be run when
// `produce()` is called.
func (p *pipeline) addMultiCallback(c multiCallback) (unregister func()) {
	p.Lock()
	defer p.Unlock()
	e := p.multiCallbacks.PushBack(c)
	return func() {
		p.Lock()
		p.multiCallbacks.Remove(e)
		p.Unlock()
	}
}

// produce calls all observable
//
// This method is safe to call concurrently.
func (p *pipeline) produce(ctx context.Context) error {
	p.Lock()
	defer p.Unlock()

	var errs multierror
	for _, c := range p.callbacks {
		// TODO make the callbacks parallel. ( #3034 )
		if err := c(ctx); err != nil {
			errs.append(err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	for e := p.multiCallbacks.Front(); e != nil; e = e.Next() {
		// TODO make the callbacks parallel. ( #3034 )
		f := e.Value.(multiCallback)
		if err := f(ctx); err != nil {
			errs.append(err)
		}
		if err := ctx.Err(); err != nil {
			// This means the context expired before we finished running callbacks.
			return err
		}
	}

	return errs.errorOrNil()
}

func (p *pipeline) registerMultiCallback(c multiCallback) metric.Registration {
	unregs := make([]func(), 1)
	unregs[0] = p.addMultiCallback(c)
	return unregisterFuncs{unregs: unregs}
}

type unregisterFuncs struct {
	embedded.Registration

	unregs []func()
}

func (u unregisterFuncs) Unregister() error {
	for _, f := range u.unregs {
		f()
	}
	return nil
}

type multierror struct {
	wrapped error
	errors  []string
}

func (m *multierror) errorOrNil() error {
	if len(m.errors) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", m.wrapped, strings.Join(m.errors, "; "))
}

func (m *multierror) append(err error) {
	m.errors = append(m.errors, err.Error())
}
