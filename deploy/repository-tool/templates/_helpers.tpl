{{- define "repository-tool.name" -}}
{{- .Chart.Name }}
{{- end }}

{{- define "repository-tool.fullname" -}}
{{- printf "%s" .Release.Name }}
{{- end }}

{{- define "repository-tool.labels" -}}
app.kubernetes.io/name: {{ include "repository-tool.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}

{{- define "repository-tool.selectorLabels" -}}
app.kubernetes.io/name: {{ include "repository-tool.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
