# CLAUDE.md

## Repository Overview

**edp-tekton** integrates Tekton stack with KubeRocketCI. The repository contains two main components:

1. **KubeRocketCI Interceptor** (Go application) - A Tekton ClusterInterceptor that enriches VCS webhook payloads (GitHub, GitLab, Gerrit, BitBucket) with KubeRocketCI Codebase metadata
2. **Tekton Pipelines** (Helm charts) - Declarative CI/CD pipelines supporting 10+ languages/frameworks with VCS-specific implementations

## Development Commands

### Building & Testing

```bash
# Build the interceptor binary
make build              # Builds Go binary to dist/edpinterceptor-{arch}

# Run all tests
make test               # Runs both Go unit tests and Helm chart tests
make test-go            # Go unit tests only (with coverage)
make test-chart         # Helm chart validation with pytest

# Documentation
make helm-docs          # Generate Helm chart documentation
```

Details are in the `Makefile`.

## Architecture Overview

### Component Architecture

```text
┌────────────┐              ┌──────────────────┐       ┌─────────────┐
│            │              │   KubeRocketCI   │       │   Tekton    │
│  VCS(Git)  ├──────────────►                  ├───────►             │
│            │  Webhook     │   Interceptor    │       │  Pipelines  │
└──────┬─────┘              └────────┬─────────┘       └─────────────┘
       │                             │
┌──────┴─────┐                       │ Enriches with
│    Repo    │                       │ Codebase data
│            │      ┌────────────────▼───────────────┐
│            │      │ apiVersion: v2.edp.epam.com/v1 │
└────────────┘      │ kind: Codebase                 │
                    │ spec: {...}                    │
                    └────────────────────────────────┘
```

### Directory Structure

```text
/cmd/interceptor/          # Go binary entry point (HTTP server, port 8443)
/pkg/
  ├── interceptor/         # Main interceptor logic, TLS cert management
  └── event_processor/     # VCS-specific event processing (github/gitlab/gerrit/bitbucket)
/charts/
  ├── pipelines-library/   # Main Helm chart (pipelines, tasks, triggers)
  ├── common-library/      # Shared Helm templates for all VCS providers
  └── tekton-cache/        # Pipelines caching service
/tests/e2e/                # KUTTL-based integration tests per VCS
/hack/                     # Development utilities (Python validation, KinD configs)
```

### Interceptor Flow

The interceptor receives VCS webhooks and enriches them with platform metadata:

1. **HTTP Server** (cmd/interceptor/main.go) listens on port 8443 with TLS
2. **EDPInterceptor.Execute()** (pkg/interceptor/edp_interceptor.go) routes to VCS-specific processor
3. **Event Processor** (pkg/event_processor/{vcs}/) parses webhook payload and extracts repository name
4. **Codebase Lookup** queries Kubernetes API for matching Codebase resource
5. **InterceptorResponse** returns enriched JSON with Codebase spec data to Tekton

Timeout: 3 seconds per request

Fetch <https://tekton.dev/docs/triggers/> if more details on Tekton Triggers is required.

### Tekton Triggers Architecture

Triggers convert VCS webhook events into PipelineRun resources through a three-stage pipeline.

#### Trigger Flow

```text
Webhook Event → EventListener → Trigger (3 interceptors) → TriggerBinding → TriggerTemplate → PipelineRun
                                    ↓
                         [VCS Validation] → [CEL Filter] → [EDP Enrichment]
```

**Resource Organization:** `/charts/pipelines-library/templates/triggers/{provider}/`

**Naming Convention:**

- Triggers: `{provider}-{type}` (e.g., `github-build`)
- TriggerBindings: `{provider}-binding-{type}`
- TriggerTemplates: `{provider}-{type}-template`

#### Parameter Flow

**EDP Interceptor Enrichment** (pkg/interceptor/edp_interceptor.go):

- Matches webhook repository → `Codebase` resource (by `GitUrlPath`)
- Matches git branch → `CodebaseBranch` resource (by `BranchName`)
- Returns `extensions` with:
  - `codebase`, `codebasebranch` (resource names)
  - `pipelines.build`, `pipelines.review` (from `CodebaseBranch.Spec`)
  - `pullRequest.*` (normalized PR metadata)

**TriggerBinding** extracts parameters from:

- `body.*` - VCS-specific webhook payload (varies per provider)
- `extensions.*` - EDP interceptor output (provider-agnostic)

**TriggerTemplate** scaffolds PipelineRun with:

- Dynamic pipeline name from `extensions.pipelines.{type}`
- Labels for UI filtering (`codebase`, `pipelinetype`, `codebasebranch`)
- Ephemeral workspace PVC per run
- VCS credentials from `ci-{provider}` secret

#### VCS Provider Differences

| Aspect | GitHub | GitLab | Gerrit | BitBucket |
|--------|--------|--------|--------|-----------|
| Build Event Filter | `merged == true` | `action: merge` | `status: NEW` | `pullrequest:fulfilled` |
| Interceptor | ClusterInterceptor | ClusterInterceptor | CEL only | Custom ClusterInterceptor |
| Secret | `ci-github` | `ci-gitlab` | `ci-gerrit` | `ci-bitbucket` |

#### Critical Facts

1. **Pipeline Selection is Dynamic** - Name from `CodebaseBranch.Spec.Pipelines.{type}`, NOT hardcoded
2. **Repository Mapping** - Webhook repo must match `Codebase.Spec.GitUrlPath` (normalized lowercase)
3. **Branch Required** - Git branch must have corresponding `CodebaseBranch` resource
4. **Ephemeral Workspaces** - Each PipelineRun gets its own PVC (`.Values.tekton.workspaceSize`)
5. **Comment Retriggering** - `/recheck` and `/ok-to-test` comments re-trigger review pipelines

