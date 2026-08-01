from .helpers import helm_template

# Roles that are still allowed to read every secret in the namespace. An entry
# here is a deliberate, temporary exception - deleting it is the definition of
# done for the work that scopes that role.
KNOWN_UNRESTRICTED_SECRET_RULES = {
    # The reporter masks secret values in the logs it publishes, and discovers
    # which secrets to mask from the env refs of whatever steps the TaskRuns
    # happen to declare. That set is not enumerable at render time, so the rule
    # cannot be scoped until the masking path can report a read it was denied.
    "tekton-reporter",
}


def roles(r):
    return list(r.get("role", {}).values()) + list(r.get("clusterrole", {}).values())


def unrestricted_secret_readers(r):
    # In RBAC an absent or empty resourceNames matches every object of that
    # resource, so a rule that loses its names does not fail closed - it
    # silently grants the whole namespace.
    found = set()
    for role in roles(r):
        for rule in role.get("rules") or []:
            if "secrets" in rule.get("resources", []) and not rule.get("resourceNames"):
                found.add(role["metadata"]["name"])
    return found


def test_no_role_reads_every_secret():
    config = """
global:
  dnsWildCard: "example.com"
    """

    r = helm_template(config)

    assert unrestricted_secret_readers(r) == KNOWN_UNRESTRICTED_SECRET_RULES


def test_no_role_reads_every_secret_with_all_components_enabled():
    # A rule can be scoped under the default values and unrestricted under
    # another combination, so the invariant is checked on a second shape too.
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - github
    - gerrit
interceptor:
  extraSecretNames:
    - custom-token
    """

    r = helm_template(config)

    assert unrestricted_secret_readers(r) == KNOWN_UNRESTRICTED_SECRET_RULES


def test_no_role_reads_every_secret_when_nothing_is_configured():
    # The shape most likely to regress: a rule whose resourceNames are computed
    # from values renders an empty list here, and an empty list is what RBAC
    # reads as "every secret". Roles must drop such rules rather than emit them.
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders: []
gitServers: {}
    """

    r = helm_template(config)

    assert unrestricted_secret_readers(r) == KNOWN_UNRESTRICTED_SECRET_RULES


def test_allowlist_entries_still_exist():
    # Guards against the allowlist outliving the role it names, which would
    # leave a stale exception silently permitting a future regression.
    config = """
global:
  dnsWildCard: "example.com"
    """

    r = helm_template(config)
    rendered = {role["metadata"]["name"] for role in roles(r)}

    assert KNOWN_UNRESTRICTED_SECRET_RULES <= rendered
