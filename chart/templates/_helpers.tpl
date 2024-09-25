{{/*
Expand the name of the chart.
*/}}
{{- define "domain-egress-policy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "domain-egress-policy.fullname" -}}
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
{{- define "domain-egress-policy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "domain-egress-policy.labels" -}}
helm.sh/chart: {{ include "domain-egress-policy.chart" . }}
{{ include "domain-egress-policy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "domain-egress-policy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "domain-egress-policy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "domain-egress-policy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "domain-egress-policy.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
admissionWebhooks secret
*/}}
{{- define "domain-egress-policy.admissionWebhookSecret" -}}
{{- if .Values.admissionWebhooks.existingSecret }}
{{- default "default" .Values.admissionWebhooks.existingSecret }}
{{- else }}
{{- printf "%s-server-cert" (include "domain-egress-policy.fullname" .) }}
{{- end }}
{{- end }}
