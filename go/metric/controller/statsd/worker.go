package statsd

import (
	"errors"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

// worker
type worker struct {
	statsdClient statsd.StatSender
	input        chan workerJob
	stop         chan struct{}
}

func newWorker(input chan workerJob, statsdClient statsd.StatSender) *worker {
	return &worker{
		statsdClient: statsdClient,
		input:        input,
		stop:         make(chan struct{}),
	}
}

func (w *worker) startReceivingMetric() {
	go w.pullMetric()
}

func (w *worker) stopReceivingMetric() {
	w.stop <- struct{}{}
}

func (w *worker) pullMetric() {
	for {
		select {
		case m := <-w.input:
			_ = m(w.statsdClient)
		case <-w.stop:
			return
		}
	}
}

// workerStatSender
type workerStatSender struct {
	statsdClient statsd.StatSender
	workers      []*worker
	input        chan workerJob
}

func newWorkerStatSender(workers int, bufferSize int, statsdClient statsd.StatSender) *workerStatSender {
	if bufferSize <= 0 {
		bufferSize = workers * 10
	}
	ret := &workerStatSender{
		statsdClient: statsdClient,
		input:        make(chan workerJob, bufferSize),
	}
	for i := 0; i < workers; i++ {
		w := newWorker(ret.input, statsdClient)
		ret.workers = append(ret.workers, w)
	}
	return ret
}

func (w *workerStatSender) Start() error {
	if len(w.workers) == 0 {
		return errors.New("no workers")
	}
	for _, w := range w.workers {
		w.startReceivingMetric()
	}
	return nil
}

func (w *workerStatSender) Stop() error {
	for _, w := range w.workers {
		w.stopReceivingMetric()
	}
	w.workers = nil
	w.flush()
	return nil
}

func (w *workerStatSender) flush() {
	for {
		select {
		case m, ok := <-w.input:
			if ok {
				_ = m(w.statsdClient)
			} else {
				// channel closed
				break
			}
		default:
			return
		}
	}
}

func (w *workerStatSender) Inc(s string, i int64, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.Inc(s, i, f, tag...)
	}
	return nil
}

func (w *workerStatSender) Dec(s string, i int64, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.Dec(s, i, f, tag...)
	}
	return nil
}

func (w *workerStatSender) Gauge(s string, i int64, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.Gauge(s, i, f, tag...)
	}
	return nil
}

func (w *workerStatSender) GaugeDelta(s string, i int64, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.GaugeDelta(s, i, f, tag...)
	}
	return nil
}

func (w *workerStatSender) Timing(s string, i int64, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.Timing(s, i, f, tag...)
	}
	return nil
}

func (w *workerStatSender) TimingDuration(s string, duration time.Duration, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.TimingDuration(s, duration, f, tag...)
	}
	return nil
}

func (w *workerStatSender) Set(s string, s2 string, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.Set(s, s2, f, tag...)
	}
	return nil
}

func (w *workerStatSender) SetInt(s string, i int64, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.SetInt(s, i, f, tag...)
	}
	return nil
}

func (w *workerStatSender) Raw(s string, s2 string, f float32, tag ...statsd.Tag) error {
	w.input <- func(sender statsd.StatSender) error {
		return sender.Raw(s, s2, f, tag...)
	}
	return nil
}

// workerJob
type workerJob func(statsd.StatSender) error
