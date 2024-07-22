from .helpers import helm_template

def test_go_pipelines_harbor_gerrit():
    config = """
global:
  gitProviders:
    - gerrit
    """

    r = helm_template(config)

    # ensure pipelines have proper steps
    for buildtool in ['go']:
        for framework in ['beego', 'gin', 'operator-sdk']:

            assert f"gerrit-go-{framework}-app-review" in r["pipeline"]
            assert f"gerrit-go-{framework}-app-build-default" in r["pipeline"]
            assert f"gerrit-go-{framework}-app-build-edp" in r["pipeline"]

            gerrit_review_pipeline = f"gerrit-{buildtool}-{framework}-app-review"
            gerrit_build_pipeline_def = f"gerrit-{buildtool}-{framework}-app-build-default"
            gerrit_build_pipeline_edp = f"gerrit-{buildtool}-{framework}-app-build-edp"

            rt = r["pipeline"][gerrit_review_pipeline]["spec"]["tasks"]
            assert "fetch-repository" in rt[0]["name"]
            assert "gerrit-notify" in rt[1]["name"]
            assert "init-values" in rt[2]["name"]
            assert "get-cache" in rt[3]["name"]
            assert "helm-docs" in rt[4]["name"]
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
            assert "security" in btd[5]["name"]
            assert "build" in btd[6]["name"]
            assert "golang" == btd[6]["taskRef"]["name"]
            assert "sonar" in btd[7]["name"]
            assert "sonarqube-general" == btd[7]["taskRef"]["name"]
            assert "kaniko-build" in btd[8]["name"]
            assert "save-cache" in btd[9]["name"]
            assert "git-tag" in btd[10]["name"]
            assert "update-cbis" in btd[11]["name"]
            assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_def]["spec"]["finally"][0]["name"]

            # build with edp versioning
            btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
            assert "fetch-repository" in btedp[0]["name"]
            assert "gerrit-notify" in btedp[1]["name"]
            assert "init-values" in btedp[2]["name"]
            assert "get-version" in btedp[3]["name"]
            assert "get-version-edp" == btedp[3]["taskRef"]["name"]
            assert "get-cache" in btedp[4]["name"]
            assert "security" in btedp[5]["name"]
            assert "build" in btedp[6]["name"]
            assert "golang" == btedp[6]["taskRef"]["name"]
            assert "sonar" in btedp[7]["name"]
            assert "sonarqube-general" == btedp[7]["taskRef"]["name"]
            assert "kaniko-build" in btedp[8]["name"]
            assert "save-cache" in btedp[9]["name"]
            assert "git-tag" in btedp[10]["name"]
            assert "update-cbis" in btedp[11]["name"]
            assert "update-cbb" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_go_pipelines_harbor_github():
    config = """
global:
  gitProviders:
    - github
    """

    r = helm_template(config)
    vcs = "github"

    # ensure pipelines have proper steps
    for buildtool in ['go']:
        for framework in ['beego', 'gin', 'operator-sdk']:

            github_review_pipeline = f"{vcs}-{buildtool}-{framework}-app-review"
            github_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-app-build-default"
            github_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-app-build-edp"

            assert github_review_pipeline in r["pipeline"]
            assert github_build_pipeline_def in r["pipeline"]
            assert github_build_pipeline_edp in r["pipeline"]

            rt = r["pipeline"][github_review_pipeline]["spec"]["tasks"]
            assert "github-set-pending-status" in rt[0]["name"]
            assert "fetch-repository" in rt[1]["name"]
            assert "init-values" in rt[2]["name"]
            assert "get-cache" in rt[3]["name"]
            assert "helm-docs" in rt[4]["name"]
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
            # ensure we have default versioning
            assert f"get-version-default" == btd[2]["taskRef"]["name"]
            assert "get-cache" in btd[3]["name"]
            assert "security" in btd[4]["name"]
            assert "build" in btd[5]["name"]
            assert "golang" == btd[5]["taskRef"]["name"]
            assert "sonar" in btd[6]["name"]
            assert "sonarqube-general" == btd[6]["taskRef"]["name"]
            assert "kaniko-build" in btd[7]["name"]
            assert "save-cache" in btd[8]["name"]
            assert "git-tag" in btd[9]["name"]
            assert "update-cbis" in btd[10]["name"]
            assert "push-to-jira" in r["pipeline"][github_build_pipeline_def]["spec"]["finally"][0]["name"]

            # build with edp versioning
            btedp = r["pipeline"][github_build_pipeline_edp]["spec"]["tasks"]
            assert "fetch-repository" in btedp[0]["name"]
            assert "init-values" in btedp[1]["name"]
            assert "get-version" in btedp[2]["name"]
            assert "get-version-edp" == btedp[2]["taskRef"]["name"]
            assert "get-cache" in btd[3]["name"]
            assert "security" in btedp[4]["name"]
            assert "build" in btedp[5]["name"]
            assert "golang" == btedp[5]["taskRef"]["name"]
            assert "sonar" in btedp[6]["name"]
            assert "sonarqube-general" == btedp[6]["taskRef"]["name"]
            assert "kaniko-build" in btedp[7]["name"]
            assert "save-cache" in btedp[8]["name"]
            assert "git-tag" in btedp[9]["name"]
            assert "update-cbis" in btedp[10]["name"]
            assert "update-cbb" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_go_pipelines_harbor_gitlab():
    config = """
global:
  gitProviders:
    - gitlab
    """

    r = helm_template(config)
    vcs = "gitlab"

    # ensure pipelines have proper steps
    for buildtool in ['go']:
        for framework in ['beego', 'gin', 'operator-sdk']:

            review_pipeline = f"{vcs}-{buildtool}-{framework}-app-review"
            build_pipeline_def = f"{vcs}-{buildtool}-{framework}-app-build-default"
            build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-app-build-edp"

            assert review_pipeline in r["pipeline"]
            assert build_pipeline_def in r["pipeline"]
            assert build_pipeline_edp in r["pipeline"]

            rt = r["pipeline"][review_pipeline]["spec"]["tasks"]
            assert "report-pipeline-start-to-gitlab" in rt[0]["name"]
            assert "fetch-repository" in rt[1]["name"]
            assert "init-values" in rt[2]["name"]
            assert "get-cache" in rt[3]["name"]
            assert "helm-docs" in rt[4]["name"]
            assert "build" in rt[5]["name"]
            assert "sonar" in rt[6]["name"]
            assert "dockerfile-lint" in rt[7]["name"]
            assert "dockerbuild-verify" in rt[8]["name"]
            assert "helm-lint" in rt[9]["name"]
            assert "save-cache" in rt[10]["name"]
            assert "gitlab-set-success-status" in r["pipeline"][review_pipeline]["spec"]["finally"][0]["name"]
            assert "gitlab-set-failure-status" in r["pipeline"][review_pipeline]["spec"]["finally"][1]["name"]

            # build with default versioning
            btd = r["pipeline"][build_pipeline_def]["spec"]["tasks"]
            assert "fetch-repository" in btd[0]["name"]
            assert "init-values" in btd[1]["name"]
            assert "get-version" in btd[2]["name"]
            # ensure we have default versioning
            assert f"get-version-default" == btd[2]["taskRef"]["name"]
            assert "get-cache" in btd[3]["name"]
            assert "security" in btd[4]["name"]
            assert "build" in btd[5]["name"]
            assert "golang" == btd[5]["taskRef"]["name"]
            assert "sonar" in btd[6]["name"]
            assert "sonarqube-general" == btd[6]["taskRef"]["name"]
            assert "kaniko-build" in btd[7]["name"]
            assert "save-cache" in btd[8]["name"]
            assert "git-tag" in btd[9]["name"]
            assert "update-cbis" in btd[10]["name"]
            assert "push-to-jira" in r["pipeline"][build_pipeline_def]["spec"]["finally"][0]["name"]

            # build with edp versioning
            btedp = r["pipeline"][build_pipeline_edp]["spec"]["tasks"]
            assert "fetch-repository" in btedp[0]["name"]
            assert "init-values" in btedp[1]["name"]
            assert "get-version" in btedp[2]["name"]
            assert "get-version-edp" == btedp[2]["taskRef"]["name"]
            assert "get-cache" in btd[3]["name"]
            assert "security" in btedp[4]["name"]
            assert "build" in btedp[5]["name"]
            assert "golang" == btedp[5]["taskRef"]["name"]
            assert "sonar" in btedp[6]["name"]
            assert "sonarqube-general" == btedp[6]["taskRef"]["name"]
            assert "kaniko-build" in btedp[7]["name"]
            assert "save-cache" in btedp[8]["name"]
            assert "git-tag" in btedp[9]["name"]
            assert "update-cbis" in btedp[10]["name"]
            assert "update-cbb" in r["pipeline"][build_pipeline_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in r["pipeline"][build_pipeline_edp]["spec"]["finally"][1]["name"]
