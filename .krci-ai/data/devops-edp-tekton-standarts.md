# Tekton Pipelines & Tasks Standards

## General Principles

- Always follow Tekton Pipelines and Tasks best practices for reliability and maintainability  
- Keep configuration declarative and version-controlled according to GitOps principles  
- Separate concerns: pipeline and task configuration should be distinct from deployment configuration  
- Validate all pipeline and task changes through proper testing before deployment  

---

## Tekton Pipelines Best Practices

- **Use stable API versions** in manifests; avoid deprecated resources  
- **Validate resource compatibility** with your cluster version  
- **Set resource limits and requests** for all Tekton workloads  
- **Apply RBAC and security best practices**: Use least privilege, audit roles regularly  
- **Use namespaces for isolation**: Each pipeline or task should have its own dedicated namespace  
- **Use labels and annotations** for better resource organization and discovery  
- **Follow secret management best practices**: Avoid hardcoding sensitive information  

---

## Helm Chart Structure & Standards for Tekton Components

- **Wrapper Chart Pattern**: For existing Helm charts, create wrapper charts with the external chart as a dependency  
- **Configuration Structure**: Use the dependency chart name as a top-level key in values.yaml  
- **Values Placement**: Keep all custom logic and configuration in the pipeline or task’s values.yaml, not in App of Apps values  
- **Directory Structure**:  
  - Shared helpers are stored in `charts/common-library/`  
  - Pipelines, tasks, triggers, and supporting resources are stored in `charts/pipelines-library/templates/{pipelines,tasks,triggers,resources}`  
  - Scripts for onboarding and maintenance are in `charts/pipelines-library/scripts/`  
- **Chart Documentation**: Update README.md and Chart.yaml with every version change  
- **Version Control**: Follow semantic versioning for charts and increment versions with changes  
- **Dependencies**: Clearly define dependencies in Chart.yaml using proper versioning  

---

## App of Apps Pattern Best Practices for Tekton

- **Minimal App of Apps Configuration**: App of Apps values.yaml should ONLY contain:  
  - Enable/disable flags for each pipeline or task  
  - Namespace settings  
  - CreateNamespace option  
- **Application Templates**: Use conditional templates to enable/disable pipelines and tasks  
- **Cluster Isolation**: Use separate cluster directories to isolate configuration between environments  

---

## Maintaining and Updating Tekton Charts

- **Review Upstream Changes**: Always check the CHANGELOG of dependency charts before updating  
- **Chart Testing**: Validate charts using `helm lint` and pytest-based tests under charts/pipelines-library/tests  
- **Version Increment**: Always increment chart version when making changes  
- **Diff Analysis**: Before applying updates, review diffs for potential breaking changes  
- **Rollback Plan**: Always have a rollback strategy ready in case of deployment issues  

---

## Repository Structure & Files

The repository includes two main Helm chart libraries: `charts/common-library` and `charts/pipelines-library`. Their directory structures and content requirements are as follows:

```
charts/
├── common-library/
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/ (helpers, shared fragments)
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

- **Onboarding Script**: An onboarding script is located at `charts/pipelines-library/scripts/onboarding-component.sh` to facilitate adding new pipelines and tasks following the standards.

### Directory Contents and Requirements

- **pipelines/**  
  Contains Tekton Pipeline manifests. Each pipeline must:  
  - Use a filename matching the pipeline's `metadata.name` field (e.g., `build-pipeline.yaml` for a pipeline named `build-pipeline`).  
  - Follow naming conventions consistent with organizational standards.  
  - Include required fields such as `spec.tasks`, `spec.workspaces` if applicable, and proper `runAfter` policies to define task order.  
  - Use feature flags in the pipeline definition to enable or disable optional steps or functionality.

- **tasks/**  
  Contains Tekton Task manifests. Each task must:  
  - Use a filename matching the task's `metadata.name` field (e.g., `build-task.yaml` for a task named `build-task`).  
  - Follow naming conventions consistent with organizational standards.  
  - Define required fields including `spec.steps`, appropriate resource requests and limits, and workspace declarations as needed.  
  - Support feature flags to toggle task behavior or optional steps.

- **triggers/**  
  Contains Tekton TriggerTemplates, TriggerBindings, and EventListeners. Each resource should:  
  - Follow naming and structuring standards consistent with the rest of the repository.  
  - Be properly annotated and labeled for discovery and management.

- **resources/**  
  Contains Tekton PipelineResource manifests (if used) or other related resource definitions. Each resource:  
  - Must follow naming conventions and include necessary metadata and spec fields.

### General Standards for All Manifests

- Use stable API versions and validate compatibility with the target cluster.  
- Include appropriate labels and annotations for resource management and tracking.  
- Ensure workspace and `runAfter` policies are used consistently to manage dependencies and data sharing between tasks.  
- Follow the generation flow and naming conventions defined in the Tekton standards documentation.  
- Validate manifests against schemas and test changes before deployment to avoid runtime errors.  
- Use feature flags to enable or disable functionality dynamically without requiring manifest changes.

This structure and these standards ensure maintainability, scalability, and consistency across all Tekton pipelines and tasks in the repository.    