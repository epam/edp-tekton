from .helpers import helm_template


def test_javascript_pipelines_harbor_gerrit():
    config = """
global:
  gitProvider: gerrit
    """

    r = helm_template(config)

    # ensure pipelines have proper steps
    for buildtool in ['npm']:
        for framework in ['react', 'angular', 'vue', 'express', 'next']:
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
                    assert "fetch-target-branch" in rt[5]["name"]
                    assert "sonar-prepare-files" in rt[6]["name"]
                    assert "sonar-prepare-files-general" == rt[6]["taskRef"]["name"]
                    assert "sonar" in rt[7]["name"]
                    assert "save-cache" in rt[8]["name"]
                if cbtype == "app":
                    assert "fetch-repository" in rt[0]["name"]
                    assert "gerrit-notify" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "get-cache" in rt[4]["name"]
                    assert "build" in rt[5]["name"]
                    assert "fetch-target-branch" in rt[6]["name"]
                    assert "sonar-prepare-files" in rt[7]["name"]
                    assert "sonar-prepare-files-general" == rt[7]["taskRef"]["name"]
                    assert "sonar" in rt[8]["name"]
                    assert "dockerfile-lint" in rt[9]["name"]
                    assert "dockerbuild-verify" in rt[10]["name"]
                    assert "helm-lint" in rt[11]["name"]
                    assert "save-cache" in rt[12]["name"]

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
                assert "sonar-cleanup" in btd[6]["name"]
                assert "security" in btd[7]["name"]
                assert "build" in btd[8]["name"]
                assert "edp-npm" == btd[8]["taskRef"]["name"]
                assert "sonar" in btd[9]["name"]
                assert "sonarqube-scanner" == btd[9]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btd[10]["name"]
                assert "get-nexus-repository-url" == btd[10]["taskRef"]["name"]
                assert "push" in btd[11]["name"]
                assert "npm" == btd[11]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[12]["name"]
                    assert "save-cache" in btd[13]["name"]
                    assert "git-tag" in btd[14]["name"]
                    assert "update-cbis" in btd[15]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btd[12]["name"]
                    assert "git-tag" in btd[13]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "gerrit-notify" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "get-cache" in btedp[4]["taskRef"]["name"]
                assert "update-build-number" in btedp[5]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[5]["taskRef"]["name"]
                assert "sonar-cleanup" in btedp[6]["name"]
                assert "security" in btedp[7]["name"]
                assert "build" in btedp[8]["name"]
                assert "edp-npm" == btedp[8]["taskRef"]["name"]
                assert "sonar" in btedp[9]["name"]
                assert "sonarqube-scanner" == btedp[9]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btedp[10]["name"]
                assert "get-nexus-repository-url" == btedp[10]["taskRef"]["name"]
                assert "push" in btedp[11]["name"]
                assert "npm" == btedp[11]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[12]["name"]
                    assert "save-cache" in btedp[13]["name"]
                    assert "git-tag" in btedp[14]["name"]
                    assert "update-cbis" in btedp[15]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btedp[12]["name"]
                    assert "git-tag" in btedp[13]["name"]
                assert "update-cbb" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_javascript_pipelines_harbor_github():
    config = """
global:
  gitProvider: github
    """

    r = helm_template(config)
    vcs = "github"

    # ensure pipelines have proper steps
    for buildtool in ['npm']:
        for framework in ['react', 'angular', 'vue', 'express', 'next']:
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

                assert "github-set-success-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][0][
                    "name"]
                assert "github-set-failure-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][1][
                    "name"]

                # build with default versioning
                btd = r["pipeline"][github_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                # ensure we have default versioning
                assert "get-version" in btd[2]["name"]
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "get-cache" in btd[3]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "security" in btd[5]["name"]
                assert "build" in btd[6]["name"]
                assert "edp-npm" == btd[6]["taskRef"]["name"]
                assert "sonar" in btd[7]["name"]
                assert "sonarqube-scanner" == btd[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btd[8]["name"]
                assert "get-nexus-repository-url" == btd[8]["taskRef"]["name"]
                assert "push" in btd[9]["name"]
                assert "npm" == btd[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[10]["name"]
                    assert "save-cache" in btd[11]["name"]
                    assert "git-tag" in btd[12]["name"]
                    assert "update-cbis" in btd[13]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][github_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "get-cache" in btedp[3]["taskRef"]["name"]
                assert "update-build-number" in btedp[4]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
                assert "security" in btedp[5]["name"]
                assert "build" in btedp[6]["name"]
                assert "edp-npm" == btedp[6]["taskRef"]["name"]
                assert "sonar" in btedp[7]["name"]
                assert "sonarqube-scanner" == btedp[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btedp[8]["name"]
                assert "get-nexus-repository-url" == btedp[8]["taskRef"]["name"]
                assert "push" in btedp[9]["name"]
                assert "npm" == btedp[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[10]["name"]
                    assert "save-cache" in btedp[11]["name"]
                    assert "git-tag" in btedp[12]["name"]
                    assert "update-cbis" in btedp[13]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                assert "update-cbb" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_javascript_pipelines_harbor_gitlab():
    config = """
global:
  gitProvider: gitlab
    """

    r = helm_template(config)
    vcs = "gitlab"

    # ensure pipelines have proper steps
    for buildtool in ['npm']:
        for framework in ['react', 'angular', 'vue', 'express', 'next']:
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
                # ensure we have default versioning
                assert "get-version" in btd[2]["name"]
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "get-cache" in btd[3]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "security" in btd[5]["name"]
                assert "build" in btd[6]["name"]
                assert "edp-npm" == btd[6]["taskRef"]["name"]
                assert "sonar" in btd[7]["name"]
                assert "sonarqube-scanner" == btd[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btd[8]["name"]
                assert "get-nexus-repository-url" == btd[8]["taskRef"]["name"]
                assert "push" in btd[9]["name"]
                assert "npm" == btd[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[10]["name"]
                    assert "save-cache" in btd[11]["name"]
                    assert "git-tag" in btd[12]["name"]
                    assert "update-cbis" in btd[13]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gitlab_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "get-cache" in btedp[3]["taskRef"]["name"]
                assert "update-build-number" in btedp[4]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
                assert "security" in btedp[5]["name"]
                assert "build" in btedp[6]["name"]
                assert "edp-npm" == btedp[6]["taskRef"]["name"]
                assert "sonar" in btedp[7]["name"]
                assert "sonarqube-scanner" == btedp[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btedp[8]["name"]
                assert "get-nexus-repository-url" == btedp[8]["taskRef"]["name"]
                assert "push" in btedp[9]["name"]
                assert "npm" == btedp[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[10]["name"]
                    assert "save-cache" in btedp[11]["name"]
                    assert "git-tag" in btedp[12]["name"]
                    assert "update-cbis" in btedp[13]["name"]
                if cbtype == "lib":
                    assert "save-cache" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                assert "update-cbb" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][1]["name"]
