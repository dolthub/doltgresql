#!/bin/bash

set -ex
set -o pipefail

root="`dirname \"$0\"`"
dir="`realpath \"$root\"`"
doltgres_root="`realpath \"$dir/../../\"`"

version=""
sed_cmd_i=""
start_marker=""
end_marker=""

dest_file="$doltgres_root/README.md"
os_type="darwin"

start_template='<!-- START_%s_RESULTS_TABLE -->'
end_template='<!-- END_%s_RESULTS_TABLE -->'

if [ "$#" -ne 3 ]; then
  echo "usage: update-perf.sh <version> <perf-path> <correctness-path>"
  exit 1;
fi

version="$1"
new_perf="$2"
new_correctness="$3"

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  os_type="linux"
fi


# update the version
if [ "$os_type" == "linux" ]
then
  sed -i 's/Here are the benchmarks for DoltgreSQL version `'".*"'`/Here are the benchmarks for DoltgreSQL version `'"$version"'`/' "$dest_file"
  sed -i 's/Here are DoltgreSQL'"'"'s sqllogictest results for version `'".*"'`./Here are DoltgreSQL'"'"'s sqllogictest results for version `'"$version"'`./' "$dest_file"
else
  sed -i '' "s/Here are the benchmarks for DoltgreSQL version \\\`.*\\\`/Here are the benchmarks for DoltgreSQL version \\\`$version\\\`/" "$dest_file"
  sed -i '' 's/Here are DoltgreSQL'"'"'s sqllogictest results for version `'".*"'`./Here are DoltgreSQL'"'"'s sqllogictest results for version `'"$version"'`./' "$dest_file"
fi

perf_start_marker=$(printf "$start_template" "LATENCY")
perf_end_marker=$(printf "$end_template" "LATENCY")

correctness_start_marker=$(printf "$start_template" "CORRECTNESS")
correctness_end_marker=$(printf "$end_template" "CORRECTNESS")

# store in variable
updated_perf_with_markers=$(printf "$perf_start_marker\n$(cat $new_perf)\n$perf_end_marker\n")
updated_correctness_with_markers=$(printf "$correctness_start_marker\n$(cat $new_correctness)\n$correctness_end_marker\n")

echo "$updated_perf_with_markers" > "$new_perf"
echo "$updated_correctness_with_markers" > "$new_correctness"

if [ "$os_type" == "linux" ]
then
  sed -i '/<!-- END_LATENCY/r '"$new_perf"'' -e '/<!-- START_LATENCY/,/<!-- END_LATENCY/d' "$dest_file"
  sed -i '/<!-- END_CORRECTNESS/r '"$new_correctness"'' -e '/<!-- START_CORRECTNESS/,/<!-- END_CORRECTNESS/d'  "$dest_file"
else
  sed -i '/\<!-- END_LATENCY/r '"$new_perf"'' -e '/\<!-- START_LATENCY/,/\<!-- END_LATENCY/d' "$dest_file"
  sed -i '/\<!-- END_CORRECTNESS/r '"$new_correctness"'' -e '/\<!-- START_CORRECTNESS/,/\<!-- END_CORRECTNESS/d' "$dest_file"
fi
