# edp-tekton

![Version: 0.12.0-SNAPSHOT](https://img.shields.io/badge/Version-0.12.0--SNAPSHOT-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.12.0-SNAPSHOT](https://img.shields.io/badge/AppVersion-0.12.0--SNAPSHOT-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart for EDP Tekton Pipelines

## Additional Information

### EDP Tekton Pipelines

Tekton Pipelines supports three VCS: Gerrit, GitHub, GitLab. To check the VCS Import strategy, please refer to the [EDP Documentation](https://epam.github.io/edp-install/operator-guide/import-strategy/).

EDP Tekton Pipelines are implemented and packaged using the [helm-chart](./charts/pipelines-library/) approach. The helm-chart contains:

- `Tasks` - basic building block for Tekton. Some of the tasks are forks from [Upstream Tekton Catalog](https://github.com/tektoncd/catalog).
- `Pipelines`, which consist of `Tasks` and implement logic for the CI flow. EDP follows the below approach for pipelines definition:
  - Each type of VCS has its own Pipelines, e.g. for Gerrit, GitHub, GitLab;
  - EDP has [two types of Pipelines](https://epam.github.io/edp-install/user-guide/ci-pipeline-details/): `CodeReview` - triggers on Review, `Build` - triggers on Merged Event.
- `Triggers`, `TriggerBindings`, `TriggerTemplates` - defines the logic for specific VCS Events (Gerrit, GitHub, GitLab) and Pipelines.
- `Resources` - Kubernetes resources, that are used from Pipelines, e.g. `ServiceAccount` with [IRSA Enablement](https://epam.github.io/edp-install/operator-guide/kaniko-irsa/), `ConfigMaps` for Maven/Gradle Pipelines, PVC to share resources between Tasks.

### EDP Interceptor

EDP Interceptor is used as a component that provides EDP data for Tekton Pipelines. The code is based on [Upstream implementation](https://github.com/tektoncd/triggers/tree/main/pkg/interceptors).

EDP Interceptor extracts information from VCS payload, like `repository_name`. The `repository_name` has 1-2-1 mapping with `EDP Codebase` (kind: Codebase; apiVersion:v2.edp.epam.com/v1). Interceptor populates Tekton Pipelines with [Codebase SPEC](https://github.com/epam/edp-codebase-operator/blob/master/docs/api.md#codebasespec) data, see the diagram below:

        ┌────────────┐              ┌─────────────────┐       ┌─────────────┐
        │            │              │ EDP Interceptor │       │   Tekton    │
        │  VCS(Git)  ├──────────────►                 ├───────►             │
        │            │              │                 │       │  Pipelines  │
        └──────┬─────┘              └────────┬────────┘       └─────────────┘
               │                             │
        ┌──────┴─────┐                       │ extract
        │    Repo    │                       │
        │            │                       │
        │            │      ┌────────────────▼───────────────┐
        └────────────┘      │ apiVersion: v2.edp.epam.com/v1 │
                            │ kind: Codebase                 │
                            │                                │
                            │ spec:                          │
                            └────────────────────────────────┘

The data, retrieved from the Codebase SPEC, is used in Tekton Pipelines logic.
The docker images for EDP Interceptor are available on the [DockerHub](https://hub.docker.com/repository/docker/epamedp/edp-tekton).
The helm-chart for interceptor deployment is in the same repository by the [charts/interceptor](./charts/interceptor) directory.
Follows [Tekton Interceptor](https://tekton.dev/vault/triggers-main/clusterinterceptors/) paradigm and enriches payload from different Version Control Systems (VCS) like Gerrit, GitHub or GitLab with EDP specific data.

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
| @epamedp | tekton-cache | 0.3.2 |
| file://../common-library | edp-tekton-common-library | 0.2.17 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| ctLint.chartSchema | string | `"name: str()\nhome: str()\nversion: str()\ntype: str()\napiVersion: str()\nappVersion: any(str(), num())\ndescription: str()\nkeywords: list(str(), required=False)\nsources: list(str(), required=True)\nmaintainers: list(include('maintainer'), required=True)\ndependencies: list(include('dependency'), required=False)\nicon: str(required=False)\nengine: str(required=False)\ncondition: str(required=False)\ntags: str(required=False)\ndeprecated: bool(required=False)\nkubeVersion: str(required=False)\nannotations: map(str(), str(), required=False)\n---\nmaintainer:\n  name: str(required=True)\n  email: str(required=False)\n  url: str(required=False)\n---\ndependency:\n  name: str()\n  version: str()\n  repository: str()\n  condition: str(required=False)\n  tags: list(str(), required=False)\n  enabled: bool(required=False)\n  import-values: any(list(str()), list(include('import-value')), required=False)\n  alias: str(required=False)\n"` |  |
| ctLint.lintconf | string | `"---\nrules:\n  braces:\n    min-spaces-inside: 0\n    max-spaces-inside: 0\n    min-spaces-inside-empty: -1\n    max-spaces-inside-empty: -1\n  brackets:\n    min-spaces-inside: 0\n    max-spaces-inside: 0\n    min-spaces-inside-empty: -1\n    max-spaces-inside-empty: -1\n  colons:\n    max-spaces-before: 0\n    max-spaces-after: 1\n  commas:\n    max-spaces-before: 0\n    min-spaces-after: 1\n    max-spaces-after: 1\n  comments:\n    require-starting-space: true\n    min-spaces-from-content: 2\n  document-end: disable\n  document-start: disable           # No --- to start a file\n  empty-lines:\n    max: 2\n    max-start: 0\n    max-end: 0\n  hyphens:\n    max-spaces-after: 1\n  indentation:\n    spaces: consistent\n    indent-sequences: whatever      # - list indentation will handle both indentation and without\n    check-multi-line-strings: false\n  key-duplicates: enable\n  line-length: disable              # Lines can be any length\n  new-line-at-end-of-file: enable\n  new-lines:\n    type: unix\n  trailing-spaces: enable\n  truthy:\n    level: warning\n"` |  |
| ctLint.validateMaintainers | bool | `false` |  |
| dashboard.affinity | object | `{}` | Affinity settings for pod assignment |
| dashboard.enabled | bool | `true` | Deploy EDP Dashboard as a part of pipeline library when true. Default: true |
| dashboard.image.repository | string | `"gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard"` | Define tekton dashboard docker image name |
| dashboard.image.tag | string | `"v0.43.1"` | Define tekton dashboard docker image tag |
| dashboard.ingress.annotations | object | `{}` | Annotations for Ingress resource |
| dashboard.ingress.enabled | bool | `true` | Enable external endpoint access. Default Ingress/Route host pattern: tekton-{{ .Release.Namespace }}.{{ .Values.global.dnsWildCard }} |
| dashboard.ingress.host | string | `""` | If not defined the will create by pattern "tekton-[namespace].[global DNS wildcard]" |
| dashboard.ingress.tls | list | `[]` | If hosts not defined the will create by pattern "tekton-[namespace].[global DNS wildcard]" |
| dashboard.nameOverride | string | `"tekton-dashboard"` |  |
| dashboard.nodeSelector | object | `{}` | Node labels for pod assignment |
| dashboard.openshift_proxy | object | `{"enabled":false,"image":{"repository":"quay.io/openshift/origin-oauth-proxy","tag":"4.9.0"},"resources":{"limits":{"cpu":"60m","memory":"70Mi"},"requests":{"cpu":"50m","memory":"40Mi"}}}` | For EKS scenario - uncomment dashboard.ingress.annotations block |
| dashboard.openshift_proxy.enabled | bool | `false` | Enable oauth-proxy to include authorization layer on tekton-dashboard. Default: false |
| dashboard.openshift_proxy.image.repository | string | `"quay.io/openshift/origin-oauth-proxy"` | oauth-proxy image repository |
| dashboard.openshift_proxy.image.tag | string | `"4.9.0"` | oauth-proxy image tag |
| dashboard.openshift_proxy.resources | object | `{"limits":{"cpu":"60m","memory":"70Mi"},"requests":{"cpu":"50m","memory":"40Mi"}}` | The resource limits and requests for the Tekton Proxy |
| dashboard.pipelinesNamespace | string | `"tekton-pipelines"` | Namespace where cluster tekton pipelines deployed. Default: tekton-pipelines |
| dashboard.readOnly | bool | `false` | Define mode for Tekton Dashboard. Enable/disaable capability to create/modify/remove Tekton objects via Tekton Dashboard. Default: false |
| dashboard.resources | object | `{"limits":{"cpu":"60m","memory":"70Mi"},"requests":{"cpu":"50m","memory":"40Mi"}}` | The resource limits and requests for the Tekton Dashboard |
| dashboard.tolerations | list | `[]` | Toleration labels for pod assignment |
| dashboard.triggersNamespace | string | `"tekton-pipelines"` | Namespace where cluster tekton triggers deployed. Default: tekton-pipelines |
| fullnameOverride | string | `""` |  |
| gitServers | object | `{}` |  |
| global.dnsWildCard | string | `""` | a cluster DNS wildcard name |
| global.dockerRegistry.type | string | `"ecr"` | Define Image Registry that will to be used in Pipelines. Can be ecr (default), harbor, dockerhub |
| global.dockerRegistry.url | string | `"<AWS_ACCOUNT_ID>.dkr.ecr.<AWS_REGION>.amazonaws.com/<registry_space>"` | Docker Registry endpoint. In dockerhub case the URL must be specified in accordance with the Kaniko name convention (docker.io/<registry_space>) |
| global.gitProviders | list | `["github","gitlab","gerrit"]` | Deploy Kubernetes Resources for the specific Git Provider. Can be gerrit, gitlab, github (default) |
| global.platform | string | `"kubernetes"` | platform type that can be "kubernetes" or "openshift" |
| interceptor.affinity | object | `{}` | Affinity settings for pod assignment |
| interceptor.enabled | bool | `true` | Deploy EDP interceptor as a part of pipeline library when true. Default: true |
| interceptor.image.pullPolicy | string | `"IfNotPresent"` |  |
| interceptor.image.repository | string | `"epamedp/edp-tekton"` |  |
| interceptor.image.tag | string | `nil` | Overrides the image tag whose default is the chart appVersion. |
| interceptor.imagePullSecrets | list | `[]` |  |
| interceptor.nameOverride | string | `"tekton-interceptor"` |  |
| interceptor.nodeSelector | object | `{}` | Node labels for pod assignment |
| interceptor.podAnnotations | object | `{}` |  |
| interceptor.podSecurityContext | object | `{}` |  |
| interceptor.resources | object | `{"limits":{"cpu":"70m","memory":"60Mi"},"requests":{"cpu":"50m","memory":"40Mi"}}` | The resource limits and requests for the Tekton Interceptor |
| interceptor.securityContext.allowPrivilegeEscalation | bool | `false` |  |
| interceptor.securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| interceptor.securityContext.readOnlyRootFilesystem | bool | `true` |  |
| interceptor.securityContext.runAsGroup | int | `65532` |  |
| interceptor.securityContext.runAsNonRoot | bool | `true` |  |
| interceptor.securityContext.runAsUser | int | `65532` |  |
| interceptor.serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| interceptor.serviceAccount.name | string | `""` | If not set, a name is generated using the fullname template |
| interceptor.tolerations | list | `[]` | Toleration labels for pod assignment |
| kaniko.customCert | bool | `false` | Save cert in secret "custom-ca-certificates" with key ca.crt |
| kaniko.image.repository | string | `"gcr.io/kaniko-project/executor"` |  |
| kaniko.image.tag | string | `"v1.12.1"` |  |
| kaniko.roleArn | string | `""` | AWS IAM role to be used for kaniko pod service account (IRSA). Format: arn:aws:iam::<AWS_ACCOUNT_ID>:role/<AWS_IAM_ROLE_NAME> |
| nameOverride | string | `""` |  |
| tekton-cache.enabled | bool | `true` | Enables the Tekton-cache subchart. |
| tekton-cache.url | string | `"http://tekton-cache:8080"` | Defines the URL to the tekton-cache. Default: http://tekton-cache:8080 |
| tekton.configs.gradleConfigMap | string | `"custom-gradle-settings"` | Default configuration maps for provisioning init.gradle file, REPOSITORY_SNAPSHOTS_PATH and REPOSITORY_RELEASES_PATH environment variables. |
| tekton.configs.mavenConfigMap | string | `"custom-maven-settings"` | Default configuration map for provisioning Maven settings.xml file. To use custom Maven settings.xml configuration file, the user should prepare another configuration map and update "mavenConfigMap". For reference see https://github.com/epam/edp-tekton/blob/master/charts/pipelines-library/templates/resources/cm-maven-settings.yaml |
| tekton.configs.npmConfigMap | string | `"custom-npm-settings"` | Default configuration maps for provisioning NPM .npmrc files. To use custom NPM .npmrc configuration file, the user should prepare another configuration map and update "npmConfigMap". For reference see https://github.com/epam/edp-tekton/blob/master/charts/pipelines-library/templates/resources/cm-npm-settings.yaml |
| tekton.configs.nugetConfigMap | string | `"custom-nuget-settings"` | Default configuration maps for provisioning nuget.config file. |
| tekton.configs.pythonConfigMap | string | `"custom-python-settings"` | Default configuration maps for provisioning PIP_TRUSTED_HOST, PIP_INDEX_PATH, PIP_INDEX_URL_PATH, REPOSITORY_SNAPSHOTS_PATH and REPOSITORY_RELEASES_PATH environment variables for Python tasks. |
| tekton.packageRegistriesSecret.enabled | bool | `false` | Set this as `true` if the secret should be available in Pipelines |
| tekton.packageRegistriesSecret.name | string | `"package-registries-auth-secret"` | Secret name that will be used in Pipelines. Default: package-registries-auth-secret |
| tekton.pruner.create | bool | `true` | Specifies whether a cronjob should be created |
| tekton.pruner.image | string | `"bitnami/kubectl:1.25"` | Docker image to run the pruner, expected to have kubectl and jq |
| tekton.pruner.recentMinutes_pods | string | `"720"` | Controls the retention period (in minutes) for Tekton pipeline run history. Pipeline runs older than this duration are automatically deleted. |
| tekton.pruner.recentMinutes_pvcs | string | `"0"` | Specifies the retention period (in minutes) for pvcs related to Tekton pipeline runs. A non-zero value aids in debugging by preventing automatic deletion of pvcs once pipeline runs complete. Default: 0 (automatic deletion enabled). |
| tekton.pruner.resources | object | `{"limits":{"cpu":"100m","memory":"70Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}` | Pod resources for Tekton pruner job |
| tekton.pruner.schedule | string | `"0 * * * *"` | How often to clean up resources |
| tekton.resources | object | `{"limits":{"cpu":"2","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | The resource limits and requests for the Tekton Tasks |
| tekton.workspaceSize | string | `"5Gi"` | Tekton workspace size. Most cases 1Gi is enough. It's common for all pipelines |
