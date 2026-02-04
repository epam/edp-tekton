#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# onboarding-component.sh
# ------------------------------------------------------------------------------
# CLI:
#   --type | -t   : build | review | task
#   --vcs  | -v   : gerrit | gitlab | github | bitbucket  (mandatory for pipelines, optional for tasks)
#   --name | -n   : resource name (string)
#   --help | -h   : print help and exit
#
# Flow:
#   main -> check_input_type -> (add_pipeline | add_task)
# ==============================================================================

SCRIPT_NAME="$(basename "$0")"
TYPE=""      # build | review | task
VCS=""       # gerrit | gitlab | github | bitbucket (required for pipelines)
NAME=""      # resource name

# ----------------------------- helpers ----------------------------------------

print_help() {
  cat <<USAGE
${SCRIPT_NAME} - onboarding helper for Tekton resources

Usage:
  ${SCRIPT_NAME} --type <build|review|task> [--vcs <gerrit|gitlab|github|bitbucket>] --name <name>

Options:
  -t, --type    Type of resource to generate:
                - build  : generate Build Pipeline
                - review : generate Review Pipeline
                - task   : generate Task
  -v, --vcs     VCS provider (pipelines: required; tasks: optional):
                gerrit | gitlab | github | bitbucket
  -n, --name    Name for the resource (string). The exact meaning may differ
                between types; later steps may enforce naming patterns.
  -h, --help    Show this help and exit.

Examples:
  ${SCRIPT_NAME} -t build  -v github -n github-python-fastapi-app-build-default
  ${SCRIPT_NAME} -t review -v gitlab -n gitlab-python-fastapi-app-review
  ${SCRIPT_NAME} -t task              -n ansible-run
USAGE
}

# macOS-safe lowercase
tolower() { printf '%s' "$1" | tr '[:upper:]' '[:lower:]'; }

# Decide which action to perform based on TYPE; returns one of: build|review|task
check_input_type() {
  local t="${1:-}"
  if [[ -z "$t" ]]; then
    echo "Error: --type is required." >&2
    return 2
  fi
  t="$(tolower "$t")"
  case "$t" in
    build|review|task)
      printf '%s' "$t"
      ;;
    *)
      echo "Error: unsupported --type '$t'. Use build|review|task." >&2
      return 2
      ;;
  esac
}

# cross-platform sed -i (GNU/BSD)
sed_inplace() {
  local pattern="$1" file="$2"
  if sed --version >/dev/null 2>&1; then
    sed -i "$pattern" "$file"
  else
    sed -i '' "$pattern" "$file"
  fi
}

# -------------------------------- pipeline constructor helpers -----------------

PARAM_BLOCKS=()
TASK_BLOCKS=()

add_param_block() { PARAM_BLOCKS+=("$1"); }
add_task_block()  { TASK_BLOCKS+=("$1"); }

join_params() { printf "%s\n" "${PARAM_BLOCKS[@]}"; }
join_tasks()  { printf "%s\n" "${TASK_BLOCKS[@]}"; }


param_pipeline_url() {
  add_param_block "$(cat <<'BLOCK'
    - name: pipelineUrl
      default: https://portal-{{ .Release.Namespace }}.{{ .Values.global.dnsWildCard }}/c/main/pipelines/pipelineruns/$(context.pipelineRun.namespace)/$(context.pipelineRun.name)
      type: string
BLOCK
)"
}

param_git_source_url() {
  add_param_block "$(cat <<'BLOCK'
    - name: git-source-url
      default: "https://github.com/epmd-edp/python-python-python-3.13"
      description: git url to clone
      type: string
BLOCK
)"
}

param_git_source_revision() {
  add_param_block "$(cat <<'BLOCK'
    - name: git-source-revision
      description: 'git revision to checkout (branch, tag, sha, ref…)'
      default: "edp"
      type: string
BLOCK
)"
}

param_git_refspec() {
  add_param_block "$(cat <<'BLOCK'
    - name: git-refspec
      description: Refspec to fetch before checking out revision.
      default: ""
      type: string
BLOCK
)"
}

