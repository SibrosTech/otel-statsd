package statsd

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/SibrosTech/otel-statsd/go/metric/provider/statsd/mocks"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/metric"
)

// A meter should be able to make instruments concurrently.
func TestMeterInstrumentConcurrency(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(12)

	m := NewMeterProvider().Meter("inst-concurrency")

	go func() {
		_, _ = m.Float64ObservableCounter("AFCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Float64ObservableUpDownCounter("AFUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Float64ObservableGauge("AFGauge")
		wg.Done()
	}()
	go func() {
		_, _ = m.Int64ObservableCounter("AICounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Int64ObservableUpDownCounter("AIUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Int64ObservableGauge("AIGauge")
		wg.Done()
	}()
	go func() {
		_, _ = m.Float64Counter("SFCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Float64UpDownCounter("SFUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Float64Histogram("SFHistogram")
		wg.Done()
	}()
	go func() {
		_, _ = m.Int64Counter("SICounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Int64UpDownCounter("SIUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.Int64Histogram("SIHistogram")
		wg.Done()
	}()

	wg.Wait()
}

var emptyCallback metric.Callback = func(context.Context, metric.Observer) error { return nil }

// A Meter Should be able register Callbacks Concurrently.
func TestMeterCallbackCreationConcurrency(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	m := NewMeterProvider().Meter("callback-concurrency")

	go func() {
		_, _ = m.RegisterCallback(emptyCallback)
		wg.Done()
	}()
	go func() {
		_, _ = m.RegisterCallback(emptyCallback)
		wg.Done()
	}()
	wg.Wait()
}

func TestNoopCallbackUnregisterConcurrency(t *testing.T) {
	m := NewMeterProvider().Meter("noop-unregister-concurrency")
	reg, err := m.RegisterCallback(emptyCallback)
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		_ = reg.Unregister()
		wg.Done()
	}()
	go func() {
		_ = reg.Unregister()
		wg.Done()
	}()
	wg.Wait()
}

func TestCallbackUnregisterConcurrency(t *testing.T) {
	meter := NewMeterProvider().Meter("unregister-concurrency")

	actr, err := meter.Float64ObservableCounter("counter")
	require.NoError(t, err)

	ag, err := meter.Int64ObservableGauge("gauge")
	require.NoError(t, err)

	regCtr, err := meter.RegisterCallback(emptyCallback, actr)
	require.NoError(t, err)

	regG, err := meter.RegisterCallback(emptyCallback, ag)
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		_ = regCtr.Unregister()
		_ = regG.Unregister()
		wg.Done()
	}()
	go func() {
		_ = regCtr.Unregister()
		_ = regG.Unregister()
		wg.Done()
	}()
	wg.Wait()
}

// Instruments should produce correct ResourceMetrics.
func TestMeterCreatesInstruments(t *testing.T) {
	testCases := []struct {
		name string
		fn   func(*testing.T, metric.Meter)
		want []mocks.MockStatSenderMethod
	}{
		{
			name: "ObservableInt64Count",
			fn: func(t *testing.T, m metric.Meter) {
				cback := func(_ context.Context, o metric.Int64Observer) error {
					o.Observe(4)
					return nil
				}
				ctr, err := m.Int64ObservableCounter("aint", metric.WithInt64Callback(cback))
				assert.NoError(t, err)
				_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveInt64(ctr, 3)
					return nil
				}, ctr)
				assert.NoError(t, err)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "aint",
					I:      3,
					F:      1.0,
				},
				{
					Method: "Inc",
					S:      "aint",
					I:      4,
					F:      1.0,
				},
			},
		},
		{
			name: "ObservableInt64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				cback := func(_ context.Context, o metric.Int64Observer) error {
					o.Observe(4)
					return nil
				}
				ctr, err := m.Int64ObservableUpDownCounter("aint", metric.WithInt64Callback(cback))
				assert.NoError(t, err)
				_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveInt64(ctr, 11)
					return nil
				}, ctr)
				assert.NoError(t, err)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "aint",
					I:      11,
					F:      1.0,
				},
				{
					Method: "Inc",
					S:      "aint",
					I:      4,
					F:      1.0,
				},
			},
		},
		{
			name: "ObservableInt64Gauge",
			fn: func(t *testing.T, m metric.Meter) {
				cback := func(_ context.Context, o metric.Int64Observer) error {
					o.Observe(4)
					return nil
				}
				gauge, err := m.Int64ObservableGauge("agauge", metric.WithInt64Callback(cback))
				assert.NoError(t, err)
				_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveInt64(gauge, 11)
					return nil
				}, gauge)
				assert.NoError(t, err)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "agauge",
					I:      11,
					F:      1.0,
				},
				{
					Method: "Inc",
					S:      "agauge",
					I:      4,
					F:      1.0,
				},
			},
		},
		{
			name: "ObservableFloat64Count",
			fn: func(t *testing.T, m metric.Meter) {
				cback := func(_ context.Context, o metric.Float64Observer) error {
					o.Observe(4)
					return nil
				}
				ctr, err := m.Float64ObservableCounter("afloat", metric.WithFloat64Callback(cback))
				assert.NoError(t, err)
				_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveFloat64(ctr, 3)
					return nil
				}, ctr)
				assert.NoError(t, err)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "afloat",
					I:      3,
					F:      1.0,
				},
				{
					Method: "Inc",
					S:      "afloat",
					I:      4,
					F:      1.0,
				},
			},
		},
		{
			name: "ObservableFloat64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				cback := func(_ context.Context, o metric.Float64Observer) error {
					o.Observe(4)
					return nil
				}
				ctr, err := m.Float64ObservableUpDownCounter("afloat", metric.WithFloat64Callback(cback))
				assert.NoError(t, err)
				_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveFloat64(ctr, 11)
					return nil
				}, ctr)
				assert.NoError(t, err)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "afloat",
					I:      11,
					F:      1.0,
				},
				{
					Method: "Inc",
					S:      "afloat",
					I:      4,
					F:      1.0,
				},
			},
		},
		{
			name: "ObservableFloat64Gauge",
			fn: func(t *testing.T, m metric.Meter) {
				cback := func(_ context.Context, o metric.Float64Observer) error {
					o.Observe(4)
					return nil
				}
				gauge, err := m.Float64ObservableGauge("agauge", metric.WithFloat64Callback(cback))
				assert.NoError(t, err)
				_, err = m.RegisterCallback(func(_ context.Context, o metric.Observer) error {
					o.ObserveFloat64(gauge, 11)
					return nil
				}, gauge)
				assert.NoError(t, err)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "agauge",
					I:      11,
					F:      1.0,
				},
				{
					Method: "Inc",
					S:      "agauge",
					I:      4,
					F:      1.0,
				},
			},
		},
		{
			name: "SyncInt64Count",
			fn: func(t *testing.T, m metric.Meter) {
				ctr, err := m.Int64Counter("sint")
				assert.NoError(t, err)

				ctr.Add(context.Background(), 3)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "sint",
					I:      3,
					F:      1.0,
				},
			},
		},
		{
			name: "SyncInt64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				ctr, err := m.Int64UpDownCounter("sint")
				assert.NoError(t, err)

				ctr.Add(context.Background(), 11)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "sint",
					I:      11,
					F:      1.0,
				},
			},
		},
		{
			name: "SyncInt64Histogram",
			fn: func(t *testing.T, m metric.Meter) {
				gauge, err := m.Int64Histogram("histogram")
				assert.NoError(t, err)

				gauge.Record(context.Background(), 7)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Timing",
					S:      "histogram",
					I:      7,
					F:      1.0,
				},
			},
		},
		{
			name: "SyncFloat64Count",
			fn: func(t *testing.T, m metric.Meter) {
				ctr, err := m.Float64Counter("sfloat")
				assert.NoError(t, err)

				ctr.Add(context.Background(), 3)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "sfloat",
					I:      3,
					F:      1.0,
				},
			},
		},
		{
			name: "SyncFloat64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				ctr, err := m.Float64UpDownCounter("sfloat")
				assert.NoError(t, err)

				ctr.Add(context.Background(), 11)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Inc",
					S:      "sfloat",
					I:      11,
					F:      1.0,
				},
			},
		},
		{
			name: "SyncFloat64Histogram",
			fn: func(t *testing.T, m metric.Meter) {
				gauge, err := m.Float64Histogram("histogram")
				assert.NoError(t, err)

				gauge.Record(context.Background(), 7)
			},
			want: []mocks.MockStatSenderMethod{
				{
					Method: "Timing",
					S:      "histogram",
					I:      7,
					F:      1.0,
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			rs := mocks.NewMockStatSender()
			rs.EXPECT(tt.want...)
			mp := NewMeterProvider(WithStatsdClient(rs))
			err := mp.Start(ctx)
			require.NoError(t, err)
			m := mp.Meter("testInstruments")

			tt.fn(t, m)

			err = mp.produce(ctx)
			require.NoError(t, err)

			err = mp.Stop(ctx)
			require.NoError(t, err)

			rs.CHECK(t)
		})
	}
}

