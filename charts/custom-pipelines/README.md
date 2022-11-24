# edp-custom-pipelines

![Version: 0.1.6](https://img.shields.io/badge/Version-0.1.6-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.6](https://img.shields.io/badge/AppVersion-0.1.6-informational?style=flat-square)
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
| https://epam.github.io/edp-helm-charts/stable | edp-tekton-common-library | 0.1.5 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| fullnameOverride | string | `""` |  |
| gerrit.enabled | bool | `true` | Deploy Gerrit related components. Default: true |
| gerrit.sshPort | int | `30003` | Gerrit port |
| global.dnsWildCard | string | `"eks-sandbox.aws.main.edp.projects.epam.com"` | a cluster DNS wildcard name |
| kaniko.roleArn | string | `"arn:aws:iam::093899590031:role/AWSIRSACoreSandboxEdpDeliveryKaniko"` | AWS IAM role to be used for kaniko pod service account (IRSA). Format: arn:aws:iam::<AWS_ACCOUNT_ID>:role/<AWS_IAM_ROLE_NAME> |
| kaniko.serviceAccount.create | bool | `false` | Specifies whether a service account should be created |
| nameOverride | string | `""` |  |
| tektonUrl | string | `"https://tekton.eks-sandbox.aws.main.edp.projects.epam.com"` | Tekton URL. Link to tekton Dashboard |
