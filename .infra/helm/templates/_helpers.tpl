{{- define "lil-poker.labels" -}}
app.kubernetes.io/name: lil-poker
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}
