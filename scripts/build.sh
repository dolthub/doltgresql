#!/bin/bash

# build.sh
#
# This script builds doltgres from source for this machine's OS and architecture.
# Requires a locally running docker server.

set -e
set -o pipefail

script_dir=$(dirname "$0")
cd $script_dir/..

go_version=`go version | cut -d" " -f 3 | sed -e 's|go||' | sed -e 's|\.[0-9]$||'`
os=`go version | cut -d" " -f 4 | sed "s|/.*||"`
arch=`go version | cut -d" " -f 4 | sed "s|.*/||"`

echo "os is $os"
echo "arch is $arch"
echo "go version is $go_version"

# Run the build script in docker, using the current working directory as the docker src
# directory. Packaged binaries will be placed in out/
docker run --rm \
       -v `pwd`:/src \
       golang:"$go_version"-trixie \
       /src/scripts/build_binaries.sh "$os-$arch"
