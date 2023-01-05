# edp-tekton-interceptor

![Version: 0.2.4](https://img.shields.io/badge/Version-0.2.4-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.2.4](https://img.shields.io/badge/AppVersion-0.2.4-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart for EDP Tekton Interceptor

## Additional Information

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
Follows [Tekton Interceptor](https://tekton.dev/vault/triggers-main/clusterinterceptors/) paradigm and enriches payload from different Version Control Systems (VCS) like Gerrit, GitHub or GitLab with EDP specific data.

**Homepage:** <https://epam.github.io/edp-install/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| epmd-edp | <SupportEPMD-EDP@epam.com> | <https://solutionshub.epam.com/solution/epam-delivery-platform> |
| sergk |  | <https://github.com/SergK> |

## Source Code

* <https://github.com/epam/edp-tekton>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| global.platform | string | `"kubernetes"` | platform type that can be "kubernetes" or "openshift" |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"epamedp/edp-tekton"` |  |
| image.tag | string | `nil` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsGroup | int | `65532` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `65532` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.name | string | `"edp-interceptor"` | If not set, a name is generated using the fullname template |
