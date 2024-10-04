import pytest
from .helpers import helm_template


def test_python_common_pipelines_harbor_gerrit():
    config = """
global:
  gitProviders:
    - gerrit
    """

    r = helm_template(config)

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['python-3.8']:
            for cbtype in ['app', 'lib']:

                assert f"gerrit-{buildtool}-{framework}-{cbtype}-review" in r["pipeline"]
                assert f"gerrit-{buildtool}-{framework}-{cbtype}-build-default" in r["pipeline"]
                assert f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp" in r["pipeline"]

                gerrit_review_pipeline = f"gerrit-{buildtool}-{framework}-{cbtype}-review"
                gerrit_build_pipeline_def = f"gerrit-{buildtool}-{framework}-{cbtype}-build-default"
                gerrit_build_pipeline_edp = f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp"

                rt = r["pipeline"][gerrit_review_pipeline]["spec"]["tasks"]
                if cbtype == "lib":
                    assert "fetch-repository" in rt[0]["name"]
                    assert "gerrit-notify" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "get-cache" in rt[3]["name"]
                    assert "build" in rt[4]["name"]
                    assert "sonar" in rt[5]["name"]
                    assert "save-cache" in rt[6]["name"]
                if cbtype == "app":
                    assert "fetch-repository" in rt[0]["name"]
                    assert "gerrit-notify" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "get-cache" in rt[4]["name"]
                    assert "build" in rt[5]["name"]
                    assert "sonar" in rt[6]["name"]
                    assert "dockerfile-lint" in rt[7]["name"]
                    assert "dockerbuild-verify" in rt[8]["name"]
                    assert "helm-lint" in rt[9]["name"]
                    assert "save-cache" in rt[10]["name"]

                assert "gerrit-vote-success" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][0]["name"]
                assert "gerrit-vote-failure" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][gerrit_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "gerrit-notify" in btd[1]["name"]
                assert "init-values" in btd[2]["name"]
                assert "get-version" in btd[3]["name"]
                # ensure we have default versioning
                assert f"get-version-default" == btd[3]["taskRef"]["name"]
                assert "get-cache" in btd[4]["name"]
                assert "update-build-number" in btd[5]["name"]
                assert "security" in btd[6]["name"]
                assert "build" in btd[7]["name"]
                assert "python" == btd[7]["taskRef"]["name"]
                assert "sonar" in btd[8]["name"]
                assert "sonarqube-general" == btd[8]["taskRef"]["name"]
                assert "push" in btd[9]["name"]
                assert buildtool == btd[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[10]["name"]
                    assert "save-cache" in btd[11]["name"]
                    assert "git-tag" in btd[12]["name"]
                    assert "update-cbis" in btd[13]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "gerrit-notify" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "get-cache" in btedp[4]["name"]
                assert "update-build-number" in btedp[5]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[5]["taskRef"]["name"]
                assert "security" in btedp[6]["name"]
                assert "build" in btedp[7]["name"]
                assert "python" == btedp[7]["taskRef"]["name"]
                assert "sonar" in btedp[8]["name"]
                assert "sonarqube-general" == btedp[8]["taskRef"]["name"]
                assert "push" in btedp[9]["name"]
                assert buildtool == btedp[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[10]["name"]
                    assert "save-cache" in btedp[11]["name"]
                    assert "git-tag" in btedp[12]["name"]
                    assert "update-cbis" in btedp[13]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                assert "update-cbb" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_python_common_pipelines_harbor_github():
    config = """
global:
  gitProviders:
    - github
    """

    r = helm_template(config)
    vcs = "github"

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['python-3.8']:
            for cbtype in ['app', 'lib']:

                github_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                github_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                github_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert github_review_pipeline in r["pipeline"]
                assert github_build_pipeline_def in r["pipeline"]
                assert github_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][github_review_pipeline]["spec"]["tasks"]
                if cbtype == "lib":
                    assert "github-set-pending-status" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "get-cache" in rt[3]["name"]
                    assert "build" in rt[4]["name"]
                    assert "sonar" in rt[5]["name"]
                    assert "save-cache" in rt[6]["name"]
                if cbtype == "app":
                    assert "github-set-pending-status" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "get-cache" in rt[4]["name"]
                    assert "build" in rt[5]["name"]
                    assert "sonar" in rt[6]["name"]
                    assert "dockerfile-lint" in rt[7]["name"]
                    assert "dockerbuild-verify" in rt[8]["name"]
                    assert "helm-lint" in rt[9]["name"]
                    assert "save-cache" in rt[10]["name"]

                assert "github-set-success-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][github_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "get-cache" in btd[3]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "security" in btd[5]["name"]
                assert "build" in btd[6]["name"]
                assert "python" == btd[6]["taskRef"]["name"]
                assert "sonar" in btd[7]["name"]
                assert "sonarqube-general" == btd[7]["taskRef"]["name"]
                assert "push" in btd[8]["name"]
                assert buildtool == btd[8]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[9]["name"]
                    assert "save-cache" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                    assert "update-cbis" in btd[12]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btd[9]["name"]
                    assert "git-tag" in btd[10]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][github_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "get-cache" in btedp[3]["name"]
                assert "update-build-number" in btedp[4]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
                assert "security" in btedp[5]["name"]
                assert "build" in btedp[6]["name"]
                assert "python" == btedp[6]["taskRef"]["name"]
                assert "sonar" in btedp[7]["name"]
                assert "sonarqube-general" == btedp[7]["taskRef"]["name"]
                assert "push" in btedp[8]["name"]
                assert buildtool == btedp[8]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[9]["name"]
                    assert "save-cache" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                    assert "update-cbis" in btedp[12]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btedp[9]["name"]
                    assert "git-tag" in btedp[10]["name"]
                assert "update-cbb" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_python_common_pipelines_harbor_gitlab():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)
    vcs = "gitlab"

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['python-3.8']:
            for cbtype in ['app', 'lib']:

                gitlab_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                gitlab_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                gitlab_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert gitlab_review_pipeline in r["pipeline"]
                assert gitlab_build_pipeline_def in r["pipeline"]
                assert gitlab_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][gitlab_review_pipeline]["spec"]["tasks"]
                if cbtype == "lib":
                    assert "report-pipeline-start-to-gitlab" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "get-cache" in rt[3]["name"]
                    assert "build" in rt[4]["name"]
                    assert "sonar" in rt[5]["name"]
                    assert "save-cache" in rt[6]["name"]
                if cbtype == "app":
                    assert "report-pipeline-start-to-gitlab" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "get-cache" in rt[4]["name"]
                    assert "build" in rt[5]["name"]
                    assert "sonar" in rt[6]["name"]
                    assert "dockerfile-lint" in rt[7]["name"]
                    assert "dockerbuild-verify" in rt[8]["name"]
                    assert "helm-lint" in rt[9]["name"]
                    assert "save-cache" in rt[10]["name"]

                assert "gitlab-set-success-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][gitlab_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "get-cache" in btd[3]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "security" in btd[5]["name"]
                assert "build" in btd[6]["name"]
                assert "python" == btd[6]["taskRef"]["name"]
                assert "sonar" in btd[7]["name"]
                assert "sonarqube-general" == btd[7]["taskRef"]["name"]
                assert "push" in btd[8]["name"]
                assert buildtool == btd[8]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[9]["name"]
                    assert "save-cache" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                    assert "update-cbis" in btd[12]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btd[9]["name"]
                    assert "git-tag" in btd[10]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gitlab_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "get-cache" in btedp[3]["name"]
                assert "update-build-number" in btedp[4]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
                assert "security" in btedp[5]["name"]
                assert "build" in btedp[6]["name"]
                assert "python" == btedp[6]["taskRef"]["name"]
                assert "sonar" in btedp[7]["name"]
                assert "sonarqube-general" == btedp[7]["taskRef"]["name"]
                assert "push" in btedp[8]["name"]
                assert buildtool == btedp[8]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[9]["name"]
                    assert "save-cache" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                    assert "update-cbis" in btedp[12]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btedp[9]["name"]
                    assert "git-tag" in btedp[10]["name"]
                assert "update-cbb" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][1]["name"]

