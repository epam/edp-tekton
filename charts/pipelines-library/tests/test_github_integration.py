import os
import sys

from .helpers import helm_template


def test_github_is_disabled():
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


def test_github_is_enabled():
    config = """
gerrit:
  enabled: false
gitlab:
  enabled: false
github:
  enabled: true
    """

    r = helm_template(config)

    glatb = r["eventlistener"]["github-listener"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    glab = r["eventlistener"]["github-listener"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]
    gllb = r["eventlistener"]["github-listener"]["spec"]["triggers"][2]["interceptors"][0]["params"][0]["value"]
    glcb = r["eventlistener"]["github-listener"]["spec"]["triggers"][3]["interceptors"][0]["params"][0]["value"]
    glaccr = r["eventlistener"]["github-listener"]["spec"]["triggers"][4]["interceptors"][0]["params"][0]["value"]
    gllatcr = r["eventlistener"]["github-listener"]["spec"]["triggers"][5]["interceptors"][0]["params"][0]["value"]

    assert "secretString" \
           == glatb["secretKey"] \
           == glab["secretKey"] \
           == gllb["secretKey"] \
           == glcb["secretKey"] \
           == glaccr["secretKey"] \
           == gllatcr["secretKey"]
    assert "github.com-config" \
           == glatb["secretName"] \
           == glab["secretName"] \
           == gllb["secretName"] \
           == glcb["secretName"] \
           == glaccr["secretName"] \
           == gllatcr["secretName"]
