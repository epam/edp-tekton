import os
import sys

from .helpers import helm_template


def test_python_pipelines_gerrit():
    config = """
gerrit:
  enabled: true
    """

    r = helm_template(config)

    assert "gerrit-python-python-3.8-review" in r["pipeline"]
    assert "gerrit-python-python-3.8-build-default" in r["pipeline"]
    assert "gerrit-python-python-3.8-build-edp" in r["pipeline"]

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['python-3.8']:

            gerrit_review_pipeline = f"gerrit-{buildtool}-{framework}-review"
            gerrit_build_pipeline_def = f"gerrit-{buildtool}-{framework}-build-default"
            gerrit_build_pipeline_edp = f"gerrit-{buildtool}-{framework}-build-edp"

            rt = r["pipeline"][gerrit_review_pipeline]["spec"]["tasks"]
            assert "fetch-repository" in rt[0]["name"]
            assert "gerrit-notify" in rt[1]["name"]
            assert "init-values" in rt[2]["name"]
            assert "compile" in rt[3]["name"]
            assert "test" in rt[4]["name"]
            assert "dockerfile-lint" in rt[5]["name"]
            assert "helm-lint" in rt[6]["name"]
            assert "gerrit-vote-success" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][0]["name"]
            assert "gerrit-vote-failure" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][1]["name"]

            # build with default versioning
            btd = r["pipeline"][gerrit_build_pipeline_def]["spec"]["tasks"]
            assert "fetch-repository" in btd[0]["name"]
            assert "gerrit-notify" in btd[1]["name"]
            assert "init-values" in btd[2]["name"]
            assert "get-version" in btd[3]["name"]
            # ensure we have default versioning
            assert f"get-version-{buildtool}-default" == btd[3]["taskRef"]["name"]
            assert "compile" in btd[4]["name"]
            assert "test" in btd[5]["name"]
            assert buildtool == btd[5]["taskRef"]["name"]
            assert "sonar" in btd[6]["name"]
            assert buildtool == btd[6]["taskRef"]["name"]
            assert "kaniko-build" in btd[7]["name"]
            assert "git-tag" in btd[8]["name"]
            assert "update-cbis" in btd[9]["name"]

            # build with edp versioning
            btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
            assert "fetch-repository" in btedp[0]["name"]
            assert "gerrit-notify" in btedp[1]["name"]
            assert "init-values" in btedp[2]["name"]
            assert "get-version" in btedp[3]["name"]
            assert "get-version-edp" == btedp[3]["taskRef"]["name"]
            assert "update-build-number" in btedp[4]["taskRef"]["name"]
            assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
            assert "compile" in btedp[5]["name"]
            assert buildtool == btedp[5]["taskRef"]["name"]
            assert "test" in btedp[6]["name"]
            assert buildtool == btedp[6]["taskRef"]["name"]
            assert "sonar" in btedp[7]["name"]
            assert buildtool == btedp[7]["taskRef"]["name"]
            assert "kaniko-build" in btedp[8]["name"]
            assert "git-tag" in btedp[9]["name"]
            assert "update-cbis" in btedp[10]["name"]
