#!/bin/bash

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
echo "pwd is " . `pwd`

# Run the build script in docker, using the current working directory as the docker src
# directory. Packaged binaries will be placed in out/
docker run --rm \
       -v `pwd`:/src \
       golang:"$go_version"-trixie \
       /src/scripts/build_binaries.sh "$os-$arch"
