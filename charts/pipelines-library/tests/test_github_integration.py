from .helpers import helm_template


def test_github_is_enabled():
    config = """
global:
  gitProviders:
    - github
    """

    r = helm_template(config)

    glb = r["eventlistener"]["edp-github"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    glr = r["eventlistener"]["edp-github"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]
    gitserver = r["gitserver"]["github"]["spec"]

    assert "secretString" \
           == glb["secretKey"] \
           == glr["secretKey"]

    assert "ci-github" \
           == glb["secretName"] \
           == glr["secretName"]

    assert "github.com" == gitserver["gitHost"]
    assert "github" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "ci-github" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
