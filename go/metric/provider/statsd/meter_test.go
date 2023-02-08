package statsd

import (
	"context"
	"sync"
	"testing"

	"github.com/SibrosTech/otel-statsd/go/metric/provider/statsd/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
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
				_, err := m.Int64ObservableCounter("aint")
				assert.Error(t, err)
			},
		},
		{
			name: "ObservableInt64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.Int64ObservableUpDownCounter("aint")
				assert.Error(t, err)
			},
		},
		{
			name: "ObservableInt64Gauge",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.Int64ObservableGauge("agauge")
				assert.Error(t, err)
			},
		},
		{
			name: "ObservableFloat64Count",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.Float64ObservableCounter("afloat")
				assert.Error(t, err)
			},
		},
		{
			name: "ObservableFloat64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.Float64ObservableUpDownCounter("afloat")
				assert.Error(t, err)
			},
		},
		{
			name: "ObservableFloat64Gauge",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.Float64ObservableGauge("agauge")
				assert.Error(t, err)
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

			err = mp.Stop(ctx)
			require.NoError(t, err)

			rs.CHECK(t)
		})
	}
}

var (
	aiCounter       instrument.Int64ObservableCounter
	aiUpDownCounter instrument.Int64ObservableUpDownCounter
	aiGauge         instrument.Int64ObservableGauge

	afCounter       instrument.Float64ObservableCounter
	afUpDownCounter instrument.Float64ObservableUpDownCounter
	afGauge         instrument.Float64ObservableGauge

	siCounter       instrument.Int64Counter
	siUpDownCounter instrument.Int64UpDownCounter
	siHistogram     instrument.Int64Histogram

	sfCounter       instrument.Float64Counter
	sfUpDownCounter instrument.Float64UpDownCounter
	sfHistogram     instrument.Float64Histogram
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