def test_python_common_pipelines_harbor_bitbucket():
    config = """
global:
  gitProviders:
    - bitbucket
    """

    r = helm_template(config)
    vcs = "bitbucket"

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['python-3.8']:
            for cbtype in ['app', 'lib']:

                bitbucket_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                bitbucket_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                bitbucket_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert bitbucket_review_pipeline in r["pipeline"]
                assert bitbucket_build_pipeline_def in r["pipeline"]
                assert bitbucket_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][bitbucket_review_pipeline]["spec"]["tasks"]
                if cbtype == "lib":
                    assert "bitbucket-set-pending-status" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "get-cache" in rt[3]["name"]
                    assert "build" in rt[4]["name"]
                    assert "sonar" in rt[5]["name"]
                    assert "save-cache" in rt[6]["name"]
                if cbtype == "app":
                    assert "bitbucket-set-pending-status" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "get-cache" in rt[4]["name"]
                    assert "build" in rt[5]["name"]
                    assert "sonar" in rt[6]["name"]
                    assert "dockerfile-lint" in rt[7]["name"]
                    assert "dockerbuild-verify" in rt[8]["name"]
                    assert "helm-lint" in rt[9]["name"]
                    assert "save-cache" in rt[10]["name"]

                assert "bitbucket-set-success-status" in r["pipeline"][bitbucket_review_pipeline]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-failure-status" in r["pipeline"][bitbucket_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][bitbucket_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "get-cache" in btd[3]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "security" in btd[5]["name"]
                assert "build" in btd[6]["name"]
                assert "python" == btd[6]["taskRef"]["name"]
                assert "sonar" in btd[7]["name"]
                assert "sonarqube-general" == btd[7]["taskRef"]["name"]
                assert "push" in btd[8]["name"]
                assert buildtool == btd[8]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[9]["name"]
                    assert "save-cache" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                    assert "update-cbis" in btd[12]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btd[9]["name"]
                    assert "git-tag" in btd[10]["name"]
                assert "push-to-jira" in r["pipeline"][bitbucket_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "get-cache" in btedp[3]["name"]
                assert "update-build-number" in btedp[4]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
                assert "security" in btedp[5]["name"]
                assert "build" in btedp[6]["name"]
                assert "python" == btedp[6]["taskRef"]["name"]
                assert "sonar" in btedp[7]["name"]
                assert "sonarqube-general" == btedp[7]["taskRef"]["name"]
                assert "push" in btedp[8]["name"]
                assert buildtool == btedp[8]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[9]["name"]
                    assert "save-cache" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                    assert "update-cbis" in btedp[12]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btedp[9]["name"]
                    assert "git-tag" in btedp[10]["name"]
                assert "update-cbb" in r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["finally"][1]["name"]


@pytest.mark.parametrize("framework", ["fastapi", "flask"])
def test_python_pipelines_harbor_gerrit(framework):
    config = """
global:
  gitProviders:
    - gerrit
    """

    ht = helm_template(config)

    buildtool = "python"

    review = f"gerrit-{buildtool}-{framework}-app-review"
    build_default = f"gerrit-{buildtool}-{framework}-app-build-default"
    build_edp = f"gerrit-{buildtool}-{framework}-app-build-edp"

    assert review in ht["pipeline"]
    assert build_default in ht["pipeline"]
    assert build_edp in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "fetch-repository" in r[0]["name"]
    assert "gerrit-notify" in r[1]["name"]
    assert "init-values" in r[2]["name"]
    assert "helm-docs" in r[3]["name"]
    assert "get-cache" in r[4]["name"]
    assert "build" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "save-cache" in r[10]["name"]
    assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "gerrit-notify" in bd[1]["name"]
    assert "init-values" in bd[2]["name"]
    assert "get-version" in bd[3]["name"]
    assert f"get-version-default" == bd[3]["taskRef"]["name"]
    assert "get-cache" in bd[4]["name"]
    assert "update-build-number" in bd[5]["name"]
    assert "security" in bd[6]["name"]
    assert "build" in bd[7]["name"]
    assert "python" == bd[7]["taskRef"]["name"]
    assert "sonar" in bd[8]["name"]
    assert "sonarqube-general" == bd[8]["taskRef"]["name"]
    assert "push" in bd[9]["name"]
    assert buildtool == bd[9]["taskRef"]["name"]
    assert "kaniko-build" in bd[10]["name"]
    assert "save-cache" in bd[11]["name"]
    assert "git-tag" in bd[12]["name"]
    assert "update-cbis" in bd[13]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "gerrit-notify" in bedp[1]["name"]
    assert "init-values" in bedp[2]["name"]
    assert "get-version" in bedp[3]["name"]
    assert "get-version-edp" == bedp[3]["taskRef"]["name"]
    assert "get-cache" in bedp[4]["name"]
    assert "update-build-number" in bedp[5]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[5]["taskRef"]["name"]
    assert "security" in bedp[6]["name"]
    assert "build" in bedp[7]["name"]
    assert "python" == bedp[7]["taskRef"]["name"]
    assert "sonar" in bedp[8]["name"]
    assert "sonarqube-general" == bedp[8]["taskRef"]["name"]
    assert "push" in bedp[9]["name"]
    assert buildtool == bedp[9]["taskRef"]["name"]
    assert "kaniko-build" in bedp[10]["name"]
    assert "save-cache" in bedp[11]["name"]
    assert "git-tag" in bedp[12]["name"]
    assert "update-cbis" in bedp[13]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


@pytest.mark.parametrize("framework", ["fastapi", "flask"])
def test_python_pipelines_harbor_gitlab(framework):
    config = """
global:
  gitProviders:
    - gitlab
    """

    ht = helm_template(config)

    buildtool = "python"

    review = f"gitlab-{buildtool}-{framework}-app-review"
    build_default = f"gitlab-{buildtool}-{framework}-app-build-default"
    build_edp = f"gitlab-{buildtool}-{framework}-app-build-edp"

    assert review in ht["pipeline"]
    assert build_default in ht["pipeline"]
    assert build_edp in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "report-pipeline-start-to-gitlab" in r[0]["name"]
    assert "fetch-repository" in r[1]["name"]
    assert "init-values" in r[2]["name"]
    assert "helm-docs" in r[3]["name"]
    assert "get-cache" in r[4]["name"]
    assert "build" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "save-cache" in r[10]["name"]
    assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "init-values" in bd[1]["name"]
    assert "get-version" in bd[2]["name"]
    assert f"get-version-default" == bd[2]["taskRef"]["name"]
    assert "get-cache" in bd[3]["name"]
    assert "update-build-number" in bd[4]["name"]
    assert "security" in bd[5]["name"]
    assert "build" in bd[6]["name"]
    assert "python" == bd[6]["taskRef"]["name"]
    assert "sonar" in bd[7]["name"]
    assert "sonarqube-general" == bd[7]["taskRef"]["name"]
    assert "push" in bd[8]["name"]
    assert buildtool == bd[8]["taskRef"]["name"]
    assert "kaniko-build" in bd[9]["name"]
    assert "save-cache" in bd[10]["name"]
    assert "git-tag" in bd[11]["name"]
    assert "update-cbis" in bd[12]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "init-values" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "get-version-edp" == bedp[2]["taskRef"]["name"]
    assert "get-cache" in bedp[3]["name"]
    assert "update-build-number" in bedp[4]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[4]["taskRef"]["name"]
    assert "security" in bedp[5]["name"]
    assert "build" in bedp[6]["name"]
    assert "python" == bedp[6]["taskRef"]["name"]
    assert "sonar" in bedp[7]["name"]
    assert "sonarqube-general" == bedp[7]["taskRef"]["name"]
    assert "push" in bedp[8]["name"]
    assert buildtool == bedp[8]["taskRef"]["name"]
    assert "kaniko-build" in bedp[9]["name"]
    assert "save-cache" in bedp[10]["name"]
    assert "git-tag" in bedp[11]["name"]
    assert "update-cbis" in bedp[12]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]

