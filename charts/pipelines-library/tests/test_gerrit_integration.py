from .helpers import helm_template


def test_gerrit_is_disabled():
    config = """
global:
  gitProviders:
    - unsupported
    """

    r = helm_template(config)

    assert "eventlistener" not in r
    assert "triggerbinding" not in r
    assert "deploy" in r["pipeline"]
    assert "deploy-with-autotests" in r["pipeline"]
    assert "gitserver" not in r


def test_gerrit_is_enabled():
    config = """
global:
  gitProviders:
    - gerrit
  gerritHost: "gerrit"
gitServers:
  my-gerrit:
    gitProvider: gerrit
    host: gerrit.com
    gitUser: ci-user
    nameSshKeySecret: gerrit-ciuser-sshkey
    sshPort: 30100
    quickLink:
      host: gerrit-external.com
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

    gitserver = r["gitserver"]["my-gerrit"]["spec"]

    assert "gerrit.com" == gitserver["gitHost"]
    assert "gerrit" == gitserver["gitProvider"]
    assert "ci-user" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "gerrit-ciuser-sshkey" == gitserver["nameSshKeySecret"]
    assert 30100 == gitserver["sshPort"]

    guicklink = r["quicklink"]["my-gerrit"]["spec"]
    assert "system" == guicklink["type"]
    assert "https://gerrit-external.com" == guicklink["url"]

def test_gerrit_is_enabled_with_custom_port():
    config = """
global:
  gitProviders:
    - gerrit
  gerritSSHPort: "777"
  gerritHost: "gerrit"
    """

    r = helm_template(config)

    gerrit_port_param = {'name': 'GERRIT_PORT', 'value': '777'}
    assert gerrit_port_param in r["pipeline"]["gerrit-go-beego-app-build-default"]["spec"]["tasks"][1]["params"]
    assert gerrit_port_param in r["pipeline"]["gerrit-maven-java11-app-build-default"]["spec"]["tasks"][1]["params"]
    assert gerrit_port_param in r["pipeline"]["gerrit-gradle-java11-app-review"]["spec"]["finally"][1]["params"]

    git_source_url_param = {'name': 'git-source-url', 'value': 'ssh://edp-ci@gerrit:777/$(tt.params.gerritproject)'}
    assert git_source_url_param in r["triggertemplate"]["gerrit-build-template"]["spec"]["resourcetemplates"][0]["spec"]["params"]
    assert git_source_url_param in r["triggertemplate"]["gerrit-review-template"]["spec"]["resourcetemplates"][0]["spec"]["params"]
