import os
import sys

from .helpers import helm_template


def test_gerrit_is_disabled():
    config = """
gerrit:
  enabled: false
github:
  enabled: false
gitlab:
  enabled: false
    """

    r = helm_template(config)

    assert "eventlistener" not in r
    assert "triggerbinding" not in r
    assert "triggertemplate" not in r
    assert "pipeline" not in r


def test_gerrit_is_enabled_with_custom_port():
    config = """
gerrit:
  enabled: true
  sshPort: 777
    """

    r = helm_template(config)

    assert {'name': 'GERRIT_PORT', 'value': 777} in r["pipeline"]["gerrit-go-beego-app-build-default"]["spec"]["tasks"][1]["params"]
    assert {'name': 'GERRIT_PORT', 'value': 777} in r["pipeline"]["gerrit-maven-java11-app-build-default"]["spec"]["tasks"][1]["params"]
    assert {'name': 'GERRIT_PORT', 'value': 777} in r["pipeline"]["gerrit-gradle-java11-app-review"]["spec"]["finally"][1]["params"]

    assert {'name': 'git-source-url', 'value': 'ssh://edp-ci@gerrit:777/$(tt.params.gerritproject)'} in r["triggertemplate"]["gerrit-build-app-template"]["spec"]["resourcetemplates"][0]["spec"]["params"]
