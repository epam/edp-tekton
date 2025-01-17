from .helpers import helm_template


def test_bitbucket_is_enabled():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - bitbucket
gitServers:
  my-bitbucket:
    gitProvider: bitbucket
    host: bitbucket.com
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

    gitserver = r["gitserver"]["my-bitbucket"]["spec"]
    assert "bitbucket.com" == gitserver["gitHost"]
    assert "bitbucket" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "ci-bitbucket" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
    assert "https://my-custom-ingress-name.example.com" == gitserver["webhookUrl"]

    guicklink = r["quicklink"]["my-bitbucket"]["spec"]
    assert "default" == guicklink["type"]
    assert "https://bitbucket.com" == guicklink["url"]

def test_bitbucket_build_trigger():
    config = """
global:
  gitProviders:
    - bitbucket
    """

    r = helm_template(config)

    # Access the bitbucket-build trigger using the new structure
    trigger = r["trigger"]["bitbucket-build"]["spec"]

    # Check if the interceptors are correctly set
    assert "bitbucket" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-bitbucket" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["pullrequest:fulfilled"] == trigger["interceptors"][0]["params"][1]["value"]

    # Check if the bindings and template are correctly set
    assert "bitbucket-binding-build" == trigger["bindings"][0]["ref"]
    assert "bitbucket-build-template" == trigger["template"]["ref"]


def test_bitbucket_review_trigger():
    config = """
global:
  gitProviders:
    - bitbucket
    """

    r = helm_template(config)

    # Access the bitbucket-review trigger using the new structure
    trigger = r["trigger"]["bitbucket-review"]["spec"]

    # Check if the interceptors are correctly set
    assert "bitbucket" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-bitbucket" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["pullrequest:created", "pullrequest:comment_created", "pullrequest:updated"] == trigger["interceptors"][0]["params"][1]["value"]

    # Check if the bindings and template are correctly set
    assert "bitbucket-binding-review" == trigger["bindings"][0]["ref"]
    assert "bitbucket-review-template" == trigger["template"]["ref"]