func TestRegisterNonSDKObserverErrors(t *testing.T) {
	meter := NewMeterProvider().Meter("scope")

	type obsrv struct{ metric.Observable }
	o := obsrv{}

	_, err := meter.RegisterCallback(
		func(context.Context, metric.Observer) error { return nil },
		o,
	)
	assert.ErrorContains(
		t,
		err,
		"invalid observable: from different implementation",
		"External instrument registred",
	)
}

func TestMeterMixingOnRegisterErrors(t *testing.T) {
	mp := NewMeterProvider()

	m1 := mp.Meter("scope1")
	m2 := mp.Meter("scope2")
	iCtr, err := m2.Int64ObservableCounter("int64 ctr")
	require.NoError(t, err)
	fCtr, err := m2.Float64ObservableCounter("float64 ctr")
	require.NoError(t, err)
	_, err = m1.RegisterCallback(
		func(context.Context, metric.Observer) error { return nil },
		iCtr, fCtr,
	)
	assert.ErrorContains(
		t,
		err,
		`invalid registration: observable "int64 ctr" from Meter "scope2", registered with Meter "scope1"`,
		"Instrument registred with non-creation Meter",
	)
	assert.ErrorContains(
		t,
		err,
		`invalid registration: observable "float64 ctr" from Meter "scope2", registered with Meter "scope1"`,
		"Instrument registred with non-creation Meter",
	)
}

