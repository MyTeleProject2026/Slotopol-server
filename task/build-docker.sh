#!/bin/bash
set -euxo pipefail

wd="$(realpath -s "$(dirname "$0")/..")"

mkdir -p "$GOPATH/bin/config" "$GOPATH/bin/sqlite"
cp -ruv "$wd/appdata/"* "$GOPATH/bin/config"

buildvers="v0.12.0"
buildtime="$(date +'%FT%T.%3NZ')"

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1

echo "===== BUILD ENVIRONMENT ====="
go version
go env GOOS GOARCH CGO_ENABLED

echo "===== RUNNING GO VET (catch early errors) ====="
go vet ./... || true  # Continue even if vet finds issues

echo "===== GO BUILD (verbose) ====="

TAGS="jsoniter prod full keno agt aristocrat betsoft ct igt megajack netent novomatic playngo playtech"

ldflags="-w -s -linkmode external -extldflags=-static"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildVers=$buildvers"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildTime=$buildtime"

# Build with -x to see all commands, and redirect stderr to stdout
set +e
build_output=$(go build \
  -x \
  -o /go/bin/app \
  -v \
  -tags="$TAGS" \
  -buildvcs=false \
  -trimpath \
  -ldflags="$ldflags" \
  "$wd" 2>&1)
status=$?
set -e

# Print the full output (this will now include all compiler errors)
echo "$build_output"

if [ $status -ne 0 ]; then
    echo "===== GO BUILD EXIT CODE: $status ====="
    echo "===== GO BUILD FAILED ====="
    exit $status
fi

echo "===== BUILD SUCCESS ====="
ls -lh /go/bin/app
