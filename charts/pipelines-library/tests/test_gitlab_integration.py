
import os
import sys

from .helpers import helm_template


def test_gitlab_is_enabled():
    config = """
global:
  gitProvider: gitlab
    """

    r = helm_template(config)

    glatb = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    glab = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]
    gllb = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][2]["interceptors"][0]["params"][0]["value"]
    glcb = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][3]["interceptors"][0]["params"][0]["value"]
    glcr = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][4]["interceptors"][0]["params"][0]["value"]
    gitserver = r["gitserver"]["gitlab"]["spec"]
    assert "secretString" == glatb["secretKey"] == glab["secretKey"] == gllb["secretKey"] == glcb["secretKey"] == glcr["secretKey"]
    assert "gitlab" == glatb["secretName"] == glab["secretName"] == gllb["secretName"] == glcb["secretName"] == glcr["secretName"]
    assert "git.epam.com" == gitserver["gitHost"]
    assert "gitlab" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "gitlab" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