func TestCallbackObserverNonRegistered(t *testing.T) {
	mp := NewMeterProvider()

	m1 := mp.Meter("scope1")
	valid, err := m1.Int64ObservableCounter("ctr")
	require.NoError(t, err)

	m2 := mp.Meter("scope2")
	iCtr, err := m2.Int64ObservableCounter("int64 ctr")
	require.NoError(t, err)
	fCtr, err := m2.Float64ObservableCounter("float64 ctr")
	require.NoError(t, err)

	type int64Obsrv struct{ metric.Int64Observable }
	int64Foreign := int64Obsrv{}
	type float64Obsrv struct{ metric.Float64Observable }
	float64Foreign := float64Obsrv{}

	_, err = m1.RegisterCallback(
		func(_ context.Context, o metric.Observer) error {
			o.ObserveInt64(valid, 1)
			o.ObserveInt64(iCtr, 1)
			o.ObserveFloat64(fCtr, 1)
			o.ObserveInt64(int64Foreign, 1)
			o.ObserveFloat64(float64Foreign, 1)
			return nil
		},
		valid,
	)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		err = mp.produce(context.Background())
	})
	assert.NoError(t, err)

	// var got metricdata.ResourceMetrics
	// assert.NotPanics(t, func() {
	// 	got, err = rdr.Collect(context.Background())
	// })
	//
	// assert.NoError(t, err)
	// want := metricdata.ResourceMetrics{
	// 	Resource: resource.Default(),
	// 	ScopeMetrics: []metricdata.ScopeMetrics{
	// 		{
	// 			Scope: instrumentation.Scope{
	// 				Name: "scope1",
	// 			},
	// 			Metrics: []metricdata.Metrics{
	// 				{
	// 					Name: "ctr",
	// 					Data: metricdata.Sum[int64]{
	// 						Temporality: metricdata.CumulativeTemporality,
	// 						IsMonotonic: true,
	// 						DataPoints: []metricdata.DataPoint[int64]{
	// 							{
	// 								Value: 1,
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}

type logSink struct {
	logr.LogSink

	messages []string
}

func newLogSink(t *testing.T) *logSink {
	return &logSink{LogSink: testr.New(t).GetSink()}
}

func (l *logSink) Info(level int, msg string, keysAndValues ...interface{}) {
	l.messages = append(l.messages, msg)
	l.LogSink.Info(level, msg, keysAndValues...)
}

func (l *logSink) Error(err error, msg string, keysAndValues ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf("%s: %s", err, msg))
	l.LogSink.Error(err, msg, keysAndValues...)
}

func (l *logSink) String() string {
	out := make([]string, len(l.messages))
	for i := range l.messages {
		out[i] = "\t-" + l.messages[i]
	}
	return strings.Join(out, "\n")
}

