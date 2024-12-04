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
Symphony Service Name
*/}}
{{- define "symphony.serviceName" -}}
{{- printf "%s-service" (include "symphony.fullname" .) }}
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
Symphony Api Container Http Port
*/}}
{{- define "symphony.apiContainerPortHttp" -}}
{{- default 8080 .Values.api.apiContainerPortHttp }}
{{- end }}

{{/*
Symphony Api Container Https Port
*/}}
{{- define "symphony.apiContainerPortHttps" -}}
{{- default 8081 .Values.api.apiContainerPortHttps }}
{{- end }}

{{/*
Symphony certificate duration time
*/}}
{{- define "symphony.certDurationTime" -}}
{{- default "2160h" .Values.cert.certDurationTime }}
{{- end }}

{{/*
Symphony certificate renew before time
*/}}
{{- define "symphony.certRenewBeforeTime" -}}
{{- default "360h" .Values.cert.certRenewBeforeTime }}
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

{{/*
Trace Middleware
*/}}
{{- define "symphony.traceMiddleware" -}}
{{- if .Values.otlpTracesEndpointGrpc }}
{{ tpl (.Files.Get "files/trace-middleware.json") .  }},
{{- end }}
{{- end }}

{{/*
Metric Middleware
*/}}
{{- define "symphony.metricMiddleware" -}}
{{- if .Values.otlpMetricsEndpointGrpc }}
{{ tpl (.Files.Get "files/metric-middleware.json") .  }},
{{- end }}
{{- end }}

{{/*
Log Middleware
*/}}
{{- define "symphony.logMiddleware" -}}
{{- if .Values.otlpLogsEndpointGrpc }}
{{ tpl (.Files.Get "files/log-middleware.json") .  }},
{{- end }}
{{- end }}

{{/*
Symphony API serving certs directory path
*/}}
{{- define "symphony.apiServingCertsDir" -}}
{{- printf "/etc/%s-api/tls" (include "symphony.fullname" .) }}
{{- end }}

{{/*
Symphony API serving certificate path
*/}}
{{- define "symphony.apiServingCert" -}}
{{- printf "%s/%s" (include "symphony.apiServingCertsDir" .) "tls.crt" }}
{{- end }}

{{/*
Symphony API serving certificate key path
*/}}
{{- define "symphony.apiServingKey" -}}
{{- printf "%s/%s" (include "symphony.apiServingCertsDir" .) "tls.key" }}
{{- end }}

{{/*
Symphony API serving certificate Name
*/}}
{{- define "symphony.apiServingCertName" -}}
{{ printf "%s%s" (include "symphony.fullname" .) "-api-serving-cert"}}
{{- end }}


{{/*
Symphony API serving certificate CA path
*/}}
{{- define "symphony.apiServingCA" -}}
{{- printf "%s/%s" (include "symphony.apiServingCertsDir" .) "ca.crt" }}
{{- end }}

{{/*
Symphony API ServingCertIssuerName
*/}}
{{- define "symphony.apiServingCertIssuerName" -}}
{{- printf "%s%s" (include "symphony.fullname" .) "-selfsigned-issuer"}}
{{- end }}

{{/*
Symphony API CAIssuerName
*/}}
{{- define "symphony.apiCAIssuerName" -}}
{{- printf "%s%s" (include "symphony.fullname" .) "-ca-issuer"}}
{{- end }}

{{/*
Symphony API trust bundle
*/}}
{{- define "symphony.apiClientCATrustBundle" -}}
{{- printf "%s%s" (include "symphony.fullname" .) "clientca-bundle"}}
{{- end }}

{{/*
Symphony API trust bundle key
*/}}
{{- define "symphony.apiClientCATrustBundleKey" -}}
{{- printf "%s%s" (include "symphony.fullname" .) "-clientca-key"}}
{{- end }}

{{/*
Symphony API client CA dir
*/}}
{{- define "symphony.apiClientCAMountPath" -}}
{{- printf "/etc/%s-api/clientca" (include "symphony.fullname" .) }}
{{- end }}

{{/*
Symphony API client CA path
*/}}
{{- define "symphony.apiClientCAPem" -}}
{{- printf "%s/%s" (include "symphony.apiClientCAMountPath" .) (include "symphony.apiClientCATrustBundleKey" .) }}
{{- end }}


