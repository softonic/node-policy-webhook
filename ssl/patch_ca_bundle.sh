#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

tmpfile=$(mktemp)

export CA_BUNDLE=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')

cat manifests/mutatingwebhook-tpl.yaml| envsubst > manifests/mutatingwebhook.yaml