# --- build/github & reuse ---
param_cb_name() {
  add_param_block "$(cat <<'BLOCK'
    - name: CODEBASE_NAME
      description: "Project name"
      type: string
BLOCK
)"
}

param_cbbranch_name() {
  add_param_block "$(cat <<'BLOCK'
    - name: CODEBASEBRANCH_NAME
      description: "Codebasebranch name"
      type: string
BLOCK
)"
}

param_change_number_def() {
  add_param_block "$(cat <<'BLOCK'
    - name: changeNumber
      description: Change number from Merge Request
      default: ""
      type: string
BLOCK
)"
}

param_gitsha() {
  add_param_block "$(cat <<'BLOCK'
    - name: gitsha
      description: "commit sha"
      type: string
BLOCK
)"
}

param_gitfullrepo() {
  add_param_block "$(cat <<'BLOCK'
    - name: gitfullrepositoryname
      description: "repository full name"
      type: string
BLOCK
)"
}

# --- build/gerrit ---
param_change_number() {
  add_param_block "$(cat <<'BLOCK'
    - name: changeNumber
      description: Change number from Merge Request
BLOCK
)"
}

param_patchset_number() {
  add_param_block "$(cat <<'BLOCK'
    - name: patchsetNumber
      description: Patchset number from Merge Request
BLOCK
)"
}

# --- review/bitbucket & gitlab shared ---
param_chart_dir_deploy() {
  add_param_block "$(cat <<'BLOCK'
    - name: CHART_DIR
      description: "The directory in source that contains the helm chart"
      default: "deploy-templates"
BLOCK
)"
}

param_ct_configs_dir() {
  add_param_block "$(cat <<'BLOCK'
    - name: CT_CONFIGS_DIR
      description: "ct-configs directory for helm-lint"
      default: "."
BLOCK
)"
}

# --- review/gerrit specific ---
param_target_branch() {
  add_param_block "$(cat <<'BLOCK'
    - name: targetBranch
      description: Target branch of Merge Request
BLOCK
)"
}

# --- review/github specific ---
param_gitfullrepo_repo() {
  add_param_block "$(cat <<'BLOCK'
    - name: gitfullrepositoryname
      description: "Repository full name"
      type: string
BLOCK
)"
}

param_gitsha_review() {
  add_param_block "$(cat <<'BLOCK'
    - name: gitsha
      description: "Commit sha"
      type: string
BLOCK
)"
}

param_chart_dir_charts() {
  add_param_block "$(cat <<'BLOCK'
    - name: CHART_DIR
      description: "The directory in source that contains the helm chart"
      default: "charts"
      type: string
BLOCK
)"
}

param_target_branch_typed() {
  add_param_block "$(cat <<'BLOCK'
    - name: targetBranch
      description: "Target branch of Merge Request"
      type: string
BLOCK
)"
}

param_chart_version_increment() {
  add_param_block "$(cat <<'BLOCK'
    - name: CHART_VERSION_INCREMENT
      description: "Check Chart version increment"
      default: 'true'
      type: string
BLOCK
)"
}

# --- build: init tasks ---
task_build_init_github() {
  add_task_block "$(cat <<'BLOCK'
    - name: init-values
      taskRef:
        kind: Task
        name: init-values
      params:
        - name: CODEBASE_NAME
          value: $(params.CODEBASE_NAME)
        - name: BRANCH_NAME
          value: $(params.git-source-revision)

BLOCK
)"
}

task_build_init_gitlab() {
  add_task_block "$(cat <<'BLOCK'
    - name: init-values
      taskRef:
        kind: Task
        name: init-values
      params:
        - name: CODEBASE_NAME
          value: $(params.CODEBASE_NAME)
        - name: BRANCH_NAME
          value: $(params.git-source-revision)
BLOCK
)"
}

