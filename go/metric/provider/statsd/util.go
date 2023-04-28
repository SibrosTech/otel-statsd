package statsd

import (
	"github.com/cactus/go-statsd-client/v5/statsd"
	"go.opentelemetry.io/otel/attribute"
)

func collectTags(provider *MeterProvider, attrs attribute.Set) []statsd.Tag {
	var ret []statsd.Tag

	for _, attr := range provider.resource.Attributes() {
		ret = append(ret, statsd.Tag{string(attr.Key), attr.Value.Emit()})
	}
	aiter := attrs.Iter()
	for aiter.Next() {
		ret = append(ret, statsd.Tag{string(aiter.Attribute().Key), aiter.Attribute().Value.Emit()})
	}

	return ret
}
