package mocks

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

type MockStatSenderMethod struct {
	Method   string
	S        string
	S2       string
	I        int64
	Duration time.Duration
	F        float32
	Tags     []statsd.Tag
	checked  bool
}

func (mm *MockStatSenderMethod) Call(sender statsd.StatSender) error {
	switch mm.Method {
	case "Inc":
		return sender.Inc(mm.S, mm.I, mm.F, mm.Tags...)
	case "Dec":
		return sender.Dec(mm.S, mm.I, mm.F, mm.Tags...)
	case "Gauge":
		return sender.Gauge(mm.S, mm.I, mm.F, mm.Tags...)
	case "GaugeDelta":
		return sender.GaugeDelta(mm.S, mm.I, mm.F, mm.Tags...)
	case "Timing":
		return sender.Timing(mm.S, mm.I, mm.F, mm.Tags...)
	case "TimingDuration":
		return sender.TimingDuration(mm.S, mm.Duration, mm.F, mm.Tags...)
	case "Set":
		return sender.Set(mm.S, mm.S2, mm.F, mm.Tags...)
	case "SetInt":
		return sender.SetInt(mm.S, mm.I, mm.F, mm.Tags...)
	case "Raw":
		return sender.Raw(mm.S, mm.S2, mm.F, mm.Tags...)
	}
	return fmt.Errorf("unknown method: %s", mm.Method)
}

func (mm *MockStatSenderMethod) Equals(other *MockStatSenderMethod) bool {
	return mm.Method == other.Method &&
		mm.S == other.S &&
		mm.S2 == other.S2 &&
		mm.I == other.I &&
		mm.Duration == other.Duration &&
		mm.F == other.F &&
		mm.EqualTags(other)
}

func (mm *MockStatSenderMethod) EqualTags(other *MockStatSenderMethod) bool {
	// each expected tag must be found, others are ignored
	for _, tag := range mm.Tags {
		found := false
		for _, checktag := range other.Tags {
			if tag[0] == checktag[0] {
				found = true
				if tag[1] != checktag[1] {
					return false
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// MockStatSender is a mock of StatSender interface.
type MockStatSender struct {
	mu      sync.Mutex
	Output  []*MockStatSenderMethod
	expects []*MockStatSenderMethod
}

// NewMockStatSender creates a new mock instance.
func NewMockStatSender() *MockStatSender {
	mock := &MockStatSender{}
	return mock
}

func (m *MockStatSender) EXPECT(method ...MockStatSenderMethod) {
	for _, e := range method {
		mm := new(MockStatSenderMethod)
		*mm = e
		m.expects = append(m.expects, mm)
	}
}

func (m *MockStatSender) CHECK(t *testing.T) {
	for _, out := range m.Output {
		found := false
		for _, mc := range m.expects {
			if !mc.checked && mc.Equals(out) {
				mc.checked = true
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("method not found: %+v", out)
		}
	}

	for _, e := range m.expects {
		if !e.checked {
			t.Errorf("method not called: %+v", e)
		}
	}
}

func (m *MockStatSender) addOutput(method *MockStatSenderMethod) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Output = append(m.Output, method)
}

func (m *MockStatSender) Inc(s string, i int64, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "Inc",
		S:      s,
		I:      i,
		F:      f,
		Tags:   tag,
	})
	return nil
}

func (m *MockStatSender) Dec(s string, i int64, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "Dec",
		S:      s,
		I:      i,
		F:      f,
		Tags:   tag,
	})
	return nil
}

func (m *MockStatSender) Gauge(s string, i int64, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "Gauge",
		S:      s,
		I:      i,
		F:      f,
		Tags:   tag,
	})
	return nil
}

func (m *MockStatSender) GaugeDelta(s string, i int64, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "GaugeDelta",
		S:      s,
		I:      i,
		F:      f,
		Tags:   tag,
	})
	return nil
}

func (m *MockStatSender) Timing(s string, i int64, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "Timing",
		S:      s,
		I:      i,
		F:      f,
		Tags:   tag,
	})
	return nil
}

func (m *MockStatSender) TimingDuration(s string, duration time.Duration, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method:   "TimingDuration",
		S:        s,
		Duration: duration,
		F:        f,
		Tags:     tag,
	})
	return nil
}

func (m *MockStatSender) Set(s string, s2 string, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "Set",
		S:      s,
		S2:     s2,
		F:      f,
		Tags:   tag,
	})
	return nil
}

func (m *MockStatSender) SetInt(s string, i int64, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "SetInt",
		S:      s,
		I:      i,
		F:      f,
		Tags:   tag,
	})
	return nil
}

func (m *MockStatSender) Raw(s string, s2 string, f float32, tag ...statsd.Tag) error {
	m.addOutput(&MockStatSenderMethod{
		Method: "Set",
		S:      s,
		S2:     s2,
		F:      f,
		Tags:   tag,
	})
	return nil
}
