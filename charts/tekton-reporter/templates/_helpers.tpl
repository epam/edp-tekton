{{/*
Expand the name of the chart.
*/}}
{{- define "tekton-reporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tekton-reporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "tekton-reporter.labels" -}}
helm.sh/chart: {{ include "tekton-reporter.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tekton-reporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tekton-reporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "tekton-reporter.serviceAccountName" -}}
{{- default (include "tekton-reporter.name" .) .Values.serviceAccount.name }}
{{- end }}

{{/*
Portal host. Defaults to the krci-portal chart's default ingress host
(krci-portal-<namespace>.<global.dnsWildCard>); override .Values.portalHost when
the portal ingress uses a custom host.
*/}}
{{- define "tekton-reporter.portalHost" -}}
{{- .Values.portalHost | default (printf "krci-portal-%s.%s" .Release.Namespace .Values.global.dnsWildCard) -}}
{{- end -}}

{{/*
Base URL of the portal PipelineRun details pages, used to build links in
pull request comments. Must match the commit-status pipeline URL scheme.
*/}}
{{- define "tekton-reporter.portalBaseUrl" -}}
https://{{ include "tekton-reporter.portalHost" . }}/c/{{ .Values.clusterName | default (.Values.global.dnsWildCard | splitList "." | first) }}/cicd/pipelineruns
{{- end -}}
