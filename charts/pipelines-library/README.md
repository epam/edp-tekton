# edp-tekton

![Version: 0.1.9](https://img.shields.io/badge/Version-0.1.9-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.9](https://img.shields.io/badge/AppVersion-0.1.9-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart for EDP Tekton Pipelines

## Additional Information

Tekton Pipelines supports three VCS: Gerrit, GitHub, GitLab. To check the VCS Import strategy, please refer to the [EDP Documentation](https://epam.github.io/edp-install/operator-guide/import-strategy/).

EDP Tekton Pipelines are implemented and packaged using the [helm-chart](./charts/pipelines-library/) approach. The helm-chart contains:

- `Tasks` - basic building block for Tekton. Some of the tasks are forks from [Upstream Tekton Catalog](https://github.com/tektoncd/catalog).
- `Pipelines`, which consist of `Tasks` and implement logic for the CI flow. EDP follows the below approach for pipelines definition:
  - Each type of VCS has its own Pipelines, e.g. for Gerrit, GitHub, GitLab;
  - EDP has [two types of Pipelines](https://epam.github.io/edp-install/user-guide/ci-pipeline-details/): `CodeReview` - triggers on Review, `Build` - triggers on Merged Event.
- `Triggers`, `TriggerBindings`, `TriggerTemplates` - defines the logic for specific VCS Events (Gerrit, GitHub, GitLab) and Pipelines.
- `Resources` - Kubernetes resources, that are used from Pipelines, e.g. `ServiceAccount` with [IRSA Enablement](https://epam.github.io/edp-install/operator-guide/kaniko-irsa/), `ConfigMaps` for Maven/Gradle Pipelines, PVC to share resources between Tasks.

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
| https://epam.github.io/edp-helm-charts/stable | edp-tekton-common-library | 0.1.9 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| buildTool.go.cache.persistentVolume.size | string | `"5Gi"` |  |
| buildTool.go.cache.persistentVolume.storageClass | string | `"ebs-sc"` |  |
| fullnameOverride | string | `""` |  |
| gerrit.enabled | bool | `true` | Deploy Gerrit related components. Default: true |
| gerrit.sshPort | int | `30003` | Gerrit port |
| github.enabled | bool | `false` |  |
| github.host | string | `"github.com"` | The GitHub host, adjust this if you run a GitHub enterprise. Default: github.com |
| github.webhook.existingSecret | string | `"github-config"` | Existing secret which holds GitHub integration credentials: Username, Access Token, Secret String and Private SSH Key |
| gitlab.enabled | bool | `false` |  |
| gitlab.host | string | `"git.epam.com"` | The GitLab host, adjust this if you run a GitLab enterprise. Default: gitlab.com |
| gitlab.webhook.existingSecret | string | `"gitlab-config"` | Existing secret which holds GitLab integration credentials: Username, Access Token, Secret String and Private SSH Key |
| global.dnsWildCard | string | `"eks-sandbox.aws.main.edp.projects.epam.com"` | a cluster DNS wildcard name |
| kaniko.roleArn | string | `"arn:aws:iam::093899590031:role/AWSIRSACoreSandboxEdpDeliveryKaniko"` | AWS IAM role to be used for kaniko pod service account (IRSA). Format: arn:aws:iam::<AWS_ACCOUNT_ID>:role/<AWS_IAM_ROLE_NAME> |
| kaniko.serviceAccount.create | bool | `false` | Specifies whether a service account should be created |
| nameOverride | string | `""` |  |
| tekton.pruner.create | bool | `true` | Specifies whether a cronjob should be created |
| tekton.pruner.keep | int | `1` | Maximum number of resources to keep while deleting removing |
| tekton.pruner.resources | string | `"pipelinerun,taskrun"` | Supported resources for auto prune are 'taskrun' and 'pipelinerun' |
| tekton.pruner.schedule | string | `"0 18 * * *"` | How often to clean up resources |
| tektonUrl | string | `"https://tekton.eks-sandbox.aws.main.edp.projects.epam.com"` | Tekton URL. Link to tekton Dashboard |