task_build_init_bitbucket() {
  add_task_block "$(cat <<'BLOCK'
    - name: init-values
      taskRef:
        kind: Task
        name: init-values
      params:
        - name: CODEBASE_NAME
          value: $(params.CODEBASE_NAME)
        - name: BRANCH_NAME
          value: $(params.git-source-revision)
BLOCK
)"
}

task_build_init_gerrit() {
  add_task_block "$(cat <<'BLOCK'
    - name: init-values
      taskRef:
        kind: Task
        name: init-values
      params:
        - name: CODEBASE_NAME
          value: $(params.CODEBASE_NAME)
        - name: BRANCH_NAME
          value: $(params.git-source-revision)
BLOCK
)"
}

task_get_version() {
  add_task_block "$(cat <<'BLOCK'
    - name: get-version
      taskRef:
        kind: Task
        name: get-version-default
      runAfter: [init-values]
      params:
        - name: BRANCH_NAME
          value: $(params.git-source-revision)
        - name: BASE_IMAGE
          value: $(params.image)
BLOCK
)"
}

task_git_tag() {
  add_task_block "$(cat <<'BLOCK'
    - name: git-tag
      taskRef:
        kind: Task
        name: git-cli
      runAfter: [helm-push]
      params:
        - name: GIT_USER_EMAIL
          value: edp-ci@edp.ci-user
        - name: GIT_USER_NAME
          value: edp-ci
        - name: GIT_SCRIPT
          value: |
            git tag -a "$(tasks.get-version.results.VCS_TAG)" -m "Tag is added automatically by CI user"
            git push --tags
      workspaces:
        - name: source
          workspace: shared-workspace
          subPath: source
        - name: ssh-directory
          workspace: ssh-creds
BLOCK
)"
}

task_update_cbis() {
  add_task_block "$(cat <<'BLOCK'
    - name: update-cbis
      taskRef:
        kind: Task
        name: update-cbis
      runAfter: [git-tag]
      params:
        - name: CODEBASEBRANCH_NAME
          value: $(params.CODEBASEBRANCH_NAME)
        - name: IMAGE_TAG
          value: $(tasks.get-version.results.VCS_TAG)
BLOCK
)"
}

# --- review/bitbucket ---
task_bb_set_pending() {
  add_task_block "$(cat <<'BLOCK'
    - name: bitbucket-set-pending-status
      taskRef:
        kind: Task
        name: bitbucket-set-status
      params:
        - name: REPO_FULL_NAME
          value: $(params.gitfullrepositoryname)
        - name: SHA
          value: "$(params.git-source-revision)"
        - name: TARGET_URL
          value: $(params.pipelineUrl)
        - name: DESCRIPTION
          value: "Review Pipeline"
        - name: STATE
          value: "INPROGRESS"
        - name: BITBUCKET_TOKEN_SECRET_NAME
          value: ci-bitbucket
        - name: BITBUCKET_TOKEN_SECRET_KEY
          value: token
        - name: KEY
          value: "review"
        - name: NAME
          value: "Pipeline"
BLOCK
)"
}

task_fetch_repo() {
  # $1 = runAfter task name (optional)
  local after="$1"
  local run_after=""
  if [[ -n "$after" ]]; then run_after="      runAfter: [${after}]"; fi
  add_task_block "$(cat <<BLOCK
    - name: fetch-repository
      taskRef:
        kind: Task
        name: git-clone
${run_after}
      params:
        - name: url
          value: \$(params.git-source-url)
        - name: revision
          value: \$(params.git-source-revision)
        - name: refspec
          value: \$(params.git-refspec)
        - name: subdirectory
          value: source
      workspaces:
        - name: output
          workspace: shared-workspace
        - name: ssh-directory
          workspace: ssh-creds
BLOCK
)"
}

task_helm_docs_default() {
  add_task_block "$(cat <<'BLOCK'
    - name: helm-docs
      taskRef:
        kind: Task
        name: helm-docs
      runAfter: [fetch-repository]
      params:
        - name: CHART_DIR
          value: $(params.CHART_DIR)
      workspaces:
        - name: source
          workspace: shared-workspace
          subPath: source
BLOCK
)"
}

