
import os
import sys

from .helpers import helm_template


def test_gitlab_is_enabled():
    config = """
global:
  gitProvider: gitlab
    """

    r = helm_template(config)

    glb = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    glr = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]
    gitserver = r["gitserver"]["gitlab"]["spec"]
    assert "secretString" == glb["secretKey"] == glr["secretKey"]
    assert "gitlab" == glb["secretName"] == glr["secretName"]
    assert "git.epam.com" == gitserver["gitHost"]
    assert "gitlab" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "gitlab" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
