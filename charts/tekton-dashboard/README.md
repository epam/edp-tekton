# edp-tekton-dashboard

![Version: 0.30.0](https://img.shields.io/badge/Version-0.30.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.30.0](https://img.shields.io/badge/AppVersion-0.30.0-informational?style=flat-square)

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
| global.image | string | `"gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard:v0.30.0@sha256:85f7d38086fadb07556052ce873d44861c29ef690f47735f32d7e6a153ca8a92"` |  |
| nameOverride | string | `"tekton-dashboard"` |  |

