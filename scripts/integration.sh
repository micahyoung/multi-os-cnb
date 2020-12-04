#!/usr/bin/env bash
set -euo pipefail

$(dirname $0)/../scripts/build.sh

OS=$(docker info --format '{{.OSType}}')
pack package-buildpack multi-os-cnb:${OS} --config package-${OS}.yml

if [[ "$OS" == "windows" ]]; then
BUILDER=cnbs/sample-builder:nanoserver-1809
else 
BUILDER=cnbs/sample-builder:bionic
fi

pack build multi-os-test:${OS} \
  --buildpack docker://multi-os-cnb:${OS} \
  --builder $BUILDER \
  --path $(dirname $0)/../integration/testdata/app \
  --pull-policy never

docker run -i --rm multi-os-test:${OS}
