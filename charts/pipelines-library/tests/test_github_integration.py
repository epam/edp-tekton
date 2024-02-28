from .helpers import helm_template


def test_github_is_enabled():
    config = """
global:
  gitProviders:
    - github
gitServers:
  my-github:
    gitProvider: github
    host: github.com
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
    el = r["eventlistener"]["edp-github"]["spec"]

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

    guicklink = r["quicklink"]["github"]["spec"]
    assert "system" == guicklink["type"]
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
