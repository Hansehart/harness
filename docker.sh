# shellcheck shell=bash
# Runs commands in pinned images with the repository mounted read-only at /src.

# Runs a command in an image.
in_image() {
    docker run --rm -v "$PWD":/src:ro -w /src "$@"
}

# Runs a command in a Go toolchain image with a build cache volume kept between runs.
in_go() {
    in_image -v harness-cache:/cache -e GOTOOLCHAIN=local -e GOFLAGS=-buildvcs=false \
        -e GOCACHE=/cache/go-build -e GOMODCACHE=/cache/go-mod -e GOLANGCI_LINT_CACHE=/cache/golangci-lint "$@"
}
