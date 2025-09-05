# Tekton Pipelines & Tasks Standards

## General Principles

- The repository follows Tekton Pipelines and Tasks best practices aimed at reliability and maintainability.
- Configurations are maintained in a declarative manner aligned with GitOps principles.
- Pipeline and task configurations are organized separately from deployment configurations.
- Changes to pipelines and tasks are validated through testing before deployment to ensure stability.

---

## Tekton Pipelines Overview

- Resource compatibility is considered to maintain smooth operation.
- Resource limits and requests are defined for Tekton workloads to optimize performance and resource usage.
- Security practices include the application of RBAC with least privilege and regular auditing of roles.
- Namespaces are used to isolate pipelines and tasks.
- Labels and annotations assist in resource organization and discovery.
- Secret management follows best practices to avoid exposing sensitive information directly.

---

## Helm Chart Structure for Tekton Components

The repository contains Helm charts organized to support Tekton components effectively:

- Configuration values are structured with the dependency chart name as a top-level key in `values.yaml`.
- Custom logic and configuration related to pipelines and tasks reside within their respective `values.yaml` files.
- Shared helper templates are located in `charts/common-library/`.
- Pipelines, tasks, triggers, and supporting resources are organized under `charts/pipelines-library/templates/` with subdirectories for each resource type: `pipelines`, `tasks`, `triggers`, and `resources`.
- Scripts for onboarding new pipelines and tasks, as well as maintenance tasks, are stored in `charts/pipelines-library/scripts/`.
- Documentation including `README.md` and `Chart.yaml` files are maintained and updated with each chart version.
- Semantic versioning is followed for chart versions, with dependencies clearly defined in `Chart.yaml`.

---

## Maintaining and Updating Tekton Charts

- Updates involve reviewing upstream changes through changelogs.
- Charts are validated using linting and pytest-based tests located under `charts/pipelines-library/tests`.
- Chart versions are incremented following semantic versioning principles.
- Diff analysis is conducted prior to applying updates to identify potential breaking changes.
- Rollback strategies are considered to manage deployment issues effectively.

---

## Repository Structure & Files

The repository includes two main Helm chart libraries: `charts/common-library` and `charts/pipelines-library`. Their directory structures and contents are organized as follows:

```
charts/
├── common-library/
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/ (shared fragments)
└── pipelines-library/
    ├── Chart.yaml
    ├── values.yaml
    ├── scripts/ (onboarding-component.sh, tekton-prune.sh)
    └── templates/
        ├── pipelines/
        ├── tasks/
        ├── triggers/
        └── resources/
```

- The onboarding script located at `charts/pipelines-library/scripts/onboarding-component.sh` is used to add new pipelines and tasks in line with the repository's structure.

### Pipelines

The `pipelines/` directory contains Tekton Pipeline manifests. Pipeline manifests are named to correspond with their `metadata.name` fields. The pipelines include definitions for tasks, workspaces if applicable, and use `runAfter` policies to establish task execution order. Feature flags may be used within pipeline definitions to enable or disable pipelines.

- Two versioning strategies are used for build pipelines:
  - *default*: currently in use
  - *edp*: reserved for extended patterns

### Tasks

The `tasks/` directory holds Tekton Task manifests. Task filenames align with their `metadata.name` fields. Task manifests define steps, resource requests and limits, and workspace declarations. Feature flags can be incorporated to adjust task behavior or toggle optional steps.

### Triggers

The `triggers/` directory includes Tekton TriggerTemplates, TriggerBindings, and EventListeners. These resources follow consistent naming and structuring conventions and include annotations and labels to support resource discovery and management.

### Resources

The `resources/` directory contains supporting resource definitions (e.g., ConfigMaps, roles, settings). Naming conventions and metadata are organized to support clarity and maintainability.

---

## General Standards for All Manifests

- Manifests throughout the repository use stable API versions.
- Labels and annotations are applied to aid resource management and tracking.
- Workspaces and `runAfter` policies are consistently used to manage dependencies and data sharing between tasks.
- Naming conventions and generation flows align with the Tekton standards documentation.
- Validation and testing are part of the workflow to improve reliability.
- Feature flags provide dynamic control over functionality without requiring manifest changes.

This structure and approach support maintainability, scalability, and consistency across all Tekton pipelines and tasks within the repository.