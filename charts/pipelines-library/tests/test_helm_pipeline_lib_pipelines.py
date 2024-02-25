from .helpers import helm_template

def test_helm_pipelines_gerrit():
    config = """
global:
  gitProviders:
    - gerrit
    """

    ht = helm_template(config)
    vcs = "gerrit"

    review = f"{vcs}-helm-pipeline-lib-review"
    build_def = f"{vcs}-helm-pipeline-lib-build-default"
    build_edp = f"{vcs}-helm-pipeline-lib-build-edp"

    assert review in ht["pipeline"]
    assert build_def in ht["pipeline"]
    assert build_edp in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "fetch-repository" in r[0]["name"]
    assert "gerrit-notify" in r[1]["name"]
    assert "helm-docs" in r[2]["name"]
    assert "helm-dependency-update" in r[3]["name"]
    assert "helm-lint" in r[4]["name"]
    assert "helm-template" in r[5]["name"]
    assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    bdef = ht["pipeline"][build_def]["spec"]["tasks"]
    assert "fetch-repository" in bdef[0]["name"]
    assert "gerrit-notify" in bdef[1]["name"]
    assert "get-version" in bdef[2]["name"]
    assert "helm-dependency-update" in bdef[3]["name"]
    assert "helm-lint" in bdef[4]["name"]
    assert "helm-template" in bdef[5]["name"]
    assert "git-tag" in bdef[6]["name"]
    assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]

    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "gerrit-notify" in bedp[1]["name"]
    assert "get-version" in bedp[2]["name"]
    assert "update-build-number" in bedp[3]["name"]
    assert "helm-dependency-update" in bedp[4]["name"]
    assert "helm-lint" in bedp[5]["name"]
    assert "helm-template" in bedp[6]["name"]
    assert "git-tag" in bedp[7]["name"]
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

    review = f"{vcs}-helm-pipeline-lib-review"
    build_def = f"{vcs}-helm-pipeline-lib-build-default"
    build_edp = f"{vcs}-helm-pipeline-lib-build-edp"

    assert review in ht["pipeline"]
    assert build_def in ht["pipeline"]
    assert build_edp in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "report-pipeline-start-to-gitlab" in r[0]["name"]
    assert "fetch-repository" in r[1]["name"]
    assert "helm-docs" in r[2]["name"]
    assert "helm-dependency-update" in r[3]["name"]
    assert "helm-lint" in r[4]["name"]
    assert "helm-template" in r[5]["name"]
    assert "gitlab-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gitlab-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    bdef = ht["pipeline"][build_def]["spec"]["tasks"]
    assert "fetch-repository" in bdef[0]["name"]
    assert "get-version" in bdef[1]["name"]
    assert "helm-dependency-update" in bdef[2]["name"]
    assert "helm-lint" in bdef[3]["name"]
    assert "helm-template" in bdef[4]["name"]
    assert "git-tag" in bdef[5]["name"]
    assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]

    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "get-version" in bedp[1]["name"]
    assert "update-build-number" in bedp[2]["name"]
    assert "helm-dependency-update" in bedp[3]["name"]
    assert "helm-lint" in bedp[4]["name"]
    assert "helm-template" in bedp[5]["name"]
    assert "git-tag" in bedp[6]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]

def test_helm_pipelines_github():
    config = """
global:
  gitProviders:
    - github
    """

    ht = helm_template(config)
    vcs = "github"

    review = f"{vcs}-helm-pipeline-lib-review"
    build_def = f"{vcs}-helm-pipeline-lib-build-default"
    build_edp = f"{vcs}-helm-pipeline-lib-build-edp"

    assert review in ht["pipeline"]
    assert build_def in ht["pipeline"]
    assert build_edp in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "github-set-pending-status" in r[0]["name"]
    assert "fetch-repository" in r[1]["name"]
    assert "helm-docs" in r[2]["name"]
    assert "helm-dependency-update" in r[3]["name"]
    assert "helm-lint" in r[4]["name"]
    assert "helm-template" in r[5]["name"]
    assert "github-set-success-status" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "github-set-failure-status" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    bdef = ht["pipeline"][build_def]["spec"]["tasks"]
    assert "fetch-repository" in bdef[0]["name"]
    assert "get-version" in bdef[1]["name"]
    assert "helm-dependency-update" in bdef[2]["name"]
    assert "helm-lint" in bdef[3]["name"]
    assert "helm-template" in bdef[4]["name"]
    assert "git-tag" in bdef[5]["name"]
    assert "push-to-jira" in ht["pipeline"][build_def]["spec"]["finally"][0]["name"]

    bedp = ht["pipeline"][build_edp]["spec"]["tasks"]
    assert "fetch-repository" in bedp[0]["name"]
    assert "get-version" in bedp[1]["name"]
    assert "update-build-number" in bedp[2]["name"]
    assert "helm-dependency-update" in bedp[3]["name"]
    assert "helm-lint" in bedp[4]["name"]
    assert "helm-template" in bedp[5]["name"]
    assert "git-tag" in bedp[6]["name"]
    assert "update-cbb" in ht["pipeline"][build_edp]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build_edp]["spec"]["finally"][1]["name"]
