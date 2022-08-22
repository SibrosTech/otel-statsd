# Copyright The OpenTelemetry Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from unittest import TestCase, mock

from opentelemetry.sdk.extension.statsd.metrics import StatsdMeterProvider
from opentelemetry.sdk.resources import Resource


class StatsdMeterProviderTest(TestCase):
    def test_counter(self):
        statsd = mock.Mock()
        resource = Resource(attributes={
            "service.name": "statsd-test",
        })

        meter_provider = StatsdMeterProvider(statsd=statsd, resource=resource)

        counter = meter_provider.get_meter("name").create_counter("name")
        counter.add(1, attributes={
            "test-counter": 3,
        })

        statsd.increment.assert_called_once_with("name", value=1, tags={'service.name': 'statsd-test', 'test-counter': '3'})

    def test_histogram(self):
        statsd = mock.Mock()
        resource = Resource(attributes={
            "service.name": "statsd-test",
        })

        meter_provider = StatsdMeterProvider(statsd=statsd, resource=resource)

        hist = meter_provider.get_meter("name").create_histogram("name")
        hist.record(2, attributes={
            "test-hist": 3,
        })

        statsd.timing.assert_called_once_with("name", value=2, tags={'service.name': 'statsd-test', 'test-hist': '3'})
