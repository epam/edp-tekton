import pytest

from .helpers import helm_template

GITOPS_PIPELINE = "gitlab-helm-gitops-sys-review"
DIFF_PREVIEW_TASK = "argocd-diff-preview"

BASE = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - gitlab
"""

ENABLED = (
    BASE
    + """
pipelines:
  argocdDiffPreview:
    enabled: true
"""
)

# The feature flag on, but the tasks themselves not installed. The pipeline
# references the diff preview Task by name, so rendering the step here would
# leave an unresolvable taskRef and fail the whole review pipeline - the
# opposite of the step's best-effort contract.
ENABLED_WITHOUT_TASKS = (
    BASE
    + """
pipelines:
  argocdDiffPreview:
    enabled: true
  deployableResources:
    tasks: false
"""
)


def _pipeline_task_names(config):
    chart = helm_template(config)
    pipeline = chart["pipeline"][GITOPS_PIPELINE]
    return chart, {task["name"] for task in pipeline["spec"]["tasks"]}


def test_diff_preview_absent_by_default():
    chart, task_names = _pipeline_task_names(BASE)

    assert DIFF_PREVIEW_TASK not in chart.get("task", {})
    assert DIFF_PREVIEW_TASK not in task_names


def test_diff_preview_wired_when_enabled():
    chart, task_names = _pipeline_task_names(ENABLED)

    assert DIFF_PREVIEW_TASK in chart["task"]
    assert DIFF_PREVIEW_TASK in task_names


def test_diff_preview_not_rendered_without_tasks():
    chart, task_names = _pipeline_task_names(ENABLED_WITHOUT_TASKS)

    assert DIFF_PREVIEW_TASK not in chart.get("task", {})
    assert DIFF_PREVIEW_TASK not in task_names


def test_diff_preview_rbac_follows_the_step():
    """RBAC is granted only when the step that uses it actually renders."""
    enabled = helm_template(ENABLED)
    without_tasks = helm_template(ENABLED_WITHOUT_TASKS)

    kubeconfig_role = "tekton-argocd-diff-preview-kubeconfig"
    assert kubeconfig_role in enabled["role"]
    assert kubeconfig_role not in without_tasks.get("role", {})

    def reaches_applications(chart):
        rules = chart["role"]["tekton-pipeline-role"]["rules"]
        return any("applications" in rule.get("resources", []) for rule in rules)

    assert reaches_applications(enabled)
    assert not reaches_applications(without_tasks)


def test_refspec_default_follows_the_step():
    """Branch heads are fetched only when the step needs them as a diff base."""

    def refspec(config):
        chart = helm_template(config)
        params = chart["pipeline"][GITOPS_PIPELINE]["spec"]["params"]
        return next(p for p in params if p["name"] == "git-refspec")["default"]

    assert refspec(ENABLED) == "+refs/heads/*:refs/remotes/origin/*"
    assert refspec(BASE) == ""
    assert refspec(ENABLED_WITHOUT_TASKS) == ""


@pytest.mark.parametrize("config", [BASE, ENABLED], ids=["default", "enabled"])
def test_gitops_review_taskrefs_all_resolve(config):
    """Guards the general failure this feature hit: a pipeline step referencing
    a Task whose own render condition is narrower than the step's.

    ENABLED_WITHOUT_TASKS is deliberately excluded: `deployableResources.tasks`
    is a chart-wide switch that renders no Tasks at all, so every pipeline's
    taskRefs dangle in that configuration - asserting otherwise would encode a
    promise the chart has never made.
    """
    chart = helm_template(config)
    rendered_tasks = set(chart.get("task", {}))

    for task in chart["pipeline"][GITOPS_PIPELINE]["spec"]["tasks"]:
        ref = task.get("taskRef")
        if ref and ref.get("kind", "Task") == "Task":
            assert ref["name"] in rendered_tasks, (
                f"pipeline {GITOPS_PIPELINE} step '{task['name']}' references "
                f"Task '{ref['name']}' which is not rendered"
            )
