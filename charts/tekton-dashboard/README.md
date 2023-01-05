# edp-tekton-dashboard

![Version: 0.31.1](https://img.shields.io/badge/Version-0.31.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.31.1](https://img.shields.io/badge/AppVersion-0.31.1-informational?style=flat-square)

Tekton dashboard

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
| global.image | string | `"gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard:v0.31.0@sha256:454a405aa4f874a0c22db7ab47ccb225a95addd3de904084e35c5de78e4f2c48"` |  |
| global.platform | string | `"kubernetes"` | platform type that can be "kubernetes" or "openshift" |
| ingress.annotations | object | `{}` | Annotations for Ingress resource |
| ingress.tls | list | `[]` | Ingress TLS configuration |
| nameOverride | string | `"tekton-dashboard"` |  |

