#!/bin/bash
set -euxo pipefail

wd=$(realpath -s "$(dirname "$0")/..")

mkdir -p "$GOPATH/bin/config" "$GOPATH/bin/sqlite"

cp -ruv "$wd/appdata/"* "$GOPATH/bin/config"

buildvers="v0.12.0"
buildtime=$(date +'%FT%T.%3NZ')

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1

go version
go env GOOS GOARCH CGO_ENABLED

go build \
  -o /go/bin/app \
  -v \
  -tags="jsoniter prod full" \
  -buildvcs=false \
  -trimpath \
  -ldflags="-w -s -linkmode external -extldflags '-static' \
-X github.com/MyTeleProject2026/Slotopol-server/config.BuildVers=$buildvers \
-X github.com/MyTeleProject2026/Slotopol-server/config.BuildTime=$buildtime" \
  "$wd"
