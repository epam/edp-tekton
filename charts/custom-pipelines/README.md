# edp-custom-pipelines

![Version: 0.6.0-SNAPSHOT](https://img.shields.io/badge/Version-0.6.0--SNAPSHOT-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.6.0-SNAPSHOT](https://img.shields.io/badge/AppVersion-0.6.0--SNAPSHOT-informational?style=flat-square)
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
| file://../common-library | edp-tekton-common-library | 0.2.11 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| fullnameOverride | string | `""` |  |
| global.dnsWildCard | string | `""` | a cluster DNS wildcard name |
| global.gerritSSHPort | string | `"30003"` | Gerrit SSH node port |
| global.gitProvider | string | `"gerrit"` | Define Git Provider to be used in Pipelines. Can be gerrit (default), gitlab, github |
| global.platform | string | `"kubernetes"` | platform type that can be "kubernetes" or "openshift" |
| imageScanReport.enable | bool | `true` |  |
| imageScanReport.grypeTemplate | string | `"<?xml version=\"1.0\" ?>\n<testsuites name=\"grype\">\n{{- $failures := len $.Matches }}\n    <testsuite tests=\"{{ $failures }}\" failures=\"{{ $failures }}\" name=\"{{ $.Distro.Name }}:{{ $.Distro.Version }}\" errors=\"0\" skipped=\"0\">\n        <properties>\n            <property name=\"type\" value=\"{{ $.Distro.Name }}\"></property>\n        </properties>\n        {{- range .Matches }}\n        <testcase classname=\"{{ .Artifact.Name }}-{{ .Artifact.Version }} ({{ .Artifact.Type }})\" name=\"[{{ .Vulnerability.Severity }}] {{ .Vulnerability.ID }}\">\n            <failure message=\"{{ .Artifact.Name }}: {{ .Vulnerability.ID }}\" type=\"description\">{{ .Vulnerability.Description }} {{ .Artifact.CPEs }} {{ .Vulnerability.DataSource }}</failure>\n        </testcase>\n        {{- end }}\n    </testsuite>\n</testsuites>\n"` |  |
| imageScanReport.trivyTemplate | string | `"<?xml version=\"1.0\" ?>\n<testsuites name=\"trivy\">\n{{- range . -}}\n{{- $failures := len .Vulnerabilities }}\n    <testsuite tests=\"{{ $failures }}\" failures=\"{{ $failures }}\" name=\"{{  .Target }}\" errors=\"0\" skipped=\"0\">\n    {{- if not (eq .Type \"\") }}\n        <properties>\n            <property name=\"type\" value=\"{{ .Type }}\"></property>\n        </properties>\n        {{- end -}}\n        {{ range .Vulnerabilities }}\n        <testcase classname=\"{{ .PkgName }}-{{ .InstalledVersion }}\" name=\"[{{ .Vulnerability.Severity }}] {{ .VulnerabilityID }}\">\n            <failure message=\"{{ escapeXML .Title }}\" type=\"description\">{{ escapeXML .Description }}</failure>\n        </testcase>\n    {{- end }}\n    </testsuite>\n{{- $failures := len .Misconfigurations }}\n{{- if gt $failures 0 }}\n    <testsuite tests=\"{{ $failures }}\" failures=\"{{ $failures }}\" name=\"{{  .Target }}\" errors=\"0\" skipped=\"0\">\n    {{- if not (eq .Type \"\") }}\n        <properties>\n            <property name=\"type\" value=\"{{ .Type }}\"></property>\n        </properties>\n        {{- end -}}\n        {{ range .Misconfigurations }}\n        <testcase classname=\"{{ .Type }}\" name=\"[{{ .Severity }}] {{ .ID }}\">\n            <failure message=\"{{ escapeXML .Title }}\" type=\"description\">{{ escapeXML .Description }}</failure>\n        </testcase>\n    {{- end }}\n    </testsuite>\n{{- end }}\n{{- end }}\n</testsuites>\n"` |  |
| nameOverride | string | `""` |  |
| tekton.resources | object | `{"limits":{"cpu":"2","memory":"3Gi"},"requests":{"cpu":"0.5","memory":"2Gi"}}` | The resource limits and requests for the Tekton Tasks |
