#!/bin/sh
# Раскатка petfamily из актуального main без потери данных.
# Манифесты deploy/k8s/ не содержат Secret-ов — apply их не трогает.
# Образы pin-ятся на SHA сборки (как раньше делал petfamily_deploy.sh),
# потому что в манифестах стоит :latest.
# Запуск на VPS с доступом к кластеру:
#   ./deploy/apply.sh <API_SHA> [ADMIN_SHA]
set -eu

API_SHA="${1:?нужен SHA образа api: apply.sh <API_SHA> [ADMIN_SHA]}"
ADMIN_SHA="${2:-$API_SHA}"
NS="petfamily"

kubectl apply -f deploy/k8s/
kubectl -n "$NS" set image deploy/petfamily-api "api=ghcr.io/alexzrn/petfamily-api:${API_SHA}"
kubectl -n "$NS" set image deploy/petfamily-admin "api=ghcr.io/alexzrn/petfamily-admin:${ADMIN_SHA}"
kubectl -n "$NS" rollout status deploy/petfamily-api --timeout=180s
kubectl -n "$NS" rollout status deploy/petfamily-admin --timeout=180s
kubectl -n "$NS" get deploy petfamily-api petfamily-admin -o custom-columns=NAME:.metadata.name,IMAGE:.spec.template.spec.containers[0].image,READY:.status.readyReplicas
