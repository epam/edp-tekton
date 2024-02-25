from .helpers import helm_template


def test_gitlab_is_enabled():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)

    glb = r["eventlistener"]["edp-gitlab"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    glr = r["eventlistener"]["edp-gitlab"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]
    gitserver = r["gitserver"]["gitlab"]["spec"]
    assert "secretString" == glb["secretKey"] == glr["secretKey"]
    assert "ci-gitlab" == glb["secretName"] == glr["secretName"]
    assert "gitlab.com" == gitserver["gitHost"]
    assert "gitlab" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "ci-gitlab" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