@pytest.mark.parametrize("framework", ["fastapi", "flask"])
def test_python_pipelines_harbor_bitbucket(framework):
    config = """
global:
  gitProviders:
    - bitbucket
    """

    ht = helm_template(config)

    buildtool = "python"

    review = f"bitbucket-{buildtool}-{framework}-app-review"
    build_default = f"bitbucket-{buildtool}-{framework}-app-build-default"
    build_edp = f"bitbucket-{buildtool}-{framework}-app-build-edp"

    assert review in ht["pipeline"]
    assert build_default in ht["pipeline"]
    assert build_edp in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "bitbucket-set-pending-status" in r[0]["name"]
    assert "fetch-repository" in r[1]["name"]
    assert "init-values" in r[2]["name"]
    assert "helm-docs" in r[3]["name"]
    assert "get-cache" in r[4]["name"]
    assert "build" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "save-cache" in r[10]["name"]
    assert "bitbucket-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "bitbucket-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "init-values" in bd[1]["name"]
    assert "get-version" in bd[2]["name"]
    assert f"get-version-default" == bd[2]["taskRef"]["name"]
    assert "get-cache" in bd[3]["name"]
    assert "update-build-number" in bd[4]["name"]
    assert "security" in bd[5]["name"]
    assert "build" in bd[6]["name"]
    assert "python" == bd[6]["taskRef"]["name"]
    assert "sonar" in bd[7]["name"]
    assert "sonarqube-general" == bd[7]["taskRef"]["name"]
    assert "push" in bd[8]["name"]
    assert buildtool == bd[8]["taskRef"]["name"]
    assert "kaniko-build" in bd[9]["name"]
    assert "save-cache" in bd[10]["name"]
    assert "git-tag" in bd[11]["name"]
    assert "update-cbis" in bd[12]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "init-values" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "get-version-edp" == bedp[2]["taskRef"]["name"]
    assert "get-cache" in bedp[3]["name"]
    assert "update-build-number" in bedp[4]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[4]["taskRef"]["name"]
    assert "security" in bedp[5]["name"]
    assert "build" in bedp[6]["name"]
    assert "python" == bedp[6]["taskRef"]["name"]
    assert "sonar" in bedp[7]["name"]
    assert "sonarqube-general" == bedp[7]["taskRef"]["name"]
    assert "push" in bedp[8]["name"]
    assert buildtool == bedp[8]["taskRef"]["name"]
    assert "kaniko-build" in bedp[9]["name"]
    assert "save-cache" in bedp[10]["name"]
    assert "git-tag" in bedp[11]["name"]
    assert "update-cbis" in bedp[12]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


