from .helpers import helm_template


def test_required_resources():
    config = """
kaniko:
  roleArn: AWSIRSARoleKaniko
    """

    r = helm_template(config)

    assert "tekton" in r["serviceaccount"]
    assert "AWSIRSARoleKaniko" in r["serviceaccount"]["tekton"]["metadata"]["annotations"]["eks.amazonaws.com/role-arn"]

    assert "tekton-triggers-sa-ns" in r["serviceaccount"]

    assert "tekton-triggers-eventlistener-binding-ns" in r["rolebinding"]
    assert "ns" in r["rolebinding"]["tekton-triggers-eventlistener-binding-ns"]["metadata"]["namespace"]
    assert "ns" in r["rolebinding"]["tekton-triggers-eventlistener-binding-ns"]["subjects"][0]["namespace"]

    assert "tekton-triggers-eventlistener-clusterbinding-ns" in r["clusterrolebinding"]
    assert "ns" in r["clusterrolebinding"]["tekton-triggers-eventlistener-clusterbinding-ns"]["subjects"][0]["namespace"]


def test_ingress_for_gitlab_el():
    config = """
global:
  dnsWildCard: "example.com"
  gitProvider: gitlab
    """

    el_Name = "event-listener"
    r = helm_template(config)

    assert el_Name in r["ingress"]
    assert "el-ns.example.com" in r["ingress"][el_Name]["spec"]["rules"][0]["host"]
    assert "el-edp" in r["ingress"][el_Name]["spec"]["rules"][0]["http"]["paths"][0]["backend"]["service"]["name"]


def test_ingress_for_github_el():
    config = """
global:
  dnsWildCard: "example.com"
  gitProvider: github
    """

    el_Name = "event-listener"
    r = helm_template(config)

    assert el_Name in r["ingress"]
    assert "el-ns.example.com" in r["ingress"][el_Name]["spec"]["rules"][0]["host"]
    assert "el-edp" in r["ingress"][el_Name]["spec"]["rules"][0]["http"]["paths"][0]["backend"]["service"]["name"]


def test_pruner_disabled():
    config = """
tekton:
  pruner:
    create: false
    """

    r = helm_template(config)

    assert "cronjob" not in r
    assert "tekton-resource-pruner" not in r["serviceaccount"]
    assert "tekton-resource-pruner" not in r["role"]
    assert "tekton-resource-pruner" not in r["rolebinding"]


def test_pruner_enabled():
    config = """
tekton:
  pruner:
    create: true
    schedule: "0 * * * *"
    recentMinutes: "30"
    """

    r = helm_template(config)

    assert "0 * * * *" in r["cronjob"]["tekton-resource-pruner"]["spec"]["schedule"]
    assert "/scripts/tekton-prune.sh" in r["cronjob"]["tekton-resource-pruner"]["spec"]["jobTemplate"]["spec"]["template"]["spec"]["containers"][0]["command"][1]
