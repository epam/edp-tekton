from .helpers import helm_template


def test_java_pipelines_gerrit():
    config = """
global:
  gitProviders:
    - gerrit
pipelines:
  deployableResources:
    java:
      java8: true
    """

    r = helm_template(config)

    # ensure pipelines have proper steps
    for buildtool in ['gradle']:
        for framework in ['java11', 'java8']:
            for cbtype in ['aut']:

                assert f"gerrit-{buildtool}-{framework}-{cbtype}-review" in r["pipeline"]
                assert f"gerrit-{buildtool}-{framework}-{cbtype}-build-default" in r["pipeline"]
                assert f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp" in r["pipeline"]

                gerrit_review_pipeline = f"gerrit-{buildtool}-{framework}-{cbtype}-review"
                gerrit_build_pipeline_def = f"gerrit-{buildtool}-{framework}-{cbtype}-build-default"
                gerrit_build_pipeline_edp = f"gerrit-{buildtool}-{framework}-{cbtype}-build-edp"

                rt = r["pipeline"][gerrit_review_pipeline]["spec"]["tasks"]
                assert "fetch-repository" in rt[0]["name"]
                assert "gerrit-notify" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "sonar" in rt[3]["name"]
                assert "gerrit-vote-success" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][0]["name"]
                assert "gerrit-vote-failure" in r["pipeline"][gerrit_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][gerrit_build_pipeline_def]["spec"]["tasks"]
                assert "fetch-repository" in btd[0]["name"]
                assert "gerrit-notify" in btd[1]["name"]
                assert "init-values" in btd[2]["name"]
                assert "get-version" in btd[3]["name"]
                assert f"get-version-default" == btd[3]["taskRef"]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "sonar" in btd[5]["name"]
                assert "git-tag" in btd[6]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gerrit_build_pipeline_edp]["spec"]["tasks"]
                assert "fetch-repository" in btedp[0]["name"]
                assert "gerrit-notify" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "sonar" in btedp[4]["name"]
                assert "git-tag" in btedp[5]["name"]

def test_java_pipelines_github():
    config = """
global:
  gitProviders:
    - github
pipelines:
  deployableResources:
    java:
      java8: true
    """

    r = helm_template(config)
    vcs = "github"

    # ensure pipelines have proper steps
    for buildtool in ['gradle']:
        for framework in ['java17', 'java11', 'java8']:
            for cbtype in ['aut']:

                github_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                github_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                github_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert github_review_pipeline in r["pipeline"]
                assert github_build_pipeline_def in r["pipeline"]
                assert github_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][github_review_pipeline]["spec"]["tasks"]
                assert "github-set-pending-status" in rt[0]["name"]
                assert "fetch-repository" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "sonar" in rt[3]["name"]
                assert "github-set-success-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in r["pipeline"][github_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][github_build_pipeline_def]["spec"]["tasks"]
                assert "github-set-pending-status" in btd[0]["name"]
                assert "fetch-repository" in btd[1]["name"]
                assert "init-values" in btd[2]["name"]
                assert "get-version" in btd[3]["name"]
                assert f"get-version-default" == btd[3]["taskRef"]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "sonar" in btd[5]["name"]
                assert "git-tag" in btd[6]["name"]
                assert "github-set-success-status" in r["pipeline"][github_build_pipeline_def]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in r["pipeline"][github_build_pipeline_def]["spec"]["finally"][1]["name"]
                # build with edp versioning
                btedp = r["pipeline"][github_build_pipeline_edp]["spec"]["tasks"]
                assert "github-set-pending-status" in btedp[0]["name"]
                assert "fetch-repository" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "sonar" in btedp[4]["name"]
                assert "git-tag" in btedp[5]["name"]
                assert "github-set-success-status" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "github-set-failure-status" in r["pipeline"][github_build_pipeline_edp]["spec"]["finally"][1]["name"]

