{{- define "calendar.name" -}}
calendar
{{- end -}}

{{- define "calendar.fullname" -}}
{{ include "calendar.name" . }}-{{ .Release.Name }}
{{- end -}}
