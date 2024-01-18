from .helpers import helm_template


def test_gerrit_is_disabled():
    config = """
global:
  gitProvider: unsupported
    """

    r = helm_template(config)

    assert "eventlistener" not in r
    assert "triggerbinding" not in r
    assert "cdpipeline" in r["pipeline"]
    assert "gitserver" not in r


def test_gerrit_is_enabled_with_custom_port():
    config = """
global:
  gitProvider: gerrit
  gerritSSHPort: "777"
    """

    r = helm_template(config)

    gerrit_port_param = {'name': 'GERRIT_PORT', 'value': '777'}
    assert gerrit_port_param in r["pipeline"]["gerrit-go-beego-app-build-default"]["spec"]["tasks"][1]["params"]
    assert gerrit_port_param in r["pipeline"]["gerrit-maven-java11-app-build-default"]["spec"]["tasks"][1]["params"]
    assert gerrit_port_param in r["pipeline"]["gerrit-gradle-java11-app-review"]["spec"]["finally"][1]["params"]

    git_source_url_param = {'name': 'git-source-url', 'value': 'ssh://edp-ci@gerrit:777/$(tt.params.gerritproject)'}
    assert git_source_url_param in r["triggertemplate"]["gerrit-build-template"]["spec"]["resourcetemplates"][0]["spec"]["params"]
    assert git_source_url_param in r["triggertemplate"]["gerrit-review-template"]["spec"]["resourcetemplates"][0]["spec"]["params"]
