# edp-tekton

![Version: 0.1.1](https://img.shields.io/badge/Version-0.1.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.1](https://img.shields.io/badge/AppVersion-0.1.1-informational?style=flat-square)

A Helm chart for EDP Tekton Pipelines

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
| gerrit.enabled | bool | `true` | Deploy Gerrit related components. Default: true |
| gerrit.sshPort | int | `30003` | Gerrit port |
| github.enabled | bool | `false` |  |
| github.webhook.existingSecret | string | `"github.com-config"` | Existing secret which holds both GitHub Access and Secret Token, default is github-configuration, which is aligned with codebase-operator |
| github.webhook.secretKeys.secretKey | string | `"secretString"` |  |
| github.webhook.secretKeys.tokenKey | string | `"token"` |  |
| gitlab.enabled | bool | `false` |  |
| gitlab.host | string | `"git.epam.com"` | The GitLab host, adjust this if you run a GitLab enterprise. Default: gitlab.com |
| gitlab.webhook.existingSecret | string | `"gitlab-configuration"` | Existing secret which holds both GitLab Access and Secret Token, default is gitlab-configuration, which is aligned with codebase-operator |
| gitlab.webhook.secretKeys.secretKey | string | `"secretString"` | Key in existingSecret. Generated on Tekton side and populated in GitLab for each Project in section: PROJECT_NAME > Settings > Webhooks > Secret Token |
| gitlab.webhook.secretKeys.tokenKey | string | `"token"` | Key in existingSecret. Generated on GitLab side in section: (User Settings) or (Project Settings) or (Group Settings) > Access Token |
| global.dnsWildCard | string | `"eks-sandbox.aws.main.edp.projects.epam.com"` | a cluster DNS wildcard name |
| kaniko.roleArn | string | `"arn:aws:iam::093899590031:role/AWSIRSACoreSandboxEdpDeliveryKaniko"` | AWS IAM role to be used for kaniko pod service account (IRSA). Format: arn:aws:iam::<AWS_ACCOUNT_ID>:role/<AWS_IAM_ROLE_NAME> |
| kaniko.serviceAccount.create | bool | `false` | Specifies whether a service account should be created |
| nameOverride | string | `""` |  |
| tekton.pruner.create | bool | `true` | Specifies whether a cronjob should be created |
| tekton.pruner.keep | int | `1` | Maximum number of resources to keep while deleting removing |
| tekton.pruner.resources | string | `"pipelinerun,taskrun"` | Supported resources for auto prune are 'taskrun' and 'pipelinerun' |
| tekton.pruner.schedule | string | `"0 18 * * *"` | How often to clean up resources |
| tektonUrl | string | `"https://tekton.eks-sandbox.aws.main.edp.projects.epam.com"` | Tekton URL. Link to tekton Dashboard |
