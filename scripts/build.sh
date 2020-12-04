#!/usr/bin/env bash

set -euo pipefail

GOOS="linux"   GOARCH="amd64" go build -ldflags='-s -w' -o bin/main -tags osusergo ./cmd/ &
GOOS="windows" GOARCH="amd64" go build -ldflags='-s -w' -o bin/main.exe            ./cmd/ &

wait

ln -fs main bin/build
ln -fs main bin/detect
ln -fs main.exe bin/build.exe
ln -fs main.exe bin/detect.exe
