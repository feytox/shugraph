{{- define "app.selectorLabels" -}}
app.kubernetes.io/name: {{ .Release.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "app.labels" -}}
{{ include "app.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}

{{- define "app.image" -}}
{{ required "image.repository is required" .Values.image.repository }}:{{ required "image.tag is required" .Values.image.tag }}
{{- end }}

{{- define "app.env" -}}
{{- range $name, $ref := .Values.secretEnv }}
- name: {{ $name }}
  valueFrom:
    secretKeyRef:
      name: {{ $ref.secret }}
      key: {{ $ref.key }}
{{- end }}
{{- range $name, $value := .Values.env }}
- name: {{ $name }}
  value: {{ $value | quote }}
{{- end }}
{{- end }}
