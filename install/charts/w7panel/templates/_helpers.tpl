{{/*
Expand the name of the chart.
*/}}
{{- define "helm.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "helm.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" $name .Release.Name  | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "helm.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "helm.labels" -}}
helm.sh/chart: {{ include "helm.chart" . }}
w7.cc/release-name: {{ .Release.Name }}
w7.cc/name: {{ include "helm.fullname" . }}
w7.cc/identifie: {{ .Chart.Name | quote }}
{{ include "helm.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "helm.annotations" -}}
w7.cc/title : "微擎面板"
title : "微擎面板"
w7.cc/description : "微擎面板"
w7.cc/icon : "https://zpk.w7.cc/zip/icon/w7panel_offline"
w7.cc/zpk-url: "https://zpk.w7.cc/respo/info/w7panel_offline"
w7.cc/identifie: {{ .Chart.Name | quote }}

{{- end }}

{{- define "helm.servicelb.annotations" -}}
{{- if .Values.servicelb.create -}}
w7.cc.app/ports: '{"8000": {{ .Values.servicelb.port | quote }}}'
{{- end -}}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "helm.selectorLabels" -}}
app: {{ include "helm.fullname" . }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "helm.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "helm.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
