from .helpers import helm_template


def secret_rules(r):
    return [
        rule
        for rule in r["role"]["tekton-triggers-edp-interceptor"]["rules"]
        if "secrets" in rule["resources"]
    ]


def granted_secret_names(r):
    return sorted(
        {name for rule in secret_rules(r) for name in rule.get("resourceNames", [])}
    )


def assert_no_unrestricted_secret_rule(r):
    # A rule that loses its names silently restores blanket access.
    unrestricted = [rule for rule in secret_rules(r) if not rule.get("resourceNames")]
    assert unrestricted == []


def test_interceptor_secret_names_follow_git_providers():
    config = """
global:
  dnsWildCard: "example.com"
    """

    r = helm_template(config)

    assert_no_unrestricted_secret_rule(r)
    assert granted_secret_names(r) == [
        "ci-bitbucket",
        "ci-gerrit",
        "ci-github",
        "ci-gitlab",
        "gerrit-ciuser-sshkey",
        "tekton-edp-interceptor-certs",
    ]


def test_interceptor_certs_secret_keeps_write_verbs():
    config = """
global:
  dnsWildCard: "example.com"
    """

    r = helm_template(config)

    certs = [
        rule
        for rule in secret_rules(r)
        if rule.get("resourceNames") == ["tekton-edp-interceptor-certs"]
    ]
    assert len(certs) == 1
    assert sorted(certs[0]["verbs"]) == ["create", "get", "list", "update", "watch"]

    git_server = [
        rule
        for rule in secret_rules(r)
        if rule.get("resourceNames") != ["tekton-edp-interceptor-certs"]
    ]
    assert len(git_server) == 1
    assert git_server[0]["verbs"] == ["get"]


def git_server_config(name, provider, host, secret=None, providers=None):
    secret_line = f"    nameSshKeySecret: {secret}\n" if secret else ""
    provider_list = "\n".join(f"    - {p}" for p in (providers or [provider]))
    return f"""
global:
  dnsWildCard: "example.com"
  gitProviders:
{provider_list}
gitServers:
  {name}:
    gitProvider: {provider}
    host: {host}
{secret_line}    quickLink:
      enabled: false
    webhook:
      skipWebhookSSLVerification: false
    eventListener:
      enabled: true
      ingress:
        enabled: false
      resources: {{}}
      nodeSelector: {{}}
      tolerations: []
      affinity: {{}}
    """


def test_interceptor_grants_the_secret_named_by_the_rendered_git_server():
    # Pins the Role to the CR it authorises by comparing the two rendered
    # objects, so the rule cannot drift away from gitserver.yaml unnoticed.
    r = helm_template(git_server_config("my-github", "github", "github.com"))

    assert_no_unrestricted_secret_rule(r)
    cr_secret = r["gitserver"]["my-github"]["spec"]["nameSshKeySecret"]
    assert cr_secret == "ci-github"
    assert cr_secret in granted_secret_names(r)


def test_interceptor_grants_custom_git_server_secret():
    r = helm_template(
        git_server_config("my-gerrit", "gerrit", "gerrit.local", secret="my-custom-token")
    )

    assert_no_unrestricted_secret_rule(r)
    cr_secret = r["gitserver"]["my-gerrit"]["spec"]["nameSshKeySecret"]
    assert cr_secret == "my-custom-token"
    assert cr_secret in granted_secret_names(r)
    assert "ci-github" not in granted_secret_names(r)


def test_interceptor_ignores_git_server_whose_provider_is_disabled():
    # gitserver.yaml only creates the CR when the provider is enabled, so
    # granting its secret would hand out access no component can use.
    r = helm_template(
        git_server_config(
            "my-github", "github", "github.com", secret="unused-token", providers=["gerrit"]
        )
    )

    assert_no_unrestricted_secret_rule(r)
    assert "gitserver" not in r
    assert "unused-token" not in granted_secret_names(r)


def test_interceptor_omits_gerrit_secret_when_gerrit_is_disabled():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - github
    """

    r = helm_template(config)

    assert granted_secret_names(r) == ["ci-github", "tekton-edp-interceptor-certs"]


def test_interceptor_extra_secret_names_are_appended():
    config = """
global:
  dnsWildCard: "example.com"
interceptor:
  extraSecretNames:
    - portal-created-token
    - another-token
    """

    r = helm_template(config)

    assert_no_unrestricted_secret_rule(r)
    names = granted_secret_names(r)
    assert "portal-created-token" in names
    assert "another-token" in names
    assert "ci-github" in names


def test_interceptor_drops_git_server_secret_rule_when_no_names_are_known():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders: []
    """

    r = helm_template(config)

    assert_no_unrestricted_secret_rule(r)
    assert granted_secret_names(r) == ["tekton-edp-interceptor-certs"]


def test_interceptor_drops_git_server_secret_rule_when_providers_are_null():
    # An explicit null reaches the template as nil rather than an empty list,
    # which ranges differently.
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    """

    r = helm_template(config)

    assert_no_unrestricted_secret_rule(r)
    assert granted_secret_names(r) == ["tekton-edp-interceptor-certs"]


def test_interceptor_keeps_rule_when_only_extra_names_are_set():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders: []
interceptor:
  extraSecretNames:
    - only-custom-token
    """

    r = helm_template(config)

    assert_no_unrestricted_secret_rule(r)
    assert granted_secret_names(r) == [
        "only-custom-token",
        "tekton-edp-interceptor-certs",
    ]


def test_interceptor_deduplicates_secret_names():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - github
interceptor:
  extraSecretNames:
    - ci-github
    """

    r = helm_template(config)

    git_server_rule = [
        rule
        for rule in secret_rules(r)
        if rule.get("resourceNames") != ["tekton-edp-interceptor-certs"]
    ][0]
    assert git_server_rule["resourceNames"] == ["ci-github"]
