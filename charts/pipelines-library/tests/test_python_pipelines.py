from .helpers import helm_template


def test_python_pipelines_harbor_gerrit():
    config = """
global:
  gitProvider: gerrit
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
                    assert "compile" in rt[3]["name"]
                    assert "test" in rt[4]["name"]
                    assert "fetch-target-branch" in rt[5]["name"]
                    assert "sonar-prepare-files" in rt[6]["name"]
                    assert "sonar-prepare-files-general" == rt[6]["taskRef"]["name"]
                    assert "sonar" in rt[7]["name"]
                if cbtype == "app":
                    assert "fetch-repository" in rt[0]["name"]
                    assert "gerrit-notify" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "compile" in rt[4]["name"]
                    assert "test" in rt[5]["name"]
                    assert "fetch-target-branch" in rt[6]["name"]
                    assert "sonar-prepare-files" in rt[7]["name"]
                    assert "sonar-prepare-files-general" == rt[7]["taskRef"]["name"]
                    assert "sonar" in rt[8]["name"]
                    assert "dockerfile-lint" in rt[9]["name"]
                    assert "dockerbuild-verify" in rt[10]["name"]
                    assert "helm-lint" in rt[11]["name"]

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
                assert "update-build-number" in btd[4]["name"]
                assert "sonar-cleanup" in btd[5]["name"]
                assert "sast" in btd[6]["name"]
                assert "compile" in btd[7]["name"]
                assert "test" in btd[8]["name"]
                assert buildtool == btd[8]["taskRef"]["name"]
                assert "sonar" in btd[9]["name"]
                assert "sonarqube-scanner" == btd[9]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btd[10]["name"]
                assert "get-nexus-repository-url" == btd[10]["taskRef"]["name"]
                assert "push" in btd[11]["name"]
                assert buildtool == btd[11]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[12]["name"]
                    assert "git-tag" in btd[13]["name"]
                    assert "update-cbis" in btd[14]["name"]
                if cbtype == "lib":
                    assert "git-tag" in btd[12]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "gerrit-notify" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "update-build-number" in btedp[4]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[4]["taskRef"]["name"]
                assert "sonar-cleanup" in btedp[5]["name"]
                assert "sast" in btedp[6]["name"]
                assert "compile" in btedp[7]["name"]
                assert buildtool == btedp[7]["taskRef"]["name"]
                assert "test" in btedp[8]["name"]
                assert buildtool == btedp[8]["taskRef"]["name"]
                assert "sonar" in btedp[9]["name"]
                assert "sonarqube-scanner" == btedp[9]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btedp[10]["name"]
                assert "get-nexus-repository-url" == btedp[10]["taskRef"]["name"]
                assert "push" in btedp[11]["name"]
                assert buildtool == btedp[11]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[12]["name"]
                    assert "git-tag" in btedp[13]["name"]
                    assert "update-cbis" in btedp[14]["name"]
                if cbtype == "lib":
                    assert "git-tag" in btedp[12]["name"]
                assert "update-cbb" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gerrit_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_python_pipelines_harbor_github():
    config = """
