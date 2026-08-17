import re

from .helpers import helm_template

CONFIG = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - github
"""

SETTINGS_CONFIGMAP = "custom-npm-settings"

VAR = re.compile(r"\$\{(\w+)\}")
USERCONFIG = re.compile(r"export\s+npm_config_userconfig=(\S+)")
EXPORTED = re.compile(r"^\s*export\s+(\w+)=", re.MULTILINE)


def _npmrc_files(rendered):
    return rendered["configmap"][SETTINGS_CONFIGMAP]["data"]


def _package_manager_steps(rendered):
    """Task steps that mount the npm settings ConfigMap and drive a package
    manager from their own script.

    Steps that run $(params.EXTRA_COMMANDS) are skipped: they hand the registry
    wiring to whichever pipeline calls them, so the invariant belongs there.
    """
    for task_name, task in sorted(rendered["task"].items()):
        volumes = {
            volume["name"]
            for volume in task["spec"].get("volumes", [])
            if volume.get("configMap", {}).get("name") == SETTINGS_CONFIGMAP
        }
        if not volumes:
            continue
        for step in task["spec"]["steps"]:
            mounted = {mount["name"] for mount in step.get("volumeMounts", [])}
            if not mounted & volumes:
                continue
            if "$(params.EXTRA_COMMANDS)" in step.get("script", ""):
                continue
            yield task_name, step


def test_mounted_npmrc_is_actually_read():
    """Mounting the ConfigMap does nothing on its own — npm and pnpm only read it
    when npm_config_userconfig names it. edp-pnpm mounted it and never pointed
    pnpm at the file, so every pnpm build resolved from the public registry while
    reporting success.
    """
    rendered = helm_template(CONFIG)
    npmrc = _npmrc_files(rendered)

    checked = []
    for task_name, step in _package_manager_steps(rendered):
        match = USERCONFIG.search(step.get("script", ""))
        assert match, (
            f"task {task_name}, step {step['name']} mounts {SETTINGS_CONFIGMAP} but "
            f"never exports npm_config_userconfig, so the file is ignored and the "
            f"package manager falls back to the public registry"
        )

        key = match.group(1).rsplit("/", 1)[-1]
        assert key in npmrc, (
            f"task {task_name}, step {step['name']} points npm_config_userconfig at "
            f"{match.group(1)}, which {SETTINGS_CONFIGMAP} does not provide"
        )
        checked.append(task_name)

    assert checked, "no task mounts the npm settings ConfigMap — the test is vacuous"


def test_every_variable_the_npmrc_interpolates_is_available():
    """.npmrc-ci interpolates ${NEXUS_HOST} and ${upBase64}, which are derived in
    the shell rather than supplied as env vars. Exporting npm_config_userconfig
    without deriving them drops the registry credentials: pnpm reports a warning,
    keeps going, and sends the request unauthenticated.
    """
    rendered = helm_template(CONFIG)
    npmrc = _npmrc_files(rendered)

    for task_name, step in _package_manager_steps(rendered):
        script = step.get("script", "")
        match = USERCONFIG.search(script)
        if not match:
            continue

        key = match.group(1).rsplit("/", 1)[-1]
        available = {env["name"] for env in step.get("env", [])}
        available |= set(EXPORTED.findall(script))

        missing = sorted(set(VAR.findall(npmrc[key])) - available)
        assert not missing, (
            f"task {task_name}, step {step['name']} reads {key}, which interpolates "
            f"{missing} — neither present in the step env nor exported by the script"
        )
