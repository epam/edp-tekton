from .helpers import helm_template


def test_dotnet_pipelines_gerrit():
    config = """
global:
  gitProvider: gerrit
  dockerRegistry:
    type: "ecr"
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"gerrit-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"gerrit-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "app":
                assert "fetch-repository" in r[0]["name"]
                assert "gerrit-notify" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "dotnet-build" in r[4]["name"]
                assert "test" in r[5]["name"]
                assert "fetch-target-branch" in r[6]["name"]
                assert "sonar-prepare-files" in r[7]["name"]
                assert "sonar-prepare-files-dotnet" == r[7]["taskRef"]["name"]
                assert "sonar" in r[8]["name"]
                assert "dotnet-publish" in r[9]["name"]
                assert "dockerfile-lint" in r[10]["name"]
                assert "dockerbuild-verify" in r[11]["name"]
                assert "helm-lint" in r[12]["name"]
            if cbtype == "lib":
                assert "fetch-repository" in r[0]["name"]
                assert "gerrit-notify" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "dotnet-build" in r[3]["name"]
                assert "test" in r[4]["name"]
                assert "fetch-target-branch" in r[5]["name"]
                assert "sonar-prepare-files" in r[6]["name"]
                assert "sonar-prepare-files-dotnet" == r[6]["taskRef"]["name"]
                assert "sonar" in r[7]["name"]

            assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "fetch-repository" in bd[0]["name"]
            assert "gerrit-notify" in bd[1]["name"]
            assert "init-values" in bd[2]["name"]
            assert "get-version" in bd[3]["name"]
            assert f"get-version-csharp-default" == bd[3]["taskRef"]["name"]
            assert "sonar-cleanup" in bd[4]["name"]
            assert "sast" in bd[5]["name"]
            assert "dotnet-build" in bd[6]["name"]
            assert "test" in bd[7]["name"]
            assert buildtool == bd[7]["taskRef"]["name"]
            assert "sonar" in bd[8]["name"]
            assert buildtool == bd[8]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bd[9]["name"]
            assert "get-nexus-repository-url" == bd[9]["taskRef"]["name"]
            assert "get-nuget-token" in bd[10]["name"]
            assert "push" in bd[11]["name"]
            assert buildtool == bd[11]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[12]["name"]
                assert "create-ecr-repository" in bd[13]["name"]
                assert "kaniko-build" in bd[14]["name"]
                assert "git-tag" in bd[15]["name"]
                assert "update-cbis" in bd[16]["name"]
            if cbtype == "lib":
                assert "git-tag" in bd[12]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "fetch-repository" in bedp[0]["name"]
            assert "gerrit-notify" in bedp[1]["name"]
            assert "init-values" in bedp[2]["name"]
            assert "get-version" in bedp[3]["name"]
            assert "get-version-edp" == bedp[3]["taskRef"]["name"]
            assert "update-build-number" in bedp[4]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[4]["taskRef"]["name"]
            assert "sonar-cleanup" in bedp[5]["name"]
            assert "sast" in bedp[6]["name"]
            assert "dotnet-build" in bedp[7]["name"]
            assert buildtool == bedp[7]["taskRef"]["name"]
            assert "test" in bedp[8]["name"]
            assert buildtool == bedp[8]["taskRef"]["name"]
            assert "sonar" in bedp[9]["name"]
            assert buildtool == bedp[9]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bedp[10]["name"]
            assert "get-nexus-repository-url" == bedp[10]["taskRef"]["name"]
            assert "get-nuget-token" in bedp[11]["name"]
            assert "push" in bedp[12]["name"]
            assert buildtool == bedp[12]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[13]["name"]
                assert "create-ecr-repository" in bedp[14]["name"]
                assert "kaniko-build" in bedp[15]["name"]
                assert "git-tag" in bedp[16]["name"]
                assert "update-cbis" in bedp[17]["name"]
            if cbtype == "lib":
                assert "git-tag" in bedp[13]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_dotnet_pipelines_gitlab():
    config = """
