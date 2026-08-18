import functools
import pathlib

import pytest
import yaml

from .helpers import helm_template

CHART_DIR = pathlib.Path("charts/pipelines-library")

BASE = """
global:
  dnsWildCard: "example.com"
"""

# A name no default could collide with, so a reference carrying it can only have
# come from the override under test.
SENTINEL = "seam-sentinel-config"


def config_map_seams():
    # Read from values.yaml rather than listed by hand, so a *ConfigMap value
    # added later is covered by these invariants the day it appears instead of
    # waiting for someone to remember this file.
    values = yaml.safe_load((CHART_DIR / "values.yaml").read_text())
    configs = values["tekton"]["configs"]
    return {key: name for key, name in sorted(configs.items()) if key.endswith("ConfigMap")}


SEAMS = config_map_seams()


def config_map_references(node):
    # A ConfigMap reaches a pipeline step three ways, and a seam is only whole
    # if the value feeds all of them: mounted as a volume, pulled in wholesale
    # with envFrom, or read key by key with configMapKeyRef.
    if isinstance(node, dict):
        for key, value in node.items():
            if key in ("configMap", "configMapRef", "configMapKeyRef"):
                if isinstance(value, dict) and "name" in value:
                    yield value["name"]
            yield from config_map_references(value)
    elif isinstance(node, list):
        for item in node:
            yield from config_map_references(item)


@functools.lru_cache(maxsize=None)
def render(config):
    # Rendering the chart costs seconds, and these invariants ask several
    # questions of the same handful of shapes. Cached on the values text, so
    # each shape is rendered once for the whole module.
    return helm_template(config)


def reference_count(config, name):
    rendered = render(config)
    return sum(
        1
        for kind in rendered.values()
        for doc in kind.values()
        for ref in config_map_references(doc)
        if ref == name
    )


def override(seam):
    return BASE + f"\ntekton:\n  configs:\n    {seam}: {SENTINEL}\n"


@pytest.mark.parametrize("seam", sorted(SEAMS))
def test_seam_has_consumers(seam):
    # A seam nothing reads is inert: the value is documented, overriding it
    # changes nothing, and no other assertion here would notice.
    assert reference_count(BASE, SEAMS[seam]) > 0


@pytest.mark.parametrize("seam", sorted(SEAMS))
def test_override_repoints_every_reference(seam):
    # Every reference has to move together. One left hardcoded is worse than a
    # seam that never worked, because the override also stops the stock
    # ConfigMap from rendering - so the straggler points at a ConfigMap that no
    # longer exists and the kubelet fails the pod at container creation.
    expected = reference_count(BASE, SEAMS[seam])

    assert reference_count(override(seam), SENTINEL) == expected


@pytest.mark.parametrize("seam", sorted(SEAMS))
def test_override_leaves_no_reference_to_the_default(seam):
    assert reference_count(override(seam), SEAMS[seam]) == 0


@pytest.mark.parametrize("seam", sorted(SEAMS))
def test_stock_config_map_renders_only_at_the_default(seam):
    # The other half of the seam. The chart ships the stock ConfigMap only
    # while the value is untouched, and never renders the user's own - naming
    # one is the user's promise to supply it.
    assert SEAMS[seam] in render(BASE).get("configmap", {})

    rendered = render(override(seam)).get("configmap", {})
    assert SEAMS[seam] not in rendered
    assert SENTINEL not in rendered
