{{/*
Expand the name of the chart.
*/}}
{{- define "gpustack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "gpustack.fullname" -}}
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
{{- define "gpustack.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "gpustack.labels" -}}
helm.sh/chart: {{ include "gpustack.chart" . }}
{{ include "gpustack.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "gpustack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gpustack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "gpustackWorker.selectorLabels" -}}
w7.cc/group-name: {{ .Release.Name }}
w7.cc/groupstack-worker: 'true'
w7.cc/groupstack-worker-name: {{ include "gpustack.fullname" . }}-worker
{{- end }}

{{- define "gpustackWorker.shell" -}}
SERVER_URL="http://{{ include "gpustack.fullname" . }}.default.svc.cluster.local"
while true; do
    if curl --output /dev/null --silent --fail "$SERVER_URL"; then
        echo "Server is reachable. Starting gpustack..."
        break
    else
        echo "Server is not reachable. Retrying in 5 seconds..."
        sleep 5
    fi
done
gpustack start --server-url "$SERVER_URL" --token "$RELEASE_NAME_SUFFIX"
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "gpustack.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "gpustack.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
