from .helpers import helm_template

BASE = """
global:
  dnsWildCard: "example.com"
"""

REGISTRY = "registry.example.com/space/edp-tekton"


def component_images(rendered):
    return {
        name: rendered["deployment"][name]["spec"]["template"]["spec"]["containers"][0]
        for name in ("tekton-interceptor", "tekton-reporter")
    }


def test_defaults_render_the_chart_appversion_for_both_components():
    r = helm_template(BASE)

    containers = component_images(r)
    images = {name: c["image"] for name, c in containers.items()}

    assert images["tekton-interceptor"] == images["tekton-reporter"]
    repository, _, tag = images["tekton-reporter"].partition(":")
    assert repository == "epamedp/edp-tekton"
    assert tag

    for container in containers.values():
        assert container["imagePullPolicy"] == "IfNotPresent"


def test_root_image_values_apply_to_both_components():
    """The deploy flow sets image.repository and image.tag only; both components must follow."""
    config = (
        BASE
        + f"""
image:
  repository: {REGISTRY}
  tag: 0.27.0-SNAPSHOT.29
    """
    )

    images = {name: c["image"] for name, c in component_images(helm_template(config)).items()}

    assert images["tekton-interceptor"] == f"{REGISTRY}:0.27.0-SNAPSHOT.29"
    assert images["tekton-reporter"] == f"{REGISTRY}:0.27.0-SNAPSHOT.29"


def test_component_values_override_the_root_per_key():
    config = (
        BASE
        + f"""
image:
  repository: {REGISTRY}
  tag: 0.27.0-SNAPSHOT.29
interceptor:
  image:
    repository: custom/interceptor
reporter:
  image:
    tag: 0.27.0-SNAPSHOT.14
    """
    )

    images = {name: c["image"] for name, c in component_images(helm_template(config)).items()}

    assert images["tekton-interceptor"] == "custom/interceptor:0.27.0-SNAPSHOT.29"
    assert images["tekton-reporter"] == f"{REGISTRY}:0.27.0-SNAPSHOT.14"


def test_root_digest_is_appended_to_both_components():
    config = (
        BASE
        + f"""
image:
  repository: {REGISTRY}
  tag: 0.27.0-SNAPSHOT.29
  digest: sha256:aaaabbbbccccdddd
    """
    )

    images = {name: c["image"] for name, c in component_images(helm_template(config)).items()}

    expected = f"{REGISTRY}:0.27.0-SNAPSHOT.29@sha256:aaaabbbbccccdddd"
    assert images["tekton-interceptor"] == expected
    assert images["tekton-reporter"] == expected


def test_pull_policy_falls_back_to_the_root_value():
    config = (
        BASE
        + """
image:
  pullPolicy: Always
reporter:
  image:
    pullPolicy: Never
    """
    )

    containers = component_images(helm_template(config))

    assert containers["tekton-interceptor"]["imagePullPolicy"] == "Always"
    assert containers["tekton-reporter"]["imagePullPolicy"] == "Never"
