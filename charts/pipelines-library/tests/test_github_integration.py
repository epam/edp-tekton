import os
import sys

from .helpers import helm_template


def test_github_is_enabled():
    config = """
global:
  gitProvider: github
    """

    r = helm_template(config)

    glatb = r["eventlistener"]["github-listener"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    glab = r["eventlistener"]["github-listener"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]
    gllb = r["eventlistener"]["github-listener"]["spec"]["triggers"][2]["interceptors"][0]["params"][0]["value"]
    glcb = r["eventlistener"]["github-listener"]["spec"]["triggers"][3]["interceptors"][0]["params"][0]["value"]
    glaccr = r["eventlistener"]["github-listener"]["spec"]["triggers"][4]["interceptors"][0]["params"][0]["value"]
    gllatcr = r["eventlistener"]["github-listener"]["spec"]["triggers"][5]["interceptors"][0]["params"][0]["value"]
    gitserver = r["gitserver"]["github"]["spec"]

    assert "secretString" \
           == glatb["secretKey"] \
           == glab["secretKey"] \
           == gllb["secretKey"] \
           == glcb["secretKey"] \
           == glaccr["secretKey"] \
           == gllatcr["secretKey"]
    assert "github" \
           == glatb["secretName"] \
           == glab["secretName"] \
           == gllb["secretName"] \
           == glcb["secretName"] \
           == glaccr["secretName"] \
           == gllatcr["secretName"]
    assert "github.com" == gitserver["gitHost"]
    assert "github" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "github" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]
