#!/bin/bash

set -e
set -o pipefail

script_dir=$(dirname "$0")
cd $script_dir/..

[ ! -z "$GO_BUILD_VERSION" ] || (echo "Must supply GO_BUILD_VERSION"; exit 1)

OS_ARCH_TUPLES="windows-amd64 linux-amd64 linux-arm64 darwin-amd64 darwin-arm64"

docker run --rm \
       -v `pwd`:/src \
       golang:"$go_version"-trixie \
       /src/scripts/build_binaries.sh "$OS_ARCH_TUPLES"
