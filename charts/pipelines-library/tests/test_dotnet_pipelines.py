from .helpers import helm_template


def test_dotnet_pipelines_harbor_gerrit():
    config = """
global:
  gitProviders:
    - gerrit
pipelines:
  deployableResources:
    cs:
      dotnet3.1: true
      dotnet6.0: true
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"gerrit-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"gerrit-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"gerrit-{buildtool}-{framework}-{cbtype}-build-semver"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "fetch-repository" in r[0]["name"]
                assert "gerrit-notify" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "get-cache" in r[3]["name"]
                assert "build" in r[4]["name"]
                assert "sonar" in r[5]["name"]
                assert "save-cache" in r[6]["name"]
            if cbtype == "app":
                assert "fetch-repository" in r[0]["name"]
                assert "gerrit-notify" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "get-cache" in r[4]["name"]
                assert "build" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]
                assert "save-cache" in r[11]["name"]

            assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "fetch-repository" in bd[0]["name"]
            assert "gerrit-notify" in bd[1]["name"]
            assert "init-values" in bd[2]["name"]
            assert "get-version" in bd[3]["name"]
            assert f"get-version-dotnet-default" == bd[3]["taskRef"]["name"]
            assert "get-cache" in bd[4]["name"]
            assert "security" in bd[5]["name"]
            assert "build" in bd[6]["name"]
            assert "sonar" in bd[7]["name"]
            assert "push" in bd[8]["name"]
            assert buildtool == bd[8]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[9]["name"]
                assert "kaniko-build" in bd[10]["name"]
                assert "save-cache" in bd[11]["name"]
                assert "git-tag" in bd[12]["name"]
                assert "update-cbis" in bd[13]["name"]
            if cbtype == "lib":
                assert "save-cache" in bd[9]["name"]
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "fetch-repository" in bedp[0]["name"]
            assert "gerrit-notify" in bedp[1]["name"]
            assert "init-values" in bedp[2]["name"]
            assert "get-version" in bedp[3]["name"]
            assert "get-version-edp" == bedp[3]["taskRef"]["name"]
            assert "get-cache" in bedp[4]["taskRef"]["name"]
            assert "update-build-number" in bedp[5]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[5]["taskRef"]["name"]
            assert "security" in bedp[6]["name"]
            assert "build" in bedp[7]["name"]
            assert "sonar" in bedp[8]["name"]
            assert "push" in bedp[9]["name"]
            assert buildtool == bedp[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[10]["name"]
                assert "kaniko-build" in bedp[11]["name"]
                assert "save-cache" in bedp[12]["name"]
                assert "git-tag" in bedp[13]["name"]
                assert "update-cbis" in bedp[14]["name"]
            if cbtype == "lib":
                assert "save-cache" in bedp[10]["name"]
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_dotnet_pipelines_harbor_gitlab():
    config = """
global:
  gitProviders:
    - gitlab
