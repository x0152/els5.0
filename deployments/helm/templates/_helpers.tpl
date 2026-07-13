{{/*
Expand the name of the chart.
*/}}
{{- define "els-expert.name" -}}
{{- default .Chart.Name .Values.global.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Fullname prefix for resources: <release>-<chart>, truncated to 63 chars.
Use this for any Kubernetes resource name to keep them unique per release.
*/}}
{{- define "els-expert.fullname" -}}
{{- if .Values.global.fullnameOverride -}}
{{- .Values.global.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.global.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Chart version label (for app.kubernetes.io/version).
*/}}
{{- define "els-expert.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels attached to every resource.
*/}}
{{- define "els-expert.labels" -}}
helm.sh/chart: {{ include "els-expert.chart" . }}
app.kubernetes.io/name: {{ include "els-expert.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end -}}
{{- end -}}

{{/*
Selector labels. Don't include helm.sh/chart / version here: mutating these
would invalidate selectors on upgrade.
*/}}
{{- define "els-expert.selectorLabels" -}}
app.kubernetes.io/name: {{ include "els-expert.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Per-app resource name. Usage: {{ include "els-expert.appName" (dict "ctx" . "app" $appName) }}
*/}}
{{- define "els-expert.appName" -}}
{{- printf "%s-%s" (include "els-expert.fullname" .ctx) .app | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Secrets reference used everywhere.
*/}}
{{- define "els-expert.secretName" -}}
{{- default (printf "%s-secrets" (include "els-expert.fullname" .)) .Values.global.secretName -}}
{{- end -}}

{{/*
ConfigMap reference used everywhere.
*/}}
{{- define "els-expert.configName" -}}
{{- default (printf "%s-config" (include "els-expert.fullname" .)) .Values.global.configName -}}
{{- end -}}

{{/*
Common env shared by every backend container.
Sensitive values come from Secret, the rest from ConfigMap.
*/}}
{{- define "els-expert.commonEnv" -}}
- name: TZ
  value: {{ .Values.global.timezone | default "Europe/Moscow" | quote }}
- name: APP_ENV
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: APP_ENV
- name: APP_NAME
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: APP_NAME
- name: LOG_LEVEL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: LOG_LEVEL
- name: LOG_FORMAT
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: LOG_FORMAT
- name: LOG_ADD_SOURCE
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: LOG_ADD_SOURCE
- name: POSTGRES_HOST
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: POSTGRES_HOST
- name: POSTGRES_PORT
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: POSTGRES_PORT
- name: POSTGRES_DATABASE
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: POSTGRES_DATABASE
- name: POSTGRES_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: POSTGRES_USER
- name: POSTGRES_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: POSTGRES_PASSWORD
- name: POSTGRES_SSLMODE
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: POSTGRES_SSLMODE
- name: POSTGRES_TIMEZONE
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: POSTGRES_TIMEZONE
- name: REDIS_ADDR
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: REDIS_ADDR
- name: REDIS_DB
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: REDIS_DB
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: REDIS_PASSWORD
- name: S3_ENDPOINT
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: S3_ENDPOINT
- name: S3_USE_SSL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: S3_USE_SSL
- name: S3_REGION
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: S3_REGION
- name: S3_AVATAR_BUCKET
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: S3_AVATAR_BUCKET
- name: S3_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: S3_ACCESS_KEY
- name: S3_SECRET_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: S3_SECRET_KEY
- name: MEDIA_PUBLIC_URL_BASE
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: MEDIA_PUBLIC_URL_BASE
- name: INVITE_SET_PASSWORD_URL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: INVITE_SET_PASSWORD_URL
- name: INVITE_MAGIC_LOGIN_URL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: INVITE_MAGIC_LOGIN_URL
- name: INVITE_RESET_PASSWORD_URL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: INVITE_RESET_PASSWORD_URL
- name: INVITE_MAGIC_LOGIN_PERSISTENT
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: INVITE_MAGIC_LOGIN_PERSISTENT
- name: SMTP_ENABLED
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: SMTP_ENABLED
- name: SMTP_HOST
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: SMTP_HOST
- name: SMTP_PORT
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: SMTP_PORT
- name: SMTP_FROM_EMAIL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: SMTP_FROM_EMAIL
- name: SMTP_FROM_NAME
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: SMTP_FROM_NAME
- name: SMTP_SECURE
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: SMTP_SECURE
- name: SMTP_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: SMTP_USER
- name: SMTP_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: SMTP_PASSWORD
- name: IMPERSONATION_ENABLED
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: IMPERSONATION_ENABLED
- name: CORE_WORKER_ENABLED
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: CORE_WORKER_ENABLED
- name: LLM_BASE_URL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: LLM_BASE_URL
- name: LLM_MODEL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: LLM_MODEL
- name: LLM_TIMEOUT_SECONDS
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: LLM_TIMEOUT_SECONDS
- name: LLM_API_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "els-expert.secretName" . }}
      key: LLM_API_KEY
- name: SPACY_URL
  valueFrom:
    configMapKeyRef:
      name: {{ include "els-expert.configName" . }}
      key: SPACY_URL
{{- end -}}
