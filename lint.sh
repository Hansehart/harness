#!/usr/bin/env bash
# Lints the Go code, shell scripts and workflows inside pinned images.
set -euo pipefail
# shellcheck source=docker.sh
source "$(dirname "$0")/docker.sh"

readonly GOLANGCI_IMAGE=golangci/golangci-lint:v2.15.0@sha256:fbfcfd12a6f8d7f1b97b3362ca80fc2a244d46863148c73cb88e7793af96c481
readonly SHELLCHECK_IMAGE=koalaman/shellcheck:v0.11.0@sha256:61862eba1fcf09a484ebcc6feea46f1782532571a34ed51fedf90dd25f925a8d
readonly ACTIONLINT_IMAGE=rhysd/actionlint:1.7.12@sha256:b1934ee5f1c509618f2508e6eb47ee0d3520686341fec936f3b79331f9315667
readonly ZIZMOR_IMAGE=ghcr.io/zizmorcore/zizmor:1.30.1@sha256:a2eb396d886c053073405c7a980f2139ba2248ec172243cfa3841e57196e8101

in_go "$GOLANGCI_IMAGE" golangci-lint run ./...
in_image "$SHELLCHECK_IMAGE" -x ./*.sh
in_image "$ACTIONLINT_IMAGE" -no-color
in_image "$ZIZMOR_IMAGE" --offline .github/workflows
