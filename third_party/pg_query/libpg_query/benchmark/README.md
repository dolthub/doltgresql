# Protobuf encode/decode benchmarks

Microbenchmarks for the protobuf serialization layer
(`pg_query_nodes_to_protobuf` / `pg_query_protobuf_to_nodes`, exercised by
`pg_query_parse_protobuf` / `pg_query_deparse_protobuf`).

## Running

```sh
make benchmark            # builds libpg_query.a and runs both harnesses
./benchmark/bench_protobuf      [iters] [label]   # default iters: 20000
./benchmark/microbench_protobuf [iters] [label]   # default iters: 50000
```

- **`bench_protobuf`** — end-to-end via the public API: `pg_query_parse_protobuf`
  (parse + encode) and `pg_query_deparse_protobuf` (decode + deparse). Reports
  ns/op, wire bytes, and peak RSS.
- **`microbench_protobuf`** — isolates *only* the protobuf layer by calling the
  internal encode/decode on pre-parsed trees / pre-encoded bytes, with a scratch
  memory context reset after each op (so per-node `palloc` churn and arena
  lifecycle are both counted, while the parse trees persist). Removes the shared
  parse/deparse cost from the measurement.

Both use the same small mixed-complexity query corpus. Point them at your own
corpus by editing the `queries[]` array.

## A/B comparing two implementations (e.g. before/after a change)

The benchmarks measure whatever `libpg_query.a` they link against. To compare a
change against a baseline ref, build the baseline in an isolated worktree and
run the same harness against each archive:

```sh
# baseline (e.g. the commit before the change)
git worktree add /tmp/libpg_base <ref>
make -C /tmp/libpg_base build

# build the harness against each archive (microbench needs the internal headers)
cc -O2 -w -I. -Isrc -Isrc/include -Isrc/postgres/include -Ivendor \
   benchmark/microbench_protobuf.c libpg_query.a -pthread -o /tmp/micro_new
cc -O2 -w -I/tmp/libpg_base -I/tmp/libpg_base/src -I/tmp/libpg_base/src/include \
   -I/tmp/libpg_base/src/postgres/include -I/tmp/libpg_base/vendor \
   benchmark/microbench_protobuf.c /tmp/libpg_base/libpg_query.a -pthread -o /tmp/micro_base

/tmp/micro_base 50000 baseline
/tmp/micro_new  50000 new
git worktree remove --force /tmp/libpg_base
```

Because the parser and deparse walk are identical across builds, the per-op
*difference* isolates the protobuf-layer change.
