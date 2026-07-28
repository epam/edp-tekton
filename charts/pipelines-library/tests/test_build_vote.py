from .helpers import helm_template

CONFIG = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - gitlab
    - github
    - bitbucket
"""

CLONE_RESULT = "$(tasks.fetch-repository.results.commit)"

REPORTERS = {
    "gitlab": ("report-pipeline-start-to-gitlab", "gitlab-report-pipeline-status"),
    "github": ("github-set-pending-status", "github-report-pipeline-status"),
    "bitbucket": ("bitbucket-set-pending-status", "bitbucket-report-pipeline-status"),
}


def _params(task):
    return {p["name"]: p["value"] for p in task.get("params", [])}


def _build_pipelines(rendered, provider):
    return {
        name: p
        for name, p in rendered["pipeline"].items()
        if name.startswith(f"{provider}-") and "-build-" in name
    }


def _review_pipelines(rendered, provider):
    return {
        name: p
        for name, p in rendered["pipeline"].items()
        if name.startswith(f"{provider}-") and name.endswith("-review")
    }


def _find_task(pipeline, task_name):
    for section in ("tasks", "finally"):
        for task in pipeline["spec"].get(section, []):
            if task["name"] == task_name:
                return task
    return None


def test_build_reporters_vote_on_clone_result_without_guards():
    r = helm_template(CONFIG)
    for provider, (start, vote) in REPORTERS.items():
        pipelines = _build_pipelines(r, provider)
        assert pipelines, f"no build pipelines rendered for {provider}"
        for name, pipeline in pipelines.items():
            for task_name in (start, vote):
                task = _find_task(pipeline, task_name)
                assert task is not None, f"{name}: task {task_name} missing"
                assert "when" not in task, f"{name}/{task_name}: must have no when guard"
                assert "runAfter" not in task, (
                    f"{name}/{task_name}: ordering must be implicit via result reference"
                )
                assert _params(task)["SHA"] == CLONE_RESULT, (
                    f"{name}/{task_name}: SHA must be the clone result"
                )


def test_review_reporters_unchanged():
    r = helm_template(CONFIG)
    expected_sha = {
        "gitlab": "$(params.git-source-revision)",
        "github": "$(params.gitsha)",
        "bitbucket": "$(params.git-source-revision)",
    }
    for provider, (_, vote) in REPORTERS.items():
        pipelines = _review_pipelines(r, provider)
        assert pipelines, f"no review pipelines rendered for {provider}"
        for name, pipeline in pipelines.items():
            task = _find_task(pipeline, vote)
            assert task is not None, f"{name}: review reporter missing"
            assert "when" not in task
            assert _params(task)["SHA"] == expected_sha[provider], (
                f"{name}: review SHA source must not change"
            )


def test_set_status_scripts_treat_completed_as_success():
    r = helm_template(CONFIG)
    for provider in REPORTERS:
        script = r["task"][f"{provider}-set-status"]["spec"]["steps"][0]["script"]
        assert 'pipeline_status == "Completed"' in script.replace("PIPELINE_STATUS", "pipeline_status"), (
            f"{provider}-set-status: Completed must have its own success branch"
        )
        assert 'in ("Failed", "Completed")' not in script, (
            f"{provider}-set-status: old Completed->canceled grouping must be gone"
        )
