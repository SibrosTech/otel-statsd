package statsd

import (
	"context"
	"sync"
	"testing"

	"github.com/SibrosTech/otel-statsd/go/metric/controller/statsd/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

// A meter should be able to make instruments concurrently.
func TestMeterInstrumentConcurrency(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(12)

	m := NewMeterProvider().Meter("inst-concurrency")

	go func() {
		_, _ = m.AsyncFloat64().Counter("AFCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.AsyncFloat64().UpDownCounter("AFUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.AsyncFloat64().Gauge("AFGauge")
		wg.Done()
	}()
	go func() {
		_, _ = m.AsyncInt64().Counter("AICounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.AsyncInt64().UpDownCounter("AIUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.AsyncInt64().Gauge("AIGauge")
		wg.Done()
	}()
	go func() {
		_, _ = m.SyncFloat64().Counter("SFCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.SyncFloat64().UpDownCounter("SFUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.SyncFloat64().Histogram("SFHistogram")
		wg.Done()
	}()
	go func() {
		_, _ = m.SyncInt64().Counter("SICounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.SyncInt64().UpDownCounter("SIUpDownCounter")
		wg.Done()
	}()
	go func() {
		_, _ = m.SyncInt64().Histogram("SIHistogram")
		wg.Done()
	}()

	wg.Wait()
}

// A Meter Should be able register Callbacks Concurrently.
func TestMeterCallbackCreationConcurrency(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	m := NewMeterProvider().Meter("callback-concurrency")

	go func() {
		_ = m.RegisterCallback([]instrument.Asynchronous{}, func(ctx context.Context) {})
		wg.Done()
	}()
	go func() {
		_ = m.RegisterCallback([]instrument.Asynchronous{}, func(ctx context.Context) {})
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
			name: "AsyncInt64Count",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.AsyncInt64().Counter("aint")
				assert.Error(t, err)
			},
		},
		{
			name: "AsyncInt64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.AsyncInt64().UpDownCounter("aint")
				assert.Error(t, err)
			},
		},
		{
			name: "AsyncInt64Gauge",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.AsyncInt64().Gauge("agauge")
				assert.Error(t, err)
			},
		},
		{
			name: "AsyncFloat64Count",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.AsyncFloat64().Counter("afloat")
				assert.Error(t, err)
			},
		},
		{
			name: "AsyncFloat64UpDownCount",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.AsyncFloat64().UpDownCounter("afloat")
				assert.Error(t, err)
			},
		},
		{
			name: "AsyncFloat64Gauge",
			fn: func(t *testing.T, m metric.Meter) {
				_, err := m.AsyncFloat64().Gauge("agauge")
				assert.Error(t, err)
			},
		},

		{
			name: "SyncInt64Count",
			fn: func(t *testing.T, m metric.Meter) {
				ctr, err := m.SyncInt64().Counter("sint")
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
				ctr, err := m.SyncInt64().UpDownCounter("sint")
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
				gauge, err := m.SyncInt64().Histogram("histogram")
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
				ctr, err := m.SyncFloat64().Counter("sfloat")
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
				ctr, err := m.SyncFloat64().UpDownCounter("sfloat")
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
				gauge, err := m.SyncFloat64().Histogram("histogram")
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
	aiCounter       asyncint64.Counter
	aiUpDownCounter asyncint64.UpDownCounter
	aiGauge         asyncint64.Gauge

	afCounter       asyncfloat64.Counter
	afUpDownCounter asyncfloat64.UpDownCounter
	afGauge         asyncfloat64.Gauge

	siCounter       syncint64.Counter
	siUpDownCounter syncint64.UpDownCounter
	siHistogram     syncint64.Histogram

	sfCounter       syncfloat64.Counter
	sfUpDownCounter syncfloat64.UpDownCounter
	sfHistogram     syncfloat64.Histogram
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
		aiCounter, _ = meter.AsyncInt64().Counter("async.int64.counter")
		aiUpDownCounter, _ = meter.AsyncInt64().UpDownCounter("async.int64.up.down.counter")
		aiGauge, _ = meter.AsyncInt64().Gauge("async.int64.gauge")

		afCounter, _ = meter.AsyncFloat64().Counter("async.float64.counter")
		afUpDownCounter, _ = meter.AsyncFloat64().UpDownCounter("async.float64.up.down.counter")
		afGauge, _ = meter.AsyncFloat64().Gauge("async.float64.gauge")

		siCounter, _ = meter.SyncInt64().Counter("sync.int64.counter")
		siUpDownCounter, _ = meter.SyncInt64().UpDownCounter("sync.int64.up.down.counter")
		siHistogram, _ = meter.SyncInt64().Histogram("sync.int64.histogram")

		sfCounter, _ = meter.SyncFloat64().Counter("sync.float64.counter")
		sfUpDownCounter, _ = meter.SyncFloat64().UpDownCounter("sync.float64.up.down.counter")
		sfHistogram, _ = meter.SyncFloat64().Histogram("sync.float64.histogram")
	}
}
