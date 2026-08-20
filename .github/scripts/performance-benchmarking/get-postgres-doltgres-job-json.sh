#!/bin/bash

set -e

if [ "$#" -lt 10 ]; then
    echo  "Usage: ./get-job-json.sh <jobname> <fromServer> <fromVersion> <toServer> <toVersion> <timeprefix> <actorprefix> <format> <issueNumber> <initBigRepo> <nomsBinFormat> <sysbenchTestTime> <withTpcc>"
    exit 1
fi

jobname="$1"
fromServer="$2"
fromVersion="$3"
toServer="$4"
toVersion="$5"
timeprefix="$6"
actorprefix="$7"
format="$8"
issueNumber="$9"
initBigRepo="${10}"
nomsBinFormat="${11}"
sysbenchTestTime="${12}"
withTpcc="${13}"
precision="1"
tpccRegex="tpcc%"

if [ -n "$initBigRepo" ]; then
  initBigRepo="\"--init-big-repo=$initBigRepo\","
fi

if [ -n "$nomsBinFormat" ]; then
  nomsBinFormat="\"--noms-bin-format=$nomsBinFormat\","
fi

if [ -n "$sysbenchTestTime" ]; then
  sysbenchTestTime="\"--sysbench-test-time=$sysbenchTestTime\","
fi

if [ -n "$withTpcc" ]; then
  withTpcc="\"--withTpcc=$withTpcc\","
fi

readTests="(
'oltp_read_only',
'oltp_point_select',
'select_random_points',
'select_random_ranges',
'covering_index_scan_postgres',
'index_scan_postgres',
'table_scan_postgres',
'groupby_scan_postgres',
'index_join_scan_postgres',
'types_table_scan_postgres',
'index_join_postgres'
)"

medianLatencyMultiplierReadsQuery="
with result(test_name, from_server, from_version, from_latency, to_server, to_version, to_latency) as (
  select
    f.test_name,
    f.server_name,
    f.server_version,
    avg(f.latency_percentile),
    t.server_name,
    t.server_version,
    avg(t.latency_percentile)
  from
    from_results as f join to_results as t
    on
      f.test_name = t.test_name
  where
    f.test_name in $readTests
  group by
    f.test_name
)
select
  test_name as read_tests,
  from_server,
  from_version,
  from_latency,
  to_server,
  to_version,
  to_latency,
  round(to_latency / (from_latency + .000001), $precision) as multiplier
from
  result;"

meanMultiplierReadsQuery="
with result(multiplier) as (
  select
    round(avg(t.latency_percentile) / (avg(f.latency_percentile) + .000001), $precision)
  from
    from_results as f join to_results as t
    on
      f.test_name = t.test_name
  where
    f.test_name in $readTests
  group by
    f.test_name
)
select
  round(avg(multiplier), $precision) as reads_mean_multiplier
from
  result;"


writeTests="(
'oltp_read_write',
'oltp_update_index',
'oltp_update_non_index',
'oltp_insert',
'oltp_write_only',
'oltp_delete_insert_postgres',
'types_delete_insert_postgres'
)"

medianLatencyMultiplierWritesQuery="
with result(test_name, from_server, from_version, from_latency, to_server, to_version, to_latency) as (
  select
    f.test_name,
    f.server_name,
    f.server_version,
    avg(f.latency_percentile),
    t.server_name,
    t.server_version,
    avg(t.latency_percentile)
  from
    from_results as f join to_results as t
    on
      f.test_name = t.test_name
  where
    f.test_name in $writeTests
  group by
    f.test_name
)
select
  test_name as write_tests,
  from_server,
  from_version,
  from_latency,
  to_server,
  to_version,
  to_latency,
  round(to_latency / (from_latency + .000001), $precision) as multiplier
from
  result;"

meanMultiplierWritesQuery="
with result(multiplier) as (
  select
    round(avg(t.latency_percentile) / (avg(f.latency_percentile) + .000001), $precision)
  from
    from_results as f join to_results as t
    on
      f.test_name = t.test_name
  where
    f.test_name in $writeTests
  group by
    f.test_name
)
select
  round(avg(multiplier), $precision) as writes_mean_multiplier
from
  result;"

meanMultiplierOverallQuery="
with result(multiplier) as (
  select
    round(avg(t.latency_percentile) / (avg(f.latency_percentile) + .000001), $precision)
  from
    from_results as f join to_results as t
    on
      f.test_name = t.test_name
  where
    f.test_name != 'bulk_insert'
  group by
    f.test_name
)
select
  round(avg(multiplier), $precision) as overall_mean_multiplier
from
  result;"

