{{/*
Expand the name of the reporter component.
*/}}
{{- define "edp-tekton-reporter.name" -}}
{{- default .Chart.Name .Values.reporter.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "edp-tekton-reporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "edp-tekton-reporter.labels" -}}
helm.sh/chart: {{ include "edp-tekton-reporter.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "edp-tekton-reporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "edp-tekton-reporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "edp-tekton-reporter.serviceAccountName" -}}
{{- default (include "edp-tekton-reporter.name" .) .Values.reporter.serviceAccount.name }}
{{- end }}

{{/*
Portal host. Defaults to the krci-portal chart's default ingress host
(krci-portal-<namespace>.<global.dnsWildCard>); override .Values.portalHost when
the portal ingress uses a custom host.
*/}}
{{- define "edp-tekton-reporter.portalHost" -}}
{{- .Values.portalHost | default (printf "krci-portal-%s.%s" .Release.Namespace .Values.global.dnsWildCard) -}}
{{- end -}}

{{/*
Base URL of the portal PipelineRun details pages, used to build links in
pull request comments. Must match the commit-status pipeline URL scheme.
*/}}
{{- define "edp-tekton-reporter.portalBaseUrl" -}}
https://{{ include "edp-tekton-reporter.portalHost" . }}/c/{{ .Values.clusterName | default (.Values.global.dnsWildCard | splitList "." | first) }}/cicd/pipelineruns
{{- end -}}
