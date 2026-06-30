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
  gitProviders:
    - gitlab
gitServers:
  my-gitlab:
    gitProvider: gitlab
    host: gitlab.com
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

    el_Name = "event-listener-my-gitlab"
    r = helm_template(config)

    assert el_Name in r["ingress"]
    assert "el-my-gitlab-ns.example.com" in r["ingress"][el_Name]["spec"]["rules"][0]["host"]
    assert "el-edp-my-gitlab" in r["ingress"][el_Name]["spec"]["rules"][0]["http"]["paths"][0]["backend"]["service"]["name"]


def test_ingress_for_github_el():
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

    el_Name = "event-listener-my-github"
    r = helm_template(config)

    assert el_Name in r["ingress"]
    assert "el-my-github-ns.example.com" in r["ingress"][el_Name]["spec"]["rules"][0]["host"]
    assert "el-edp-my-github" in r["ingress"][el_Name]["spec"]["rules"][0]["http"]["paths"][0]["backend"]["service"]["name"]


def test_ingress_for_gerrit_el():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - gerrit
gitServers:
  my-gerrit:
    gitProvider: gerrit
    host: gerrit.com
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

    el_Name = "event-listener-my-gerrit"
    r = helm_template(config)

    assert el_Name in r["ingress"]
    assert "el-my-gerrit-ns.example.com" in r["ingress"][el_Name]["spec"]["rules"][0]["host"]
    assert "el-edp-my-gerrit" in r["ingress"][el_Name]["spec"]["rules"][0]["http"]["paths"][0]["backend"]["service"]["name"]


def test_pipeline_url_uses_cluster_name():
    config = """
global:
  gitProviders:
    - github
  dnsWildCard: "apps.example.com"
clusterName: "my-cluster"
    """

    r = helm_template(config)
    params = r["pipeline"]["github-maven-java17-app-build-default"]["spec"]["params"]
    param = next(p for p in params if p["name"] == "pipelineUrl")

    assert "krci-portal-ns.apps.example.com" in param["default"]
    assert "/c/my-cluster/" in param["default"]


def test_pipeline_url_falls_back_to_dns_wildcard_first_segment():
    config = """
global:
  gitProviders:
    - github
  dnsWildCard: "apps.example.com"
    """

    r = helm_template(config)
    params = r["pipeline"]["github-maven-java17-app-build-default"]["spec"]["params"]
    param = next(p for p in params if p["name"] == "pipelineUrl")

    assert "krci-portal-ns.apps.example.com" in param["default"]
    assert "/c/apps/" in param["default"]


def test_httproute_for_github_el():
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
      enabled: true
      resources:
        requests:
          memory: "64Mi"
          cpu: "50m"
        limits:
          memory: "128Mi"
          cpu: "500m"
      nodeSelector: {}
      tolerations: []
      affinity: {}
      httproute:
        enabled: true
        annotations: {}
    """

    el_Name = "event-listener-my-github"
    r = helm_template(config)

    assert el_Name in r["httproute"]
    spec = r["httproute"][el_Name]["spec"]
    assert spec["parentRefs"][0]["name"] == "main-gateway"
    assert spec["parentRefs"][0]["namespace"] == "envoy-gateway-system"
    assert spec["hostnames"][0] == "el-my-github-ns.example.com"
    assert spec["rules"][0]["backendRefs"][0]["name"] == "el-edp-my-github"
    assert spec["rules"][0]["backendRefs"][0]["port"] == 8080


def test_httproute_for_gitlab_el():
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
      skipWebhookSSLVerification: false
    eventListener:
      enabled: true
      resources:
        requests:
          memory: "64Mi"
          cpu: "50m"
        limits:
          memory: "128Mi"
          cpu: "500m"
      nodeSelector: {}
      tolerations: []
      affinity: {}
      httproute:
        enabled: true
        annotations: {}
    """

    el_Name = "event-listener-my-gitlab"
    r = helm_template(config)

    assert el_Name in r["httproute"]
    spec = r["httproute"][el_Name]["spec"]
    assert spec["parentRefs"][0]["name"] == "main-gateway"
    assert spec["parentRefs"][0]["namespace"] == "envoy-gateway-system"
    assert spec["hostnames"][0] == "el-my-gitlab-ns.example.com"
    assert spec["rules"][0]["backendRefs"][0]["name"] == "el-edp-my-gitlab"
    assert spec["rules"][0]["backendRefs"][0]["port"] == 8080


def test_httproute_for_gerrit_el():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - gerrit
gitServers:
  my-gerrit:
    gitProvider: gerrit
    host: gerrit.com
    quickLink:
      enabled: true
    webhook:
      skipWebhookSSLVerification: false
    eventListener:
      enabled: true
      resources:
        requests:
          memory: "64Mi"
          cpu: "50m"
        limits:
          memory: "128Mi"
          cpu: "500m"
      nodeSelector: {}
      tolerations: []
      affinity: {}
      httproute:
        enabled: true
        annotations: {}
    """

    el_Name = "event-listener-my-gerrit"
    r = helm_template(config)

    assert el_Name in r["httproute"]
    spec = r["httproute"][el_Name]["spec"]
    assert spec["parentRefs"][0]["name"] == "main-gateway"
    assert spec["parentRefs"][0]["namespace"] == "envoy-gateway-system"
    assert spec["hostnames"][0] == "el-my-gerrit-ns.example.com"
    assert spec["rules"][0]["backendRefs"][0]["name"] == "el-edp-my-gerrit"
    assert spec["rules"][0]["backendRefs"][0]["port"] == 8080


def test_httproute_disabled_by_default():
    # HTTPRoute is opt-in: with only the Ingress enabled, no HTTPRoute is rendered
    # and the existing Ingress keeps working (backward compatibility).
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
      enabled: true
      resources:
        requests:
          memory: "64Mi"
          cpu: "50m"
        limits:
          memory: "128Mi"
          cpu: "500m"
      nodeSelector: {}
      tolerations: []
      affinity: {}
      ingress:
        enabled: true
        annotations: {}
        tls: []
    """

    r = helm_template(config)

    assert "httproute" not in r
    assert "event-listener-my-github" in r["ingress"]


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
