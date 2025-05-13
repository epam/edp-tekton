# tekton-cache

![Version: 0.4.1](https://img.shields.io/badge/Version-0.4.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.4.1](https://img.shields.io/badge/AppVersion-0.4.1-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart for KubeRocketCI Tekton Cache

## Additional Information

The Tekton Cache Helm chart is designed to enhance your CI/CD pipeline by providing a caching mechanism that is both efficient and easy to integrate.
It leverages the power of Kubernetes/Tekton to cache dependencies and build outputs, reducing build times.

**Homepage:** <https://docs.kuberocketci.io>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| epmd-edp | <SupportEPMD-EDP@epam.com> | <https://solutionshub.epam.com/solution/kuberocketci> |
| sergk |  | <https://github.com/SergK> |

## Source Code

* <https://github.com/epam/edp-tekton>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Pod affinity. |
| cacheSize | string | `"5Gi"` | Defines size of the Persistent Volume that is used for cache. |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Defines the policy with which pod will request its image. |
| image.registry | string | `"docker.io"` | Set the image registry, default to Docker Hub; can be customized to use an alternative provider |
| image.repository | string | `"epamedp/tekton-cache"` | Tekton-cache container image. |
| image.tag | string | `"0.1.3"` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | Specifies secrets for pulling Docker images. |
| initContainers.repository | string | `"busybox"` | Defines the repository. |
| initContainers.tag | string | `"1.36.1"` | InitContainer image. |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` | Node labels for Tekton-cache pod assignment. |
| podAnnotations | object | `{}` | Annotations to be added to Tekton-cache pods. |
| podSecurityContext | object | `{}` | Specifies privilege and access control configurations for Tekton-cache pods. |
| resources | object | `{}` | Tekton-cache pod resource requests and limits. |
| securityContext | object | `{}` | Security context to be added to Tekton-cache pods. |
| service | object | `{"name":"tekton-cache","port":8080,"type":"ClusterIP"}` | Tekton-cache service configurations. |
| tolerations | list | `[]` | Node tolerations for pod scheduling to nodes with taints. |
