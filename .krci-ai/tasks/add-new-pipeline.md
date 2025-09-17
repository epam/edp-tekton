---
dependencies:
  data:
    - devops-edp-tekton-standarts.md
    - krci-ai/core-framework-standards.yaml
---

# Task: Add New Tekton Pipelines (Build & Review)

<task_header>
<description>Automate the process of adding new Tekton **Build** and **Review** Pipelines to the *edp-tekton* repository.
The agent must use the onboarding script to generate **two pipelines** (build and review) under
'./charts/pipelines-library/templates/pipelines', follow repository conventions, and apply the enforced naming pattern.</description>
</task_header>

---

### Reference Assets

<prerequisites>
Dependencies:
- Tekton overview: https://docs.kuberocketci.io/docs/operator-guide/ci/tekton-overview
- User guide: https://docs.kuberocketci.io/docs/user-guide/tekton-pipelines
- Custom pipelines flow: https://docs.kuberocketci.io/docs/use-cases/custom-pipelines-flow

Validation: Verify the dependency exists at the specified path before proceeding. HALT if it is missing.

**Repository structure & constants:**

- Directory for new pipelines: './charts/pipelines-library/templates/pipelines'
- Onboarding script: './charts/pipelines-library/scripts/onboarding-component.sh'

**Validation (mandatory before starting):**

1. Ensure the directory './charts/pipelines-library/templates/pipelines' exists. If not — **HALT**.
2. Ensure the onboarding script exists at './charts/pipelines-library/scripts/onboarding-component.sh'. If missing — **HALT**.
3. Verify online documentation URLs are reachable. If not, continue with a warning.
</prerequisites>

## Overview

<task_overview>
The user provides:

- *vcs* — VCS type (e.g., github, gitlab, bitbucket)
- *language* — programming language (e.g., python, java, javascript)
- *framework* — framework/tool (e.g., fastapi, springboot, npm)

The agent then generates **two pipelines** using the onboarding script:

- one with type `build-pipeline`
- one with type `review-pipeline`

There are two types of versioning suffixes used in the naming convention: `edp` and `default`. This suffix is placed at the end of the build pipeline name. Currently, only the `default` suffix is actively used, but both exist as part of the standard naming scheme. The agent must not confuse these suffixes and must strictly follow the naming rule as specified.

Both must follow the enforced naming patterns:

- Build:

  ```
  <vcs>-<language>-<framework>-app-build-default
  ```

- Review:

  ```
  <vcs>-<language>-<framework>-app-review
  ```

After generation, the agent validates file existence and reports results.
The **file name** and the **metadata.name** inside the pipeline YAML must match the same pattern.
</task_overview>

## Instructions

<instructions>
1. Review the reference documentation (links above).
2. Collect input values from the user (see *Required Inputs*).
3. Derive final pipeline names based on the patterns above.
   - Example: `gitlab-python-fastapi-app-build-default`, `gitlab-python-fastapi-app-review`.
4. Run the onboarding script twice:

```sh
./charts/pipelines-library/scripts/onboarding-component.sh -type build-pipeline -n <vcs>-<language>-<framework>-app-build-default --vcs <vcs>
./charts/pipelines-library/scripts/onboarding-component.sh -type review-pipeline -n <vcs>-<language>-<framework>-app-review --vcs <vcs>
```

5. Verify that both files are created under './charts/pipelines-library/templates/pipelines/'.
6. Verify that the `metadata.name` inside each pipeline matches the file name.
7. **Important rule**:
   - Do **not** remove configurations created by the onboarding script.
   - Allowed changes:
     - Update parameter descriptions
     - Update default values
     - Update metadata (labels, annotations, descriptions)
     - Add new parameters, steps, or sections if required
   - Forbidden changes:
     - Removing parameters
     - Removing steps
     - Removing or restructuring core sections of the pipelines
   - Make changes incrementally, based on existing pipelines.
8. Report back the file paths and confirm creation/update.
</instructions>

## Output Format

<output_format>

- created_files:
  - "./charts/pipelines-library/templates/pipelines/<vcs>-<language>-<framework>-app-build-default.yaml"
  - "./charts/pipelines-library/templates/pipelines/<vcs>-<language>-<framework>-app-review.yaml"
- executed_commands:
  - "./charts/pipelines-library/scripts/onboarding-component.sh -type build-pipeline -n <name> --vcs <vcs>"
  - "./charts/pipelines-library/scripts/onboarding-component.sh -type review-pipeline -n <name> --vcs <vcs>"
- validation:
  - metadata_name_matches_filename: true
  - contains_apiVersion_and_kind: ["apiVersion: tekton.dev/v1", "kind: Pipeline"]
</output_format>

## Execution Checklist

<execution_checklist>

1. Collect inputs: `vcs`, `language`, `framework`.
2. Derive pipeline names following patterns.
3. Run onboarding script for build pipeline (see `executed_commands`).
4. Run onboarding script for review pipeline (see `executed_commands`).
5. Verify both files exist under `./charts/pipelines-library/templates/pipelines/`.
6. Open both files and confirm `metadata.name` equals filename (without .yaml).
7. Confirm both manifests include `apiVersion: tekton.dev/v1` and `kind: Pipeline`.
8. Report executed commands and created file paths.
</execution_checklist>

## Required Inputs

<user_inputs>
**Mandatory:**

- *vcs* — VCS provider, e.g., github, gitlab, bitbucket
- *language* — programming language, e.g., python, java, javascript
- *framework* — framework/tool, e.g., fastapi, springboot, npm

**Example questions to the user:**

- *Which VCS provider are you targeting (github, gitlab, bitbucket)?*
- *What is the programming language (python, java, javascript)?*
- *What framework or build tool is used (fastapi, springboot, npm)?*
</user_inputs>

## Usage Examples

<usage_examples>

### Example 1: GitLab Python FastAPI

```sh
./charts/pipelines-library/scripts/onboarding-component.sh -type build-pipeline -n gitlab-python-fastapi-app-build-default --vcs gitlab
./charts/pipelines-library/scripts/onboarding-component.sh -type review-pipeline -n gitlab-python-fastapi-app-review --vcs gitlab
```

### Example 2: GitHub Java SpringBoot

```sh
./charts/pipelines-library/scripts/onboarding-component.sh -type build-pipeline -n github-java-springboot-app-build-default --vcs github
./charts/pipelines-library/scripts/onboarding-component.sh -type review-pipeline -n github-java-springboot-app-review --vcs github
```

</usage_examples>

## Acceptance Criteria

<success_criteria>

- [ ] The onboarding script is executed twice: once with `-type build-pipeline`, once with `-type review-pipeline`.
- [ ] Pipeline names strictly follow the patterns:
  - `<vcs>-<language>-<framework>-app-build-default`
  - `<vcs>-<language>-<framework>-app-review`
- [ ] File names match the patterns exactly.
- [ ] `metadata.name` inside each pipeline YAML matches its file name.
- [ ] Two new Pipeline files exist under './charts/pipelines-library/templates/pipelines/'.
- [ ] Both files contain `apiVersion: tekton.dev/v1`, `kind: Pipeline`.
- [ ] Existing structures generated by the script remain intact.
- [ ] Only non-destructive updates are applied (metadata, descriptions, default values).
- [ ] The agent reports back executed commands and created/updated file path(s).
</success_criteria>

## Post-Implementation Steps

<post_implementation>

- Render & lint:

```sh
helm template charts/pipelines-library | yq
yamllint .
```

- (Optional) Create or update PipelineRuns for testing.
- (Optional) Integrate the pipelines into your application flow or App of Apps.
</post_implementation>
