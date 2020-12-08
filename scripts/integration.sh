#!/bin/bash
set -o errexit -o pipefail -o nounset

$(dirname $0)/../scripts/build.sh

OS=$(docker info --format '{{.OSType}}')
pack package-buildpack multi-os-cnb-${OS} --config package-${OS}.yml

if [[ "$OS" == "windows" ]]; then
STACK_ID=io.buildpacks.samples.stacks.dotnet-framework-1809
BUILD_IMAGE=cnbs/sample-stack-build:dotnet-framework-1809
RUN_IMAGE=cnbs/sample-stack-run:dotnet-framework-1809
else
STACK_ID=io.buildpacks.samples.stacks.bionic
BUILD_IMAGE=cnbs/sample-stack-build:bionic
RUN_IMAGE=cnbs/sample-stack-run:bionic
fi

pack create-builder multi-os-builder-${OS} --config <(cat <<EOF
[[buildpacks]]
uri = "docker://multi-os-cnb-${OS}"

[[order]]
[[order.group]]
id = "multi-os"
version = "0.0.1"

[stack]
id = "${STACK_ID}"
run-image = "${RUN_IMAGE}"
build-image = "${BUILD_IMAGE}"
EOF
)

pack build multi-os-test:${OS} \
  --builder multi-os-builder-${OS} \
  --path $(dirname $0)/../integration/testdata/app \
  --env "BP_GO_TARGETS=./cmd/app" \
  --trust-builder

docker run -i --rm multi-os-test:${OS}
