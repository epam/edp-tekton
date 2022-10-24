
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
    assert "pipeline" not in r


def test_gitlab_is_enabled():
    config = """
gerrit:
  enabled: false
github:
  enabled: true
gitlab:
  enabled: true
    """

    r = helm_template(config)

    sr = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][0]["interceptors"][0]["params"][0]["value"]
    sm = r["eventlistener"]["gitlab-listener"]["spec"]["triggers"][1]["interceptors"][0]["params"][0]["value"]

    assert "secretString" == sr["secretKey"] == sm["secretKey"]
    assert "gitlab-configuration" == sr["secretName"] == sm["secretName"]
