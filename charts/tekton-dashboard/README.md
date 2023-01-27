# edp-tekton-dashboard

![Version: 0.32.0](https://img.shields.io/badge/Version-0.32.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.32.0](https://img.shields.io/badge/AppVersion-0.32.0-informational?style=flat-square)

A Helm chart for EDP Tekton Dashboard

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
| fullnameOverride | string | `""` |  |
| global.dnsWildCard | string | `""` | a cluster DNS wildcard name |
| global.edpName | string | `""` | namespace or a project name |
| global.platform | string | `"kubernetes"` | platform type that can be "kubernetes" or "openshift" |
| image.repository | string | `"gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard"` | Define tekton dashboard docker image name |
| image.tag | string | `"v0.32.0"` | Define tekton dashboard docker image tag |
| ingress.annotations | object | `{}` | Annotations for Ingress resource |
| ingress.tls | list | `[]` | Ingress TLS configuration |
| nameOverride | string | `"tekton-dashboard"` |  |

