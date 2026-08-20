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

echo "===== GO BUILD ====="

# Include all game provider tags
TAGS="jsoniter prod full keno agt aristocrat betsoft ct igt megajack netent novomatic playngo playtech"

ldflags="-w -s -linkmode external -extldflags=-static"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildVers=$buildvers"
ldflags="$ldflags -X github.com/MyTeleProject2026/Slotopol-server/config.BuildTime=$buildtime"

go build \
  -o /go/bin/app \
  -v \
  -tags="$TAGS" \
  -buildvcs=false \
  -trimpath \
  -ldflags="$ldflags" \
  "$wd"

echo "===== BUILD SUCCESS ====="
ls -lh /go/bin/app