{{/*
Symphony full url Endpoint
*/}}
{{- define "symphony.httpsUrl" -}}
{{- printf "https://%s:%s/v1alpha2/" (include "symphony.serviceName" .)  (include "symphony.apiContainerPortHttps" .) }}
{{- end }}

{{/*
Symphony full url Endpoint
*/}}
{{- define "symphony.httpUrl" -}}
{{- printf "http://%s:%s/v1alpha2/" (include "symphony.serviceName" .)  (include "symphony.apiContainerPortHttp" .) }}
{{- end }}

{{/* Symphony Env Config Name */}}
{{- define "symphony.envConfigName" -}}
{{- printf "%s-env-config" (include "symphony.fullname" .) }}
{{- end }}

{{/* Symphony Redis host*/}}
{{- define "symphony.redisHost" -}}
{{- if .Values.redis.asSidecar }}
{{- printf "localhost:%d" (.Values.redis.port | int) }}
{{- else }}
{{- printf "%s-redis:%d" (include "symphony.name" .) (.Values.redis.port | int)}}
{{- end }}
{{- end }}

{{/* Symphony Redis protected-mode*/}}
{{- define "symphony.protectedMode" -}}
{{- if .Values.redis.asSidecar}}
{{- printf "yes" }}
{{- else }}
{{- printf "no" }}
{{- end }}
{{- end }}

{{/* Symphony Emit Time Field in User Logs */}}
{{- define "symphony.emitTimeFieldInUserLogs" -}}
{{- default "false" .Values.observability.log.emitTimeFieldInUserLogs | quote }}
{{- end }}

{{- define "RedisPVCStorageClassName" -}}
{{- $pvcName := "redis-pvc" -}}
{{- $existingPVC := (lookup "v1" "PersistentVolumeClaim" .Release.Namespace $pvcName) -}}
{{- if .Values.redis.persistentVolume.storageClass }}
{{- $storageClass := .Values.redis.persistentVolume.storageClass }}
{{- $sc := lookup "storage.k8s.io/v1" "StorageClass" "" $storageClass }}
{{- if not $sc }}
{{- fail (printf "Error: StorageClass '%s' not found. Please create it before installing." $storageClass)}}
{{- end  }}
{{- .Values.redis.persistentVolume.storageClass -}}
{{- else if $existingPVC  }}
{{- $existingPVC.spec.storageClassName -}}
{{- else }}
{{- $defaultStorageClass := "" -}}
{{- range $sc := (lookup "storage.k8s.io/v1" "StorageClass" "" "").items -}}
{{- if (hasKey $sc.metadata.annotations "storageclass.kubernetes.io/is-default-class") -}}
{{- $annotations := $sc.metadata.annotations -}}
{{- $labelValue := index $annotations "storageclass.kubernetes.io/is-default-class" -}}
{{- if eq $labelValue "true" -}}
{{- $defaultStorageClass = $sc.metadata.name -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- if eq $defaultStorageClass "" -}}
{{- fail (printf "Error: No default storage class found. Please ensure a storage class with the label 'is-default-class' set to 'true' exists.")}}
{{- end -}}
{{- $defaultStorageClass -}}
{{- end -}}
{{- end -}}

{{- define "CheckRedisPvSetting" -}}
{{- $configMap := (lookup "v1" "ConfigMap" .Release.Namespace "redis-config-map") -}}
{{- if not $configMap }}
true
{{- else if eq ($configMap.data.pvEnabled | quote) ""}}
true
{{- else if ne ($configMap.data.pvEnabled | quote) (.Values.redis.persistentVolume.enabled | quote)}}
{{- fail (printf ".Values.redis.persistentVolume.enabled is immutable. Unable to change %s to %s" ($configMap.data.pvEnabled | quote) (.Values.redis.persistentVolume.enabled | quote))}}
{{- else}}
true
{{- end -}}
{{- end -}}

{{/*Observability.OtelCollector.caBundleLabelValue */}}
{{- define "symphony.otelcollector.caBundleLabelValue" -}}
{{- default "false" .Values.observability.otelCollector.caBundleLabelValue }}
{{- end }}

{{/* Otel collector's CA trust bundle label value */}}
{{- define "symphony.tls.caBundleLabelValue" -}}
{{- default "false" .Values.observability.tls.caBundleLabelValue }}
{{- end }}
