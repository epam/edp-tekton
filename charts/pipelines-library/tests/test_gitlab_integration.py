
import os
import sys

from .helpers import helm_template


def test_gitlab_is_disabled():
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
    assert "cdpipeline" in r["pipeline"]


def test_gitlab_is_enabled():
    config = """
gerrit:
  enabled: false
github:
  enabled: false
gitlab:
  enabled: true
    """

    r = helm_template(config)

    glatb = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    glab = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]
    gllb = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][2]["interceptors"][0]["params"][0]["value"]
    glcb = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][3]["interceptors"][0]["params"][0]["value"]
    glcr = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][4]["interceptors"][0]["params"][0]["value"]
    assert "secretString" == glatb["secretKey"] == glab["secretKey"] == gllb["secretKey"] == glcb["secretKey"] == glcr["secretKey"]
    assert "gitlab.com-config" == glatb["secretName"] == glab["secretName"] == gllb["secretName"] == glcb["secretName"] == glcr["secretName"]
