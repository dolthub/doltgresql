#!/usr/bin/env bash
# Records what Dolt does in Doltlite's oracle tests, which run the same version control scenarios on Doltlite and Dolt:
# record.sh <doltlite checkout> <doltlite binary> <dolt binary> <record file> [test name...]
# Each test is a test/vc_oracle_*_test.sh script of the checkout, run one at a time with record.py standing in for
# dolt, which appends each of Dolt's sessions to the record file for `dolt-oracle generate`. Doltlite's own results
# are not used. The tests' databases go under ORACLE_WORK_ROOT (the system's temporary directory by default), and the
# run stops before a test when that disk has less than ORACLE_MIN_FREE_GB (10 by default) free.
set -u
here=$(cd "$(dirname "$0")" && pwd)
checkout=$(cd "$1" && pwd)
doltlite=$(cd "$(dirname "$2")" && pwd)/$(basename "$2")
export DOLT_REAL=$(cd "$(dirname "$3")" && pwd)/$(basename "$3")
export ORACLE_RECORD_LOG=$(cd "$(dirname "$4")" && pwd)/$(basename "$4")
shift 4
work=$(mktemp -d "${ORACLE_WORK_ROOT:-${TMPDIR:-/tmp}}/oracle.XXXXXX")
trap 'rm -rf "$work"' EXIT
cp -R "$checkout/test" "$work/test"
"$DOLT_REAL" version | awk '$1=="dolt" && $2=="version" {print "v" $3; exit}' > "$work/.dolt-oracle-version"
tests=("$@")
if [ ${#tests[@]} -eq 0 ]; then
  tests=($(cd "$work/test" && ls vc_oracle_*_test.sh | sed 's/_test\.sh$//'))
fi
for test in "${tests[@]}"; do
  free_kb=$(df -Pk "$work" | awk 'NR == 2 { print $4 }')
  if [ "$free_kb" -lt $(( ${ORACLE_MIN_FREE_GB:-10} * 1024 * 1024 )) ]; then
    echo "stopping: less than ${ORACLE_MIN_FREE_GB:-10}GB free under $work"
    exit 2
  fi
  export ORACLE_SUITE="$test"
  (cd "$work" && TMPDIR="$work" perl -e 'alarm shift; exec @ARGV' 1800 bash "test/${test}_test.sh" "$doltlite" "$here/record.py" > "$work/$test.log" 2>&1)
  echo "$test rc=$? sessions=$(grep -c "\"suite\": \"$test\"" "$ORACLE_RECORD_LOG")"
  rm -rf "$work/tmp."*
done
