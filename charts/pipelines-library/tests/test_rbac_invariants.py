import pathlib
import re
import subprocess
import tempfile

import pytest
import yaml

from .helpers import helm_template

CHART_DIR = pathlib.Path("charts/pipelines-library")

# Every flag that gates a role, so no role template is invisible to the checks
# below just because its feature is off by default.
EVERYTHING_ENABLED = """
global:
  dnsWildCard: "example.com"
  platform: openshift
interceptor:
  enabled: true
reporter:
  enabled: true
pipelines:
  argocdDiffPreview:
    enabled: true
  autotestsServiceAccount:
    enabled: true
"""

# Resources no role may reach through the Kubernetes API at all, mapped to the
# roles allowed an exception. ConfigMaps reach pipeline steps as volumes,
# envFrom or configMapKeyRef, which the kubelet resolves on the pod's behalf.
FORBIDDEN_RESOURCES = {
    "configmaps": set(),
}

NAME_SCOPED_RESOURCES = {
    "secrets": set(),
}

CONFIGS = {
    "defaults": """
global:
  dnsWildCard: "example.com"
    """,
    "components-enabled": """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - github
    - gerrit
interceptor:
  extraSecretNames:
    - custom-token
    """,
    # The shape most likely to regress: rules whose resourceNames are computed
    # from values render empty here.
    "nothing-configured": """
global:
  dnsWildCard: "example.com"
  gitProviders: []
gitServers: {}
    """,
    # Every optional role needs a shape that renders it, or the invariants
    # never see it.
    "argocd-diff-preview": """
global:
  dnsWildCard: "example.com"
pipelines:
  argocdDiffPreview:
    enabled: true
    """,
    "openshift": """
global:
  dnsWildCard: "example.com"
  platform: openshift
    """,
    "autotests-service-account": """
global:
  dnsWildCard: "example.com"
pipelines:
  autotestsServiceAccount:
    enabled: true
    """,
}


def role_templates():
    for path in sorted(CHART_DIR.glob("templates/**/*.yaml")):
        if re.search(r"^kind:\s*(Cluster)?Role\s*$", path.read_text(), re.MULTILINE):
            yield path.relative_to(CHART_DIR)


def render_template(path, config):
    with tempfile.NamedTemporaryFile("w", suffix=".yaml") as values:
        values.write(config)
        values.flush()
        result = subprocess.run(
            ["helm", "template", "release-name", "-f", values.name, str(CHART_DIR),
             "--namespace=ns", "-s", str(path)],
            capture_output=True,
            text=True,
        )
    if result.returncode:
        return []
    return [d for d in yaml.safe_load_all(result.stdout) if d]


def roles(r):
    return list(r.get("role", {}).values()) + list(r.get("clusterrole", {}).values())


def rule_grants(rule, resource):
    # A wildcard matches the resource just as an explicit entry does, so
    # matching on the literal name alone would wave through the broadest rules
    # of all. Secrets and ConfigMaps live in the core group, spelled as the
    # empty string.
    groups = rule.get("apiGroups") or []
    resources = rule.get("resources") or []
    return ("*" in groups or "" in groups) and ("*" in resources or resource in resources)


def roles_granting(r, resource, unscoped_only=False):
    found = set()
    for role in roles(r):
        for rule in role.get("rules") or []:
            if not rule_grants(rule, resource):
                continue
            # In RBAC an absent or empty resourceNames matches every object of
            # that resource, so a rule that loses its names does not fail
            # closed - it silently grants the whole namespace.
            if unscoped_only and rule.get("resourceNames"):
                continue
            found.add(role["metadata"]["name"])
    return found


@pytest.mark.parametrize("resource", sorted(FORBIDDEN_RESOURCES))
@pytest.mark.parametrize("shape", sorted(CONFIGS))
def test_no_role_reaches_forbidden_resource(shape, resource):
    r = helm_template(CONFIGS[shape])

    assert roles_granting(r, resource) == FORBIDDEN_RESOURCES[resource]


@pytest.mark.parametrize("resource", sorted(NAME_SCOPED_RESOURCES))
@pytest.mark.parametrize("shape", sorted(CONFIGS))
def test_no_role_reads_every_object(shape, resource):
    r = helm_template(CONFIGS[shape])

    assert roles_granting(r, resource, unscoped_only=True) == NAME_SCOPED_RESOURCES[resource]


@pytest.mark.parametrize("path", [str(p) for p in role_templates()])
def test_every_role_template_is_reachable(path):
    # The invariants can only judge a role they have rendered, so a role gated
    # behind a flag none of the shapes sets would never be examined. Reading
    # the templates from disk keeps that reachable set honest without a list
    # to maintain by hand.
    rendered = render_template(path, EVERYTHING_ENABLED)

    assert [d for d in rendered if d.get("kind") in ("Role", "ClusterRole")]


def test_allowlist_entries_still_exist():
    # Guards against an allowlist outliving the role it names, which would
    # leave a stale exception silently permitting a future regression.
    r = helm_template(CONFIGS["defaults"])
    rendered = {role["metadata"]["name"] for role in roles(r)}

    expected = set().union(*FORBIDDEN_RESOURCES.values(), *NAME_SCOPED_RESOURCES.values())
    assert expected <= rendered
