#!/usr/bin/env bash

set -eu

# keep version in sync with k8s.io/api and friends in go.mod
codegeneratorversion=v0.23.5
codegeneratorrepo=target/code-generator-${codegeneratorversion}
generategroupsbin=${codegeneratorrepo}/generate-groups.sh

if [ ! -d "$codegeneratorrepo" ]; then
  mkdir -p $codegeneratorrepo
  git clone --depth 1 --branch $codegeneratorversion https://github.com/kubernetes/code-generator $codegeneratorrepo
  chmod +x "${generategroupsbin}"
fi

# ROOT_PACKAGE :: the package that is the target for code generation
ROOT_PACKAGE=github.com/buoyantio/linkerd-buoyant

crds=(operator.buoyant.io:v1alpha1)

# run the code-generator entrypoint script
GOPATH='' "${generategroupsbin}" \
  'all' \
  "${ROOT_PACKAGE}/operator/generated" \
  "${ROOT_PACKAGE}/operator/apis" \
  "${crds[*]}" \
  --go-header-file "${codegeneratorrepo}/hack/boilerplate.go.txt"

# copy generated code out of GOPATH
cp -R "${ROOT_PACKAGE}/operator" .
rm -rf ./$ROOT_PACKAGE
