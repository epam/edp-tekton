# tekton-reporter

![Version: 0.25.0-SNAPSHOT](https://img.shields.io/badge/Version-0.25.0--SNAPSHOT-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.25.0-SNAPSHOT](https://img.shields.io/badge/AppVersion-0.25.0--SNAPSHOT-informational?style=flat-square)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/epmdedp)](https://artifacthub.io/packages/search?repo=epmdedp)

A Helm chart for KubeRocketCI Tekton Reporter that publishes review PipelineRun results as pull request comments

## Additional Information

The Tekton Reporter is a Kubernetes controller that watches finished review PipelineRuns and publishes their result
as a pull request comment: a per-task status table plus the trailing logs of any failed step. It supports GitHub,
GitLab and Bitbucket. It is packaged as a separate, independently toggleable deployment unit (`enabled`) that shares
the `edp-tekton` container image.

> **Security note:** published logs originate from pipeline steps that execute the pull request's own code. Secret
> values are masked on a best-effort basis only. Enable this chart only for pipelines whose secret-bearing steps do
> not execute untrusted code, or whose injected secrets have a blast radius limited to the build itself.

**Homepage:** <https://docs.kuberocketci.io>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| epmd-edp | <SupportEPMD-EDP@epam.com> | <https://solutionshub.epam.com/solution/kuberocketci> |
| sergk |  | <https://github.com/SergK> |

## Source Code

* <https://github.com/epam/edp-tekton>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings for pod assignment |
| clusterName | string | `""` | Cluster name used to construct the krci-portal pipeline URL (/c/<clusterName>/...). Must match krci-portal configEnv.DEFAULT_CLUSTER_NAME. If left empty, falls back to the first segment of global.dnsWildCard. |
| commentStrategy | string | `"update"` | Report comment strategy: 'update' edits the previous report comment of the same pull request, 'new' always creates a new comment |
| extraVolumeMounts | list | `[]` | Additional volume mounts, e.g. mount a private CA into /etc/ssl/certs to trust it |
| extraVolumes | list | `[]` | Additional volumes, e.g. a ConfigMap with a private CA certificate for git servers with self-signed TLS |
| global.dnsWildCard | string | `""` | Wildcard DNS used to build the default portal host |
| global.platform | string | `"kubernetes"` | Platform type: "kubernetes" or "openshift" |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"epamedp/edp-tekton"` |  |
| image.tag | string | `nil` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `"tekton-reporter"` |  |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| portalHost | string | `""` | Portal host used to build pipeline links in reporter pull request comments. Set this when the krci-portal ingress uses a custom host. If left empty, defaults to the krci-portal chart's default host krci-portal-<namespace>.<global.dnsWildCard>. |
| resources | object | `{"limits":{"cpu":"100m","memory":"128Mi"},"requests":{"cpu":"50m","memory":"64Mi"}}` | The resource limits and requests for the Tekton Reporter |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsGroup | int | `65532` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `65532` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.name | string | `""` | If not set, a name is generated using the fullname template |
| tailLines | int | `100` | Number of trailing log lines published for every failed step |
| tolerations | list | `[]` | Toleration labels for pod assignment |
