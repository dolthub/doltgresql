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
  echo "Must supply version and type, eg update-perf.sh 'v0.39.0' 'latency|correctness' '/path/to/file'"
  exit 1;
fi

version="$1"
type="$2"
new_table="$3"

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  os_type="linux"
fi


if [ "$type" == "latency" ]
then
  # update the version
  if [ "$os_type" == "linux" ]
  then
    sed -i 's/Here are the benchmarks for DoltgreSQL version `'".*"'`/Here are the benchmarks for DoltgreSQL version `'"$version"'`/' "$dest_file"
  else
    sed -i '' "s/Here are the benchmarks for DoltgreSQL version \\\`.*\\\`/Here are the benchmarks for DoltgreSQL version \\\`$version\\\`/" "$dest_file"
  fi

  start_marker=$(printf "$start_template" "LATENCY")
  end_marker=$(printf "$end_template" "LATENCY")

else
  # update the version
  if [ "$os_type" == "linux" ]
  then
    sed -i 's/Here are DoltgreSQL'"'"'s sqllogictest results for version `'".*"'`./Here are DoltgreSQL'"'"'s sqllogictest results for version `'"$version"'`./' "$dest_file"
  else
    sed -i '' 's/Here are DoltgreSQL'"'"'s sqllogictest results for version `'".*"'`./Here are DoltgreSQL'"'"'s sqllogictest results for version `'"$version"'`./' "$dest_file"
  fi

  start_marker=$(printf "$start_template" "CORRECTNESS")
  end_marker=$(printf "$end_template" "CORRECTNESS")
fi

# store in variable
updated=$(cat "$new_table")
updated_with_markers=$(printf "$start_marker\n$updated\n$end_marker\n")

echo "$updated_with_markers" > "$new_table"

if [ "$type" == "latency" ]
then
  if [ "$os_type" == "linux" ]
  then
    sed -e '/<!-- END_LATENCY/r '"$new_table"'' -e '/<!-- START_LATENCY/,/<!-- END_LATENCY/d' "$dest_file" > temp.md
  else
    sed -e '/\<!-- END_LATENCY/r '"$new_table"'' -e '/\<!-- START_LATENCY/,/\<!-- END_LATENCY/d' "$dest_file" > temp.md
  fi
else
  if [ "$os_type" == "linux" ]
  then
    sed -e '/<!-- END_CORRECTNESS/r '"$new_table"'' -e '/<!-- START_CORRECTNESS/,/<!-- END_CORRECTNESS/d' "$dest_file" > temp.md
  else
    sed -e '/\<!-- END_CORRECTNESS/r '"$new_table"'' -e '/\<!-- START_CORRECTNESS/,/\<!-- END_CORRECTNESS/d' "$dest_file" > temp.md
  fi
fi

mv temp.md "$dest_file"