global:
  gitProvider: github
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
                    assert "compile" in rt[3]["name"]
                    assert "test" in rt[4]["name"]
                    assert "sonar" in rt[5]["name"]
                if cbtype == "app":
                    assert "github-set-pending-status" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "compile" in rt[4]["name"]
                    assert "test" in rt[5]["name"]
                    assert "sonar" in rt[6]["name"]
                    assert "dockerfile-lint" in rt[7]["name"]
                    assert "dockerbuild-verify" in rt[8]["name"]
                    assert "helm-lint" in rt[9]["name"]

                assert "github-set-success-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][github_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "update-build-number" in btd[3]["name"]
                assert "sast" in btd[4]["name"]
                assert "compile" in btd[5]["name"]
                assert "test" in btd[6]["name"]
                assert buildtool == btd[6]["taskRef"]["name"]
                assert "sonar" in btd[7]["name"]
                assert "sonarqube-scanner" == btd[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btd[8]["name"]
                assert "get-nexus-repository-url" == btd[8]["taskRef"]["name"]
                assert "push" in btd[9]["name"]
                assert buildtool == btd[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                    assert "update-cbis" in btd[12]["name"]
                if cbtype == "lib":
                    assert "git-tag" in btd[10]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][github_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "update-build-number" in btedp[3]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[3]["taskRef"]["name"]
                assert "sast" in btedp[4]["name"]
                assert "compile" in btedp[5]["name"]
                assert buildtool == btedp[5]["taskRef"]["name"]
                assert "test" in btedp[6]["name"]
                assert buildtool == btedp[6]["taskRef"]["name"]
                assert "sonar" in btedp[7]["name"]
                assert "sonarqube-scanner" == btedp[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btedp[8]["name"]
                assert "get-nexus-repository-url" == btedp[8]["taskRef"]["name"]
                assert "push" in btedp[9]["name"]
                assert buildtool == btedp[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                    assert "update-cbis" in btedp[12]["name"]
                if cbtype == "lib":
                    assert "git-tag" in btedp[10]["name"]
                assert "update-cbb" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][1]["name"]


def test_python_pipelines_harbor_gitlab():
    config = """
global:
  gitProvider: gitlab
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
                    assert "compile" in rt[3]["name"]
                    assert "test" in rt[4]["name"]
                    assert "sonar" in rt[5]["name"]
                if cbtype == "app":
                    assert "report-pipeline-start-to-gitlab" in rt[0]["name"]
                    assert "fetch-repository" in rt[1]["name"]
                    assert "init-values" in rt[2]["name"]
                    assert "helm-docs" in rt[3]["name"]
                    assert "compile" in rt[4]["name"]
                    assert "test" in rt[5]["name"]
                    assert "sonar" in rt[6]["name"]
                    assert "dockerfile-lint" in rt[7]["name"]
                    assert "dockerbuild-verify" in rt[8]["name"]
                    assert "helm-lint" in rt[9]["name"]

                assert "gitlab-set-success-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][gitlab_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "init-values" in btd[1]["name"]
                assert "get-version" in btd[2]["name"]
                assert f"get-version-default" == btd[2]["taskRef"]["name"]
                assert "update-build-number" in btd[3]["name"]
                assert "sast" in btd[4]["name"]
                assert "compile" in btd[5]["name"]
                assert "test" in btd[6]["name"]
                assert buildtool == btd[6]["taskRef"]["name"]
                assert "sonar" in btd[7]["name"]
                assert "sonarqube-scanner" == btd[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btd[8]["name"]
                assert "get-nexus-repository-url" == btd[8]["taskRef"]["name"]
                assert "push" in btd[9]["name"]
                assert buildtool == btd[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btd[10]["name"]
                    assert "git-tag" in btd[11]["name"]
                    assert "update-cbis" in btd[12]["name"]
                if cbtype == "lib":
                    assert "git-tag" in btd[10]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_def]["spec"]["finally"][0]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gitlab_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "init-values" in btedp[1]["name"]
                assert "get-version" in btedp[2]["name"]
                assert "get-version-edp" == btedp[2]["taskRef"]["name"]
                assert "update-build-number" in btedp[3]["taskRef"]["name"]
                assert f"update-build-number-{buildtool}" == btedp[3]["taskRef"]["name"]
                assert "sast" in btedp[4]["name"]
                assert "compile" in btedp[5]["name"]
                assert buildtool == btedp[5]["taskRef"]["name"]
                assert "test" in btedp[6]["name"]
                assert buildtool == btedp[6]["taskRef"]["name"]
                assert "sonar" in btedp[7]["name"]
                assert "sonarqube-scanner" == btedp[7]["taskRef"]["name"]
                assert "get-nexus-repository-url" in btedp[8]["name"]
                assert "get-nexus-repository-url" == btedp[8]["taskRef"]["name"]
                assert "push" in btedp[9]["name"]
                assert buildtool == btedp[9]["taskRef"]["name"]
                if cbtype == "app":
                    assert "kaniko-build" in btedp[10]["name"]
                    assert "git-tag" in btedp[11]["name"]
                    assert "update-cbis" in btedp[12]["name"]
                if cbtype == "lib":
                    assert "git-tag" in btedp[10]["name"]
                assert "update-cbb" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][1]["name"]


# FastApi
def test_python_fastapi_pipelines_harbor_gerrit():
    config = """
global:
  gitProvider: gerrit
    """

    ht = helm_template(config)

    buildtool = "python"
    framework = "fastapi"

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
    assert "test" in r[4]["name"]
    assert "lint" in r[5]["name"]
    assert "fetch-target-branch" in r[6]["name"]
    assert "sonar-prepare-files" in r[7]["name"]
    assert "sonar-prepare-files-general" == r[7]["taskRef"]["name"]
    assert "sonar" in r[8]["name"]
    assert "dockerfile-lint" in r[9]["name"]
    assert "dockerbuild-verify" in r[10]["name"]
    assert "helm-lint" in r[11]["name"]
    assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "gerrit-notify" in bd[1]["name"]
    assert "init-values" in bd[2]["name"]
    assert "get-version" in bd[3]["name"]
    assert f"get-version-default" == bd[3]["taskRef"]["name"]
    assert "update-build-number" in bd[4]["name"]
    assert "sonar-cleanup" in bd[5]["name"]
    assert "sast" in bd[6]["name"]
    assert "test" in bd[7]["name"]
    assert buildtool == bd[7]["taskRef"]["name"]
    assert "lint" in bd[8]["name"]
    assert buildtool == bd[8]["taskRef"]["name"]
    assert "compile" in bd[9]["name"]
    assert buildtool == bd[9]["taskRef"]["name"]
    assert "sonar" in bd[10]["name"]
    assert "sonarqube-scanner" == bd[10]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bd[11]["name"]
    assert "get-nexus-repository-url" == bd[11]["taskRef"]["name"]
    assert "push" in bd[12]["name"]
    assert buildtool == bd[12]["taskRef"]["name"]
    assert "kaniko-build" in bd[13]["name"]
    assert "git-tag" in bd[14]["name"]
    assert "update-cbis" in bd[15]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "gerrit-notify" in bedp[1]["name"]
    assert "init-values" in bedp[2]["name"]
    assert "get-version" in bedp[3]["name"]
    assert "get-version-edp" == bedp[3]["taskRef"]["name"]
    assert "update-build-number" in bedp[4]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[4]["taskRef"]["name"]
    assert "sonar-cleanup" in bedp[5]["name"]
    assert "sast" in bedp[6]["name"]
    assert "test" in bedp[7]["name"]
    assert buildtool == bedp[7]["taskRef"]["name"]
    assert "lint" in bedp[8]["name"]
    assert buildtool == bedp[8]["taskRef"]["name"]
    assert "compile" in bedp[9]["name"]
    assert buildtool == bedp[9]["taskRef"]["name"]
    assert "sonar" in bedp[10]["name"]
    assert "sonarqube-scanner" == bedp[10]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bedp[11]["name"]
    assert "get-nexus-repository-url" == bedp[11]["taskRef"]["name"]
    assert "push" in bedp[12]["name"]
    assert buildtool == bedp[12]["taskRef"]["name"]
    assert "kaniko-build" in bedp[13]["name"]
    assert "git-tag" in bedp[14]["name"]
    assert "update-cbis" in bedp[15]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


# Flask
def test_python_flask_pipelines_harbor_gerrit():
    config = """
global:
  gitProvider: gerrit
    """

    ht = helm_template(config)

    buildtool = "python"
    framework = "flask"

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
    assert "test" in r[4]["name"]
    assert "lint" in r[5]["name"]
    assert "fetch-target-branch" in r[6]["name"]
    assert "sonar-prepare-files" in r[7]["name"]
    assert "sonar-prepare-files-general" == r[7]["taskRef"]["name"]
    assert "sonar" in r[8]["name"]
    assert "dockerfile-lint" in r[9]["name"]
    assert "dockerbuild-verify" in r[10]["name"]
    assert "helm-lint" in r[11]["name"]
    assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "gerrit-notify" in bd[1]["name"]
    assert "init-values" in bd[2]["name"]
    assert "get-version" in bd[3]["name"]
    assert f"get-version-default" == bd[3]["taskRef"]["name"]
    assert "update-build-number" in bd[4]["name"]
    assert "sonar-cleanup" in bd[5]["name"]
    assert "sast" in bd[6]["name"]
    assert "test" in bd[7]["name"]
    assert buildtool == bd[7]["taskRef"]["name"]
    assert "lint" in bd[8]["name"]
    assert buildtool == bd[8]["taskRef"]["name"]
    assert "compile" in bd[9]["name"]
    assert buildtool == bd[9]["taskRef"]["name"]
    assert "sonar" in bd[10]["name"]
    assert "sonarqube-scanner" == bd[10]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bd[11]["name"]
    assert "get-nexus-repository-url" == bd[11]["taskRef"]["name"]
    assert "push" in bd[12]["name"]
    assert buildtool == bd[12]["taskRef"]["name"]
    assert "kaniko-build" in bd[13]["name"]
    assert "git-tag" in bd[14]["name"]
    assert "update-cbis" in bd[15]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "gerrit-notify" in bedp[1]["name"]
    assert "init-values" in bedp[2]["name"]
    assert "get-version" in bedp[3]["name"]
    assert "get-version-edp" == bedp[3]["taskRef"]["name"]
    assert "update-build-number" in bedp[4]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[4]["taskRef"]["name"]
    assert "sonar-cleanup" in bedp[5]["name"]
    assert "sast" in bedp[6]["name"]
    assert "test" in bedp[7]["name"]
    assert buildtool == bedp[7]["taskRef"]["name"]
    assert "lint" in bedp[8]["name"]
    assert buildtool == bedp[8]["taskRef"]["name"]
    assert "compile" in bedp[9]["name"]
    assert buildtool == bedp[9]["taskRef"]["name"]
    assert "sonar" in bedp[10]["name"]
    assert "sonarqube-scanner" == bedp[10]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bedp[11]["name"]
    assert "get-nexus-repository-url" == bedp[11]["taskRef"]["name"]
    assert "push" in bedp[12]["name"]
    assert buildtool == bedp[12]["taskRef"]["name"]
    assert "kaniko-build" in bedp[13]["name"]
    assert "git-tag" in bedp[14]["name"]
    assert "update-cbis" in bedp[15]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_python_flask_pipelines_harbor_gitlab():
    config = """
global:
  gitProvider: gitlab
    """

    ht = helm_template(config)

    buildtool = "python"
    framework = "flask"

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
    assert "test" in r[4]["name"]
    assert "lint" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "init-values" in bd[1]["name"]
    assert "get-version" in bd[2]["name"]
    assert f"get-version-default" == bd[2]["taskRef"]["name"]
    assert "update-build-number" in bd[3]["name"]
    assert "sast" in bd[4]["name"]
    assert "test" in bd[5]["name"]
    assert buildtool == bd[5]["taskRef"]["name"]
    assert "lint" in bd[6]["name"]
    assert buildtool == bd[6]["taskRef"]["name"]
    assert "compile" in bd[7]["name"]
    assert buildtool == bd[7]["taskRef"]["name"]
    assert "sonar" in bd[8]["name"]
    assert "sonarqube-scanner" == bd[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bd[9]["name"]
    assert "get-nexus-repository-url" == bd[9]["taskRef"]["name"]
    assert "push" in bd[10]["name"]
    assert buildtool == bd[10]["taskRef"]["name"]
    assert "kaniko-build" in bd[11]["name"]
    assert "git-tag" in bd[12]["name"]
    assert "update-cbis" in bd[13]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "init-values" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "get-version-edp" == bedp[2]["taskRef"]["name"]
    assert "update-build-number" in bedp[3]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[3]["taskRef"]["name"]
    assert "sast" in bd[4]["name"]
    assert "test" in bd[5]["name"]
    assert buildtool == bd[5]["taskRef"]["name"]
    assert "lint" in bd[6]["name"]
    assert buildtool == bd[6]["taskRef"]["name"]
    assert "compile" in bedp[7]["name"]
    assert buildtool == bedp[7]["taskRef"]["name"]
    assert "sonar" in bedp[8]["name"]
    assert "sonarqube-scanner" == bedp[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bedp[9]["name"]
    assert "get-nexus-repository-url" == bedp[9]["taskRef"]["name"]
    assert "push" in bedp[10]["name"]
    assert buildtool == bedp[10]["taskRef"]["name"]
    assert "kaniko-build" in bedp[11]["name"]
    assert "git-tag" in bedp[12]["name"]
    assert "update-cbis" in bedp[13]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_python_fastapi_pipelines_harbor_gitlab():
    config = """
global:
  gitProvider: gitlab
    """

    ht = helm_template(config)

    buildtool = "python"
    framework = "fastapi"

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
    assert "test" in r[4]["name"]
    assert "lint" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "init-values" in bd[1]["name"]
    assert "get-version" in bd[2]["name"]
    assert f"get-version-default" == bd[2]["taskRef"]["name"]
    assert "update-build-number" in bd[3]["name"]
    assert "sast" in bd[4]["name"]
    assert "test" in bd[5]["name"]
    assert buildtool == bd[5]["taskRef"]["name"]
    assert "lint" in bd[6]["name"]
    assert buildtool == bd[6]["taskRef"]["name"]
    assert "compile" in bd[7]["name"]
    assert buildtool == bd[7]["taskRef"]["name"]
    assert "sonar" in bd[8]["name"]
    assert "sonarqube-scanner" == bd[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bd[9]["name"]
    assert "get-nexus-repository-url" == bd[9]["taskRef"]["name"]
    assert "push" in bd[10]["name"]
    assert buildtool == bd[10]["taskRef"]["name"]
    assert "kaniko-build" in bd[11]["name"]
    assert "git-tag" in bd[12]["name"]
    assert "update-cbis" in bd[13]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "init-values" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "get-version-edp" == bedp[2]["taskRef"]["name"]
    assert "update-build-number" in bedp[3]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[3]["taskRef"]["name"]
    assert "sast" in bedp[4]["name"]
    assert "test" in bedp[5]["name"]
    assert "lint" in bedp[6]["name"]
    assert buildtool == bedp[6]["taskRef"]["name"]
    assert "compile" in bedp[7]["name"]
    assert buildtool == bedp[7]["taskRef"]["name"]
    assert "sonar" in bedp[8]["name"]
    assert "sonarqube-scanner" == bedp[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bedp[9]["name"]
    assert "get-nexus-repository-url" == bedp[9]["taskRef"]["name"]
    assert "push" in bedp[10]["name"]
    assert buildtool == bedp[10]["taskRef"]["name"]
    assert "kaniko-build" in bedp[11]["name"]
    assert "git-tag" in bedp[12]["name"]
    assert "update-cbis" in bedp[13]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_python_flask_pipelines_harbor_github():
    config = """
global:
  gitProvider: github
    """

    ht = helm_template(config)

    buildtool = "python"
    framework = "flask"

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
    assert "test" in r[4]["name"]
    assert "lint" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "init-values" in bd[1]["name"]
    assert "get-version" in bd[2]["name"]
    assert f"get-version-default" == bd[2]["taskRef"]["name"]
    assert "update-build-number" in bd[3]["name"]
    assert "sast" in bd[4]["name"]
    assert "test" in bd[5]["name"]
    assert buildtool == bd[5]["taskRef"]["name"]
    assert "lint" in bd[6]["name"]
    assert buildtool == bd[6]["taskRef"]["name"]
    assert "compile" in bd[7]["name"]
    assert buildtool == bd[7]["taskRef"]["name"]
    assert "sonar" in bd[8]["name"]
    assert "sonarqube-scanner" == bd[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bd[9]["name"]
    assert "get-nexus-repository-url" == bd[9]["taskRef"]["name"]
    assert "push" in bd[10]["name"]
    assert buildtool == bd[10]["taskRef"]["name"]
    assert "kaniko-build" in bd[11]["name"]
    assert "git-tag" in bd[12]["name"]
    assert "update-cbis" in bd[13]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "init-values" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "get-version-edp" == bedp[2]["taskRef"]["name"]
    assert "update-build-number" in bedp[3]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[3]["taskRef"]["name"]
    assert "sast" in bedp[4]["name"]
    assert "test" in bedp[5]["name"]
    assert buildtool == bedp[5]["taskRef"]["name"]
    assert "lint" in bedp[6]["name"]
    assert buildtool == bedp[6]["taskRef"]["name"]
    assert "compile" in bedp[7]["name"]
    assert buildtool == bedp[7]["taskRef"]["name"]
    assert "sonar" in bedp[8]["name"]
    assert "sonarqube-scanner" == bedp[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bedp[9]["name"]
    assert "get-nexus-repository-url" == bedp[9]["taskRef"]["name"]
    assert "push" in bedp[10]["name"]
    assert buildtool == bedp[10]["taskRef"]["name"]
    assert "kaniko-build" in bedp[11]["name"]
    assert "git-tag" in bedp[12]["name"]
    assert "update-cbis" in bedp[13]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_python_fastapi_pipelines_harbor_github():
    config = """
global:
  gitProvider: github
    """

    ht = helm_template(config)

    buildtool = "python"
    framework = "fastapi"

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
    assert "test" in r[4]["name"]
    assert "lint" in r[5]["name"]
    assert "sonar" in r[6]["name"]
    assert "dockerfile-lint" in r[7]["name"]
    assert "dockerbuild-verify" in r[8]["name"]
    assert "helm-lint" in r[9]["name"]
    assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    # build with default versioning
    bd = ht["pipeline"][build_default]["spec"]["tasks"]
    assert "fetch-repository" in bd[0]["name"]
    assert "init-values" in bd[1]["name"]
    assert "get-version" in bd[2]["name"]
    assert f"get-version-default" == bd[2]["taskRef"]["name"]
    assert "update-build-number" in bd[3]["name"]
    assert "sast" in bd[4]["name"]
    assert "test" in bd[5]["name"]
    assert buildtool == bd[5]["taskRef"]["name"]
    assert "lint" in bd[6]["name"]
    assert buildtool == bd[6]["taskRef"]["name"]
    assert "compile" in bd[7]["name"]
    assert buildtool == bd[7]["taskRef"]["name"]
    assert "sonar" in bd[8]["name"]
    assert "sonarqube-scanner" == bd[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bd[9]["name"]
    assert "get-nexus-repository-url" == bd[9]["taskRef"]["name"]
    assert "push" in bd[10]["name"]
    assert buildtool == bd[10]["taskRef"]["name"]
    assert "kaniko-build" in bd[11]["name"]
    assert "git-tag" in bd[12]["name"]
    assert "update-cbis" in bd[13]["name"]
    assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

    # build with edp versioning
    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "init-values" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "get-version-edp" == bedp[2]["taskRef"]["name"]
    assert "update-build-number" in bedp[3]["taskRef"]["name"]
    assert f"update-build-number-{buildtool}" == bedp[3]["taskRef"]["name"]
    assert "sast" in bedp[4]["name"]
    assert "test" in bedp[5]["name"]
    assert "lint" in bedp[6]["name"]
    assert buildtool == bedp[6]["taskRef"]["name"]
    assert "compile" in bedp[7]["name"]
    assert buildtool == bedp[7]["taskRef"]["name"]
    assert "sonar" in bedp[8]["name"]
    assert "sonarqube-scanner" == bedp[8]["taskRef"]["name"]
    assert "get-nexus-repository-url" in bedp[9]["name"]
    assert "get-nexus-repository-url" == bedp[9]["taskRef"]["name"]
    assert "push" in bedp[10]["name"]
    assert buildtool == bedp[10]["taskRef"]["name"]
    assert "kaniko-build" in bedp[11]["name"]
    assert "git-tag" in bedp[12]["name"]
    assert "update-cbis" in bedp[13]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
