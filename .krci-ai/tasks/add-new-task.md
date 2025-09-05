# Task: Add New Tekton Task (Helm-templated)

<task_header>
<description>Automate the process of adding a new Tekton **Task** to the `edp-tekton` repository.  
The agent should use the onboarding script to generate a Helm-templated Task file under  
`./charts/pipeline-library/templates/tasks`, follow repository conventions, request  
mandatory fields from the user (e.g., task name), and prepare the file for review.</description>
</task_header>

---

## Reference Assets (Prerequisites)

<prerequisites>
Dependencies:
- Tekton overview: [tekton-overview](https://docs.kuberocketci.io/docs/operator-guide/ci/tekton-overview)
- User guide: [tekton-pipelines](https://docs.kuberocketci.io/docs/user-guide/tekton-pipelines)
- Custom pipelines flow: [custom-pipelines-flow](https://docs.kuberocketci.io/docs/use-cases/custom-pipelines-flow)
- Best practices for `edp-tekton` repository: [repository-best-practices.md](./.krci-ai/data/devops-edp-tekton-standarts.md)

**Repository structure & constants:**
- Directory for new tasks: `./charts/pipeline-library/templates/tasks`
- Onboarding script: `./charts/pipelines-library/scripts/onboarding-component.sh`

**Validation (mandatory before starting):**
1. Ensure the directory `./charts/pipeline-library/templates/tasks` exists. If not — **HALT**.
2. Ensure the onboarding script exists at `./charts/pipelines-library/scripts/onboarding-component.sh`. If missing — **HALT**.
3. Verify online documentation URLs are reachable. If not, continue with a warning.
</prerequisites>

## Overview

<task_overview>
The user provides minimal input (**task name**).  
The agent runs the onboarding script with type `task`, which generates the new  
Tekton Task under `templates/tasks/`.  
After generation, the agent validates file existence and reports back results.
</task_overview>

## Instructions

<instructions>
1. Review the reference documentation (links above).  
2. Collect the task name from the user (kebab-case).  
3. Run the onboarding script with `-type task -n <task_name>`.  
4. Verify the new file is created under `./charts/pipeline-library/templates/tasks/`.  
5. Report back the file path and confirm creation.
</instructions>

## Required Inputs

<user_inputs>
**Mandatory:**
- `task_name` — task name (kebab-case), used in the `-n` flag and the generated file.

**Example question for the user:**  
_"Please provide the name of the new Tekton Task (kebab-case), for example: `ansible-run`."_
</user_inputs>

## Usage Examples

<usage_examples>
```sh
./charts/pipelines-library/scripts/onboarding-component.sh --type task -n ansible-run
./charts/pipelines-library/scripts/onboarding-component.sh --type task -n maven-build
```
</usage_examples>

## Acceptance Criteria

<success_criteria>
- [ ] The onboarding script is executed with `-type task` parameter
- [ ] Task name follows kebab-case convention
- [ ] New Task file exists under './charts/pipeline-library/templates/tasks/'
- [ ] Task file contains proper Helm template wrapping
- [ ] Task file contains `apiVersion: tekton.dev/v1`, `kind: Task`
- [ ] Task file has proper metadata, labels and descriptions
- [ ] The agent reports back the created file path
</success_criteria>

## Post-Implementation Steps

<post_implementation>
- Validate Helm template syntax:
```sh
helm template charts/pipelines-library | yq
```
- Review generated task structure for completeness
- Verify task integration with existing pipelines if applicable
</post_implementation>