# --- review/gerrit ---
task_gerrit_notify() {
  add_task_block "$(cat <<'BLOCK'
    - name: gerrit-notify
      taskRef:
        kind: Task
        name: gerrit-ssh-cmd
      params:
        - name: GERRIT_PORT
          value: '{{ .Values.global.gerritSSHPort }}'
        - name: SSH_GERRIT_COMMAND
          value: review --verified 0 --message 'Build Started $(params.pipelineUrl)' $(params.changeNumber),$(params.patchsetNumber)
      workspaces:
        - name: ssh-directory
          workspace: ssh-creds
BLOCK
)"
}

# --- review/gitlab ---
task_gitlab_set_status() {
  add_task_block "$(cat <<'BLOCK'
    - name: report-pipeline-start-to-gitlab
      taskRef:
        kind: Task
        name: gitlab-set-status
      params:
        - name: "STATE"
          value: "pending"
        - name: "GITLAB_HOST_URL"
          value: "$(params.git-source-url)"
        - name: "REPO_FULL_NAME"
          value: "$(params.gitfullrepositoryname)"
        - name: "GITLAB_TOKEN_SECRET_NAME"
          value: ci-gitlab
        - name: "GITLAB_TOKEN_SECRET_KEY"
          value: token
        - name: "SHA"
          value: "$(params.git-source-revision)"
        - name: "TARGET_URL"
          value: $(params.pipelineUrl)
        - name: "CONTEXT"
          value: "Review Pipeline"
        - name: "DESCRIPTION"
          value: "Managed by KubeRocketCI. Run with Tekton"
BLOCK
)"
}

task_check_chart_name() {
  add_task_block "$(cat <<'BLOCK'
    - name: check-chart-name
      taskRef:
        name: check-helm-chart-name
      runAfter: [fetch-repository]
      params:
        - name: codebase_name
          value: $(params.CODEBASE_NAME)
        - name: chart_dir
          value: $(params.CHART_DIR)
      workspaces:
        - name: source
          subPath: source
          workspace: shared-workspace
BLOCK
)"
}

# --- review/github ---
task_github_set_pending() {
  add_task_block "$(cat <<'BLOCK'
    - name: github-set-pending-status
      taskRef:
        kind: Task
        name: github-set-status
      params:
        - name: REPO_FULL_NAME
          value: $(params.gitfullrepositoryname)
        - name: DESCRIPTION
          value: "Pipeline (IN PROGRESS)"
        - name: STATE
          value: "pending"
        - name: CONTEXT
          value: "Build Pipeline"
        - name: AUTH_TYPE
          value: Token
        - name: GITHUB_TOKEN_SECRET_NAME
          value: ci-github
        - name: GITHUB_TOKEN_SECRET_KEY
          value: token
        - name: SHA
          value: $(params.gitsha)
        - name: TARGET_URL
          value: $(params.pipelineUrl)
BLOCK
)"
}

