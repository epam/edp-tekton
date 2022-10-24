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

    sr = r["eventlistener"]["github-listener"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    sm = r["eventlistener"]["github-listener"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]

    assert "secretString" == sr["secretKey"] == sm["secretKey"]
    assert "github.com-config" == sr["secretName"] == sm["secretName"]
