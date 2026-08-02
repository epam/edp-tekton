from .helpers import helm_template

ALL_PROVIDERS = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - github
    - gitlab
    - gerrit
    - bitbucket
"""

PROVIDERS = ["github", "gitlab", "gerrit", "bitbucket"]


def pipeline_run_sa(trigger_template):
    resource = trigger_template["spec"]["resourcetemplates"][0]
    return resource["spec"]["taskRunTemplate"]["serviceAccountName"]


def tt_param(trigger_template, name):
    # Exactly-once: Tekton rejects duplicate param names, and a first-match
    # lookup would hide a duplicated declaration.
    matches = [p for p in trigger_template["spec"]["params"] if p["name"] == name]
    assert len(matches) == 1, trigger_template["metadata"]["name"]
    return matches[0]


def tb_param(trigger_binding, name):
    for param in trigger_binding["spec"]["params"]:
        if param["name"] == name:
            return param
    return None


def test_runs_use_interceptor_resolved_service_account():
    r = helm_template(ALL_PROVIDERS)

    for p in PROVIDERS:
        for kind in ("review", "build"):
            tt = r["triggertemplate"][f"{p}-{kind}-template"]
            assert pipeline_run_sa(tt) == "$(tt.params.serviceAccount)"
            assert tt_param(tt, "serviceAccount")["default"] == "tekton-unprivileged"

            tb = r["triggerbinding"][f"{p}-binding-{kind}"]
            assert (
                tb_param(tb, "serviceAccount")["value"]
                == f"$(extensions.serviceAccounts.{kind})"
            )


def test_fixed_trigger_templates_use_scoped_service_accounts():
    r = helm_template(ALL_PROVIDERS)

    expected = {"tekton-cd": ["clean", "deploy", "deploy-ansible", "deploy-ansible-awx",
                              "deploy-diff-approve", "deploy-with-approve", "deploy-with-autotests"],
                "tekton-security": [f"{p}-security-template" for p in PROVIDERS],
                "tekton-unprivileged": [f"{p}-run-autotests" for p in PROVIDERS] + ["image-scan-remote"]}
    for sa, templates in expected.items():
        for name in templates:
            assert pipeline_run_sa(r["triggertemplate"][name]) == sa, name

    # tekton keeps only the build bookkeeping: no trigger references it directly
    for tt in r["triggertemplate"].values():
        assert pipeline_run_sa(tt) != "tekton", tt["metadata"]["name"]


def test_build_pipelines_are_annotated_with_tekton():
    r = helm_template(ALL_PROVIDERS)

    builds = reviews = 0
    for pipeline in r["pipeline"].values():
        labels = pipeline["metadata"].get("labels") or {}
        annotations = pipeline["metadata"].get("annotations") or {}
        sa = annotations.get("app.edp.epam.com/service-account")
        if labels.get("app.edp.epam.com/pipelinetype") == "build":
            builds += 1
            assert sa == "tekton", pipeline["metadata"]["name"]
        elif labels.get("app.edp.epam.com/pipelinetype") == "review":
            reviews += 1
            # No annotation: review runs stay on the unprivileged default.
            assert sa is None, pipeline["metadata"]["name"]

    assert builds > 100
    assert reviews > 100


def test_interceptor_reads_pipelines():
    r = helm_template(ALL_PROVIDERS)

    rules = r["role"]["tekton-triggers-edp-interceptor"]["rules"]
    pipeline_rules = [
        rule
        for rule in rules
        if "pipelines" in rule["resources"] and "tekton.dev" in rule["apiGroups"]
    ]
    assert len(pipeline_rules) == 1
    assert pipeline_rules[0]["verbs"] == ["get"]

    container = r["deployment"]["tekton-interceptor"]["spec"]["template"]["spec"]["containers"][0]
    env = {e["name"]: e.get("value") for e in container["env"]}
    assert env["TEKTON_SA_DEFAULT"] == "tekton-unprivileged"


def test_unprivileged_service_account_has_no_token_and_no_bindings():
    r = helm_template(ALL_PROVIDERS)

    sa = r["serviceaccount"]["tekton-unprivileged"]
    assert sa["automountServiceAccountToken"] is False
    assert "secrets" not in sa

    bindings = list(r.get("rolebinding", {}).values()) + list(
        r.get("clusterrolebinding", {}).values()
    )
    for binding in bindings:
        for subject in binding.get("subjects") or []:
            assert subject["name"] != "tekton-unprivileged", binding["metadata"]["name"]


def test_default_service_account_is_overridable():
    r = helm_template(
        ALL_PROVIDERS
        + """
tekton:
  defaultServiceAccount: tekton
"""
    )

    for p in PROVIDERS:
        for kind in ("review", "build"):
            tt = r["triggertemplate"][f"{p}-{kind}-template"]
            assert tt_param(tt, "serviceAccount")["default"] == "tekton"

    container = r["deployment"]["tekton-interceptor"]["spec"]["template"]["spec"]["containers"][0]
    env = {e["name"]: e.get("value") for e in container["env"]}
    assert env["TEKTON_SA_DEFAULT"] == "tekton"
