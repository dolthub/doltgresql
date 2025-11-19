#!/opt/homebrew/bin/bash
#!/bin/bash

set -e
set -o pipefail

script_dir=$(dirname "$0")
cd $script_dir/..

GO_BUILD_VERSION=1.25

os=`go version | cut -d" " -f 4 | sed "s|/.*||"`
arch=`go version | cut -d" " -f 4 | sed "s|.*/||"`

echo "os is $os"
echo "arch is $arch"

docker run --rm -v `pwd`:/src golang:"$GO_BUILD_VERSION"-trixie /bin/bash -c '
set -e
set -o pipefail
set -x

OS=$1
ARCH=$2

echo "os is $OS"
echo "arch is $ARCH"

apt-get update && apt-get install -y p7zip-full pigz curl xz-utils mingw-w64 clang-19

cd /
curl -o optcross.tar.xz https://dolthub-tools.s3.us-west-2.amazonaws.com/optcross/"$(uname -m)"-linux_20250327_0.0.3_trixie.tar.xz
tar Jxf optcross.tar.xz
curl -o icustatic.tar.xz https://dolthub-tools.s3.us-west-2.amazonaws.com/icustatic/20250327_0.0.3_trixie.tar.xz
tar Jxf icustatic.tar.xz
export PATH=/opt/cross/bin:"$PATH"

cd /src

declare -A platform_cc
platform_cc["linux-arm64"]="aarch64-linux-musl-gcc"
platform_cc["linux-amd64"]="x86_64-linux-musl-gcc"
platform_cc["darwin-arm64"]="clang-19 --target=aarch64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0"
platform_cc["darwin-amd64"]="clang-19 --target=x86_64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0"
platform_cc["windows-amd64"]="x86_64-w64-mingw32-gcc"

declare -A platform_cxx
platform_cxx["linux-arm64"]="aarch64-linux-musl-g++"
platform_cxx["linux-amd64"]="x86_64-linux-musl-g++"
platform_cxx["darwin-arm64"]="clang++-19 --target=aarch64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0 --stdlib=libc++"
platform_cxx["darwin-amd64"]="clang++-19 --target=x86_64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0 --stdlib=libc++"
platform_cxx["windows-amd64"]="x86_64-w64-mingw32-g++"

declare -A platform_as
platform_as["linux-arm64"]="aarch64-linux-musl-as"
platform_as["linux-amd64"]="x86_64-linux-musl-as"
platform_as["darwin-arm64"]="clang-19 --target=aarch64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0"
platform_as["darwin-amd64"]="clang-19 --target=x86_64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0"
platform_as["windows-amd64"]="x86_64-w64-mingw32-as"

declare -A platform_go_ldflags
platform_go_ldflags["linux-arm64"]="-linkmode external -s -w"
platform_go_ldflags["linux-amd64"]="-linkmode external -s -w"
platform_go_ldflags["darwin-arm64"]="-s -w -compressdwarf=false -extldflags -Wl,-platform_version,macos,12.0,14.4"
platform_go_ldflags["darwin-amd64"]="-s -w -compressdwarf=false -extldflags -Wl,-platform_version,macos,12.0,14.4"
platform_go_ldflags["windows-amd64"]="-s -w"

declare -A platform_cgo_ldflags
platform_cgo_ldflags["linux-arm64"]="-static -s"
platform_cgo_ldflags["linux-amd64"]="-static -s"
platform_cgo_ldflags["darwin-arm64"]=""
platform_cgo_ldflags["darwin-amd64"]=""

# Stack smash protection lib is built into clang for unix platforms,
# but on Windows we need to pull in the separate ssp library
platform_cgo_ldflags["windows-amd64"]="-static-libgcc -static-libstdc++ -Wl,-Bstatic -lssp"

tuple="$OS-$ARCH"
o="out/doltgresql-$OS-$ARCH"

mkdir -p "$o/bin"
mkdir -p "$o/licenses"
cp -r ./licenses "$o/licenses"
cp LICENSE "$o/licenses"

echo Building "$o/$bin"
obin="$bin"
if [ "$OS" = windows ]; then
    obin="$bin.exe"
fi
CGO_ENABLED=1 \
    GOOS="$OS" \
    GOARCH="$ARCH" \
    CC="${platform_cc[${tuple}]}" \
    CXX="${platform_cxx[${tuple}]}" \
    AS="${platform_as[${tuple}]}" \
    CGO_LDFLAGS="${platform_cgo_ldflags[${tuple}]}" \
    go build -buildvcs=false -trimpath -ldflags="${platform_go_ldflags[${tuple}]}" -tags icu_static -o "$o/bin/$obin" ./cmd/doltgres

' " " $os $arch
