#!/bin/bash
set -o errexit -o pipefail -o nounset

$(dirname $0)/../scripts/build.sh

OS=$(docker info --format '{{.OSType}}')
pack package-buildpack multi-os-cnb:${OS} --config package-${OS}.yml

if [[ "$OS" == "windows" ]]; then
BUILDER=cnbs/sample-builder:nanoserver-1809
else 
BUILDER=paketobuildpacks/builder:tiny
fi

pack build multi-os-test:${OS} \
  --buildpack docker://multi-os-cnb:${OS} \
  --builder $BUILDER \
  --path $(dirname $0)/../integration/testdata/app \
  --env "BP_GO_TARGETS=./cmd/app" \
  --trust-builder

docker run -i --rm multi-os-test:${OS}
