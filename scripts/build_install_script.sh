#!/bin/bash

# This script generates an install.sh file with the $DOLTGRES_VERSION provided via ENV.

set -e
set -o pipefail

script_dir=$(dirname "$0")
cd $script_dir/..

sed 's|__DOLTGRES_VERSION__|'"$DOLTGRES_VERSION"'|' scripts/install.sh > out/install.sh
chmod 755 out/install.sh
