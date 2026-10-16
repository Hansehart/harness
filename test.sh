#!/usr/bin/env bash
# Runs the Go tests inside the pinned toolchain image.
set -euo pipefail

readonly GO_IMAGE=golang:1.27.2-bookworm@sha256:5cf287a799e6b94384bad13d16b14904c531f51ba65792237e122ce42b392f61

docker run --rm --user "$(id -u):$(id -g)" -e GOCACHE=/tmp/go-cache -e GOMODCACHE=/tmp/go-mod \
    -e GOTOOLCHAIN=local -v "$PWD":/src -w /src "$GO_IMAGE" go test ./... "$@"
