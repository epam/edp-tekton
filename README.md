[![codecov](https://codecov.io/gh/epam/edp-tekton/branch/master/graph/badge.svg?token=P2RDX1F68O)](https://codecov.io/gh/epam/edp-tekton)

# KubeRocketCI Tekton
<!-- TOC -->

- [KubeRocketCI Tekton](#kuberocketci-tekton)
  - [KubeRocketCI Interceptor](#edp-interceptor)
  - [Tekton Pipelines](#tekton-pipelines)

<!-- /TOC -->

The edp-tekton repository consolidates elements for Tekton integration with [KubeRocketCI](https://docs.kuberocketci.io) (former EPAM Delivery Platform (EDP)).
and disposes of two main components:

- **KubeRocketCI Interceptor**. Follows [Tekton Interceptor](https://tekton.dev/vault/triggers-main/clusterinterceptors/) paradigm and enriches payload from different Version Control Systems (VCS) like Gerrit, GitHub, GitLab or BitBucket with the platform specific metadata.
- **Tekton Pipelines**. Consists of [Tekton Tasks, Pipelines, Triggers](https://tekton.dev/docs/pipelines/) and implements KubeRocketCI Pipelines logic. Some of the tasks are forks from [origin source](https://github.com/tektoncd/catalog), the others are platform specific.

## KubeRocketCI Interceptor

KubeRocketCI Interceptor is used as a component that provides KubeRocketCI metadata for Tekton Pipelines. The code is based on [Upstream implementation](https://github.com/tektoncd/triggers/tree/main/pkg/interceptors).

KubeRocketCI Interceptor extracts information from VCS payload, like `repository_name`. The `repository_name` has 1-2-1 mapping with the KubeRocketCI `Codebase` (kind: Codebase; apiVersion:v2.edp.epam.com/v1). Interceptor populates Tekton Pipelines with [Codebase SPEC](https://github.com/epam/edp-codebase-operator/blob/master/docs/api.md#codebasespec) data, see the diagram below:

        ┌────────────┐              ┌──────────────────┐       ┌─────────────┐
        │            │              │   KubeRocketCI   │       │   Tekton    │
        │  VCS(Git)  ├──────────────►                  ├───────►             │
        │            │              │   Interceptor    │       │  Pipelines  │
        └──────┬─────┘              └────────┬─────────┘       └─────────────┘
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
The docker images for Interceptor are available on the [DockerHub](https://hub.docker.com/repository/docker/epamedp/edp-tekton).
The helm-chart for interceptor deployment is in the same repository by the [charts/interceptor](./charts/pipelines-library/values.yaml#L146) directory.

## Tekton Pipelines

Tekton Pipelines supports four VCS: Gerrit, GitHub, GitLab and BitBucket. To check the VCS Import strategy, please refer to the [KubeRocketCI Documentation](https://docs.kuberocketci.io)).

Tekton Pipelines are implemented and packaged using the [helm-chart](./charts/pipelines-library/) approach. The helm-chart contains:

- `Tasks` - basic building block for Tekton. Some of the tasks are forks from [Upstream Tekton Catalog](https://github.com/tektoncd/catalog).
- `Pipelines`, which consist of `Tasks` and implement logic for the CI flow. KubeRocketCI follows the below approach for pipelines definition:
  - Each type of VCS has its own Pipelines, e.g. for Gerrit, GitHub, GitLab and BitBucket;
  - KubeRocketCI has [two types of Pipelines](https://docs.kuberocketci.io/docs/operator-guide/ci/tekton-overview): `CodeReview` - triggers on Review, `Build` - triggers on Merged Event.
- `Triggers`, `TriggerBindings`, `TriggerTemplates` - defines the logic for specific VCS Events (Gerrit, GitHub, GitLab, BitBucket) and Pipelines.
- `Resources` - Kubernetes resources, that are used from Pipelines, e.g. `ServiceAccount` with [IRSA Enablement](https://docs.kuberocketci.io/docs/developer-guide/aws-reference-architecture#iam-roles-for-service-accounts-irsa), `ConfigMaps` for Maven/Gradle Pipelines, Tekton cache, CodeNarc, CTLint, and PVC to share resources between Tasks.
- `Tekton Pipeline pruner` - created as a cron job, it is designed to clear outdated pipelines.
