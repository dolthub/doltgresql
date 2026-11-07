#!/bin/bash

# build_spatial_libraries.sh
#
# Builds the static GEOS and PROJ libraries that the PostGIS emulation links, and places them under
# third_party/spatial. PROJ is built with its database embedded in the library, and with SQLite compiled into it, since
# PROJ cannot read its database without SQLite. Nothing else links SQLite.
#
# Releases link the musl, macOS, and msvcrt Windows libraries using the same Debian release and cross toolchains as
# build_binaries.sh. Builds on a Linux machine link the glibc libraries instead, which are built on an older Debian
# release so that they link against any glibc from 2.33 onward. Builds on a Windows machine link the UCRT libraries,
# which match the MSYS2 UCRT64 toolchain. Run both from the repository root:
#
#   docker run --rm --platform linux/amd64 -v `pwd`:/src -w /src debian:trixie /src/scripts/build_spatial_libraries.sh windows-amd64 windows-amd64-msvcrt linux-amd64-musl linux-arm64-musl darwin-amd64 darwin-arm64
#   docker run --rm -v `pwd`:/src -w /src debian:bookworm /src/scripts/build_spatial_libraries.sh linux-amd64 linux-arm64

set -e
set -o pipefail

GEOS_VERSION=3.11.1
PROJ_VERSION=9.8.1
SQLITE_VERSION=3460100
SQLITE_YEAR=2024

OS_ARCH_TUPLES="$*"
OUT="$(pwd)/third_party/spatial"

declare -A platform_cc
platform_cc["linux-arm64"]="aarch64-linux-gnu-gcc"
platform_cc["linux-amd64"]="x86_64-linux-gnu-gcc"
platform_cc["linux-arm64-musl"]="aarch64-linux-musl-gcc"
platform_cc["linux-amd64-musl"]="x86_64-linux-musl-gcc"
platform_cc["darwin-arm64"]="clang-19"
platform_cc["darwin-amd64"]="clang-19"
platform_cc["windows-amd64"]="x86_64-w64-mingw32ucrt-gcc"
platform_cc["windows-amd64-msvcrt"]="x86_64-w64-mingw32-gcc"

declare -A platform_cxx
platform_cxx["linux-arm64"]="aarch64-linux-gnu-g++"
platform_cxx["linux-amd64"]="x86_64-linux-gnu-g++"
platform_cxx["linux-arm64-musl"]="aarch64-linux-musl-g++"
platform_cxx["linux-amd64-musl"]="x86_64-linux-musl-g++"
platform_cxx["darwin-arm64"]="clang++-19"
platform_cxx["darwin-amd64"]="clang++-19"
platform_cxx["windows-amd64"]="x86_64-w64-mingw32ucrt-g++"
platform_cxx["windows-amd64-msvcrt"]="x86_64-w64-mingw32-g++"

declare -A platform_flags
platform_flags["darwin-arm64"]="--target=aarch64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0"
platform_flags["darwin-amd64"]="--target=x86_64-darwin --sysroot=/opt/cross/darwin-sysroot -mmacosx-version-min=12.0"

declare -A platform_system
platform_system["linux-arm64"]="Linux"
platform_system["linux-amd64"]="Linux"
platform_system["linux-arm64-musl"]="Linux"
platform_system["linux-amd64-musl"]="Linux"
platform_system["darwin-arm64"]="Darwin"
platform_system["darwin-amd64"]="Darwin"
platform_system["windows-amd64"]="Windows"
platform_system["windows-amd64-msvcrt"]="Windows"

declare -A platform_processor
platform_processor["linux-arm64"]="aarch64"
platform_processor["linux-amd64"]="x86_64"
platform_processor["linux-arm64-musl"]="aarch64"
platform_processor["linux-amd64-musl"]="x86_64"
platform_processor["darwin-arm64"]="arm64"
platform_processor["darwin-amd64"]="x86_64"
platform_processor["windows-amd64"]="x86_64"
platform_processor["windows-amd64-msvcrt"]="x86_64"

declare -A platform_ar
platform_ar["linux-arm64"]="aarch64-linux-gnu-ar"
platform_ar["linux-amd64"]="x86_64-linux-gnu-ar"
platform_ar["linux-arm64-musl"]="aarch64-linux-musl-ar"
platform_ar["linux-amd64-musl"]="x86_64-linux-musl-ar"
platform_ar["darwin-arm64"]="llvm-ar-19"
platform_ar["darwin-amd64"]="llvm-ar-19"
platform_ar["windows-amd64"]="x86_64-w64-mingw32ucrt-ar"
platform_ar["windows-amd64-msvcrt"]="x86_64-w64-mingw32-ar"

declare -A platform_ranlib
platform_ranlib["linux-arm64"]="aarch64-linux-gnu-ranlib"
platform_ranlib["linux-amd64"]="x86_64-linux-gnu-ranlib"
platform_ranlib["linux-arm64-musl"]="aarch64-linux-musl-ranlib"
platform_ranlib["linux-amd64-musl"]="x86_64-linux-musl-ranlib"
platform_ranlib["darwin-arm64"]="llvm-ranlib-19"
platform_ranlib["darwin-amd64"]="llvm-ranlib-19"
platform_ranlib["windows-amd64"]="x86_64-w64-mingw32ucrt-ranlib"
platform_ranlib["windows-amd64-msvcrt"]="x86_64-w64-mingw32-ranlib"

