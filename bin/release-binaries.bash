#!/bin/bash

set -eu
set -o pipefail

function run() {
    local arch os output
    arch="${1:?Please provide an architecture}"
    os="${2:?Please provide an OS}"
    version="${3:?Please provide a version}"
    output="${4:?Please provide an output directory}"


    GOARCH="${arch}" GOOS="${os}" \
    go build \
    -o "${output}/grace" \

    local tarball
    tarball="grace-${os}-${arch}-${version}.tgz"

    tar -czvf "${output}/${tarball}" -C "${output}" grace
    local sha1sum=$(sha1sum "${output}/${tarball}" | cut -d' ' -f1)

    mkdir -p "${output}/docker"

cat <<HERE > "${output}/docker/grace-opsfile.yml"
---
- type: replace
  path: /instance_groups/name=vizzini/jobs/name=vizzini/properties/vizzini?/grace_tarball_url?
  value: https://storage.googleapis.com/grace-assets/${tarball}
- type: replace
  path: /instance_groups/name=vizzini/jobs/name=vizzini/properties/vizzini?/grace_tarball_checksum?
  value: ${sha1sum}
- type: replace
  path: /instance_groups/name=vizzini/jobs/name=vizzini/properties/vizzini?/grace_busybox_image_url?
  value: docker:///cloudfoundry/grace#${version}
HERE

    GOARCH="${arch}" GOOS="${os}" \
    go build \
    -o "${output}/docker/grace" \
    -tags "busybox"

    cp Dockerfile "${output}/docker/Dockerfile"
}
run "$@"