@pytest.mark.parametrize("framework", ["fastapi", "flask"])
def test_python_pipelines_harbor_github(framework):
    config = """
global:
  gitProviders:
    - github
    """

    ht = helm_template(config)

    buildtool = "python"

    review = f"github-{buildtool}-{framework}-app-review"
    build_default = f"github-{buildtool}-{framework}-app-build-default"
    build_edp = f"github-{buildtool}-{framework}-app-build-edp"

    assert review in ht["pipeline"]
    assert build_default in ht["pipeline"]
    assert build_edp in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "github-set-pending-status" in r[0]["name"]
    assert "fetch-repository" in r[1]["name"]
    assert "init-values" in r[2]["name"]
    assert "helm-docs" in r[3]["name"]
    assert "get-cache" in r[4]["name"]
    assert "build" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "save-cache" in r[10]["name"]
    assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "init-values" in bd[1]["name"]
    assert "get-version" in bd[2]["name"]
    assert f"get-version-default" == bd[2]["taskRef"]["name"]
    assert "get-cache" in bd[3]["name"]
    assert "update-build-number" in bd[4]["name"]
    assert "security" in bd[5]["name"]
    assert "build" in bd[6]["name"]
    assert "python" == bd[6]["taskRef"]["name"]
    assert "sonar" in bd[7]["name"]
    assert "sonarqube-general" == bd[7]["taskRef"]["name"]
    assert "push" in bd[8]["name"]
    assert buildtool == bd[8]["taskRef"]["name"]
    assert "kaniko-build" in bd[9]["name"]
    assert "save-cache" in bd[10]["name"]
    assert "git-tag" in bd[11]["name"]
    assert "update-cbis" in bd[12]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "init-values" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "get-version-edp" == bedp[2]["taskRef"]["name"]
    assert "get-cache" in bedp[3]["name"]
    assert "update-build-number" in bedp[4]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[4]["taskRef"]["name"]
    assert "security" in bedp[5]["name"]
    assert "build" in bedp[6]["name"]
    assert "python" == bedp[6]["taskRef"]["name"]
    assert "sonar" in bedp[7]["name"]
    assert "sonarqube-general" == bedp[7]["taskRef"]["name"]
    assert "push" in bedp[8]["name"]
    assert buildtool == bedp[8]["taskRef"]["name"]
    assert "kaniko-build" in bedp[9]["name"]
    assert "save-cache" in bedp[10]["name"]
    assert "git-tag" in bedp[11]["name"]
    assert "update-cbis" in bedp[12]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]

