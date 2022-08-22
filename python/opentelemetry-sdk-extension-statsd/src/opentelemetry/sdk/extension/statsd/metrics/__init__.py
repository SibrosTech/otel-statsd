from typing import Optional, Union, Sequence, Dict
from logging import getLogger
from threading import Lock
from statsd import StatsdClient

from opentelemetry.metrics import (
    MeterProvider,
    Meter,
    NoOpMeter,
    CallbackT,
    Counter,
    Histogram,
    ObservableCounter,
    ObservableGauge,
    ObservableUpDownCounter,
    UpDownCounter,
)
from opentelemetry.sdk.resources import Resource
from opentelemetry.util.types import Attributes
from opentelemetry.sdk.util.instrumentation import InstrumentationScope


_logger = getLogger(__name__)


class StatsdMeterProvider(MeterProvider):
    r"""Statsd MeterProvider.

    Args:
        statsd: StatsD client
        resource: The resource representing what the metrics emitted from the SDK pertain to.
    """

    def __init__(
        self,
        statsd: StatsdClient,
        resource: Resource = Resource.create({}),
    ):
        self._lock = Lock()
        self._meter_lock = Lock()
        self._statsd = statsd
        self._resource = resource

        self._meters = {}

    def get_meter(
            self,
            name: str,
            version: Optional[str] = None,
            schema_url: Optional[str] = None,
    ) -> Meter:
        """Returns a StatsdMeter."""
        if not name:
            _logger.warning("Meter name cannot be None or empty.")
            return NoOpMeter(name, version=version, schema_url=schema_url)

        info = InstrumentationScope(name, version, schema_url)
        with self._meter_lock:
            if not self._meters.get(info):
                self._meters[info] = StatsdMeter(
                    self._statsd,
                    self._resource,
                    info,
                )
            return self._meters[info]


class StatsdMeter(Meter):
    """
    Statsd meter implementation.
    """

    def __init__(
            self,
            statsd: StatsdClient,
            resource: Resource,
            instrumentation_scope: InstrumentationScope
    ):
        super().__init__(instrumentation_scope.name, instrumentation_scope.version, instrumentation_scope.schema_url)
        self._statsd = statsd
        self._resource = resource
        self._instrumentation_scope = instrumentation_scope

    def create_counter(
            self,
            name: str,
            unit: str = "",
            description: str = "",
    ) -> Counter:
        if self._is_instrument_registered(
                name, StatsdCounter, unit, description
        )[0]:
            _logger.warning(
                "An instrument with name %s, type %s, unit %s and "
                "description %s has been created already.",
                name,
                Counter.__name__,
                unit,
                description,
            )
        return StatsdCounter(self._statsd, self._resource,
                             name, unit=unit, description=description)

    def create_up_down_counter(
            self,
            name: str,
            unit: str = "",
            description: str = "",
    ) -> UpDownCounter:
        if self._is_instrument_registered(
                name, StatsdUpDownCounter, unit, description
        )[0]:
            _logger.warning(
                "An instrument with name %s, type %s, unit %s and "
                "description %s has been created already.",
                name,
                UpDownCounter.__name__,
                unit,
                description,
            )
        return StatsdUpDownCounter(self._statsd, self._resource,
                                   name, unit=unit, description=description)

    def create_observable_counter(
            self,
            name: str,
            callbacks: Optional[Sequence[CallbackT]] = None,
            unit: str = "",
            description: str = "",
    ) -> ObservableCounter:
        if self._is_instrument_registered(
                name, StatsdObservableCounter, unit, description
        )[0]:
            _logger.warning(
                "An instrument with name %s, type %s, unit %s and "
                "description %s has been created already.",
                name,
                ObservableCounter.__name__,
                unit,
                description,
            )
        return StatsdObservableCounter(
            self._statsd,
            self._resource,
            name,
            callbacks,
            unit=unit,
            description=description,
        )

    def create_histogram(
            self,
            name: str,
            unit: str = "",
            description: str = "",
    ) -> Histogram:
        if self._is_instrument_registered(
                name, StatsdHistogram, unit, description
        )[0]:
            _logger.warning(
                "An instrument with name %s, type %s, unit %s and "
                "description %s has been created already.",
                name,
                Histogram.__name__,
                unit,
                description,
            )
        return StatsdHistogram(self._statsd, self._resource,
                               name, unit=unit, description=description)

    def create_observable_gauge(
            self,
            name: str,
            callbacks: Optional[Sequence[CallbackT]] = None,
            unit: str = "",
            description: str = "",
    ) -> ObservableGauge:
        if self._is_instrument_registered(
                name, StatsdObservableGauge, unit, description
        )[0]:
            _logger.warning(
                "An instrument with name %s, type %s, unit %s and "
                "description %s has been created already.",
                name,
                ObservableGauge.__name__,
                unit,
                description,
            )
        return StatsdObservableGauge(
            self._statsd, self._resource,
            name,
            callbacks,
            unit=unit,
            description=description,
        )

    def create_observable_up_down_counter(
            self,
            name: str,
            callbacks: Optional[Sequence[CallbackT]] = None,
            unit: str = "",
            description: str = "",
    ) -> ObservableUpDownCounter:
        if self._is_instrument_registered(
                name, StatsdObservableUpDownCounter, unit, description
        )[0]:
            _logger.warning(
                "An instrument with name %s, type %s, unit %s and "
                "description %s has been created already.",
                name,
                ObservableUpDownCounter.__name__,
                unit,
                description,
            )
        return StatsdObservableUpDownCounter(
            self._statsd, self._resource,
            name,
            callbacks,
            unit=unit,
            description=description,
        )


