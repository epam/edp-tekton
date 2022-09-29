import os
import sys

from .helpers import helm_template


def test_java_pipelines_gerrit():
    config = """
gerrit:
  enabled: true
    """

    r = helm_template(config)

    assert "gerrit-maven-java11-review" in r["pipeline"]
    assert "gerrit-maven-java11-build-default" in r["pipeline"]
    assert "gerrit-maven-java11-build-edp" in r["pipeline"]

    assert "gerrit-gradle-java11-review" in r["pipeline"]
    assert "gerrit-gradle-java11-build-default" in r["pipeline"]
    assert "gerrit-gradle-java11-build-edp" in r["pipeline"]

    # ensure pipelines have proper steps
    for buildtool in ['maven', 'gradle']:
        for framework in ['java11']:

            gerrit_review_pipeline = f"gerrit-{buildtool}-{framework}-review"
            gerrit_build_pipeline_def = f"gerrit-{buildtool}-{framework}-build-default"
            gerrit_build_pipeline_edp = f"gerrit-{buildtool}-{framework}-build-edp"

            rt = r["pipeline"][gerrit_review_pipeline]["spec"]["tasks"]
            assert "fetch-repository" in rt[0]["name"]
            assert "gerrit-notify" in rt[1]["name"]
            assert "init-values" in rt[2]["name"]
            assert "compile" in rt[3]["name"]
            assert "test" in rt[4]["name"]
            assert "sonar" in rt[5]["name"]
            assert "dockerfile-lint" in rt[6]["name"]
            assert "dockerbuild-verify" in rt[7]["name"]
            assert "helm-lint" in rt[8]["name"]
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
            assert "sast" in btd[4]["name"]
            assert "compile" in btd[5]["name"]
            assert buildtool == btd[5]["taskRef"]["name"]
            assert "test" in btd[6]["name"]
            assert buildtool == btd[6]["taskRef"]["name"]
            assert "sonar" in btd[7]["name"]
            assert buildtool == btd[7]["taskRef"]["name"]
            assert "build" in btd[8]["name"]
            assert buildtool == btd[8]["taskRef"]["name"]
            assert "push" in btd[9]["name"]
            assert buildtool == btd[9]["taskRef"]["name"]
            assert "create-ecr-repository" in btd[10]["name"]
            assert "kaniko-build" in btd[11]["name"]
            assert "git-tag" in btd[12]["name"]
            assert "update-cbis" in btd[13]["name"]

            # build with edp versioning
            btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
            assert "fetch-repository" in btedp[0]["name"]
            assert "gerrit-notify" in btedp[1]["name"]
            assert "init-values" in btedp[2]["name"]
            assert "get-version" in btedp[3]["name"]
            assert "get-version-edp" == btedp[3]["taskRef"]["name"]
            idx = 3
            # we have update-build-number only for gradle
            if buildtool == "gradle":
                assert "update-build-number" in btedp[4]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
                idx = 4

            assert "sast" in btedp[idx+1]["name"]
            assert "compile" in btedp[idx+2]["name"]
            assert buildtool == btedp[idx+2]["taskRef"]["name"]
            assert "test" in btedp[idx+3]["name"]
            assert buildtool == btedp[idx+3]["taskRef"]["name"]
            assert "sonar" in btedp[idx+4]["name"]
            assert buildtool == btedp[idx+4]["taskRef"]["name"]
            assert "build" in btedp[idx+5]["name"]
            assert buildtool == btedp[idx+5]["taskRef"]["name"]
            assert "push" in btedp[idx+6]["name"]
            assert buildtool == btedp[idx+6]["taskRef"]["name"]
            assert "create-ecr-repository" in btedp[idx+7]["name"]
            assert "kaniko-build" in btedp[idx+8]["name"]
            assert "git-tag" in btedp[idx+9]["name"]
            assert "update-cbis" in btedp[idx+10]["name"]