tpccLatencyQuery="
with result(test_name, from_latency, to_latency) as (
  select
    f.test_name,
    avg(f.latency_percentile),
    avg(t.latency_percentile)
  from
    from_results as f join to_results as t on f.test_name = t.test_name
  where
    f.test_name LIKE '$tpccRegex'
  group by
    f.test_name
)
select
  test_name,
  from_latency as from_latency_p95,
  to_latency as to_latency_p95,
  round(to_latency / (from_latency + .000001), $precision) as multiplier
from
  result;"

tpccTpsQuery="
with result(test_name, from_server_name, from_server_version, from_tps, to_server_name, to_server_version, to_tps) as (
  select
    f.test_name,
    f.server_name,
    f.server_version,
    avg(f.sql_transactions_per_second),
    t.server_name,
    t.server_version,
    avg(t.sql_transactions_per_second)
  from
    from_results as f join to_results as t
    on
      f.test_name = t.test_name
    where
      f.test_name LIKE '$tpccRegex'
  group by
    f.test_name
)
select
  test_name,
  from_server_name,
  from_server_version,
  from_tps,
  to_server_name,
  to_server_version,
  to_tps,
    round(to_tps / (from_tps + .000001), $precision) as percent_change
from
  result;"

# Replace new lines with spaces, so it is valid JSON later
medianLatencyMultiplierReadsQuery=${medianLatencyMultiplierReadsQuery//$'\n'/ }
meanMultiplierReadsQuery=${meanMultiplierReadsQuery//$'\n'/ }
medianLatencyMultiplierWritesQuery=${medianLatencyMultiplierWritesQuery//$'\n'/ }
meanMultiplierWritesQuery=${meanMultiplierWritesQuery//$'\n'/ }
meanMultiplierOverallQuery=${meanMultiplierOverallQuery//$'\n'/ }
tpccLatencyQuery=${tpccLatencyQuery//$'\n'/ }
tpccTpsQuery=${tpccTpsQuery//$'\n'/ }

echo '
{
  "apiVersion": "batch/v1",
  "kind": "Job",
  "metadata": {
    "name": "'$jobname'",
    "namespace": "performance-benchmarking"
  },
  "spec": {
    "backoffLimit": 3,
    "template": {
      "metadata": {
        "annotations": {
          "alert_recipients": "'$ACTOR_EMAIL'"
        },
        "labels": {
          "app": "performance-benchmarking",
          "k8s-liquidata-inc-monitored-job": "created-by-static-config"
        }
      },
      "spec": {
        "serviceAccountName": "performance-benchmarking",
        "containers": [
          {
            "name": "performance-benchmarking",
            "image": "407903926827.dkr.ecr.us-west-2.amazonaws.com/liquidata/performance-benchmarking:latest",
            "resources": {
              "limits": {
                "cpu": "7000m"
              }
            },
            "env": [
              { "name": "GOMAXPROCS", "value": "7" },
              { "name": "ACTOR", "value": "'$ACTOR'" },
              { "name": "ACTOR_EMAIL", "value": "'$ACTOR_EMAIL'" },
              { "name": "REPO_ACCESS_TOKEN", "value": "'$REPO_ACCESS_TOKEN'" }
            ],
            "imagePullPolicy": "Always",
            "args": [
              "--schema=/app/schema.sql",
              "--useDoltHubLuaScriptsRepo",
              "--repo=doltgresql",
              "--output='$format'",
              "--email-template=DoltgresPerformanceBenchmarkingReleaseTemplate",
              "--postgres-exec=/usr/lib/postgresql/15/bin/postgres",
              "--init-db-exec=/usr/lib/postgresql/15/bin/initdb",
              "--from-server='$fromServer'",
              "--from-version='$fromVersion'",
              "--to-server='$toServer'",
              "--to-version='$toVersion'",
              "--bucket=performance-benchmarking-github-actions-results",
              "--region=us-west-2",
              "--results-dir='$timeprefix'",
              "--results-prefix='$actorprefix'",
              '"$sysbenchTestTime"'
              '"$withTpcc"'
              '"$initBigRepo"'
              '"$nomsBinFormat"'
              "--sysbenchQueries='"$medianLatencyMultiplierReadsQuery"'",
              "--sysbenchQueries='"$meanMultiplierReadsQuery"'",
              "--sysbenchQueries='"$medianLatencyMultiplierWritesQuery"'",
              "--sysbenchQueries='"$meanMultiplierWritesQuery"'",
              "--sysbenchQueries='"$meanMultiplierOverallQuery"'",
              "--tpccQueries='"$tpccLatencyQuery"'",
              "--tpccQueries='"$tpccTpsQuery"'"
            ]
          }
        ],
        "restartPolicy": "Never",
        "nodeSelector": {
          "performance-benchmarking-worker": "true"
        },
        "tolerations": [
          {
              "effect": "NoSchedule",
              "key": "dedicated",
              "operator": "Equal",
              "value": "performance-benchmarking-worker"
          }
        ]
      }
    }
  }
}
'
