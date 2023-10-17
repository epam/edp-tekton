from .helpers_custom import helm_template

def test_operator():
    config = """
global:
  gitProvider: gerrit
    """

    ht = helm_template(config)
    vcs = "gerrit"

    review = f"{vcs}-kaniko-other-app-review"
    build = f"{vcs}-kaniko-other-app-build-edp"

    assert review in ht["pipeline"]
    assert build in ht["pipeline"]

    b = ht["pipeline"][build]["spec"]["tasks"]
    assert "fetch-repository" in b[0]["name"]
    assert "gerrit-notify" in b[1]["name"]
    assert "init-values" in b[2]["name"]
    assert "get-version" in b[3]["name"]
    assert "dockerfile-lint" in b[4]["name"]
    assert "kaniko-build" in b[5]["name"]
    assert "update-cbb" in ht["pipeline"][build]["spec"]["finally"][0]["name"]
    assert "push-to-jira" in ht["pipeline"][build]["spec"]["finally"][1]["name"]
    assert "send-to-microsoft-teams-failed" in ht["pipeline"][build]["spec"]["finally"][2]["name"]

    r = ht["pipeline"][review]["spec"]["tasks"]
    assert "fetch-repository" in r[0]["name"]
    assert "gerrit-notify" in r[1]["name"]
    assert "init-values" in r[2]["name"]
    assert "commit-validate" in r[3]["name"]
    assert "dockerfile-lint" in r[4]["name"]
    assert "dockerbuild-verify" in r[5]["name"]
    assert "dependency-check" in r[6]["name"]
    assert "gerrit-vote-success" in ht["pipeline"][review]["spec"]["finally"][0]["name"]
    assert "gerrit-vote-failure" in ht["pipeline"][review]["spec"]["finally"][1]["name"]
