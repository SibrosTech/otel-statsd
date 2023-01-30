package statsd

import (
	"context"
	"testing"

	"github.com/SibrosTech/otel-statsd/go/metric/controller/statsd/mocks"
	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
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

	counter, err := meter.SyncInt64().Counter("a.b.c")
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

	counter, err := meter.SyncInt64().Counter("a.b.c")
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