from .helpers import helm_template


def test_ansible_pipelines_gerrit():
    config = """
global:
  gitProviders:
    - gerrit
    """

    r = helm_template(config)

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['ansible']:
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
                assert "ansible-lint" in rt[3]["name"]
                assert "ansible-tests" in rt[4]["name"]

                assert "gerrit-vote-success" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][0]["name"]
                assert "gerrit-vote-failure" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][gerrit_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "gerrit-notify" in btd[1]["name"]
                assert "init-values" in btd[2]["name"]
                assert "get-version" in btd[3]["name"]
                # ensure we have default versioning
                assert f"get-version-default" == btd[3]["taskRef"]["name"]
                assert "ansible-lint" in btd[4]["name"]
                assert "ansible-tests" in btd[5]["name"]
                assert "git-tag" in btd[6]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "gerrit-notify" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "ansible-lint" in btedp[4]["name"]
                assert "ansible-tests" in btedp[5]["name"]
                assert "git-tag" in btedp[6]["name"]
                assert "update-cbb" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][1]["name"]

def test_ansible_pipelines_github():
    config = """
global:
  gitProviders:
    - github
    """

    r = helm_template(config)
    vcs = "github"

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['ansible']:
            for cbtype in ['lib']:

                github_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                github_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                github_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert github_review_pipeline in r["pipeline"]
                assert github_build_pipeline_def in r["pipeline"]
                assert github_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][github_review_pipeline]["spec"]["tasks"]
                assert "github-set-pending-status" in rt[0]["name"]
                assert "fetch-repository" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "ansible-lint" in rt[3]["name"]
                assert "ansible-tests" in rt[4]["name"]

                assert "github-set-success-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][github_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                # ensure we have default versioning
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "ansible-lint" in btd[3]["name"]
                assert "ansible-tests" in btd[4]["name"]
                assert "git-tag" in btd[5]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][github_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "ansible-lint" in btedp[3]["name"]
                assert "ansible-tests" in btedp[4]["name"]
                assert "git-tag" in btedp[5]["name"]
                assert "update-cbb" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][1]["name"]

