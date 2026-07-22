from .helpers import helm_template


def test_bitbucket_review_trigger_labels_pipelinerun_with_head_sha():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - bitbucket
    """

    r = helm_template(config)

    trigger_template = r["triggertemplate"]["bitbucket-review-template"]
    pipelinerun = trigger_template["spec"]["resourcetemplates"][0]
    labels = pipelinerun["metadata"]["labels"]

    # Bitbucket fires pullrequest:updated for both code pushes and metadata-only
    # edits with no distinguishing payload field. The krci interceptor tells them
    # apart by comparing head SHAs of review PipelineRuns already created for the
    # same change (EPMDEDP-17224), which requires the head SHA to be label-selectable.
    assert labels["app.edp.epam.com/git-commit-sha"] == "$(tt.params.gitrevision)"

    # Keep the existing change-number label pairing locked in - the guard's List
    # selector filters on both.
    assert labels["app.edp.epam.com/git-change-number"] == "$(tt.params.changeNumber)"