func TestGlobalInstRegisterCallback(t *testing.T) {
	l := newLogSink(t)
	otel.SetLogger(logr.New(l))

	const mtrName = "TestGlobalInstRegisterCallback"
	preMtr := otel.GetMeterProvider().Meter(mtrName)
	preInt64Ctr, err := preMtr.Int64ObservableCounter("pre.int64.counter")
	require.NoError(t, err)
	preFloat64Ctr, err := preMtr.Float64ObservableCounter("pre.float64.counter")
	require.NoError(t, err)

	mp := NewMeterProvider(WithResource(resource.Empty()))
	otel.SetMeterProvider(mp)

	postMtr := otel.GetMeterProvider().Meter(mtrName)
	postInt64Ctr, err := postMtr.Int64ObservableCounter("post.int64.counter")
	require.NoError(t, err)
	postFloat64Ctr, err := postMtr.Float64ObservableCounter("post.float64.counter")
	require.NoError(t, err)

	cb := func(_ context.Context, o metric.Observer) error {
		o.ObserveInt64(preInt64Ctr, 1)
		o.ObserveFloat64(preFloat64Ctr, 2)
		o.ObserveInt64(postInt64Ctr, 3)
		o.ObserveFloat64(postFloat64Ctr, 4)
		return nil
	}

	_, err = preMtr.RegisterCallback(cb, preInt64Ctr, preFloat64Ctr, postInt64Ctr, postFloat64Ctr)
	assert.NoError(t, err)

	_, err = preMtr.RegisterCallback(cb, preInt64Ctr, preFloat64Ctr, postInt64Ctr, postFloat64Ctr)
	assert.NoError(t, err)

	err = mp.produce(context.Background())
	assert.NoError(t, err)

	// got, err := rdr.Collect(context.Background())
	// assert.NoError(t, err)
	// assert.Lenf(t, l.messages, 0, "Warnings and errors logged:\n%s", l)
	// metricdatatest.AssertEqual(t, metricdata.ResourceMetrics{
	// 	ScopeMetrics: []metricdata.ScopeMetrics{
	// 		{
	// 			Scope: instrumentation.Scope{Name: "TestGlobalInstRegisterCallback"},
	// 			Metrics: []metricdata.Metrics{
	// 				{
	// 					Name: "pre.int64.counter",
	// 					Data: metricdata.Sum[int64]{
	// 						Temporality: metricdata.CumulativeTemporality,
	// 						IsMonotonic: true,
	// 						DataPoints:  []metricdata.DataPoint[int64]{{Value: 1}},
	// 					},
	// 				},
	// 				{
	// 					Name: "pre.float64.counter",
	// 					Data: metricdata.Sum[float64]{
	// 						DataPoints:  []metricdata.DataPoint[float64]{{Value: 2}},
	// 						Temporality: metricdata.CumulativeTemporality,
	// 						IsMonotonic: true,
	// 					},
	// 				},
	// 				{
	// 					Name: "post.int64.counter",
	// 					Data: metricdata.Sum[int64]{
	// 						Temporality: metricdata.CumulativeTemporality,
	// 						IsMonotonic: true,
	// 						DataPoints:  []metricdata.DataPoint[int64]{{Value: 3}},
	// 					},
	// 				},
	// 				{
	// 					Name: "post.float64.counter",
	// 					Data: metricdata.Sum[float64]{
	// 						DataPoints:  []metricdata.DataPoint[float64]{{Value: 4}},
	// 						Temporality: metricdata.CumulativeTemporality,
	// 						IsMonotonic: true,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }, got, metricdatatest.IgnoreTimestamp())
}

