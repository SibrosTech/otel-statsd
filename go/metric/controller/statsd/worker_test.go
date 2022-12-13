package statsd

import (
	"testing"

	"github.com/SibrosTech/otel-statsd/go/metric/controller/statsd/mocks"
	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/stretchr/testify/require"
)

func TestWorkerStatSender(t *testing.T) {
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

	rs := mocks.NewMockStatSender(len(tests))
	rs.EXPECT(tests...)

	sender := newWorkerStatSender(2, 10, rs)
	err := sender.Start()
	require.NoError(t, err)
	defer sender.Stop()

	for _, test := range tests {
		err := test.Call(sender)
		require.NoError(t, err)
	}

	rs.CHECK(t)
}