### Pipeline Organization

Pipelines are declarative compositions of reusable Tekton Tasks, organized by VCS provider, language, and type.

**Location:** `/charts/pipelines-library/templates/pipelines/{language}/{provider}-{type}-{version}.yaml`

**Naming Pattern:** `{vcs}-{language}-{app-type}-{pipeline-type}-{version}.yaml`

- Example: `github-maven-java17-app-build-default.yaml`

#### Task Composition Pattern

All pipelines follow a common execution flow:

```text
Build Pipelines:
  Init (set status pending) → get-version → get-cache
    → [Language Tasks: compile/test/sonar/build] → push-artifact
    → container-build (kaniko) → save-cache
    → git-tag → update-codebasebranch
    → finally: report-status (JIRA, VCS)

Review Pipelines:
  Init (fetch PR) → get-cache
    → [Language Tasks: compile/test/sonar] → docker-lint → helm-lint
    → save-cache
    → finally: set-review-status (success/failure)
```

**Key Differences:**

| Aspect | Build Pipeline | Review Pipeline |
|--------|---------------|-----------------|
| Trigger | Merge to branch | PR/MR creation or update |
| Versioning | `get-version` task sets release version | No versioning |
| Artifact Push | Pushes to registry (Maven, npm, PyPI) | No push (validation only) |
| Container Build | Builds and pushes container image | Skipped |
| Git Operations | Creates VCS tag, updates CodebaseBranch | No git modifications |
| Status Reporting | JIRA ticket update | VCS status update (GitHub status, GitLab MR comment) |

#### Pipeline Structure

Pipelines define:

```yaml
spec:
  workspaces:
    - name: shared-workspace    # Shared across all tasks (source, cache subdirs)
    - name: ssh-creds          # Git credentials
  params:
    - name: git-source-url     # From TriggerTemplate
    - name: CODEBASE_NAME      # From Codebase resource
    - name: image              # Language runtime image (e.g., maven:3.9-jdk-17)
  tasks:
    - name: init
      taskRef: {provider}-init
    - name: compile
      taskRef: maven
      runAfter: [init]         # Task ordering
      params:
        - name: GOALS
          value: [compile, test, package]
  finally:                     # Runs regardless of task success/failure
    - name: report
      taskRef: push-to-jira
```

**Reusable Components:**

- **Common Task Includes** (from `charts/common-library/`):
  - `github-build-start`, `gitlab-review-start` - VCS-specific initialization
  - `get-cache`, `save-cache` - Artifact caching
  - `build-pipeline-end` - Git tagging and CodebaseBranch updates
  - `finally-block-default`, `finally-block-semver` - Status reporting
- **Language-Specific Includes** (from `charts/pipelines-library/templates/pipelines/`):
  - `_common_java.yaml` - Maven/Gradle task sequences
  - `_common_javascript.yaml` - npm/pnpm task sequences
  - Tasks parameterized via `values.yaml` ConfigMaps

#### Extending Pipelines: Adding a New Language

1. **Pipeline YAMLs** — add `{provider}-{language}-app-build-default.yaml` and `{provider}-{language}-app-review.yaml` under `charts/pipelines-library/templates/pipelines/{language}/`. Guard with `{{ if has "{provider}" .Values.global.gitProviders }}` and `range` over the framework→image map.
2. **Task sequence** — define reusable steps in `_common_{language}.yaml` as `{{- define "edp-tekton.{language}-build-common" -}}`; pass language commands via `$(params.ci-{language})` (sourced from a `values.yaml` ConfigMap).
3. **Feature flags** — register frameworks under `deployableResources.{language}` in `values.yaml`.
4. **Image mapping** — add `edp-tekton.resourceMapping.{language}` in `_helpers.tpl` to map enabled frameworks → runtime images.
5. **Tasks** — reuse generic tasks in `templates/tasks/` (`maven`, `gradle`, `npm`, `python`, `golang`) or add `{language}.yaml`.

**Principles:** DRY via Helm includes · language commands as `values.yaml` ConfigMaps · one template covers all VCS via conditionals · isolate dirs with workspace `subPath` (source, cache). Mirror an existing language dir (e.g. `java/`) as the working reference.

### Helm Chart Configuration

Key `values.yaml` sections in `charts/pipelines-library/`:

- `deployableResources` - Toggle which pipelines/tasks to install
- `global.gitProviders` - Select VCS systems (array: bitbucket, gerrit, github, gitlab)
- `tekton-cache.enabled` - Enable artifact caching
- `kaniko.*` - Container image build configuration
- `tekton.configs.*` - Maven/Gradle/npm/Python settings (ConfigMaps)

## Testing Strategy

1. **Unit Tests** (Go test files in /pkg): Test interceptor logic and event processors
2. **Chart Tests** (pytest): Validate Helm template rendering and pipeline definitions
3. **E2E Tests** (KUTTL): Full integration tests with Kind cluster, per VCS provider

## Repository Standards

Key standards summary:

- Pipeline/task filenames match `metadata.name` fields
- Two versioning strategies: `default` (current) and `semver` (extended patterns)
- Feature flags control pipeline/task enablement via values.yaml
- Stable Tekton API versions (v1, not v1beta1)
- Consistent use of workspaces and `runAfter` for task dependencies

## Key Dependencies

**Go Libraries:**

- `github.com/tektoncd/triggers` - Tekton interceptor framework
- `github.com/epam/edp-codebase-operator` - Provides Codebase CRD
- `sigs.k8s.io/controller-runtime` - Kubernetes client

**External Tools:**

- Helm 3
- kubectl-kuttl (E2E testing)
- Python 3.11+ (chart tests)