def test_ansible_pipelines_gitlab():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)
    vcs = "gitlab"

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['ansible']:
            for cbtype in ['lib']:

                gitlab_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                gitlab_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                gitlab_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert gitlab_review_pipeline in r["pipeline"]
                assert gitlab_build_pipeline_def in r["pipeline"]
                assert gitlab_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][gitlab_review_pipeline]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in rt[0]["name"]
                assert "fetch-repository" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "ansible-lint" in rt[3]["name"]
                assert "ansible-tests" in rt[4]["name"]

                assert "gitlab-set-success-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][gitlab_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                # ensure we have default versioning
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "ansible-lint" in btd[3]["name"]
                assert "ansible-tests" in btd[4]["name"]
                assert "git-tag" in btd[5]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gitlab_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "ansible-lint" in btedp[3]["name"]
                assert "ansible-tests" in btedp[4]["name"]
                assert "git-tag" in btedp[5]["name"]
                assert "update-cbb" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][1]["name"]

def test_ansible_pipelines_bitbucket():
    config = """
global:
  gitProviders:
    - bitbucket
    """

    r = helm_template(config)
    vcs = "bitbucket"

    # ensure pipelines have proper steps
    for buildtool in ['python']:
        for framework in ['ansible']:
            for cbtype in ['lib']:

                bitbucket_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                bitbucket_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                bitbucket_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert bitbucket_review_pipeline in r["pipeline"]
                assert bitbucket_build_pipeline_def in r["pipeline"]
                assert bitbucket_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][bitbucket_review_pipeline]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in rt[0]["name"]
                assert "fetch-repository" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "ansible-lint" in rt[3]["name"]
                assert "ansible-tests" in rt[4]["name"]

                assert "bitbucket-set-success-status" in r["pipeline"][bitbucket_review_pipeline]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-failure-status" in r["pipeline"][bitbucket_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][bitbucket_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                # ensure we have default versioning
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "ansible-lint" in btd[3]["name"]
                assert "ansible-tests" in btd[4]["name"]
                assert "git-tag" in btd[5]["name"]
                assert "push-to-jira" in r["pipeline"][bitbucket_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "ansible-lint" in btedp[3]["name"]
                assert "ansible-tests" in btedp[4]["name"]
                assert "git-tag" in btedp[5]["name"]
                assert "update-cbb" in r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["finally"][1]["name"]
