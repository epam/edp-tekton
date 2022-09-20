
import os
import sys

from .helpers import helm_template


def test_gitlab_is_disabled():
    config = """
gerrit:
  enabled: false
gitlab:
  enabled: false
    """

    r = helm_template(config)

    assert "eventlistener" not in r
    assert "triggerbinding" not in r
    assert "triggertemplate" not in r
    assert "pipeline" not in r
