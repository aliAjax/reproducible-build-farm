#!/usr/bin/env bash
set -euo pipefail
base="${BASE_URL:-http://127.0.0.1:8080}"
curl -fsS "$base/healthz" >/dev/null
curl -fsS -X POST "$base/api/v1/projects" -H 'content-type: application/json' -d '{"id":"smoke-project","name":"smoke","owner":"qa"}' >/dev/null
curl -fsS -X POST "$base/api/v1/build-definitions" -H 'content-type: application/json' -d '{"id":"smoke-definition","projectID":"smoke-project","DSL":{"name":"smoke","toolchain_id":"go123","steps":[{"id":"compile","args":["compile"],"outputs":["bin/app"]}]}}' >/dev/null
curl -fsS -X POST "$base/api/v1/executions" -H 'content-type: application/json' -d '{"id":"smoke-execution","definitionID":"smoke-definition","idempotencyKey":"smoke-key"}' >/dev/null
echo "smoke ok"
