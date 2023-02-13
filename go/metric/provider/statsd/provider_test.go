package statsd

import (
	"context"
	"testing"
	"time"

	"github.com/SibrosTech/otel-statsd/go/metric/provider/statsd/mocks"
	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
)

func TestProvider(t *testing.T) {
	tests := []mocks.MockStatSenderMethod{
		{
			Method: "Inc",
			S:      "a.b.c",
			I:      10,
			F:      1.0,
			Tags:   []statsd.Tag{{"x", "y"}},
		},
		{
			Method: "Inc",
			S:      "a.b.c",
			I:      25,
			F:      1.0,
			Tags:   []statsd.Tag{{"x", "y"}},
		},
	}

	ctx := context.Background()

	rs := mocks.NewMockStatSender()
	rs.EXPECT(tests...)

	cont := NewMeterProvider(WithStatsdClient(rs))
	err := cont.Start(ctx)
	require.NoError(t, err)
	defer cont.Stop(ctx)

	meter := cont.Meter("")

	counter, err := meter.Int64Counter("a.b.c")
	require.NoError(t, err)

	for _, test := range tests {
		var attr []attribute.KeyValue
		for _, tag := range test.Tags {
			attr = append(attr, attribute.String(tag[0], tag[1]))
		}
		counter.Add(ctx, test.I, attr...)
	}

	rs.CHECK(t)
}

func TestProviderWithWorkers(t *testing.T) {
	tests := []mocks.MockStatSenderMethod{
		{
			Method: "Inc",
			S:      "a.b.c",
			I:      10,
			F:      1.0,
			Tags:   []statsd.Tag{{"x", "y"}},
		},
		{
			Method: "Inc",
			S:      "a.b.c",
			I:      25,
			F:      1.0,
			Tags:   []statsd.Tag{{"x", "y"}},
		},
	}

	ctx := context.Background()

	rs := mocks.NewMockStatSender()
	rs.EXPECT(tests...)

	cont := NewMeterProvider(WithStatsdClient(rs), WithWorkers(2))
	err := cont.Start(ctx)
	require.NoError(t, err)

	meter := cont.Meter("")

	counter, err := meter.Int64Counter("a.b.c")
	require.NoError(t, err)

	for _, test := range tests {
		var attr []attribute.KeyValue
		for _, tag := range test.Tags {
			attr = append(attr, attribute.String(tag[0], tag[1]))
		}
		counter.Add(ctx, test.I, attr...)
	}

	err = cont.Stop(ctx)
	require.NoError(t, err)

	rs.CHECK(t)
}

func triggerTicker(t *testing.T) chan time.Time {
	t.Helper()

	// Override the ticker C chan so tests are not flaky and rely on timing.
	orig := newTicker
	t.Cleanup(func() { newTicker = orig })

	// Keep this at size zero so when triggered with a send it will hang until
	// the select case is selected and the collection loop is started.
	trigger := make(chan time.Time)
	newTicker = func(d time.Duration) *time.Ticker {
		ticker := time.NewTicker(d)
		ticker.C = trigger
		return ticker
	}
	return trigger
}

func TestProviderRun(t *testing.T) {
	trigger := triggerTicker(t)

	ctx := context.Background()

	rs := mocks.NewMockStatSender()
	rs.EXPECT(mocks.MockStatSenderMethod{
		Method: "Inc",
		S:      "aint",
		I:      4,
		F:      1.0,
	})

	mp := NewMeterProvider(WithStatsdClient(rs))

	m := mp.Meter("testInstruments")

	cback := func(_ context.Context, o instrument.Int64Observer) error {
		o.Observe(4)
		return nil
	}

	_, err := m.Int64ObservableCounter("aint", instrument.WithInt64Callback(cback))
	require.NoError(t, err)

	err = mp.Start(ctx)
	require.NoError(t, err)

	trigger <- time.Now()

	// Cleanup
	err = mp.Stop(ctx)
	require.NoError(t, err)

	// assert
	rs.CHECK(t)
}
