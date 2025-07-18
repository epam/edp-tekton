# edp-tekton

![Version: 0.19.0-SNAPSHOT](https://img.shields.io/badge/Version-0.19.0--SNAPSHOT-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.19.0-SNAPSHOT](https://img.shields.io/badge/AppVersion-0.19.0--SNAPSHOT-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart for KubeRocketCI Tekton Pipelines

## Additional Information

## Tekton Pipelines

Tekton Pipelines supports four VCS: Gerrit, GitHub, GitLab and BitBucket. To check the VCS Import strategy, please refer to the [KubeRocketCI Documentation](https://docs.kuberocketci.io)).

Tekton Pipelines are implemented and packaged using the [helm-chart](./charts/pipelines-library/) approach. The helm-chart contains:

- `Tasks` - basic building block for Tekton. Some of the tasks are forks from [Upstream Tekton Catalog](https://github.com/tektoncd/catalog).
- `Pipelines`, which consist of `Tasks` and implement logic for the CI flow. KubeRocketCI follows the below approach for pipelines definition:
  - Each type of VCS has its own Pipelines, e.g. for Gerrit, GitHub, GitLab and BitBucket;
  - KubeRocketCI has [two types of Pipelines](https://docs.kuberocketci.io/docs/operator-guide/ci/tekton-overview): `CodeReview` - triggers on Review, `Build` - triggers on Merged Event.
- `Triggers`, `TriggerBindings`, `TriggerTemplates` - defines the logic for specific VCS Events (Gerrit, GitHub, GitLab, BitBucket) and Pipelines.
- `Resources` - Kubernetes resources, that are used from Pipelines, e.g. `ServiceAccount` with [IRSA Enablement](https://docs.kuberocketci.io/docs/developer-guide/aws-reference-architecture#iam-roles-for-service-accounts-irsa), `ConfigMaps` for Maven/Gradle Pipelines, Tekton cache, CodeNarc, CTLint, and PVC to share resources between Tasks.
- `Tekton Pipeline pruner` - created as a cron job, it is designed to clear outdated pipelines.

### EDP Interceptor

EDP Interceptor is used as a component that provides KubeRocketCI metadata for Tekton Pipelines. The code is based on [Upstream implementation](https://github.com/tektoncd/triggers/tree/main/pkg/interceptors).

EDP Interceptor extracts information from VCS payload, like `repository_name`. The `repository_name` has 1-2-1 mapping with `Codebase` (kind: Codebase; apiVersion:v2.edp.epam.com/v1). Interceptor populates Tekton Pipelines with [Codebase SPEC](https://github.com/epam/edp-codebase-operator/blob/master/docs/api.md#codebasespec) data, see the diagram below:

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
The docker images for Interceptor are available on the [DockerHub](https://hub.docker.com/repository/docker/epamedp/edp-tekton).
The helm-chart for interceptor deployment is in the same repository by the [charts/interceptor](./charts/interceptor) directory.
Follows [Tekton Interceptor](https://tekton.dev/vault/triggers-main/clusterinterceptors/) paradigm and enriches payload from different Version Control Systems (VCS) like Gerrit, GitHub, GitLab or BitBucket with KubeRocketCI specific data.

**Homepage:** <https://docs.kuberocketci.io>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| epmd-edp | <SupportEPMD-EDP@epam.com> | <https://solutionshub.epam.com/solution/kuberocketci> |
| sergk |  | <https://github.com/SergK> |

## Source Code

* <https://github.com/epam/edp-tekton>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| @epamedp | tekton-cache | 0.4.2 |
| file://../common-library | edp-tekton-common-library | 0.3.15 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| ctLint.chartSchema | string | `"name: str()\nhome: str()\nversion: str()\ntype: str()\napiVersion: str()\nappVersion: any(str(), num())\ndescription: str()\nkeywords: list(str(), required=False)\nsources: list(str(), required=True)\nmaintainers: list(include('maintainer'), required=True)\ndependencies: list(include('dependency'), required=False)\nicon: str(required=False)\nengine: str(required=False)\ncondition: str(required=False)\ntags: str(required=False)\ndeprecated: bool(required=False)\nkubeVersion: str(required=False)\nannotations: map(str(), str(), required=False)\n---\nmaintainer:\n  name: str(required=True)\n  email: str(required=False)\n  url: str(required=False)\n---\ndependency:\n  name: str()\n  version: str()\n  repository: str()\n  condition: str(required=False)\n  tags: list(str(), required=False)\n  enabled: bool(required=False)\n  import-values: any(list(str()), list(include('import-value')), required=False)\n  alias: str(required=False)\n"` |  |
| ctLint.lintconf | string | `"---\nrules:\n  braces:\n    min-spaces-inside: 0\n    max-spaces-inside: 0\n    min-spaces-inside-empty: -1\n    max-spaces-inside-empty: -1\n  brackets:\n    min-spaces-inside: 0\n    max-spaces-inside: 0\n    min-spaces-inside-empty: -1\n    max-spaces-inside-empty: -1\n  colons:\n    max-spaces-before: 0\n    max-spaces-after: 1\n  commas:\n    max-spaces-before: 0\n    min-spaces-after: 1\n    max-spaces-after: 1\n  comments:\n    require-starting-space: true\n    min-spaces-from-content: 2\n  document-end: disable\n  document-start: disable           # No --- to start a file\n  empty-lines:\n    max: 2\n    max-start: 0\n    max-end: 0\n  hyphens:\n    max-spaces-after: 1\n  indentation:\n    spaces: consistent\n    indent-sequences: whatever      # - list indentation will handle both indentation and without\n    check-multi-line-strings: false\n  key-duplicates: enable\n  line-length: disable              # Lines can be any length\n  new-line-at-end-of-file: enable\n  new-lines:\n    type: unix\n  trailing-spaces: enable\n  truthy:\n    level: warning\n"` |  |
| ctLint.validateMaintainers | bool | `false` |  |
| fullnameOverride | string | `""` |  |
| gitServers | object | `{}` |  |
| githubOwners | object | `{"checkType":"all","enabled":true}` | Enabling this feature ensures that Tekton pipelines trigger only when the repo owner creates a PR. More information: https://tekton.dev/docs/triggers/interceptors/#owners-validation-for-pull-requests |
| global.dnsWildCard | string | `""` | a cluster DNS wildcard name |
| global.dockerRegistry.type | string | `"ecr"` | Define Image Registry that will to be used in Pipelines. Can be ecr (default), harbor, dockerhub |
| global.dockerRegistry.url | string | `"<AWS_ACCOUNT_ID>.dkr.ecr.<AWS_REGION>.amazonaws.com/<registry_space>"` | Docker Registry endpoint. In dockerhub case the URL must be specified in accordance with the Kaniko name convention (docker.io/<registry_space>) |
| global.gerritHost | string | `"gerrit"` | Gerrit Host URL, must be specified if gerrit is enabled |
| global.gitProviders | list | `["bitbucket","gerrit","github","gitlab"]` | Deploy Kubernetes Resources for the specific Git Provider. Can be gerrit, gitlab, github (default) |
| global.platform | string | `"kubernetes"` | platform type that can be "kubernetes" or "openshift" |
| grafana.dashboards.labelKey | string | `"grafana_dashboard"` |  |
| grafana.dashboards.labelValue | string | `"1"` |  |
| grafana.enabled | bool | `false` |  |
| grafana.serviceMonitor.prometheusReleaseLabels.release | string | `"prom"` |  |
| interceptor.affinity | object | `{}` | Affinity settings for pod assignment |
| interceptor.enabled | bool | `true` | Deploy KubeRocketCI interceptor as a part of pipeline library when true. Default: true |
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
| pipelines.deployableResources | object | `{"autotests":true,"c":{"cmake":true,"make":true},"cs":{"dotnet3.1":false,"dotnet6.0":false},"deploy":true,"docker":true,"gitops":true,"go":{"beego":true,"gin":true,"operatorsdk":true},"groovy":true,"helm":true,"helm-pipeline":true,"infrastructure":true,"java":{"java11":true,"java17":true,"java21":true,"java8":false},"js":{"angular":true,"antora":true,"express":true,"next":true,"react":true,"vue":true},"opa":false,"python":{"ansible":true,"fastapi":true,"flask":true,"python3.8":false},"security":true,"tasks":true,"terraform":true}` | This section contains the list of pipelines and tasks that will be installed. |
| pipelines.deployableResources.c | object | `{"cmake":true,"make":true}` | This section control the installation of the review and build pipelines. |
| pipelines.deployableResources.deploy | bool | `true` | This flag control the installation of the Deploy pipelines. |
| pipelines.deployableResources.tasks | bool | `true` | This flag control the installation of the tasks. |
| pipelines.image.registry | string | `"docker.io"` | Registry for tekton pipelines images. Default: docker.io |
| pipelines.imagePullSecrets | list | `[]` | List of image pull secrets used by the Tekton ServiceAccount for pulling images from private registries. Example: imagePullSecrets:   - name: regcred |
| pipelines.podTemplate | list | `[]` | This section allows to determine on which nodes to run tekton pipelines |
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
| tekton.pruner.imagePullSecrets | list | `[]` | List of ImagePullSecrets to be used by the pruner CronJob |
| tekton.pruner.resources | object | `{"limits":{"cpu":"100m","memory":"70Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}` | Pod resources for Tekton pruner job |
| tekton.pruner.schedule | string | `"0 10 */1 * *"` | How often to clean up resources |
| tekton.resources | object | `{"limits":{"cpu":"2","memory":"3Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | The resource limits and requests for the Tekton Tasks |
| tekton.workspaceSize | string | `"5Gi"` | Tekton workspace size. Most cases 1Gi is enough. It's common for all pipelines |
