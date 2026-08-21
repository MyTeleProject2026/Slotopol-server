#!/bin/bash
set -euxo pipefail

wd="$(realpath -s "$(dirname "$0")/..")"

mkdir -p "$GOPATH/bin/config" "$GOPATH/bin/sqlite"
cp -ruv "$wd/appdata/"* "$GOPATH/bin/config"

buildvers="v0.12.0"
buildtime="$(date +'%FT%T.%3NZ')"

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

echo "===== BUILD ENVIRONMENT ====="
go version
go env GOOS GOARCH CGO_ENABLED

echo "===== GO BUILD ====="

# This line already contains ALL your slot game providers!
TAGS="jsoniter prod agt aristocrat betsoft ct igt megajack netent novomatic playngo playtech"

ldflags="-w -s"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildVers=$buildvers"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildTime=$buildtime"

# Build with error output captured
set +e
go build \
  -o /go/bin/app \
  -tags="$TAGS" \
  -buildvcs=false \
  -trimpath \
  -ldflags="$ldflags" \
  "$wd" 2>&1
status=$?
set -e

if [ $status -ne 0 ]; then
    echo "===== GO BUILD EXIT CODE: $status ====="
    echo "===== GO BUILD FAILED ====="
    exit $status
fi

echo "===== BUILD SUCCESS ====="
ls -lh /go/bin/app
