import os
import sys

from .helpers_custom import helm_template

print(helm_template)

def test_operator():
    config = """
global:
  gitProvider: gerrit
    """

    ht = helm_template(config)
    vcs = "gerrit"

    review = f"{vcs}-go-other-app-review"
    build = f"{vcs}-go-other-app-build-edp"

    assert review in ht["pipeline"]
    assert build in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "fetch-repository" in r[0]["name"]
    assert "gerrit-notify" in r[1]["name"]
    assert "init-values" in r[2]["name"]
    assert "commit-validate" in r[3]["name"]
    assert "helm-docs" in r[4]["name"]
    assert "helm-lint" in r[5]["name"]
    assert "build" in r[6]["name"]
    assert "test" in r[7]["name"]
    assert "lint" in r[8]["name"]
    assert "fetch-target-branch" in r[9]["name"]
    assert "sonar-prepare-files" in r[10]["name"]
    assert "sonar" in r[11]["name"]
    assert "dockerfile-lint" in r[12]["name"]
    assert "dockerbuild-verify" in r[13]["name"]
    assert "image-scan" in r[14]["name"]
    assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    b = ht["pipeline"][build]["spec"]["tasks"]
    assert "fetch-repository" in b[0]["name"]
    assert "gerrit-notify" in b[1]["name"]
    assert "init-values" in b[2]["name"]
    assert "get-version" in b[3]["name"]
    assert "sonar-cleanup" in b[4]["name"]
    assert "sast" in b[5]["name"]
    assert "test" in b[6]["name"]
    assert "sonar" in b[7]["name"]
    assert "build" in b[8]["name"]
    assert "create-ecr-repository" in b[9]["name"]
    assert "kaniko-build" in b[10]["name"]
    assert "ecr-to-docker" in b[11]["name"]
    assert "set-version" in b[12]["name"]
    assert "wait-for" in b[13]["name"]
    assert "helm-push-gh-pages" in b[14]["name"]
    assert "git-tag" in b[15]["name"]
    assert "update-cbis" in b[16]["name"]
    assert "update-cbb" in ht["pipeline"][build]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build]["spec"]["finally"][1]["name"]
    assert "send-to-microsoft-teams-failed" in ht["pipeline"][build]["spec"]["finally"][2]["name"]
