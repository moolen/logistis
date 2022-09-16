#!/bin/bash
set -euo pipefail

SERVICE_NAME=${1}

openssl genrsa -out ca.key 2048

openssl req -new -x509 -days 365 -key ca.key \
  -subj "/C=AU/CN=logistis"\
  -out ca.crt

openssl req -newkey rsa:2048 -nodes -keyout server.key \
  -subj "/C=AU/CN=logistis" \
  -out server.csr

openssl x509 -req \
  -extfile <(printf "subjectAltName=DNS:$SERVICE_NAME") \
  -days 365 \
  -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt

yq -i e ".tls.cert=\"$(cat ./server.crt | base64)\"" ./dev/values.dev.yaml
yq -i e ".tls.key=\"$(cat ./server.key  | base64)\"" ./dev/values.dev.yaml
yq -i e ".tls.ca=\"$(cat ./ca.crt       | base64)\"" ./dev/values.dev.yaml

rm ca.crt ca.key ca.srl server.crt server.csr server.key
