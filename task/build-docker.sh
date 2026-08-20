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
go env GOOS GOARCH CGO_ENABLED CC CXX

echo "===== C COMPILER ====="
which gcc || true
gcc --version || true

ldflags="-w -s -linkmode external -extldflags=-static"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildVers=$buildvers"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildTime=$buildtime"

echo "===== GO BUILD ====="

set +e

go build \
  -o /go/bin/app \
  -v \
  -tags="jsoniter prod full" \
  -buildvcs=false \
  -trimpath \
  -ldflags="$ldflags" \
  "$wd"

status=$?

echo "===== GO BUILD EXIT CODE: $status ====="

if [ "$status" -ne 0 ]; then
    echo "===== GO BUILD FAILED ====="
    exit "$status"
fi

echo "===== BUILD SUCCESS ====="
ls -lh /go/bin/app
