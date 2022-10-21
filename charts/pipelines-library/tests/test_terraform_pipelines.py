import os
import sys

from .helpers import helm_template


def test_terraform_pipelines_gerrit():
    config = """
gerrit:
  enabled: true
    """

    r = helm_template(config)

    # ensure pipelines have proper steps
    for buildtool in ['terraform']:
        for framework in ['terraform']:
            for cbtype in ['lib']:

                assert f"gerrit-{buildtool}-{framework}-{cbtype}-review" in r["pipeline"]
                assert f"gerrit-{buildtool}-{framework}-{cbtype}-build-default" in r["pipeline"]
                assert f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp" in r["pipeline"]

                gerrit_review_pipeline = f"gerrit-{buildtool}-{framework}-{cbtype}-review"
                gerrit_build_pipeline_def = f"gerrit-{buildtool}-{framework}-{cbtype}-build-default"
                gerrit_build_pipeline_edp = f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp"

                rt = r["pipeline"][gerrit_review_pipeline]["spec"]["tasks"]
                assert "fetch-repository" in rt[0]["name"]
                assert "gerrit-notify" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "terraform-format" in rt[3]["name"]
                assert "terraform-lint" in rt[4]["name"]

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
                assert "terraform-format" in btd[4]["name"]
                assert buildtool == btd[4]["taskRef"]["name"]
                assert "terraform-lint" in btd[5]["name"]
                assert buildtool == btd[5]["taskRef"]["name"]
                assert "git-tag" in btd[6]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "gerrit-notify" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "terraform-format" in btedp[4]["name"]
                assert buildtool == btedp[4]["taskRef"]["name"]
                assert "terraform-lint" in btedp[5]["name"]
                assert buildtool == btedp[5]["taskRef"]["name"]
                assert "git-tag" in btedp[6]["name"]
