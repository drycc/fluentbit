{{/* Generate fluentbit envs */}}
{{- define "fluentbit.envs" }}
env:
- name: NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
{{- end }}