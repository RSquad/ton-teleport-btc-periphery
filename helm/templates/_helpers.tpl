{{- define "app.name" -}}
{{ .Values.app.name | default .Chart.Name | replace "-" "" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "app.fullname" -}}
{{ .Release.Name }}-{{ include "app.name" . }}
{{- end }}
