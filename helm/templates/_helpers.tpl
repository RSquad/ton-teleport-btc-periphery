{{- define "app.name" -}}
{{ .Values.app.name | default .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}