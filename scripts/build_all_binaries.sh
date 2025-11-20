#!/bin/bash

# build_all_binaries.sh
#
# This script builds doltgres from source for all supported OSes and architectures.
# Requires a locally running docker server.
#
# Used as part of automated releases. To build your current OS and architecture, use build.sh.
#
# GO_BUILD_VERSION is the major version of Go to target, e.g. 1.25. Must be set in ENV.

set -e
set -o pipefail

script_dir=$(dirname "$0")
cd $script_dir/..

[ ! -z "$GO_BUILD_VERSION" ] || (echo "Must supply GO_BUILD_VERSION"; exit 1)

OS_ARCH_TUPLES="windows-amd64 linux-amd64 linux-arm64 darwin-amd64 darwin-arm64"

docker run --rm \
       -v `pwd`:/src \
       golang:"$GO_BUILD_VERSION"-trixie \
       /src/scripts/build_binaries.sh "$OS_ARCH_TUPLES"
