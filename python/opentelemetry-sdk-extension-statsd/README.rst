OpenTelemetry SDK Extension for StatsD
=======================================================

This library provides components necessary to send OpenTelemetry
metrics to a StatsD backend.

Installation
------------

::

    pip install sb-opentelemetry-sdk-extension-statsd

Usage
-----

Install the OpenTelemetry SDK package.

::

    pip install opentelemetry-sdk

Next, use the provided `StatsdMeterProvider` instead of the default `MeterProvider`.

.. code-block:: python

    from statsd import StatsdClient

    from opentelemetry import metrics
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.sdk.extension.statsd.metrics import StatsdMeterProvider

    statsd = StatsdClient(max_buffer_size=0)
    provider = StatsdMeterProvider(statsd=statsd, resource=Resource(attributes={
        "service.name": "service-Name",
    }))
    metrics.set_meter_provider(provider)


References
----------

* `OpenTelemetry Project <https://opentelemetry.io/>`_
* `OpenTelemetry Specification for Resource Attributes <https://github.com/open-telemetry/opentelemetry-specification/tree/main/specification/resource/semantic_conventions>`_
