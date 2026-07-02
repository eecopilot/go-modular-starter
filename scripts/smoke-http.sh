#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"
stamp="$(date +%s)"
email="smoke-${stamp}-$$@example.com"
username="smoke${stamp}$$"

curl -fsS "${BASE_URL}/healthz" >/dev/null
curl -fsS "${BASE_URL}/readyz" >/dev/null
curl -fsS "${BASE_URL}/version" >/dev/null
curl -fsS "${BASE_URL}/" | grep -q "Go Modular Starter"

register_response="$(curl -fsS -X POST "${BASE_URL}/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${email}\",\"username\":\"${username}\",\"password\":\"password123\",\"full_name\":\"Smoke User\"}")"
printf '%s' "${register_response}" | grep -q '"token"'

login_response="$(curl -fsS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"identifier\":\"${email}\",\"password\":\"password123\"}")"
token="$(printf '%s' "${login_response}" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"
if [ -z "${token}" ]; then
  echo "missing login token" >&2
  exit 1
fi

curl -fsS "${BASE_URL}/api/v1/me" \
  -H "Authorization: Bearer ${token}" | grep -q "${email}"

curl -fsS "${BASE_URL}/api/v1/protected/example" \
  -H "Authorization: Bearer ${token}" | grep -q "${email}"

curl -fsS -X POST "${BASE_URL}/api/v1/examples" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Smoke item","description":"Docker smoke test"}' | grep -q '"Smoke item"'

curl -fsS "${BASE_URL}/api/v1/examples" | grep -q '"Smoke item"'

echo "http smoke test passed"
