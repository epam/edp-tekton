# edp-custom-pipelines

![Version: 0.4.0](https://img.shields.io/badge/Version-0.4.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.4.0](https://img.shields.io/badge/AppVersion-0.4.0-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart for EDP4EDP Tekton Pipelines

## Additional Information

Custom library EDP4EDP delivers custom Tekton pipelines used by the EDP Platform itself. This library is an example of EDP Tekton Pipelines customization. All the functionality which extends the EDP pipelines core logic should be placed in a separate chart to guarantee proper platform upgrades in the future.

**Homepage:** <https://epam.github.io/edp-install/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| epmd-edp | <SupportEPMD-EDP@epam.com> | <https://solutionshub.epam.com/solution/epam-delivery-platform> |
| sergk |  | <https://github.com/SergK> |

## Source Code

* <https://github.com/epam/edp-tekton>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../common-library | edp-tekton-common-library | 0.2.6 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| fullnameOverride | string | `""` |  |
| global.dnsWildCard | string | `""` | a cluster DNS wildcard name |
| global.gerritSSHPort | string | `"30003"` | Gerrit SSH node port |
| global.gitProvider | string | `"gerrit"` | Define Git Provider to be used in Pipelines. Can be gerrit (default), gitlab, github |
| global.platform | string | `"kubernetes"` | platform type that can be "kubernetes" or "openshift" |
| nameOverride | string | `""` |  |
| tekton.resources | object | `{"limits":{"cpu":"2","memory":"3Gi"},"requests":{"cpu":"0.5","memory":"2Gi"}}` | The resource limits and requests for the Tekton Tasks |
