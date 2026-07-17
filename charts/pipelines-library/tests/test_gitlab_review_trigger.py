from .helpers import helm_template


def test_gitlab_review_trigger_retriggers_only_on_code_changes_and_comments():
    config = """
global:
  dnsWildCard: "example.com"
  gitProviders:
    - gitlab
    """

    r = helm_template(config)

    trigger = r["trigger"]["gitlab-review"]
    cel = next(i for i in trigger["spec"]["interceptors"] if i["ref"]["name"] == "cel")
    cel_filter = next(p["value"] for p in cel["params"] if p["name"] == "filter")

    # MR events must be limited to open/reopen and code pushes: an 'update'
    # without oldrev is a metadata-only edit (thread resolution, labels, title)
    # and re-running review on those loops with bots like SonarQube PR
    # decoration (EPMDEDP-17223).
    assert "body.object_attributes.action in ['open', 'reopen']" in cel_filter
    assert (
        "body.object_attributes.action == 'update' && has(body.object_attributes.oldrev)"
        in cel_filter
    )
    assert "'update']" not in cel_filter, "plain 'update' action must not pass the filter"

    # The old assignees/reviewers exclusion was a partial fix for the same
    # problem; the oldrev guard subsumes it (those are oldrev-less updates).
    assert "body.changes" not in cel_filter

    # Comment retriggering (/recheck, /ok-to-test) flows through the note
    # branch and is narrowed by the EDP interceptor; it must stay intact.
    assert "body.object_kind == 'note' && has(body.merge_request)" in cel_filter