func TestMetersProvideScope(t *testing.T) {
	mp := NewMeterProvider()

	m1 := mp.Meter("scope1")
	ctr1, err := m1.Float64ObservableCounter("ctr1")
	assert.NoError(t, err)
	_, err = m1.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveFloat64(ctr1, 5)
		return nil
	}, ctr1)
	assert.NoError(t, err)

	m2 := mp.Meter("scope2")
	ctr2, err := m2.Int64ObservableCounter("ctr2")
	assert.NoError(t, err)
	_, err = m2.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveInt64(ctr2, 7)
		return nil
	}, ctr2)
	assert.NoError(t, err)

	err = mp.produce(context.Background())
	assert.NoError(t, err)

	// want := metricdata.ResourceMetrics{
	// 	Resource: resource.Default(),
	// 	ScopeMetrics: []metricdata.ScopeMetrics{
	// 		{
	// 			Scope: instrumentation.Scope{
	// 				Name: "scope1",
	// 			},
	// 			Metrics: []metricdata.Metrics{
	// 				{
	// 					Name: "ctr1",
	// 					Data: metricdata.Sum[float64]{
	// 						Temporality: metricdata.CumulativeTemporality,
	// 						IsMonotonic: true,
	// 						DataPoints: []metricdata.DataPoint[float64]{
	// 							{
	// 								Value: 5,
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 		{
	// 			Scope: instrumentation.Scope{
	// 				Name: "scope2",
	// 			},
	// 			Metrics: []metricdata.Metrics{
	// 				{
	// 					Name: "ctr2",
	// 					Data: metricdata.Sum[int64]{
	// 						Temporality: metricdata.CumulativeTemporality,
	// 						IsMonotonic: true,
	// 						DataPoints: []metricdata.DataPoint[int64]{
	// 							{
	// 								Value: 7,
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	//
	// got, err := rdr.Collect(context.Background())
	// assert.NoError(t, err)
	// metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}

func TestUnregisterUnregisters(t *testing.T) {
	mp := NewMeterProvider()
	m := mp.Meter("TestUnregisterUnregisters")

	int64Counter, err := m.Int64ObservableCounter("int64.counter")
	require.NoError(t, err)

	int64UpDownCounter, err := m.Int64ObservableUpDownCounter("int64.up_down_counter")
	require.NoError(t, err)

	int64Gauge, err := m.Int64ObservableGauge("int64.gauge")
	require.NoError(t, err)

	floag64Counter, err := m.Float64ObservableCounter("floag64.counter")
	require.NoError(t, err)

	floag64UpDownCounter, err := m.Float64ObservableUpDownCounter("floag64.up_down_counter")
	require.NoError(t, err)

	floag64Gauge, err := m.Float64ObservableGauge("floag64.gauge")
	require.NoError(t, err)

	var called bool
	reg, err := m.RegisterCallback(
		func(context.Context, metric.Observer) error {
			called = true
			return nil
		},
		int64Counter,
		int64UpDownCounter,
		int64Gauge,
		floag64Counter,
		floag64UpDownCounter,
		floag64Gauge,
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = mp.produce(ctx)
	require.NoError(t, err)
	assert.True(t, called, "callback not called for registered callback")

	called = false
	require.NoError(t, reg.Unregister(), "unregister")

	err = mp.produce(ctx)
	require.NoError(t, err)
	assert.False(t, called, "callback called for unregistered callback")
}

var (
	aiCounter       metric.Int64ObservableCounter
	aiUpDownCounter metric.Int64ObservableUpDownCounter
	aiGauge         metric.Int64ObservableGauge

	afCounter       metric.Float64ObservableCounter
	afUpDownCounter metric.Float64ObservableUpDownCounter
	afGauge         metric.Float64ObservableGauge

	siCounter       metric.Int64Counter
	siUpDownCounter metric.Int64UpDownCounter
	siHistogram     metric.Int64Histogram

	sfCounter       metric.Float64Counter
	sfUpDownCounter metric.Float64UpDownCounter
	sfHistogram     metric.Float64Histogram
)

func BenchmarkInstrumentCreation(b *testing.B) {
	ctx := context.Background()

	rs := mocks.NewMockStatSender()
	provider := NewMeterProvider(WithStatsdClient(rs))
	err := provider.Start(ctx)
	require.NoError(b, err)
	defer provider.Stop(ctx)
	meter := provider.Meter("BenchmarkInstrumentCreation")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		aiCounter, _ = meter.Int64ObservableCounter("observable.int64.counter")
		aiUpDownCounter, _ = meter.Int64ObservableUpDownCounter("observable.int64.up.down.counter")
		aiGauge, _ = meter.Int64ObservableGauge("observable.int64.gauge")

		afCounter, _ = meter.Float64ObservableCounter("observable.float64.counter")
		afUpDownCounter, _ = meter.Float64ObservableUpDownCounter("observable.float64.up.down.counter")
		afGauge, _ = meter.Float64ObservableGauge("observable.float64.gauge")

		siCounter, _ = meter.Int64Counter("sync.int64.counter")
		siUpDownCounter, _ = meter.Int64UpDownCounter("sync.int64.up.down.counter")
		siHistogram, _ = meter.Int64Histogram("sync.int64.histogram")

		sfCounter, _ = meter.Float64Counter("sync.float64.counter")
		sfUpDownCounter, _ = meter.Float64UpDownCounter("sync.float64.up.down.counter")
		sfHistogram, _ = meter.Float64Histogram("sync.float64.histogram")
	}
}
