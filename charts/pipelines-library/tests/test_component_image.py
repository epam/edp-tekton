from .helpers import helm_template

BASE = """
global:
  dnsWildCard: "example.com"
"""

REGISTRY = "registry.example.com/space/edp-tekton"

COMPONENTS = ("tekton-interceptor", "tekton-reporter")


def containers(rendered):
    """The interceptor and the reporter containers, which run the same image."""
    return {
        name: rendered["deployment"][name]["spec"]["template"]["spec"]["containers"][0]
        for name in COMPONENTS
    }


def images(config):
    return {name: c["image"] for name, c in containers(helm_template(config)).items()}


def test_defaults_render_the_chart_appversion_for_both_components():
    rendered = containers(helm_template(BASE))

    resolved = {name: c["image"] for name, c in rendered.items()}
    assert resolved["tekton-interceptor"] == resolved["tekton-reporter"]

    repository, _, tag = resolved["tekton-reporter"].partition(":")
    assert repository == "epamedp/edp-tekton"
    assert tag

    for container in rendered.values():
        assert container["imagePullPolicy"] == "IfNotPresent"


def test_image_values_apply_to_both_components():
    """The deploy flow sets image.repository and image.tag only; both components must follow."""
    resolved = images(
        BASE
        + f"""
image:
  repository: {REGISTRY}
  tag: 0.27.0-SNAPSHOT.29
"""
    )

    assert resolved["tekton-interceptor"] == f"{REGISTRY}:0.27.0-SNAPSHOT.29"
    assert resolved["tekton-reporter"] == f"{REGISTRY}:0.27.0-SNAPSHOT.29"


def test_digest_is_appended_to_both_components():
    resolved = images(
        BASE
        + f"""
image:
  repository: {REGISTRY}
  tag: 0.27.0-SNAPSHOT.29
  digest: sha256:aaaabbbbccccdddd
"""
    )

    expected = f"{REGISTRY}:0.27.0-SNAPSHOT.29@sha256:aaaabbbbccccdddd"
    assert resolved["tekton-interceptor"] == expected
    assert resolved["tekton-reporter"] == expected


def test_pull_policy_applies_to_both_components():
    rendered = containers(
        helm_template(
            BASE
            + """
image:
  pullPolicy: Always
"""
        )
    )

    for container in rendered.values():
        assert container["imagePullPolicy"] == "Always"
