# tekton-cache

![Version: 0.3.3](https://img.shields.io/badge/Version-0.3.3-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.3.3](https://img.shields.io/badge/AppVersion-0.3.3-informational?style=flat-square)

A Helm chart for EDP Tekton Cache

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
| affinity | object | `{}` | Pod affinity. |
| cacheSize | string | `"5Gi"` | Defines size of the Persistent Volume that is used for cache. |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Defines the policy with which pod will request its image. |
| image.repository | string | `"epamedp/tekton-cache"` | Tekton-cache container image. |
| image.tag | string | `"0.1.2"` | Overrides the image tag whose default is the chart appVersion. |
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

