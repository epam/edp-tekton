# edp-tekton-common-library

![Version: 0.2.18](https://img.shields.io/badge/Version-0.2.18-informational?style=flat-square) ![Type: library](https://img.shields.io/badge/Type-library-informational?style=flat-square) ![AppVersion: 0.2.18](https://img.shields.io/badge/AppVersion-0.2.18-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart library with common components for EDP Tekton Pipelines

## Additional Information

Common library that holds steps for EDP pipelines

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
| global.gerritSSHPort | string | `"30003"` | Gerrit SSH node port |
| tekton.resources | object | `{"limits":{"cpu":"2","memory":"3Gi"},"requests":{"cpu":"0.5","memory":"2Gi"}}` | The resource limits and requests for the Tekton Tasks |
