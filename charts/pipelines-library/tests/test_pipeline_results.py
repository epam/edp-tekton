import re

import pytest

from .helpers import helm_template

BASE = """
global:
  dnsWildCard: "example.com"
"""

TASK_RESULT_REF = re.compile(r"\$\(tasks\.([\w-]+)\.results\.([\w-]+)\)")
BRANCH_START = re.compile(r"(^|;)\s*then\s*$|^\s*else\s*$")


def result_producers():
    # A pipeline result that references a task result fails the whole
    # PipelineRun (CouldntGetPipelineResult) when the task never writes it,
    # even if every task succeeded.
    rendered = helm_template(BASE)
    producers = set()
    for pipeline in rendered.get("pipeline", {}).values():
        spec = pipeline["spec"]
        pipeline_tasks = {t["name"]: t for t in spec.get("tasks", []) + spec.get("finally", [])}
        for result in spec.get("results", []):
            for pipeline_task, result_name in TASK_RESULT_REF.findall(str(result.get("value", ""))):
                task_name = pipeline_tasks[pipeline_task].get("taskRef", {}).get("name")
                task = rendered.get("task", {}).get(task_name)
                if task is None:
                    continue
                for step in task["spec"]["steps"]:
                    if f"results.{result_name}.path" in step.get("script", ""):
                        producers.add((task_name, step["name"], result_name, step["script"]))
    return sorted(producers)


PRODUCERS = result_producers()


def early_exits_without_result(script, result_name):
    lines = script.splitlines()
    missing = []
    for index, line in enumerate(lines):
        if line.strip() != "exit 0":
            continue
        start = index - 1
        while start >= 0 and not BRANCH_START.search(lines[start]):
            start -= 1
        if not any(f"results.{result_name}.path" in branch_line for branch_line in lines[start + 1 : index]):
            missing.append(index + 1)
    return missing


def test_security_scan_report_url_is_covered():
    assert ("security", "upload-report", "SCAN_REPORT_URL") in {p[:3] for p in PRODUCERS}


@pytest.mark.parametrize(
    "task_name,step_name,result_name,script",
    PRODUCERS,
    ids=[f"{p[0]}/{p[1]}/{p[2]}" for p in PRODUCERS],
)
def test_pipeline_result_is_written_before_early_exit(task_name, step_name, result_name, script):
    assert early_exits_without_result(script, result_name) == [], (
        f"{task_name}/{step_name} exits 0 without writing {result_name}"
    )
