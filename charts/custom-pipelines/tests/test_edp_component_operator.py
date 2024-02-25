from .helpers_custom import helm_template

def test_operator():
    config = """
global:
  gitProviders:
    - gerrit
    """

    ht = helm_template(config)
    vcs = "gerrit"

    review = f"{vcs}-co-go-other-app-review"
    build = f"{vcs}-co-go-other-app-build-edp"

    assert review in ht["pipeline"]
    assert build in ht["pipeline"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "fetch-repository" in r[0]["name"]
    assert "gerrit-notify" in r[1]["name"]
    assert "init-values" in r[2]["name"]
    assert "commit-validate" in r[3]["name"]
    assert "helm-docs" in r[4]["name"]
    assert "helm-lint" in r[5]["name"]
    assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]

    b = ht["pipeline"][build]["spec"]["tasks"]
    assert "fetch-repository" in b[0]["name"]
    assert "gerrit-notify" in b[1]["name"]
    assert "init-values" in b[2]["name"]
    assert "get-version" in b[3]["name"]
    assert "set-version" in b[4]["name"]
    assert "wait-for" in b[5]["name"]
    assert "helm-push-gh-pages" in b[6]["name"]
    assert "git-tag" in b[7]["name"]
    assert "update-cbis" in b[8]["name"]
    assert "update-cbb" in ht["pipeline"][build]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build]["spec"]["finally"][1]["name"]
    assert "send-to-microsoft-teams-failed" in ht["pipeline"][build]["spec"]["finally"][2]["name"]
