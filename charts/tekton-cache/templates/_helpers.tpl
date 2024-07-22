{{/*
Expand the name of the chart.
*/}}
{{- define "tekton-cache.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tekton-cache.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "tekton-cache.labels" -}}
helm.sh/chart: {{ include "tekton-cache.chart" . }}
{{ include "tekton-cache.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tekton-cache.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tekton-cache.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
