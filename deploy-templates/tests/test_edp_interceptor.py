from importlib.metadata import metadata
import os
import sys

from .helpers import helm_template

PROJECT = "edp-tekton"
NAME = f"release-name-{PROJECT}"


def test_interceptor_deployment_is_enabled():
    config = """
interceptor:
  deploy: true
    """

    r = helm_template(config)

    assert NAME in r["deployment"]
    assert "deployment" in r

    m = r["deployment"][NAME]["metadata"]
    assert m["name"] == NAME
    assert m["namespace"] == "tekton-pipelines"

    assert r["deployment"][NAME]["spec"]["template"]["spec"]["serviceAccountName"] == "edp-interceptor"

    c = r["deployment"][NAME]["spec"]["template"]["spec"]["containers"][0]
    assert c["name"] == "tekton-triggers-edp-interceptor"
    assert c["image"].startswith("epamedp/" + PROJECT + ":")

    assert c["env"][0]["name"] == "SYSTEM_NAMESPACE"
    assert c["env"][0]["valueFrom"]["fieldRef"]["fieldPath"] == "metadata.namespace"
    assert c["env"][3]["name"] == "METRICS_DOMAIN"
    assert c["env"][3]["value"] == "tekton.dev/triggers"

    assert c["securityContext"]["allowPrivilegeEscalation"] == False
    assert c["securityContext"]["runAsUser"] == 65532
    assert c["securityContext"]["runAsGroup"] == 65532
    # Check ClusterRole and ClusterRoleBinding
    assert "tekton-triggers-edp-interceptor" in r["clusterrole"]
    assert "tekton-pipelines" in r["clusterrolebinding"]["tekton-triggers-edp-interceptor"]["subjects"][0]["namespace"]
    # Check Service
    assert "tekton-pipelines" in r["service"]["tekton-triggers-edp-interceptor"]["metadata"]["namespace"]
    # Check ServiceAccount
    assert "tekton-pipelines" in r["serviceaccount"]["edp-interceptor"]["metadata"]["namespace"]
    # Check interceptor
    assert "tekton-pipelines" in r["clusterinterceptor"]["edp"]["spec"]["clientConfig"]["service"]["namespace"]


def test_interceptor_deployment_is_disabled():
    config = """
interceptor:
  deploy: false
    """

    r = helm_template(config)

    assert "deployment" not in r
    assert "clusterrole" not in r
    assert "tekton-triggers-edp-interceptor" not in r["clusterrolebinding"]
    assert "service" not in r
    assert "tekton-pipelines" not in r["serviceaccount"]
    assert "clusterinterceptor" not in r
