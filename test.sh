#!/usr/bin/env bash
# Runs the Go tests inside the pinned toolchain image.
set -euo pipefail
# Works on the repository the script belongs to, wherever it is called from.
cd "$(dirname "$0")"
# shellcheck source=docker.sh
source ./docker.sh

readonly GO_IMAGE=golang:1.27.2-bookworm@sha256:5cf287a799e6b94384bad13d16b14904c531f51ba65792237e122ce42b392f61

in_go "$GO_IMAGE" go test ./... "$@"
