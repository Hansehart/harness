#!/usr/bin/env bash
# Lints the Go code and the shell scripts inside pinned images.
set -euo pipefail

readonly GOLANGCI_IMAGE=golangci/golangci-lint:v2.15.0@sha256:fbfcfd12a6f8d7f1b97b3362ca80fc2a244d46863148c73cb88e7793af96c481
readonly SHELLCHECK_IMAGE=koalaman/shellcheck:v0.11.0@sha256:61862eba1fcf09a484ebcc6feea46f1782532571a34ed51fedf90dd25f925a8d

docker run --rm --user "$(id -u):$(id -g)" -e GOCACHE=/tmp/go-cache -e GOMODCACHE=/tmp/go-mod \
    -e GOLANGCI_LINT_CACHE=/tmp/lint-cache -e GOTOOLCHAIN=local -v "$PWD":/src -w /src \
    "$GOLANGCI_IMAGE" golangci-lint run ./...
docker run --rm -v "$PWD":/src -w /src "$SHELLCHECK_IMAGE" ./*.sh
