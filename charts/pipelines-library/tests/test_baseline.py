import os
import sys

from .helpers import helm_template


def test_required_resources():
    config = """
kaniko:
  serviceAccount:
    create: true
  roleArn: AWSIRSARoleKaniko
    """

    r = helm_template(config)

    assert "edp-kaniko" in r["serviceaccount"]
    assert "AWSIRSARoleKaniko" in r["serviceaccount"]["edp-kaniko"]["metadata"]["annotations"]["eks.amazonaws.com/role-arn"]

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

    el_Name = "el-gitlab-listener"
    r = helm_template(config)

    assert el_Name in r["ingress"]
    assert "el-gitlab-listener-ns.example.com" in r["ingress"][el_Name]["spec"]["rules"][0]["host"]
    assert el_Name in r["ingress"][el_Name]["spec"]["rules"][0]["http"]["paths"][0]["backend"]["service"]["name"]


def test_ingress_for_github_el():
    config = """
global:
  dnsWildCard: "example.com"
  gitProvider: github
    """

    el_Name = "el-github-listener"
    r = helm_template(config)

    assert el_Name in r["ingress"]
    assert "el-github-listener-ns.example.com" in r["ingress"][el_Name]["spec"]["rules"][0]["host"]
    assert el_Name in r["ingress"][el_Name]["spec"]["rules"][0]["http"]["paths"][0]["backend"]["service"]["name"]


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
    schedule: "0 5 * * *"
    keep: 5
    resources: pipelinerun
    """

    r = helm_template(config)

    assert "0 5 * * *" in r["cronjob"]["tekton-resource-pruner"]["spec"]["schedule"]
    assert " ns;--keep=5;pipelinerun" in r["cronjob"]["tekton-resource-pruner"]["spec"]["jobTemplate"]["spec"]["template"]["spec"]["containers"][0]["args"][1]