pipelines:
  deployableResources:
    cs:
      dotnet3.1: true
      dotnet6.0: true
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"gitlab-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"gitlab-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"gitlab-{buildtool}-{framework}-{cbtype}-build-semver"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "get-cache" in r[3]["name"]
                assert "build" in r[4]["name"]
                assert "sonar" in r[5]["name"]
                assert "save-cache" in r[6]["name"]
            if cbtype == "app":
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "get-cache" in r[4]["name"]
                assert "build" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]
                assert "save-cache" in r[11]["name"]

            assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "report-pipeline-start-to-gitlab" in bd[0]["name"]
            assert "fetch-repository" in bd[1]["name"]
            assert "init-values" in bd[2]["name"]
            assert "get-version" in bd[3]["name"]
            assert f"get-version-dotnet-default" == bd[3]["taskRef"]["name"]
            assert "get-cache" in bd[4]["name"]
            assert "security" in bd[5]["name"]
            assert "build" in bd[6]["name"]
            assert "sonar" in bd[7]["name"]
            assert "push" in bd[8]["name"]
            assert buildtool == bd[8]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[9]["name"]
                assert "kaniko-build" in bd[10]["name"]
                assert "save-cache" in bd[11]["name"]
                assert "git-tag" in bd[12]["name"]
                assert "update-cbis" in bd[13]["name"]
            if cbtype == "lib":
                assert "save-cache" in bd[9]["name"]
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]
            assert "gitlab-set-success-status" in ht["pipeline"][build_default]["spec"]["finally"][1]["name"]
            assert "gitlab-set-failure-status" in ht["pipeline"][build_default]["spec"]["finally"][2]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "report-pipeline-start-to-gitlab" in bedp[0]["name"]
            assert "fetch-repository" in bedp[1]["name"]
            assert "init-values" in bedp[2]["name"]
            assert "get-version" in bedp[3]["name"]
            assert "get-version-edp" == bedp[3]["taskRef"]["name"]
            assert "get-cache" in bedp[4]["taskRef"]["name"]
            assert "update-build-number" in bedp[5]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[5]["taskRef"]["name"]
            assert "security" in bedp[6]["name"]
            assert "build" in bedp[7]["name"]
            assert "sonar" in bedp[8]["name"]
            assert "push" in bedp[9]["name"]
            assert buildtool == bedp[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[10]["name"]
                assert "kaniko-build" in bedp[11]["name"]
                assert "save-cache" in bedp[12]["name"]
                assert "git-tag" in bedp[13]["name"]
                assert "update-cbis" in bedp[14]["name"]
            if cbtype == "lib":
                assert "save-cache" in bedp[10]["name"]
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
            assert "gitlab-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
            assert "gitlab-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]


def test_dotnet_pipelines_harbor_github():
    config = """
global:
  gitProviders:
    - github
pipelines:
  deployableResources:
    cs:
      dotnet3.1: true
      dotnet6.0: true
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"github-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"github-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"github-{buildtool}-{framework}-{cbtype}-build-semver"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "get-cache" in r[3]["name"]
                assert "build" in r[4]["name"]
                assert "sonar" in r[5]["name"]
                assert "save-cache" in r[6]["name"]
            if cbtype == "app":
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "get-cache" in r[4]["name"]
                assert "build" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]
                assert "save-cache" in r[11]["name"]

            assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "github-set-pending-status" in bd[0]["name"]
            assert "fetch-repository" in bd[1]["name"]
            assert "init-values" in bd[2]["name"]
            assert "get-version" in bd[3]["name"]
            assert f"get-version-dotnet-default" == bd[3]["taskRef"]["name"]
            assert "get-cache" in bd[4]["name"]
            assert "security" in bd[5]["name"]
            assert "build" in bd[6]["name"]
            assert "sonar" in bd[7]["name"]
            assert "push" in bd[8]["name"]
            assert buildtool == bd[8]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[9]["name"]
                assert "kaniko-build" in bd[10]["name"]
                assert "save-cache" in bd[11]["name"]
                assert "git-tag" in bd[12]["name"]
                assert "update-cbis" in bd[13]["name"]
            if cbtype == "lib":
                assert "save-cache" in bd[9]["name"]
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]
            assert "github-set-success-status" in ht["pipeline"][build_default]["spec"]["finally"][1]["name"]
            assert "github-set-failure-status" in ht["pipeline"][build_default]["spec"]["finally"][2]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "github-set-pending-status" in bedp[0]["name"]
            assert "fetch-repository" in bedp[1]["name"]
            assert "init-values" in bedp[2]["name"]
            assert "get-version" in bedp[3]["name"]
            assert "get-version-edp" == bedp[3]["taskRef"]["name"]
            assert "get-cache" in bedp[4]["taskRef"]["name"]
            assert "update-build-number" in bedp[5]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[5]["taskRef"]["name"]
            assert "security" in bedp[6]["name"]
            assert "build" in bedp[7]["name"]
            assert "sonar" in bedp[8]["name"]
            assert "push" in bedp[9]["name"]
            assert buildtool == bedp[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[10]["name"]
                assert "kaniko-build" in bedp[11]["name"]
                assert "save-cache" in bedp[12]["name"]
                assert "git-tag" in bedp[13]["name"]
                assert "update-cbis" in bedp[14]["name"]
            if cbtype == "lib":
                assert "save-cache" in bedp[10]["name"]
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
            assert "github-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
            assert "github-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]

def test_dotnet_pipelines_harbor_bitbucket():
    config = """
global:
  gitProviders:
    - bitbucket
pipelines:
  deployableResources:
    cs:
      dotnet3.1: true
      dotnet6.0: true
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"bitbucket-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"bitbucket-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"bitbucket-{buildtool}-{framework}-{cbtype}-build-semver"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "bitbucket-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "get-cache" in r[3]["name"]
                assert "build" in r[4]["name"]
                assert "sonar" in r[5]["name"]
                assert "save-cache" in r[6]["name"]
            if cbtype == "app":
                assert "bitbucket-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "get-cache" in r[4]["name"]
                assert "build" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]
                assert "save-cache" in r[11]["name"]

            assert "bitbucket-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "bitbucket-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "bitbucket-set-pending-status" in bd[0]["name"]
            assert "fetch-repository" in bd[1]["name"]
            assert "init-values" in bd[2]["name"]
            assert "get-version" in bd[3]["name"]
            assert f"get-version-dotnet-default" == bd[3]["taskRef"]["name"]
            assert "get-cache" in bd[4]["name"]
            assert "security" in bd[5]["name"]
            assert "build" in bd[6]["name"]
            assert "sonar" in bd[7]["name"]
            assert "push" in bd[8]["name"]
            assert buildtool == bd[8]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[9]["name"]
                assert "kaniko-build" in bd[10]["name"]
                assert "save-cache" in bd[11]["name"]
                assert "git-tag" in bd[12]["name"]
                assert "update-cbis" in bd[13]["name"]
            if cbtype == "lib":
                assert "save-cache" in bd[9]["name"]
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]
            assert "bitbucket-set-success-status" in ht["pipeline"][build_default]["spec"]["finally"][1]["name"]
            assert "bitbucket-set-failure-status" in ht["pipeline"][build_default]["spec"]["finally"][2]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "bitbucket-set-pending-status" in bedp[0]["name"]
            assert "fetch-repository" in bedp[1]["name"]
            assert "init-values" in bedp[2]["name"]
            assert "get-version" in bedp[3]["name"]
            assert "get-version-edp" == bedp[3]["taskRef"]["name"]
            assert "get-cache" in bedp[4]["taskRef"]["name"]
            assert "update-build-number" in bedp[5]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[5]["taskRef"]["name"]
            assert "security" in bedp[6]["name"]
            assert "build" in bedp[7]["name"]
            assert "sonar" in bedp[8]["name"]
            assert "push" in bedp[9]["name"]
            assert buildtool == bedp[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[10]["name"]
                assert "kaniko-build" in bedp[11]["name"]
                assert "save-cache" in bedp[12]["name"]
                assert "git-tag" in bedp[13]["name"]
                assert "update-cbis" in bedp[14]["name"]
            if cbtype == "lib":
                assert "save-cache" in bedp[10]["name"]
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
            assert "bitbucket-set-success-status" in ht["pipeline"][build_edp]["spec"]["finally"][2]["name"]
            assert "bitbucket-set-failure-status" in ht["pipeline"][build_edp]["spec"]["finally"][3]["name"]
