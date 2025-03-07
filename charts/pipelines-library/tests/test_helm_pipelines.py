from .helpers import helm_template

def test_helm_pipelines_gerrit():
    config = """
global:
  gitProviders:
    - gerrit
    """

    ht = helm_template(config)
    vcs = "gerrit"


    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['helm']:
            for cbtype in ['app']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "fetch-repository" in r[0]["name"]
                assert "check-chart-name" in r[1]["name"]
                assert "gerrit-notify" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "helm-dependency-update" in r[4]["name"]
                assert "helm-lint" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "fetch-repository" in bdef[0]["name"]
                assert "gerrit-notify" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "check-chart-name" in bdef[3]["name"]
                assert "get-version" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-lint" in bdef[6]["name"]
                assert "helm-template" in bdef[7]["name"]
                assert "helm-push" in bdef[8]["name"]
                assert "git-tag" in bdef[9]["name"]
                assert "update-cbis" in bdef[10]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "fetch-repository" in bedp[0]["name"]
                assert "gerrit-notify" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "check-chart-name" in bedp[3]["name"]
                assert "get-version" in bedp[4]["name"]
                assert "update-build-number" in bedp[5]["name"]
                assert "helm-dependency-update" in bedp[6]["name"]
                assert "helm-lint" in bedp[7]["name"]
                assert "helm-template" in bedp[8]["name"]
                assert "helm-push" in bedp[9]["name"]
                assert "git-tag" in bedp[10]["name"]
                assert "update-cbis" in bedp[11]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]

    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['charts']:
            for cbtype in ['lib']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "fetch-repository" in r[0]["name"]
                assert "gerrit-notify" in r[1]["name"]
                assert "helm-docs" in r[2]["name"]
                assert "fetch-target-branch" in r[3]["name"]
                assert "helm-lint" in r[4]["name"]
                assert "helm-dependency-update" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "fetch-repository" in bdef[0]["name"]
                assert "gerrit-notify" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "get-version" in bdef[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "fetch-repository" in bedp[0]["name"]
                assert "gerrit-notify" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "get-version" in bedp[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]

def test_helm_pipelines_gitlab():
    config = """
global:
  gitProviders:
    - gitlab
    """

    ht = helm_template(config)
    vcs = "gitlab"

    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['helm']:
            for cbtype in ['app']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "check-chart-name" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "helm-dependency-update" in r[4]["name"]
                assert "helm-lint" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in bdef[0]["name"]
                assert "fetch-repository" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "check-chart-name" in bdef[3]["name"]
                assert "get-version" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-lint" in bdef[6]["name"]
                assert "helm-template" in bdef[7]["name"]
                assert "helm-push" in bdef[8]["name"]
                assert "git-tag" in bdef[9]["name"]
                assert "update-cbis" in bdef[10]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]
                assert "gitlab-set-success-status" in ht["pipeline"][build_def]["spec"]["finally"][1]["name"]
                assert "gitlab-set-failure-status" in ht["pipeline"][build_def]["spec"]["finally"][2]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in bedp[0]["name"]
                assert "fetch-repository" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "check-chart-name" in bedp[3]["name"]
                assert "get-version" in bedp[4]["name"]
                assert "update-build-number" in bedp[5]["name"]
                assert "helm-dependency-update" in bedp[6]["name"]
                assert "helm-lint" in bedp[7]["name"]
                assert "helm-template" in bedp[8]["name"]
                assert "helm-push" in bedp[9]["name"]
                assert "git-tag" in bedp[10]["name"]
                assert "update-cbis" in bedp[11]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
                assert "gitlab-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
                assert "gitlab-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]

    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['charts']:
            for cbtype in ['lib']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "helm-docs" in r[2]["name"]
                assert "fetch-target-branch" in r[3]["name"]
                assert "helm-lint" in r[4]["name"]
                assert "helm-dependency-update" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in bdef[0]["name"]
                assert "fetch-repository" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "get-version" in bdef[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]
                assert "gitlab-set-success-status" in ht["pipeline"][build_def]["spec"]["finally"][1]["name"]
                assert "gitlab-set-failure-status" in ht["pipeline"][build_def]["spec"]["finally"][2]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in bedp[0]["name"]
                assert "fetch-repository" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "get-version" in bedp[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
                assert "gitlab-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
                assert "gitlab-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]

def test_helm_pipelines_github():
    config = """
global:
  gitProviders:
    - github
    """

    ht = helm_template(config)
    vcs = "github"

    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['helm']:
            for cbtype in ['app']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "check-chart-name" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "helm-dependency-update" in r[4]["name"]
                assert "helm-lint" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "github-set-pending-status" in bdef[0]["name"]
                assert "fetch-repository" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "check-chart-name" in bdef[3]["name"]
                assert "get-version" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-lint" in bdef[6]["name"]
                assert "helm-template" in bdef[7]["name"]
                assert "helm-push" in bdef[8]["name"]
                assert "git-tag" in bdef[9]["name"]
                assert "update-cbis" in bdef[10]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]
                assert "github-set-success-status" in ht["pipeline"][build_def]["spec"]["finally"][1]["name"]
                assert "github-set-failure-status" in ht["pipeline"][build_def]["spec"]["finally"][2]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "github-set-pending-status" in bedp[0]["name"]
                assert "fetch-repository" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "check-chart-name" in bedp[3]["name"]
                assert "get-version" in bedp[4]["name"]
                assert "update-build-number" in bedp[5]["name"]
                assert "helm-dependency-update" in bedp[6]["name"]
                assert "helm-lint" in bedp[7]["name"]
                assert "helm-template" in bedp[8]["name"]
                assert "helm-push" in bedp[9]["name"]
                assert "git-tag" in bedp[10]["name"]
                assert "update-cbis" in bedp[11]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
                assert "github-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
                assert "github-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]

    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['charts']:
            for cbtype in ['lib']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "helm-docs" in r[2]["name"]
                assert "fetch-target-branch" in r[3]["name"]
                assert "helm-lint" in r[4]["name"]
                assert "helm-dependency-update" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "github-set-pending-status" in bdef[0]["name"]
                assert "fetch-repository" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "get-version" in bdef[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]
                assert "github-set-success-status" in ht["pipeline"][build_def]["spec"]["finally"][1]["name"]
                assert "github-set-failure-status" in ht["pipeline"][build_def]["spec"]["finally"][2]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "github-set-pending-status" in bedp[0]["name"]
                assert "fetch-repository" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "get-version" in bedp[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
                assert "github-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
                assert "github-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]

def test_helm_pipelines_bitbucket():
    config = """
global:
  gitProviders:
    - bitbucket
    """

    ht = helm_template(config)
    vcs = "bitbucket"

    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['helm']:
            for cbtype in ['app']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "check-chart-name" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "helm-dependency-update" in r[4]["name"]
                assert "helm-lint" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "bitbucket-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in bdef[0]["name"]
                assert "fetch-repository" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "check-chart-name" in bdef[3]["name"]
                assert "get-version" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-lint" in bdef[6]["name"]
                assert "helm-template" in bdef[7]["name"]
                assert "helm-push" in bdef[8]["name"]
                assert "git-tag" in bdef[9]["name"]
                assert "update-cbis" in bdef[10]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-success-status" in ht["pipeline"][build_def]["spec"]["finally"][1]["name"]
                assert "bitbucket-set-failure-status" in ht["pipeline"][build_def]["spec"]["finally"][2]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in bedp[0]["name"]
                assert "fetch-repository" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "check-chart-name" in bedp[3]["name"]
                assert "get-version" in bedp[4]["name"]
                assert "update-build-number" in bedp[5]["name"]
                assert "helm-dependency-update" in bedp[6]["name"]
                assert "helm-lint" in bedp[7]["name"]
                assert "helm-template" in bedp[8]["name"]
                assert "helm-push" in bedp[9]["name"]
                assert "git-tag" in bedp[10]["name"]
                assert "update-cbis" in bedp[11]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
                assert "bitbucket-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
                assert "bitbucket-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]

    # ensure pipelines have proper steps
    for buildtool in ['helm']:
        for framework in ['charts']:
            for cbtype in ['lib']:

                review = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                build_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                build_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-semver"

                assert review in ht["pipeline"]
                assert build_def in ht["pipeline"]
                assert build_edp in ht["pipeline"]

                r = ht["pipeline"][review]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "helm-docs" in r[2]["name"]
                assert "fetch-target-branch" in r[3]["name"]
                assert "helm-lint" in r[4]["name"]
                assert "helm-dependency-update" in r[5]["name"]
                assert "helm-template" in r[6]["name"]
                assert "bitbucket-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

                bdef = ht["pipeline"][build_def]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in bdef[0]["name"]
                assert "fetch-repository" in bdef[1]["name"]
                assert "init-values" in bdef[2]["name"]
                assert "get-version" in bdef[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-success-status" in ht["pipeline"][build_def]["spec"]["finally"][1]["name"]
                assert "bitbucket-set-failure-status" in ht["pipeline"][build_def]["spec"]["finally"][2]["name"]

                bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in bedp[0]["name"]
                assert "fetch-repository" in bedp[1]["name"]
                assert "init-values" in bedp[2]["name"]
                assert "get-version" in bedp[3]["name"]
                assert "helm-lint" in bdef[4]["name"]
                assert "helm-dependency-update" in bdef[5]["name"]
                assert "helm-template" in bdef[6]["name"]
                assert "helm-push" in bdef[7]["name"]
                assert "git-tag" in bdef[8]["name"]
                assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
                assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
                assert "bitbucket-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
                assert "bitbucket-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]