# The glibc compilers come from Debian's cross packages, except for the machine's own architecture, whose compiler is
# the native one
declare -A platform_packages
platform_packages["linux-arm64"]="g++-aarch64-linux-gnu"
platform_packages["linux-amd64"]="g++-x86-64-linux-gnu"
platform_packages["linux-$(dpkg --print-architecture)"]="g++"
platform_packages["windows-amd64"]="g++-mingw-w64-ucrt64"

packages="cmake curl xz-utils sqlite3 unzip bzip2"
for tuple in $OS_ARCH_TUPLES; do
  if [ -n "${platform_packages[$tuple]}" ]; then
    packages="$packages ${platform_packages[$tuple]}"
  else
    packages="$packages mingw-w64 clang-19 llvm-19"
    use_optcross=1
  fi
done
apt-get update && apt-get install -y $packages

mkdir -p /tmp/build-deps
pushd /tmp/build-deps
if [ -n "$use_optcross" ]; then
  curl -o optcross.tar.xz https://dolthub-tools.s3.us-west-2.amazonaws.com/optcross/"$(uname -m)"-linux_20250327_0.0.3_trixie.tar.xz
  tar Jxf optcross.tar.xz -C /
  export PATH=/opt/cross/bin:"$PATH"
fi
curl -L -o geos.tar.bz2 https://download.osgeo.org/geos/geos-"$GEOS_VERSION".tar.bz2
tar jxf geos.tar.bz2
curl -L -o proj.tar.gz https://download.osgeo.org/proj/proj-"$PROJ_VERSION".tar.gz
tar zxf proj.tar.gz
curl -L -o sqlite.zip https://www.sqlite.org/"$SQLITE_YEAR"/sqlite-amalgamation-"$SQLITE_VERSION".zip
unzip -q -o sqlite.zip
popd

for tuple in $OS_ARCH_TUPLES; do
  echo "Building $tuple"
  work="/tmp/build-$tuple"
  rm -rf "$work"
  mkdir -p "$work"
  cc="${platform_cc[$tuple]}"
  cxx="${platform_cxx[$tuple]}"
  # Fused multiply-adds round differently from separate operations, so they are disabled to compute the same results
  # on every platform
  flags="${platform_flags[$tuple]} -ffp-contract=off"
  cxxflags="$flags"
  if [ "${platform_system[$tuple]}" = Darwin ]; then
    cxxflags="$cxxflags --stdlib=libc++"
  fi
  cmake_args=(
    -DCMAKE_SYSTEM_NAME="${platform_system[$tuple]}"
    -DCMAKE_SYSTEM_PROCESSOR="${platform_processor[$tuple]}"
    -DCMAKE_C_COMPILER="$cc"
    -DCMAKE_CXX_COMPILER="$cxx"
    -DCMAKE_C_FLAGS="$flags"
    -DCMAKE_CXX_FLAGS="$cxxflags"
    -DCMAKE_AR="$(command -v "${platform_ar[$tuple]}")"
    -DCMAKE_RANLIB="$(command -v "${platform_ranlib[$tuple]}")"
    -DCMAKE_BUILD_TYPE=Release
    -DCMAKE_INSTALL_PREFIX="$work/install"
    -DCMAKE_INSTALL_LIBDIR=lib
    -DCMAKE_FIND_ROOT_PATH="$work/install"
    -DBUILD_SHARED_LIBS=OFF
    -DBUILD_TESTING=OFF
  )

  # SQLite, only for PROJ
  mkdir -p "$work/install/lib" "$work/install/include"
  $cc $flags -O2 -DSQLITE_THREADSAFE=1 -DSQLITE_OMIT_LOAD_EXTENSION -c /tmp/build-deps/sqlite-amalgamation-"$SQLITE_VERSION"/sqlite3.c -o "$work/sqlite3.o"
  "${platform_ar[$tuple]}" rcs "$work/install/lib/libsqlite3.a" "$work/sqlite3.o"
  cp /tmp/build-deps/sqlite-amalgamation-"$SQLITE_VERSION"/sqlite3.h "$work/install/include/"

  # GEOS 3.11 relies on <cstdint> arriving through other headers, which newer C++ libraries no longer do
  cmake -S /tmp/build-deps/geos-"$GEOS_VERSION" -B "$work/geos" "${cmake_args[@]}" -DBUILD_GEOSOP=OFF -DBUILD_BENCHMARKS=OFF \
    -DCMAKE_CXX_FLAGS="$cxxflags -include cstdint"
  cmake --build "$work/geos" --parallel "$(nproc)" --target install

  cmake -S /tmp/build-deps/proj-"$PROJ_VERSION" -B "$work/proj" "${cmake_args[@]}" \
    -DSQLite3_INCLUDE_DIR="$work/install/include" -DSQLite3_LIBRARY="$work/install/lib/libsqlite3.a" \
    -DEXE_SQLITE3="$(command -v sqlite3)" -DENABLE_TIFF=OFF -DENABLE_CURL=OFF -DBUILD_APPS=OFF \
    -DEMBED_RESOURCE_FILES=ON -DUSE_ONLY_EMBEDDED_RESOURCE_FILES=ON
  cmake --build "$work/proj" --parallel "$(nproc)" --target install

  mkdir -p "$OUT/lib/$tuple" "$OUT/include/geos"
  for lib in libgeos_c.a libgeos.a libproj.a libsqlite3.a; do
    cp "$work/install/lib/$lib" "$OUT/lib/$tuple/$lib"
  done
  cp "$work/install/include/geos_c.h" "$work/install/include/geodesic.h" "$work/install/include/"proj*.h "$OUT/include/"
  cp "$work/install/include/geos/export.h" "$OUT/include/geos/"
done