add_pipeline() {
  # --- checks ---
  if [[ -z "${NAME}" ]]; then echo "Error: --name is required for pipelines." >&2; exit 1; fi
  if [[ -z "${VCS}" ]];  then echo "Error: --vcs is required for pipelines."  >&2; exit 1; fi
  case "$TYPE" in build|review) ;; *) echo "Error: --type must be build or review for pipelines." >&2; exit 1;; esac
  case "$VCS" in gerrit|gitlab|github|bitbucket) ;; *) echo "Error: --vcs must be one of: gerrit|gitlab|github|bitbucket." >&2; exit 1;; esac

  local outdir="./charts/pipelines-library/templates/pipelines"
  local outfile="${outdir}/${NAME}.yaml"
  mkdir -p "${outdir}"

  PARAM_BLOCKS=(); TASK_BLOCKS=()

  param_pipeline_url
  param_git_source_url
  param_git_source_revision
  param_git_refspec

  local TEMPLATE=""
  local PIPELINETYPE_LABEL=""
  local TYPE_HUMAN=""
  local VOTE_INCLUDE=""

  if [[ "$TYPE" == "build" ]]; then
    PIPELINETYPE_LABEL="build"; TYPE_HUMAN="Build"
    case "$VCS" in
      github)
        TEMPLATE="github-build-template"
        param_cb_name
        param_cbbranch_name
        param_change_number_def
        param_gitsha
        param_gitfullrepo
        task_build_init_github
        task_get_version
        task_git_tag
        task_update_cbis
        VOTE_INCLUDE='{{ include "github-build-vote" . | nindent 2 }}'
        ;;
      gitlab)
        TEMPLATE="gitlab-build-template"
        param_gitfullrepo
        param_cb_name
        param_cbbranch_name
        param_change_number_def
        task_build_init_gitlab
        task_get_version
        task_git_tag
        task_update_cbis
        VOTE_INCLUDE='{{ include "gitlab-build-vote" . | nindent 2 }}'
        ;;
      bitbucket)
        TEMPLATE="bitbucket-build-template"
        param_cb_name
        param_cbbranch_name
        param_change_number_def
        param_gitfullrepo
        task_build_init_bitbucket
        task_get_version
        task_git_tag
        task_update_cbis
        VOTE_INCLUDE='{{ include "bitbucket-build-vote" . | nindent 2 }}'
        ;;
      gerrit)
        TEMPLATE="gerrit-build-template"
        param_change_number
        param_patchset_number
        param_cb_name
        param_cbbranch_name
        task_build_init_gerrit
        task_get_version
        task_git_tag
        task_update_cbis
        VOTE_INCLUDE=''
        ;;
    esac
  else
    PIPELINETYPE_LABEL="review"; TYPE_HUMAN="Review"
    case "$VCS" in
      bitbucket)
        TEMPLATE="bitbucket-review-template"
        param_cb_name
        param_change_number_def
        param_gitfullrepo
        param_chart_dir_deploy
        param_ct_configs_dir
        task_bb_set_pending
        task_fetch_repo "bitbucket-set-pending-status"
        task_helm_docs_default
        VOTE_INCLUDE='{{ include "bitbucket-review-vote" . | nindent 2 }}'
        ;;
      gerrit)
        TEMPLATE="gerrit-review-template"
        param_change_number
        param_patchset_number
        param_cb_name
        param_target_branch
        param_chart_dir_deploy
        param_ct_configs_dir
        task_fetch_repo ""
        task_gerrit_notify
        task_helm_docs_default
        VOTE_INCLUDE='{{ include "gerrit-review-vote" . | nindent 2 }}'
        ;;
      github)
        TEMPLATE="github-review-template"
        param_cb_name
        param_change_number_def
        param_gitfullrepo_repo
        param_gitsha_review
        param_chart_dir_charts
        param_ct_configs_dir
        param_target_branch_typed
        param_chart_version_increment
        task_github_set_pending
        task_fetch_repo "github-set-pending-status"
        task_helm_docs_default
        VOTE_INCLUDE='{{ include "github-review-vote" . | nindent 2 }}'
        ;;
      gitlab)
        TEMPLATE="gitlab-review-template"
        param_cb_name
        param_change_number_def
        param_gitfullrepo
        param_chart_dir_deploy
        param_ct_configs_dir
        task_gitlab_set_status
        task_fetch_repo "report-pipeline-start-to-gitlab"
        task_check_chart_name
        task_helm_docs_default
        VOTE_INCLUDE='{{ include "gitlab-review-vote" . | nindent 2 }}'
        ;;
    esac
  fi

  local PARAMS_JOINED; PARAMS_JOINED="$(join_params)"
  local TASKS_JOINED; TASKS_JOINED="$(join_tasks)"

  cat > "${outfile}" <<EOF
{{ if .Values.pipelines.deployableResources.${NAME} }}
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: ${NAME}
  labels:
    app.edp.epam.com/pipelinetype: ${PIPELINETYPE_LABEL}
    app.edp.epam.com/triggertemplate: ${TEMPLATE}
    {{- include "edp-tekton.labels" . | nindent 4 }}
