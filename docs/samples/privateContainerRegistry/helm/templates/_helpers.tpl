{{/*
Expand the name of the chart.
*/}}
{{- define "privateContainerRegistry.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

