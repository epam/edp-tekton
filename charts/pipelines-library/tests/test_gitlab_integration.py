from .helpers import helm_template


def test_gitlab_is_enabled():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - gitlab
gitServers:
  my-gitlab:
    gitProvider: gitlab
    host: gitlab.com
    quickLink:
      enabled: true
    webhook:
      url: https://my-custom-ingress-name.example.com
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

    gitserver = r["gitserver"]["my-gitlab"]["spec"]
    assert "gitlab.com" == gitserver["gitHost"]
    assert "gitlab" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "ci-gitlab" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
    assert "https://my-custom-ingress-name.example.com" == gitserver["webhookUrl"]

    guicklink = r["quicklink"]["my-gitlab"]["spec"]
    assert "default" == guicklink["type"]
    assert "https://gitlab.com" == guicklink["url"]

def test_gitlab_build_trigger():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)

    # Access the gitlab-build trigger using the new structure
    trigger = r["trigger"]["gitlab-build"]["spec"]

    # Check if the interceptors are correctly set
    assert "gitlab" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-gitlab" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["Merge Request Hook"] == trigger["interceptors"][0]["params"][1]["value"]

    # Check if the bindings and template are correctly set
    assert "gitlab-binding-build" == trigger["bindings"][0]["ref"]
    assert "gitlab-build-template" == trigger["template"]["ref"]


def test_gitlab_review_trigger():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)

    # Access the gitlab-review trigger using the new structure
    trigger = r["trigger"]["gitlab-review"]["spec"]

    # Check if the interceptors are correctly set
    assert "gitlab" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-gitlab" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["Merge Request Hook", "Note Hook"] == trigger["interceptors"][0]["params"][1]["value"]

    # Check if the bindings and template are correctly set
    assert "gitlab-binding-review" == trigger["bindings"][0]["ref"]
    assert "gitlab-review-template" == trigger["template"]["ref"]
