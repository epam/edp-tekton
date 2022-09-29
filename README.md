[![codecov](https://codecov.io/gh/epam/edp-tekton/branch/master/graph/badge.svg?token=P2RDX1F68O)](https://codecov.io/gh/epam/edp-tekton)

# EDP Tekton
<!-- TOC -->

- [EDP Tekton](#edp-tekton)
  - [EDP Interceptor](#edp-interceptor)
  - [Tekton Pipelines](#tekton-pipelines)

<!-- /TOC -->

The edp-tekton repository consolidates elements for Tekton integration with EDP [EPAM Delivery Platform (EDP)](https://epam.github.io/edp-install/).
and disposes of two main components:

- **EDP Interceptor**. Follows [Tekton Interceptor](https://tekton.dev/vault/triggers-main/clusterinterceptors/) paradigm and enriches payload from different Version Control Systems (VCS) like Gerrit, GitHub or GitLab with EDP specific data.
- **Tekton Pipelines**. Consists of [Tekton Tasks, Pipelines, Triggers](https://tekton.dev/docs/pipelines/) and implements EDP CI Pipelines logic. Some of the tasks are forks from [origin source](https://github.com/tektoncd/catalog), the others are EDP specific.

## EDP Interceptor

EDP Interceptor is used as a component that provides EDP data for Tekton Pipelines. The code is based on [Upstream implementation](https://github.com/tektoncd/triggers/tree/main/pkg/interceptors).

EDP Interceptor extracts information from VCS payload, like `repository_name`. The `repository_name` has 1-2-1 mapping with `EDP Codebase` (kind: Codebase; apiVersion:v2.edp.epam.com/v1). Interceptor populates Tekton Pipelines with [Codebase SPEC](https://github.com/epam/edp-codebase-operator/blob/master/docs/api.md#codebasespec) data, see the diagram below:

        ┌────────────┐              ┌─────────────────┐       ┌─────────────┐
        │            │              │ EDP Interceptor │       │   Tekton    │
        │  VCS(Git)  ├──────────────►                 ├───────►             │
        │            │              │                 │       │  Pipelines  │
        └──────┬─────┘              └────────┬────────┘       └─────────────┘
               │                             │
        ┌──────┴─────┐                       │ extract
        │    Repo    │                       │
        │            │                       │
        │            │      ┌────────────────▼───────────────┐
        └────────────┘      │ apiVersion: v2.edp.epam.com/v1 │
                            │ kind: Codebase                 │
                            │                                │
                            │ spec:                          │
                            └────────────────────────────────┘

The data, retrieved from the Codebase SPEC, is used in Tekton Pipelines logic.
The docker images for EDP Interceptor are available on the [DockerHub](https://hub.docker.com/repository/docker/epamedp/edp-tekton).
The helm-chart for interceptor deployment is in the same repository by the [charts/interceptor](./charts/interceptor) directory.

## Tekton Pipelines

Tekton Pipelines supports three VCS: Gerrit, GitHub, GitLab. To check the VCS Import strategy, please refer to the [EDP Documentation](https://epam.github.io/edp-install/operator-guide/import-strategy/).

EDP Tekton Pipelines are implemented and packaged using the [helm-chart](./charts/pipelines-library/) approach. The helm-chart contains:

- `Tasks` - basic building block for Tekton. Some of the tasks are forks from [Upstream Tekton Catalog](https://github.com/tektoncd/catalog).
- `Pipelines`, which consist of `Tasks` and implement logic for the CI flow. EDP follows the below approach for pipelines definition:
  - Each type of VCS has its own Pipelines, e.g. for Gerrit, GitHub, GitLab;
  - EDP has [two types of Pipelines](https://epam.github.io/edp-install/user-guide/ci-pipeline-details/): `CodeReview` - triggers on Review, `Build` - triggers on Merged Event.
- `Triggers`, `TriggerBindings`, `TriggerTemplates` - defines the logic for specific VCS Events (Gerrit, GitHub, GitLab) and Pipelines.
- `Resources` - Kubernetes resources, that are used from Pipelines, e.g. `ServiceAccount` with [IRSA Enablement](https://epam.github.io/edp-install/operator-guide/kaniko-irsa/), `ConfigMaps` for Maven/Gradle Pipelines, PVC to share resources between Tasks.
