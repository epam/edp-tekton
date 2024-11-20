{{/*
Expand the name of the chart.
*/}}
{{- define "edp-tekton.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "edp-tekton.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "edp-tekton.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "edp-tekton.labels" -}}
helm.sh/chart: {{ include "edp-tekton.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{/*
Validate values of gitProviders
*/}}
{{- define "edp-tekton.validateGitProviders" -}}
{{- $allowedProviders := list "github" "gitlab" "gerrit" "bitbucket" -}}
{{- range .Values.global.gitProviders }}
  {{- if not (has . $allowedProviders) }}
    {{- printf "Error: Invalid gitProvider %s. The gitProvider must be one of: %s" . (join ", " $allowedProviders) | fail }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Define registry for pipelines images
*/}}
{{- define "edp-tekton.registry" -}}
{{- .Values.pipelines.image.registry -}}
{{- end -}}

# Mapping for Java Maven (application and library) pipeliens
{{- define "edp-tekton.resourceMapping.maven" -}}
{{- $registry := .Values.pipelines.image.registry -}}
{{- $mavenVersions := dict -}}
{{- with .Values.pipelines.deployableResources.java -}}
  {{- if .java8 }}
    {{- $mavenVersions = set $mavenVersions "java8" (printf "%s/maven:3.9.0-eclipse-temurin-8" $registry)  }}
  {{- end }}
  {{- if .java11 }}
    {{- $mavenVersions = set $mavenVersions "java11" (printf "%s/maven:3.9.0-eclipse-temurin-11" $registry)  }}
  {{- end }}
  {{- if .java17 }}
    {{- $mavenVersions = set $mavenVersions "java17" (printf "%s/maven:3.9.0-eclipse-temurin-17" $registry)  }}
  {{- end }}
{{- end }}
{{- $mavenVersions | toYaml -}}
{{- end }}

{{- define "edp-tekton.resourceMapping.mavenSonar" -}}
{{- $registry := .Values.pipelines.image.registry -}}
{{- $sonarVersions := dict -}}
{{- with .Values.pipelines.deployableResources.java -}}
  {{- if .java8 }}
    {{- $sonarVersions = set $sonarVersions "java8" (printf "%s/maven:3.9.0-eclipse-temurin-11" $registry) }}
  {{- end }}
  {{- if .java11 }}
    {{- $sonarVersions = set $sonarVersions "java11" (printf "%s/maven:3.9.0-eclipse-temurin-11" $registry) }}
  {{- end }}
  {{- if .java17 }}
    {{- $sonarVersions = set $sonarVersions "java17" (printf "%s/maven:3.9.0-eclipse-temurin-17" $registry) }}
  {{- end }}
{{- end }}
{{- $sonarVersions | toYaml -}}
{{- end }}

# Mapping for Java Gradle (application and library) pipeliens
{{- define "edp-tekton.resourceMapping.gradle" -}}
{{- $registry := .Values.pipelines.image.registry -}}
{{- $gradleVersions := dict -}}
{{- with .Values.pipelines.deployableResources.java -}}
  {{- if .java8 }}
    {{- $gradleVersions = set $gradleVersions "java8" (printf "%s/gradle:7.5.1-jdk8" $registry)  }}
  {{- end }}
  {{- if .java11 }}
    {{- $gradleVersions = set $gradleVersions "java11" (printf "%s/gradle:7.5.1-jdk11" $registry)  }}
  {{- end }}
  {{- if .java17 }}
    {{- $gradleVersions = set $gradleVersions "java17" (printf "%s/gradle:7.5.1-jdk17" $registry)  }}
  {{- end }}
{{- end }}
{{- $gradleVersions | toYaml -}}
{{- end }}

{{- define "edp-tekton.resourceMapping.gradleSonar" -}}
{{- $registry := .Values.pipelines.image.registry -}}
{{- $sonarVersions := dict -}}
{{- with .Values.pipelines.deployableResources.java -}}
  {{- if .java8 }}
    {{- $sonarVersions = set $sonarVersions "java8" (printf "%s/gradle:7.5.1-jdk11" $registry) }}
  {{- end }}
  {{- if .java11 }}
    {{- $sonarVersions = set $sonarVersions "java11" (printf "%sgradle:7.5.1-jdk11" $registry) }}
  {{- end }}
  {{- if .java17 }}
    {{- $sonarVersions = set $sonarVersions "java17" (printf "%s/gradle:7.5.1-jdk17" $registry) }}
  {{- end }}
{{- end }}
{{- $sonarVersions | toYaml -}}
{{- end }}


# Mapping for Go pipelines
{{- define "edp-tekton.resourceMapping.go" -}}
{{- $go := list -}}
{{- with .Values.pipelines.deployableResources.go -}}
  {{- if .beego }}
    {{- $go = append $go "beego" }}
  {{- end }}
  {{- if .gin }}
    {{- $go = append $go "gin" }}
  {{- end }}
  {{- if .operatorsdk }}
    {{- $go = append $go "operator-sdk" }}
  {{- end }}
{{- end }}
{{- $go  -}}
{{- end }}

# Mapping for JS pipelines
{{- define "edp-tekton.resourceMapping.js" -}}
{{- $js := list -}}
{{- with .Values.pipelines.deployableResources.js -}}
  {{- if .vue }}
    {{- $js = append $js "vue" }}
  {{- end }}
  {{- if .angular }}
    {{- $js = append $js "angular" }}
  {{- end }}
  {{- if .express }}
    {{- $js = append $js "express" }}
  {{- end }}
  {{- if .next }}
    {{- $js = append $js "next" }}
  {{- end }}
  {{- if .react }}
    {{- $js = append $js "react" }}
  {{- end }}
{{- end }}
{{- $js  -}}
{{- end }}

# Mapping for Pyhton pipelines
{{- define "edp-tekton.resourceMapping.python" -}}
{{- $python := list -}}
{{- with .Values.pipelines.deployableResources.python -}}
  {{- if .fastapi }}
    {{- $python = append $python "fastapi" }}
  {{- end }}
  {{- if .flask }}
    {{- $python = append $python "flask" }}
  {{- end }}
{{- end }}
{{- $python  -}}
{{- end }}


# Mapping for Csharp pipeliens
{{- define "edp-tekton.resourceMapping.cs" -}}
{{- $registry := .Values.pipelines.image.registry -}}
{{- $csVersions := dict -}}
{{- with .Values.pipelines.deployableResources.cs -}}
  {{- if (index . "dotnet3.1" )}}
    {{- $csVersions = set $csVersions "dotnet-3.1" "mcr.microsoft.com/dotnet/sdk:3.1.423-alpine3.16"  }}
  {{- end }}
  {{- if (index . "dotnet6.0" ) }}
    {{- $csVersions = set $csVersions "dotnet-6.0" "mcr.microsoft.com/dotnet/sdk:6.0.407-alpine3.17"  }}
  {{- end }}
{{- end }}
{{- $csVersions | toYaml -}}
{{- end }}
