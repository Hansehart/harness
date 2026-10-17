# shellcheck shell=bash
# Runs the tools in pinned images, isolated from the host and the files.

# Runs one tool against the repository.
in_image() {
    docker run --rm -v "$PWD":/src:ro -w /src "$@"
}

# Runs a Go tool with caches that survive between runs.
in_go() {
    in_image -v harness-cache:/cache -e GOTOOLCHAIN=local -e GOFLAGS=-buildvcs=false \
        -e GOCACHE=/cache/go-build -e GOMODCACHE=/cache/go-mod -e GOLANGCI_LINT_CACHE=/cache/golangci-lint "$@"
}
