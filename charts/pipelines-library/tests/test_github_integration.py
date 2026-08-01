from subprocess import CalledProcessError

import pytest

from .helpers import helm_template


def test_github_is_enabled():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - github
gitServers:
  my-github:
    gitProvider: github
    host: github.com
    quickLink:
      enabled: true
    webhook:
      skipWebhookSSLVerification: false
    eventListener:
      # -- Enable EventListener
      enabled: true
      # -- EventListener resources
      resources:
        requests:
          memory: "64Mi"
          cpu: "50m"
        limits:
          memory: "128Mi"
          cpu: "500m"
      # -- Node labels for EventListener pod assignment
      nodeSelector: {}
      # -- Tolerations for EventListener pod assignment
      tolerations: []
      # -- Affinity for EventListener pod assignment
      affinity: {}

      ingress:
        # -- Enable ingress controller resource
        enabled: true
        # -- Ingress annotations
        annotations: {}
        # -- Ingress TLS configuration
        tls: []
    """

    r = helm_template(config)

    # Access the event listener using the new structure
    el = r["eventlistener"]["edp-my-github"]["spec"]

    # Check if the triggers are correctly set
    assert "github-build" == el["triggers"][0]["triggerRef"]
    assert "github-review" == el["triggers"][1]["triggerRef"]

    gitserver = r["gitserver"]["my-github"]["spec"]

    assert "github.com" == gitserver["gitHost"]
    assert "github" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "ci-github" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
    assert "webhookUrl" not in gitserver

    guicklink = r["quicklink"]["my-github"]["spec"]
    assert "default" == guicklink["type"]
    assert "https://github.com" == guicklink["url"]

def test_github_build_trigger():
    config = """
global:
  gitProviders:
    - github
    """

    r = helm_template(config)

    # Access the github-build trigger using the new structure
    trigger = r["trigger"]["github-build"]["spec"]

    # Check if the interceptors are correctly set
    assert "github" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-github" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["pull_request"] == trigger["interceptors"][0]["params"][1]["value"]

    # Check if the bindings and template are correctly set
    assert "github-binding-build" == trigger["bindings"][0]["ref"]
    assert "github-build-template" == trigger["template"]["ref"]


def test_github_review_trigger():
    config = """
global:
  gitProviders:
    - github
    """

    r = helm_template(config)

    # Access the github-review trigger using the new structure
    trigger = r["trigger"]["github-review"]["spec"]

    # Check if the interceptors are correctly set
    assert "github" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-github" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["pull_request", "issue_comment"] == trigger["interceptors"][0]["params"][1][
        "value"
    ]

    # Check if the bindings and template are correctly set
    assert "github-binding-review" == trigger["bindings"][0]["ref"]
    assert "github-review-template" == trigger["template"]["ref"]


def test_github_review_trigger_acl_default_filter():
    config = """
global:
  gitProviders:
    - github
    """

    r = helm_template(config)

    trigger = r["trigger"]["github-review"]["spec"]
    cel_filter = trigger["interceptors"][1]["params"][0]["value"]

    assert cel_filter == (
        "(body.action in ['opened', 'synchronize'] && has(body.pull_request)"
        ' && body.pull_request.author_association in ["OWNER","MEMBER","COLLABORATOR"])'
        " || (body.action == 'created' && has(body.comment) && has(body.issue.pull_request)"
        ' && body.comment.author_association in ["OWNER","MEMBER","COLLABORATOR"])'
    )

    # githubOwners is deprecated and disabled by default
    github_params = [p["name"] for p in trigger["interceptors"][0]["params"]]
    assert "githubOwners" not in github_params


def test_github_review_trigger_acl_custom_associations():
    config = """
global:
  gitProviders:
    - github
githubAcl:
  enabled: true
  allowedAssociations:
    - OWNER
    - COLLABORATOR
    """

    r = helm_template(config)

    trigger = r["trigger"]["github-review"]["spec"]
    cel_filter = trigger["interceptors"][1]["params"][0]["value"]

    assert 'in ["OWNER","COLLABORATOR"]' in cel_filter
    assert '"MEMBER"' not in cel_filter


def test_github_review_trigger_acl_disabled_preserves_legacy_filter():
    config = """
global:
  gitProviders:
    - github
githubAcl:
  enabled: false
    """

    r = helm_template(config)

    trigger = r["trigger"]["github-review"]["spec"]
    cel_filter = trigger["interceptors"][1]["params"][0]["value"]

    assert cel_filter == "body.action in ['opened', 'synchronize', 'created']"


def test_github_review_trigger_owners_opt_in():
    config = """
global:
  gitProviders:
    - github
githubOwners:
  enabled: true
    """

    r = helm_template(config)

    trigger = r["trigger"]["github-review"]["spec"]
    github_params = {p["name"]: p["value"] for p in trigger["interceptors"][0]["params"]}

    assert "githubOwners" in github_params
    assert github_params["githubOwners"]["checkType"] == "all"


def test_github_review_trigger_acl_empty_associations_fails_render():
    config = """
global:
  gitProviders:
    - github
githubAcl:
  enabled: true
  allowedAssociations: []
    """

    with pytest.raises(CalledProcessError):
        helm_template(config)


def test_github_review_trigger_acl_null_associations_fails_render():
    config = """
global:
  gitProviders:
    - github
githubAcl:
  enabled: true
  allowedAssociations: null
    """

    with pytest.raises(CalledProcessError):
        helm_template(config)


def test_github_review_trigger_acl_disabled_ignores_empty_associations():
    config = """
global:
  gitProviders:
    - github
githubAcl:
  enabled: false
  allowedAssociations: []
    """

    r = helm_template(config)

    trigger = r["trigger"]["github-review"]["spec"]
    cel_filter = trigger["interceptors"][1]["params"][0]["value"]

    assert cel_filter == "body.action in ['opened', 'synchronize', 'created']"


def test_github_review_trigger_owners_and_acl_combined():
    config = """
global:
  gitProviders:
    - github
githubOwners:
  enabled: true
githubAcl:
  enabled: true
    """

    r = helm_template(config)

    trigger = r["trigger"]["github-review"]["spec"]
    github_params = [p["name"] for p in trigger["interceptors"][0]["params"]]
    cel_filter = trigger["interceptors"][1]["params"][0]["value"]

    assert "githubOwners" in github_params
    assert "author_association" in cel_filter


def test_github_build_trigger_unaffected_by_acl():
    config = """
global:
  gitProviders:
    - github
githubAcl:
  enabled: true
    """

    r = helm_template(config)

    trigger = r["trigger"]["github-build"]["spec"]
    cel_filter = trigger["interceptors"][1]["params"][0]["value"]

    assert cel_filter == "body.action in ['closed'] && body.pull_request.merged == true"
