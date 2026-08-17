from .helpers import helm_template

CONFIG = """
global:
  gitProviders:
    - github
"""


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