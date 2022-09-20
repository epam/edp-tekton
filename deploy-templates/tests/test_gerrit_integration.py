import os
import sys

from .helpers import helm_template


def test_gerrit_is_disabled():
    config = """
gerrit:
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

    assert {'name': 'GERRIT_PORT', 'value': 777} in r["pipeline"]["gerrit-go-beego-build-default"]["spec"]["tasks"][1]["params"]
    assert {'name': 'GERRIT_PORT', 'value': 777} in r["pipeline"]["gerrit-maven-java11-build-default"]["spec"]["tasks"][1]["params"]
    assert {'name': 'GERRIT_PORT', 'value': 777} in r["pipeline"]["gerrit-gradle-java11-review"]["spec"]["finally"][1]["params"]

    assert {'name': 'git-source-url', 'value': 'ssh://jenkins@gerrit:777/$(tt.params.gerritproject)'} in r["triggertemplate"]["gerrit-build-template"]["spec"]["resourcetemplates"][0]["spec"]["params"]
