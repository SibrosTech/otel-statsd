package statsd

import (
	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel/attribute"
)

func collectTags(provider *MeterProvider, attrs []attribute.KeyValue) []statsd.Tag {
	var ret []statsd.Tag

	for _, attr := range provider.resource.Attributes() {
		ret = append(ret, statsd.Tag{string(attr.Key), attr.Value.Emit()})
	}
	for _, attr := range attrs {
		ret = append(ret, statsd.Tag{string(attr.Key), attr.Value.Emit()})
	}

	return ret
}
