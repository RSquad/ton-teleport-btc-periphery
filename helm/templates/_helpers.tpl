{{- define "app.name" -}}
{{ .Chart.Name | replace "-" "" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "app.fullname" -}}
{{ .Release.Name }}-{{ include "app.name" . }}
{{- end }}