class StatsdCounter(Counter):
    """StatsD implementation of `Counter`."""

    def __init__(
            self,
            statsd: StatsdClient,
            resource: Resource,
            name: str,
            unit: str = "",
            description: str = "",
    ) -> None:
        super().__init__(name, unit=unit, description=description)
        self._statsd = statsd
        self._resource = resource
        self._name = name

    def add(
            self,
            amount: Union[int, float],
            attributes: Optional[Attributes] = None,
    ) -> None:
        if amount < 0:
            _logger.warning(
                "Add amount must be non-negative on Counter %s.", self._name
            )
            return
        self._statsd.increment(self._name, value=amount, tags=resource_to_tags(self._resource, attributes))


class StatsdUpDownCounter(UpDownCounter):
    """StatsD implementation of `UpDownCounter`."""

    def __init__(
            self,
            statsd: StatsdClient,
            resource: Resource,
            name: str,
            unit: str = "",
            description: str = "",
    ) -> None:
        super().__init__(name, unit=unit, description=description)
        self._statsd = statsd
        self._resource = resource
        self._name = name

    def add(
            self,
            amount: Union[int, float],
            attributes: Optional[Attributes] = None,
    ) -> None:
        self._statsd.increment(self._name, value=amount, tags=resource_to_tags(self._resource, attributes))


class StatsdObservableCounter(ObservableCounter):
    """StatsD implementation of `ObservableCounter`."""

    def __init__(
            self,
            statsd: StatsdClient,
            resource: Resource,
            name: str,
            callbacks: Optional[Sequence[CallbackT]] = None,
            unit: str = "",
            description: str = "",
    ) -> None:
        super().__init__(name, callbacks, unit=unit, description=description)
        _logger.warning(
            "Observable not supported for Statsd."
        )


class StatsdObservableUpDownCounter(ObservableUpDownCounter):
    """No-op implementation of `ObservableUpDownCounter`."""

    def __init__(
            self,
            statsd: StatsdClient,
            resource: Resource,
            name: str,
            callbacks: Optional[Sequence[CallbackT]] = None,
            unit: str = "",
            description: str = "",
    ) -> None:
        super().__init__(name, callbacks, unit=unit, description=description)
        _logger.warning(
            "Observable not supported for Statsd."
        )


class StatsdHistogram(Histogram):
    """No-op implementation of `Histogram`."""

    def __init__(
            self,
            statsd: StatsdClient,
            resource: Resource,
            name: str,
            unit: str = "",
            description: str = "",
    ) -> None:
        super().__init__(name, unit=unit, description=description)
        self._statsd = statsd
        self._resource = resource
        self._name = name

    def record(
            self,
            amount: Union[int, float],
            attributes: Optional[Attributes] = None,
    ) -> None:
        self._statsd.timing(self._name, value=amount, tags=resource_to_tags(self._resource, attributes))


class StatsdObservableGauge(ObservableGauge):
    """No-op implementation of `ObservableGauge`."""

    def __init__(
            self,
            statsd: StatsdClient,
            resource: Resource,
            name: str,
            callbacks: Optional[Sequence[CallbackT]] = None,
            unit: str = "",
            description: str = "",
    ) -> None:
        super().__init__(name, callbacks, unit=unit, description=description)
        _logger.warning(
            "Observable not supported for Statsd."
        )


def resource_to_tags(resource: Resource, attributes: Optional[Attributes] = None) -> Optional[Dict[str, str]]:
    tags = {}
    for key, value in resource.attributes.items():
        tags[str(key)] = str(value)
    if attributes is not None:
        for key, value in attributes.items():
            tags[str(key)] = str(value)
    if len(tags) == 0:
        return None
    return tags
