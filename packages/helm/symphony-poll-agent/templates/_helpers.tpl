{{/*
Expand the name of the chart.
*/}}
{{- define "symphony.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "symphony.fullname" -}}
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
{{- define "symphony.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "symphony.labels" -}}
helm.sh/chart: {{ include "symphony.chart" . }}
{{ include "symphony.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "symphony.selectorLabels" -}}
app.kubernetes.io/name: {{ include "symphony.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "symphony.serviceAccountName" -}}
{{- printf "%s-api-sp" (include "symphony.fullname" .) }}
{{- end }}

{{/*
Configmap Name
*/}}
{{- define "symphony.configmapName" -}}
{{- printf "%s-api-config" (include "symphony.fullname" .) }}
{{- end }}

{{/*
App Selector
*/}}
{{- define "symphony.appSelector" -}}
{{- printf "%s-api" (include "symphony.fullname" .)  }}
{{- end }}

{{/*
Zipkin Middleware
*/}}
{{- define "symphony.zipkinMiddleware" -}}
{{- if .Values.observability.tracing.exporter.zipkin }}
{{ tpl (.Files.Get "files/zipkin-middleware.json") .  }},
{{- end }}
{{- end }}
