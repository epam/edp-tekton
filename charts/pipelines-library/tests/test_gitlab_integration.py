from .helpers import helm_template


def test_gitlab_is_enabled():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)

    gitserver = r["gitserver"]["my-gitlab"]["spec"]
    assert "gitlab.com" == gitserver["gitHost"]
    assert "gitlab" == gitserver["gitProvider"]
    assert "git" == gitserver["gitUser"]
    assert 443 == gitserver["httpsPort"]
    assert "ci-gitlab" == gitserver["nameSshKeySecret"]
    assert 22 == gitserver["sshPort"]


def test_gitlab_build_trigger():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)

    # Access the gitlab-build trigger using the new structure
    trigger = r["trigger"]["gitlab-build"]["spec"]

    # Check if the interceptors are correctly set
    assert "gitlab" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-gitlab" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["Merge Request Hook"] == trigger["interceptors"][0]["params"][1]["value"]

    # Check if the bindings and template are correctly set
    assert "gitlab-binding-build" == trigger["bindings"][0]["ref"]
    assert "gitlab-build-template" == trigger["template"]["ref"]


def test_gitlab_review_trigger():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)

    # Access the gitlab-review trigger using the new structure
    trigger = r["trigger"]["gitlab-review"]["spec"]

    # Check if the interceptors are correctly set
    assert "gitlab" == trigger["interceptors"][0]["ref"]["name"]
    assert "ci-gitlab" == trigger["interceptors"][0]["params"][0]["value"]["secretName"]
    assert ["Merge Request Hook", "Note Hook"] == trigger["interceptors"][0]["params"][1]["value"]

    # Check if the bindings and template are correctly set
    assert "gitlab-binding-review" == trigger["bindings"][0]["ref"]
    assert "gitlab-review-template" == trigger["template"]["ref"]
