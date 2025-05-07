{{/* Generate fluentbit envs */}}
{{- define "fluentbit.envs" }}
env:
{{- if (.Values.valkeyUrl) }}
- name: DRYCC_VALKEY_URL
  valueFrom:
    secretKeyRef:
      name: logger-fluentbit-creds
      key: valkey-url
{{- else if .Values.valkey.enabled }}
- name: VALKEY_PASSWORD
  valueFrom:
    secretKeyRef:
      name: valkey-creds
      key: password
- name: DRYCC_VALKEY_URL
  value: "redis://:$(VALKEY_PASSWORD)@drycc-valkey.{{.Release.Namespace}}.svc.{{.Values.global.clusterDomain}}:16379/2"
{{- end }}
{{- end }}