spec:
  description: "The ${TYPE_HUMAN} pipeline for building Helm"
  workspaces:
    - name: shared-workspace
    - name: ssh-creds
  params:
${PARAMS_JOINED}
  tasks:
${TASKS_JOINED}

{{ include "finally-block-default" . | nindent 2 }}

${VOTE_INCLUDE}

{{ end }}
EOF

  echo "✓ Pipeline created: ${outfile}"
  echo "  metadata.name: ${NAME}"
  echo "  provider: ${VCS}, type: ${TYPE}"
}

# -------------------------------- TASK GENERATION ------------------------------

add_task() {
  if [[ -z "${NAME}" ]]; then
    echo "Error: --name is required for task." >&2
    exit 1
  fi

  local outdir="./charts/pipelines-library/templates/tasks"
  local outfile="${outdir}/${NAME}.yaml"

  mkdir -p "${outdir}"

  cat <<'EOF' > "${outfile}"
{{ if .Values.pipelines.deployableResources.tasks }}
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: <TASK_NAME>
  labels:<LABELS_BLOCK>
spec:
  description: >-
    <TASK_DESCRIPTION>
  workspaces:
    - name: source
  params:
    - name: PROJECT_DIR
      description: The directory containing ansible files
      type: string
      default: "<PROJECT_DIR_DEFAULT>"
    - name: EXTRA_COMMANDS
      type: string
    - name: BASE_IMAGE
      type: string
      default: {{ include "edp-tekton.registry" . }}/<BASE_IMAGE_SUFFIX>
      description: The ansible image.
  steps:
    - name: <STEP_NAME>
      image: $(params.BASE_IMAGE)
      workingDir: $(workspaces.source.path)/$(params.PROJECT_DIR)
      script: |
        set -ex
        $(params.EXTRA_COMMANDS)
{{- include "resources" . | nindent 6 }}
{{ end }}
EOF

  sed_inplace "s|<TASK_NAME>|${NAME}|g" "${outfile}"
  sed_inplace "s|<STEP_NAME>|${NAME}|g" "${outfile}"

  echo "✓ Task created: ${outfile}"
  echo "  metadata.name: ${NAME}"
  echo "  TODO: replace <LABELS_BLOCK>, <TASK_DESCRIPTION>, <PROJECT_DIR_DEFAULT>, <BASE_IMAGE_SUFFIX>"
}

# --------------------------------- main ---------------------------------------

main() {
  # Parse CLI args
  if [[ $# -eq 0 ]]; then
    print_help; exit 1
  fi

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --type|-t)
        TYPE="${2:-}"; shift 2;;
      --vcs|-v|-vcs)
        VCS="${2:-}"; shift 2;;
      --name|-n|-name)
        NAME="${2:-}"; shift 2;;
      --help|-h)
        print_help; exit 0;;
      *)
        echo "Unknown argument: $1" >&2
        print_help; exit 1;;
    esac
  done

  # Basic required flags
  if [[ -z "${TYPE}" ]]; then
    echo "Error: --type is required." >&2; print_help; exit 1
  fi
  if [[ -z "${NAME}" ]]; then
    echo "Error: --name is required." >&2; print_help; exit 1
  fi

  # Normalize
  TYPE="$(tolower "$TYPE")"
  if [[ -n "${VCS}" ]]; then VCS="$(tolower "$VCS")"; fi

  # Decide and dispatch
  local kind
  kind="$(check_input_type "${TYPE}")" || { print_help; exit 1; }

  case "$kind" in
    build|review)
      add_pipeline
      ;;
    task)
      add_task
      ;;
    *)
      echo "Internal error: unknown kind '$kind'." >&2
      exit 1
      ;;
  esac
}

# Entrypoint
main "$@"