def test_java_pipelines_gitlab():
    config = """
global:
  gitProviders:
    - gitlab
pipelines:
  deployableResources:
    java:
      java8: true
    """

    r = helm_template(config)
    vcs = "gitlab"

    # ensure pipelines have proper steps
    for buildtool in ['gradle']:
        for framework in ['java17', 'java11', 'java8']:
            for cbtype in ['aut']:

                gitlab_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                gitlab_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                gitlab_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert gitlab_review_pipeline in r["pipeline"]
                assert gitlab_build_pipeline_def in r["pipeline"]
                assert gitlab_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][gitlab_review_pipeline]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in rt[0]["name"]
                assert "fetch-repository" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "sonar" in rt[3]["name"]
                assert "gitlab-set-success-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in r["pipeline"][gitlab_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][gitlab_build_pipeline_def]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in btd[0]["name"]
                assert "fetch-repository" in btd[1]["name"]
                assert "init-values" in btd[2]["name"]
                assert "get-version" in btd[3]["name"]
                assert f"get-version-default" == btd[3]["taskRef"]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "sonar" in btd[5]["name"]
                assert "git-tag" in btd[6]["name"]
                assert "gitlab-set-success-status" in r["pipeline"][gitlab_build_pipeline_def]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in r["pipeline"][gitlab_build_pipeline_def]["spec"]["finally"][1]["name"]

                # build with edp versioning
                btedp = r["pipeline"][gitlab_build_pipeline_edp]["spec"]["tasks"]
                assert "report-pipeline-start-to-gitlab" in btedp[0]["name"]
                assert "fetch-repository" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "sonar" in btedp[4]["name"]
                assert "git-tag" in btedp[5]["name"]
                assert "gitlab-set-success-status" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "gitlab-set-failure-status" in r["pipeline"][gitlab_build_pipeline_edp]["spec"]["finally"][1]["name"]

def test_java_pipelines_bitbucket():
    config = """
global:
  gitProviders:
    - bitbucket
pipelines:
  deployableResources:
    java:
      java8: true
    """

    r = helm_template(config)
    vcs = "bitbucket"

    # ensure pipelines have proper steps
    for buildtool in ['gradle']:
        for framework in ['java17', 'java11', 'java8']:
            for cbtype in ['aut']:

                bitbucket_review_pipeline = f"{vcs}-{buildtool}-{framework}-{cbtype}-review"
                bitbucket_build_pipeline_def = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-default"
                bitbucket_build_pipeline_edp = f"{vcs}-{buildtool}-{framework}-{cbtype}-build-edp"

                assert bitbucket_review_pipeline in r["pipeline"]
                assert bitbucket_build_pipeline_def in r["pipeline"]
                assert bitbucket_build_pipeline_edp in r["pipeline"]

                rt = r["pipeline"][bitbucket_review_pipeline]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in rt[0]["name"]
                assert "fetch-repository" in rt[1]["name"]
                assert "init-values" in rt[2]["name"]
                assert "sonar" in rt[3]["name"]
                assert "bitbucket-set-success-status" in r["pipeline"][bitbucket_review_pipeline]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-failure-status" in r["pipeline"][bitbucket_review_pipeline]["spec"]["finally"][1]["name"]

                # build with default versioning
                btd = r["pipeline"][bitbucket_build_pipeline_def]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in btd[0]["name"]
                assert "fetch-repository" in btd[1]["name"]
                assert "init-values" in btd[2]["name"]
                assert "get-version" in btd[3]["name"]
                assert f"get-version-default" == btd[3]["taskRef"]["name"]
                assert "update-build-number" in btd[4]["name"]
                assert "sonar" in btd[5]["name"]
                assert "git-tag" in btd[6]["name"]
                assert "bitbucket-set-success-status" in r["pipeline"][bitbucket_build_pipeline_def]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-failure-status" in r["pipeline"][bitbucket_build_pipeline_def]["spec"]["finally"][1]["name"]

                # build with edp versioning
                btedp = r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["tasks"]
                assert "bitbucket-set-pending-status" in btedp[0]["name"]
                assert "fetch-repository" in btedp[1]["name"]
                assert "init-values" in btedp[2]["name"]
                assert "get-version" in btedp[3]["name"]
                assert "get-version-edp" == btedp[3]["taskRef"]["name"]
                assert "sonar" in btedp[4]["name"]
                assert "git-tag" in btedp[5]["name"]
                assert "bitbucket-set-success-status" in r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["finally"][0]["name"]
                assert "bitbucket-set-failure-status" in r["pipeline"][bitbucket_build_pipeline_edp]["spec"]["finally"][1]["name"]
