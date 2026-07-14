from .helpers import helm_template


def test_reporter_enabled_by_default():
    config = """
global:
  dnsWildCard: "example.com"
    """

    r = helm_template(config)

    assert "tekton-reporter" in r["deployment"]
    assert "tekton-reporter" in r["serviceaccount"]
    assert "tekton-reporter" in r["role"]
    assert "tekton-reporter" in r["rolebinding"]

    deployment = r["deployment"]["tekton-reporter"]
    container = deployment["spec"]["template"]["spec"]["containers"][0]

    assert container["command"] == ["/edpreporter"]
    assert container["image"].startswith("epamedp/edp-tekton:")

    env = {e["name"]: e.get("value") for e in container["env"]}
    assert env["REPORTER_TAIL_LINES"] == "100"
    assert env["REPORTER_COMMENT_STRATEGY"] == "update"
    assert env["PORTAL_BASE_URL"] == "https://krci-portal-ns.example.com/c/example/cicd/pipelineruns"

    assert deployment["spec"]["template"]["spec"]["serviceAccountName"] == "tekton-reporter"


def test_reporter_disabled():
    config = """
global:
  dnsWildCard: "example.com"
reporter:
  enabled: false
    """

    r = helm_template(config)

    assert "tekton-reporter" not in r.get("deployment", {})
    assert "tekton-reporter" not in r.get("serviceaccount", {})
    assert "tekton-reporter" not in r.get("role", {})
    assert "tekton-reporter" not in r.get("rolebinding", {})


def test_reporter_custom_configuration():
    config = """
global:
  dnsWildCard: "example.com"
clusterName: "prod-cluster"
reporter:
  tailLines: 50
  commentStrategy: new
  image:
    repository: custom/reporter
    tag: 1.2.3
    """

    r = helm_template(config)

    container = r["deployment"]["tekton-reporter"]["spec"]["template"]["spec"]["containers"][0]

    assert container["image"] == "custom/reporter:1.2.3"

    env = {e["name"]: e.get("value") for e in container["env"]}
    assert env["REPORTER_TAIL_LINES"] == "50"
    assert env["REPORTER_COMMENT_STRATEGY"] == "new"
    assert env["PORTAL_BASE_URL"] == "https://krci-portal-ns.example.com/c/prod-cluster/cicd/pipelineruns"


def test_reporter_custom_portal_host():
    config = """
global:
  dnsWildCard: "example.com"
portalHost: "portal.example.com"
    """

    r = helm_template(config)

    container = r["deployment"]["tekton-reporter"]["spec"]["template"]["spec"]["containers"][0]
    env = {e["name"]: e.get("value") for e in container["env"]}
    assert env["PORTAL_BASE_URL"] == "https://portal.example.com/c/example/cicd/pipelineruns"


def test_reporter_extra_volumes():
    config = """
global:
  dnsWildCard: "example.com"
reporter:
  extraVolumes:
    - name: git-ca
      configMap:
        name: git-ca
  extraVolumeMounts:
    - name: git-ca
      mountPath: /etc/ssl/certs/git-ca.crt
      subPath: ca.crt
      readOnly: true
    """

    r = helm_template(config)

    pod = r["deployment"]["tekton-reporter"]["spec"]["template"]["spec"]
    assert pod["volumes"][0]["configMap"]["name"] == "git-ca"

    mounts = pod["containers"][0]["volumeMounts"]
    assert mounts[0]["mountPath"] == "/etc/ssl/certs/git-ca.crt"
    assert mounts[0]["readOnly"] is True


def test_reporter_rbac_rules():
    config = """
global:
  dnsWildCard: "example.com"
    """

    r = helm_template(config)

    rules = r["role"]["tekton-reporter"]["rules"]
    by_resource = {}
    for rule in rules:
        for resource in rule["resources"]:
            by_resource[resource] = sorted(rule["verbs"])

    assert by_resource["pipelineruns"] == ["get", "list", "patch", "watch"]
    assert by_resource["taskruns"] == ["get", "list", "watch"]
    assert by_resource["pods/log"] == ["get"]
    assert by_resource["secrets"] == ["get"]
    assert by_resource["codebases"] == ["get", "list", "watch"]
    assert by_resource["gitservers"] == ["get", "list", "watch"]
    assert "leases" in by_resource

    binding = r["rolebinding"]["tekton-reporter"]
    assert binding["roleRef"]["name"] == "tekton-reporter"
    assert binding["subjects"][0]["name"] == "tekton-reporter"