global:
  gitProvider: gitlab
  dockerRegistry:
    type: "ecr"
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"gitlab-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"gitlab-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"gitlab-{buildtool}-{framework}-{cbtype}-build-edp"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "dotnet-build" in r[3]["name"]
                assert "test" in r[4]["name"]
                assert "sonar" in r[5]["name"]
            if cbtype == "app":
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "dotnet-build" in r[4]["name"]
                assert "test" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]

            assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "fetch-repository" in bd[0]["name"]
            assert "init-values" in bd[1]["name"]
            assert "get-version" in bd[2]["name"]
            assert f"get-version-csharp-default" == bd[2]["taskRef"]["name"]
            assert "sast" in bd[3]["name"]
            assert "dotnet-build" in bd[4]["name"]
            assert "test" in bd[5]["name"]
            assert buildtool == bd[5]["taskRef"]["name"]
            assert "sonar" in bd[6]["name"]
            assert buildtool == bd[6]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bd[7]["name"]
            assert "get-nexus-repository-url" == bd[7]["taskRef"]["name"]
            assert "get-nuget-token" in bd[8]["name"]
            assert "push" in bd[9]["name"]
            assert buildtool == bd[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[10]["name"]
                assert "create-ecr-repository" in bd[11]["name"]
                assert "kaniko-build" in bd[12]["name"]
                assert "git-tag" in bd[13]["name"]
                assert "update-cbis" in bd[14]["name"]
            if cbtype == "lib":
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "fetch-repository" in bedp[0]["name"]
            assert "init-values" in bedp[1]["name"]
            assert "get-version" in bedp[2]["name"]
            assert "get-version-edp" == bedp[2]["taskRef"]["name"]
            assert "update-build-number" in bedp[3]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[3]["taskRef"]["name"]
            assert "sast" in bedp[4]["name"]
            assert "dotnet-build" in bedp[5]["name"]
            assert buildtool == bedp[5]["taskRef"]["name"]
            assert "test" in bedp[6]["name"]
            assert buildtool == bedp[6]["taskRef"]["name"]
            assert "sonar" in bedp[7]["name"]
            assert buildtool == bedp[7]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bedp[8]["name"]
            assert "get-nexus-repository-url" == bedp[8]["taskRef"]["name"]
            assert "get-nuget-token" in bedp[9]["name"]
            assert "push" in bedp[10]["name"]
            assert buildtool == bedp[10]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[11]["name"]
                assert "create-ecr-repository" in bedp[12]["name"]
                assert "kaniko-build" in bedp[13]["name"]
                assert "git-tag" in bedp[14]["name"]
                assert "update-cbis" in bedp[15]["name"]
            if cbtype == "lib":
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_dotnet_pipelines_github():
    config = """
global:
  gitProvider: github
  dockerRegistry:
    type: "ecr"
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"github-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"github-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"github-{buildtool}-{framework}-{cbtype}-build-edp"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "dotnet-build" in r[3]["name"]
                assert "test" in r[4]["name"]
                assert "sonar" in r[5]["name"]
            if cbtype == "app":
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "dotnet-build" in r[4]["name"]
                assert "test" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]

            assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "fetch-repository" in bd[0]["name"]
            assert "init-values" in bd[1]["name"]
            assert "get-version" in bd[2]["name"]
            assert f"get-version-csharp-default" == bd[2]["taskRef"]["name"]
            assert "sast" in bd[3]["name"]
            assert "dotnet-build" in bd[4]["name"]
            assert "test" in bd[5]["name"]
            assert buildtool == bd[5]["taskRef"]["name"]
            assert "sonar" in bd[6]["name"]
            assert buildtool == bd[6]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bd[7]["name"]
            assert "get-nexus-repository-url" == bd[7]["taskRef"]["name"]
            assert "get-nuget-token" in bd[8]["name"]
            assert "push" in bd[9]["name"]
            assert buildtool == bd[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[10]["name"]
                assert "create-ecr-repository" in bd[11]["name"]
                assert "kaniko-build" in bd[12]["name"]
                assert "git-tag" in bd[13]["name"]
                assert "update-cbis" in bd[14]["name"]
            if cbtype == "lib":
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "fetch-repository" in bedp[0]["name"]
            assert "init-values" in bedp[1]["name"]
            assert "get-version" in bedp[2]["name"]
            assert "get-version-edp" == bedp[2]["taskRef"]["name"]
            assert "update-build-number" in bedp[3]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[3]["taskRef"]["name"]
            assert "sast" in bedp[4]["name"]
            assert "dotnet-build" in bedp[5]["name"]
            assert buildtool == bedp[5]["taskRef"]["name"]
            assert "test" in bedp[6]["name"]
            assert buildtool == bedp[6]["taskRef"]["name"]
            assert "sonar" in bedp[7]["name"]
            assert buildtool == bedp[7]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bedp[8]["name"]
            assert "get-nexus-repository-url" == bedp[8]["taskRef"]["name"]
            assert "get-nuget-token" in bedp[9]["name"]
            assert "push" in bedp[10]["name"]
            assert buildtool == bedp[10]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[11]["name"]
                assert "create-ecr-repository" in bedp[12]["name"]
                assert "kaniko-build" in bedp[13]["name"]
                assert "git-tag" in bedp[14]["name"]
                assert "update-cbis" in bedp[15]["name"]
            if cbtype == "lib":
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_dotnet_pipelines_harbor_gerrit():
    config = """
global:
  gitProvider: gerrit
  dockerRegistry:
    type: "harbor"
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"gerrit-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"gerrit-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "fetch-repository" in r[0]["name"]
                assert "gerrit-notify" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "dotnet-build" in r[3]["name"]
                assert "test" in r[4]["name"]
                assert "fetch-target-branch" in r[5]["name"]
                assert "sonar-prepare-files" in r[6]["name"]
                assert "sonar-prepare-files-dotnet" == r[6]["taskRef"]["name"]
                assert "sonar" in r[7]["name"]
            if cbtype == "app":
                assert "fetch-repository" in r[0]["name"]
                assert "gerrit-notify" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "dotnet-build" in r[4]["name"]
                assert "test" in r[5]["name"]
                assert "fetch-target-branch" in r[6]["name"]
                assert "sonar-prepare-files" in r[7]["name"]
                assert "sonar-prepare-files-dotnet" == r[7]["taskRef"]["name"]
                assert "sonar" in r[8]["name"]
                assert "dotnet-publish" in r[9]["name"]
                assert "dockerfile-lint" in r[10]["name"]
                assert "dockerbuild-verify" in r[11]["name"]
                assert "helm-lint" in r[12]["name"]

            assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "fetch-repository" in bd[0]["name"]
            assert "gerrit-notify" in bd[1]["name"]
            assert "init-values" in bd[2]["name"]
            assert "get-version" in bd[3]["name"]
            assert f"get-version-csharp-default" == bd[3]["taskRef"]["name"]
            assert "sonar-cleanup" in bd[4]["name"]
            assert "sast" in bd[5]["name"]
            assert "dotnet-build" in bd[6]["name"]
            assert "test" in bd[7]["name"]
            assert buildtool == bd[7]["taskRef"]["name"]
            assert "sonar" in bd[8]["name"]
            assert buildtool == bd[8]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bd[9]["name"]
            assert "get-nexus-repository-url" == bd[9]["taskRef"]["name"]
            assert "get-nuget-token" in bd[10]["name"]
            assert "push" in bd[11]["name"]
            assert buildtool == bd[11]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[12]["name"]
                assert "kaniko-build" in bd[13]["name"]
                assert "git-tag" in bd[14]["name"]
                assert "update-cbis" in bd[15]["name"]
            if cbtype == "lib":
                assert "git-tag" in bd[12]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "fetch-repository" in bedp[0]["name"]
            assert "gerrit-notify" in bedp[1]["name"]
            assert "init-values" in bedp[2]["name"]
            assert "get-version" in bedp[3]["name"]
            assert "get-version-edp" == bedp[3]["taskRef"]["name"]
            assert "update-build-number" in bedp[4]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[4]["taskRef"]["name"]
            assert "sonar-cleanup" in bedp[5]["name"]
            assert "sast" in bedp[6]["name"]
            assert "dotnet-build" in bedp[7]["name"]
            assert buildtool == bedp[7]["taskRef"]["name"]
            assert "test" in bedp[8]["name"]
            assert buildtool == bedp[8]["taskRef"]["name"]
            assert "sonar" in bedp[9]["name"]
            assert buildtool == bedp[9]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bedp[10]["name"]
            assert "get-nexus-repository-url" == bedp[10]["taskRef"]["name"]
            assert "get-nuget-token" in bedp[11]["name"]
            assert "push" in bedp[12]["name"]
            assert buildtool == bedp[12]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[13]["name"]
                assert "kaniko-build" in bedp[14]["name"]
                assert "git-tag" in bedp[15]["name"]
                assert "update-cbis" in bedp[16]["name"]
            if cbtype == "lib":
                assert "git-tag" in bedp[13]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_dotnet_pipelines_harbor_gitlab():
    config = """
global:
  gitProvider: gitlab
  dockerRegistry:
    type: "harbor"
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"gitlab-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"gitlab-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"gitlab-{buildtool}-{framework}-{cbtype}-build-edp"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "dotnet-build" in r[3]["name"]
                assert "test" in r[4]["name"]
                assert "sonar" in r[5]["name"]
            if cbtype == "app":
                assert "report-pipeline-start-to-gitlab" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "dotnet-build" in r[4]["name"]
                assert "test" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]

            assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "fetch-repository" in bd[0]["name"]
            assert "init-values" in bd[1]["name"]
            assert "get-version" in bd[2]["name"]
            assert f"get-version-csharp-default" == bd[2]["taskRef"]["name"]
            assert "sast" in bd[3]["name"]
            assert "dotnet-build" in bd[4]["name"]
            assert "test" in bd[5]["name"]
            assert buildtool == bd[5]["taskRef"]["name"]
            assert "sonar" in bd[6]["name"]
            assert buildtool == bd[6]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bd[7]["name"]
            assert "get-nexus-repository-url" == bd[7]["taskRef"]["name"]
            assert "get-nuget-token" in bd[8]["name"]
            assert "push" in bd[9]["name"]
            assert buildtool == bd[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[10]["name"]
                assert "kaniko-build" in bd[11]["name"]
                assert "git-tag" in bd[12]["name"]
                assert "update-cbis" in bd[13]["name"]
            if cbtype == "lib":
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "fetch-repository" in bedp[0]["name"]
            assert "init-values" in bedp[1]["name"]
            assert "get-version" in bedp[2]["name"]
            assert "get-version-edp" == bedp[2]["taskRef"]["name"]
            assert "update-build-number" in bedp[3]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[3]["taskRef"]["name"]
            assert "sast" in bedp[4]["name"]
            assert "dotnet-build" in bedp[5]["name"]
            assert buildtool == bedp[5]["taskRef"]["name"]
            assert "test" in bedp[6]["name"]
            assert buildtool == bedp[6]["taskRef"]["name"]
            assert "sonar" in bedp[7]["name"]
            assert buildtool == bedp[7]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bedp[8]["name"]
            assert "get-nexus-repository-url" == bedp[8]["taskRef"]["name"]
            assert "get-nuget-token" in bedp[9]["name"]
            assert "push" in bedp[10]["name"]
            assert buildtool == bedp[10]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[11]["name"]
                assert "kaniko-build" in bedp[12]["name"]
                assert "git-tag" in bedp[13]["name"]
                assert "update-cbis" in bedp[14]["name"]
            if cbtype == "lib":
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]


def test_dotnet_pipelines_harbor_github():
    config = """
global:
  gitProvider: github
  dockerRegistry:
    type: "harbor"
    """

    ht = helm_template(config)

    buildtool = "dotnet"

    for framework in ['dotnet-3.1', 'dotnet-6.0']:
        for cbtype in ['app', 'lib']:

            review = f"github-{buildtool}-{framework}-{cbtype}-review"
            build_default = f"github-{buildtool}-{framework}-{cbtype}-build-default"
            build_edp = f"github-{buildtool}-{framework}-{cbtype}-build-edp"

            assert review in ht["pipeline"]
            assert build_default in ht["pipeline"]
            assert build_edp in ht["pipeline"]

            r = ht["pipeline"][review]["spec"]["tasks"]
            if cbtype == "lib":
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "dotnet-build" in r[3]["name"]
                assert "test" in r[4]["name"]
                assert "sonar" in r[5]["name"]
            if cbtype == "app":
                assert "github-set-pending-status" in r[0]["name"]
                assert "fetch-repository" in r[1]["name"]
                assert "init-values" in r[2]["name"]
                assert "helm-docs" in r[3]["name"]
                assert "dotnet-build" in r[4]["name"]
                assert "test" in r[5]["name"]
                assert "sonar" in r[6]["name"]
                assert "dotnet-publish" in r[7]["name"]
                assert "dockerfile-lint" in r[8]["name"]
                assert "dockerbuild-verify" in r[9]["name"]
                assert "helm-lint" in r[10]["name"]

            assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
            assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

            # build with default versioning
            bd = ht["pipeline"][build_default]["spec"]["tasks"]
            assert "fetch-repository" in bd[0]["name"]
            assert "init-values" in bd[1]["name"]
            assert "get-version" in bd[2]["name"]
            assert f"get-version-csharp-default" == bd[2]["taskRef"]["name"]
            assert "sast" in bd[3]["name"]
            assert "dotnet-build" in bd[4]["name"]
            assert "test" in bd[5]["name"]
            assert buildtool == bd[5]["taskRef"]["name"]
            assert "sonar" in bd[6]["name"]
            assert buildtool == bd[6]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bd[7]["name"]
            assert "get-nexus-repository-url" == bd[7]["taskRef"]["name"]
            assert "get-nuget-token" in bd[8]["name"]
            assert "push" in bd[9]["name"]
            assert buildtool == bd[9]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bd[10]["name"]
                assert "kaniko-build" in bd[11]["name"]
                assert "git-tag" in bd[12]["name"]
                assert "update-cbis" in bd[13]["name"]
            if cbtype == "lib":
                assert "git-tag" in bd[10]["name"]
            assert "push-to-jira" in ht["pipeline"][build_default]["spec"]["finally"][0]["name"]

            # build with edp versioning
            bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
            assert "fetch-repository" in bedp[0]["name"]
            assert "init-values" in bedp[1]["name"]
            assert "get-version" in bedp[2]["name"]
            assert "get-version-edp" == bedp[2]["taskRef"]["name"]
            assert "update-build-number" in bedp[3]["taskRef"]["name"]
            assert f"update-build-number-csharp" == bedp[3]["taskRef"]["name"]
            assert "sast" in bedp[4]["name"]
            assert "dotnet-build" in bedp[5]["name"]
            assert buildtool == bedp[5]["taskRef"]["name"]
            assert "test" in bedp[6]["name"]
            assert buildtool == bedp[6]["taskRef"]["name"]
            assert "sonar" in bedp[7]["name"]
            assert buildtool == bedp[7]["taskRef"]["name"]
            assert "get-nexus-repository-url" in bedp[8]["name"]
            assert "get-nexus-repository-url" == bedp[8]["taskRef"]["name"]
            assert "get-nuget-token" in bedp[9]["name"]
            assert "push" in bedp[10]["name"]
            assert buildtool == bedp[10]["taskRef"]["name"]
            if cbtype == "app":
                assert "dotnet-publish" in bedp[11]["name"]
                assert "kaniko-build" in bedp[12]["name"]
                assert "git-tag" in bedp[13]["name"]
                assert "update-cbis" in bedp[14]["name"]
            if cbtype == "lib":
                assert "git-tag" in bedp[11]["name"]
            assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
            assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
