import functools
import posixpath
import re

from .helpers import helm_template

CONFIG = """
global:
  gitProviders:
    - github
"""

ALL_PROVIDERS = """
global:
  gitProviders:
    - bitbucket
    - gerrit
    - github
    - gitlab
"""

WORKSPACE_PATH = "$(workspaces.source.path)"
WORKSPACE_MOUNT = "/workspace/source"
RESULT_REF = re.compile(r"\$\(tasks\.([\w-]+)\.results\.")


@functools.lru_cache(maxsize=None)
def render(config):
    # Rendering all providers costs seconds; the tests share one render per config.
    return helm_template(config)


def _params(task):
    return {p["name"]: p.get("default") for p in task["spec"].get("params", [])}


def _env(step):
    return {e["name"]: e.get("value") for e in step.get("env", [])}


def test_npm_task_exports_npm_cache_dir():
    r = helm_template(CONFIG)
    task = r["task"]["npm"]

    params = _params(task)
    assert params.get("CACHE_DIR") == "/workspace/source/cache", (
        "npm Task must default CACHE_DIR to the shared workspace cache path"
    )

    step = task["spec"]["steps"][0]
    env = _env(step)
    assert env.get("NPM_CACHE_DIR") == "$(params.CACHE_DIR)", (
        "npm Task must export NPM_CACHE_DIR from CACHE_DIR, otherwise npm/pnpm "
        "publish resolve cache=${NPM_CACHE_DIR} literally and pack a stray "
        "${NPM_CACHE_DIR} directory into the published tarball"
    )


def _in_shared_workspace(path, params, sub_path):
    """Directory of shared-workspace that a path of the Task resolves to.

    The Task mounts its source workspace at /workspace/source; the pipeline task
    binds it to sub_path of shared-workspace.
    """
    for name, value in params.items():
        path = path.replace(f"$(params.{name})", value)
    path = posixpath.normpath(path.replace(WORKSPACE_PATH, WORKSPACE_MOUNT))
    relative = posixpath.relpath(path, WORKSPACE_MOUNT)
    assert not relative.startswith(".."), f"{path} is outside the source workspace"
    return posixpath.normpath(posixpath.join(sub_path or ".", relative))


def _inside(path, directory):
    return directory == "." or path == directory or path.startswith(f"{directory}/")


def _cache_subpaths(pipeline):
    """shared-workspace directories that get-cache restores and save-cache uploads."""
    return {
        workspace.get("subPath", ".")
        for task in pipeline["spec"]["tasks"]
        if task.get("taskRef", {}).get("name") in ("get-cache", "save-cache")
        for workspace in task["workspaces"]
        if workspace["name"] == "cache"
    }


def _cache_dir_tasks(rendered):
    """Pipeline tasks that run a Task with a CACHE_DIR param, with the
    shared-workspace directories of their package root, HOME and cache."""
    for pipeline_name, pipeline in sorted(rendered["pipeline"].items()):
        for pipeline_task in pipeline["spec"]["tasks"]:
            task = rendered["task"].get(pipeline_task.get("taskRef", {}).get("name"))
            if task is None or "CACHE_DIR" not in _params(task):
                continue

            params = {k: v for k, v in _params(task).items() if isinstance(v, str)}
            params.update(
                (p["name"], p["value"])
                for p in pipeline_task.get("params", [])
                if isinstance(p["value"], str)
            )
            binding = next(
                w for w in pipeline_task["workspaces"] if w["name"] == "source"
            )
            assert binding["workspace"] == "shared-workspace", pipeline_name
            sub_path = binding.get("subPath")
            step = task["spec"]["steps"][0]
            env = _env(step)
            paths = {"package": step["workingDir"], "CACHE_DIR": params["CACHE_DIR"]}
            if "HOME" in env:
                paths["HOME"] = env["HOME"]

            yield pipeline_name, pipeline, pipeline_task, {
                name: _in_shared_workspace(path, params, sub_path)
                for name, path in paths.items()
            }


def _upstream(pipeline, task_name):
    """Tasks that task_name waits for, directly or transitively."""
    after = {}
    for task in pipeline["spec"]["tasks"]:
        refs = set(RESULT_REF.findall(str(task.get("params", []))))
        after[task["name"]] = set(task.get("runAfter", [])) | refs

    seen, pending = set(), list(after[task_name])
    while pending:
        name = pending.pop()
        if name not in seen:
            seen.add(name)
            pending.extend(after.get(name, ()))
    return seen


def test_cache_dir_is_the_shared_cache_outside_the_package():
    """npm and pnpm publish pack every file under the package root that the
    package's own ignore rules do not exclude. CACHE_DIR and HOME resolve to
    shared-workspace directories outside the package root, and CACHE_DIR is the
    directory get-cache restores and save-cache uploads."""
    r = render(ALL_PROVIDERS)

    publishing = set()
    for pipeline_name, pipeline, pipeline_task, dirs in _cache_dir_tasks(r):
        where = f"{pipeline_name}/{pipeline_task['name']}"
        for name in sorted(dirs.keys() - {"package"}):
            assert not _inside(dirs[name], dirs["package"]), (
                f"{where}: {name} resolves to shared-workspace/{dirs[name]}, inside "
                f"the package root shared-workspace/{dirs['package']}"
            )

        shared = _cache_subpaths(pipeline)
        if shared:
            assert {dirs["CACHE_DIR"]} == shared, (
                f"{where}: CACHE_DIR resolves to shared-workspace/{dirs['CACHE_DIR']}, "
                f"get-cache/save-cache use shared-workspace/{sorted(shared)}"
            )

        if pipeline_task["name"] == "push":
            publishing.add(pipeline_name.split("-")[1])

    assert {"npm", "pnpm"} <= publishing, (
        f"no npm/pnpm publish task checked, only {sorted(publishing)}"
    )


def test_save_cache_waits_for_every_task_writing_the_cache():
    """save-cache archives the shared cache directory; every task whose CACHE_DIR
    resolves to that directory finishes before save-cache starts."""
    r = render(ALL_PROVIDERS)

    writers = set()
    for pipeline_name, pipeline, pipeline_task, dirs in _cache_dir_tasks(r):
        names = {task["name"] for task in pipeline["spec"]["tasks"]}
        shared = _cache_subpaths(pipeline)
        if "save-cache" not in names or dirs["CACHE_DIR"] not in shared:
            continue

        assert pipeline_task["name"] in _upstream(pipeline, "save-cache"), (
            f"{pipeline_name}: save-cache runs in parallel with {pipeline_task['name']}, "
            f"which writes shared-workspace/{dirs['CACHE_DIR']}"
        )
        writers.add(pipeline_task["name"])

    assert {"build", "push"} <= writers, (
        f"no build and push cache writers checked, only {sorted(writers)}"
    )
