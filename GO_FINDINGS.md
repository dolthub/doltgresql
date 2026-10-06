# Findings in the Go implementation

Surprising behavior in Go Doltgres and Dolt, and problems that would be very hard to change there, found while
porting. They are evidence for the Rust port beyond performance. Small bugs that are easy to fix in Go don't belong
here. "Confirmed" means reproduced or traced to the Go source. "Observed" means seen but not yet isolated.

## Chunk boundaries depend on the CPU architecture (confirmed)

Dolt decides where a prolly tree node ends with a Weibull check that calls `math.Expm1`. Go's compiler fuses
multiply-adds into FMA instructions on arm64 but not on amd64, so `math.Expm1` rounds differently on the two. Over
every input the chunker can pass (`-(size/4096)^4` for sizes 0 through 16,384), 25 of 16,385 results differ by one
ulp. Found by hashing all results from a Go program built for each architecture (amd64 under Rosetta), and confirmed
in the disassembly of `math.expm1` (FMADDD, FMSUBD, and FNMSUBD instructions).

Impact: a boundary flips only when a key's hash lands within an ulp of the threshold, so it is very rare, but then the
same data has a different tree, and so a different root hash, on the two architectures. Files stay readable
everywhere, but hash comparisons across machines report differences that aren't there, and push, pull, and merge
share fewer chunks.

Hard to change in Go: the rounding comes from the compiler, and making every platform agree changes hashes for
existing data on one of them.

Rust: always rounds as Go does on amd64 (`crates/prolly/src/chunker.rs`), so it chunks identically on every platform.
Checked exhaustively against Go's amd64 results.

## Commit closures depend on their history (confirmed)

A commit closure editor adds new keys with a one-byte value (`emptyCommitClosureValue`), but closure leaf nodes store
no values, so keys re-read from an existing node have empty values. The chunker counts value lengths toward where a
node ends, so the same closure can split differently depending on which keys were added in this edit and which were
already there. Found in `store/prolly/commit_closure.go` and `store/prolly/message/commit_closure.go`.

Impact: closures with the same commits can have different hashes, against the history independence that prolly trees
are meant to have. Small closures stay below the minimum node size, so it shows only in long histories.

Hard to change in Go: fixing it changes the hashes of existing closures.

Rust: mirrors it exactly for byte compatibility (`commit_closure` in `crates/doltdb/src/database.rs`).

## GC output isn't deterministic (confirmed)

Running `dolt_gc('--archive-level', '0')` on two copies of the same database (the `rich` fixture, with
`testing/go/regression/out/gc_twice.sh`) wrote the same chunks to each generation, and the same new-generation table
file, but old-generation table files with different names because their chunks were in a different order.

Impact: identical databases end up with different files after GC, which defeats file-level comparison, caching, and
deduplication between replicas.

Rust: GC isn't ported yet. It should write a deterministic order.

## The journal index can't give back full addresses (confirmed)

`journal.idx` keys chunks by the first 16 bytes of their addresses, so `journalChunkSource.iterateAllChunks` returns
the chunks it found through the index with their addresses padded with zeros (a comment in Dolt's code says so).
Found when Go's store oracle printed such addresses for a large Rust-written journal.

Impact: anything that iterates every chunk of an indexed journal and trusts `Chunk.Hash()` sees wrong addresses.

Hard to change in Go: the 16-byte keys are part of the index file format.

Rust: the journal reader keeps full addresses.

## Query outcomes vary between runs (observed)

The full sqllogictest corpus was run twice against fresh Go servers, once per runner
(`testing/go/regression/out/slt-full.out`), and 17,458 records had different outcomes. The samples look like
nondeterminism in the Go server rather than differences between the runners: the same statement fails with
`the expression ... could not be found from the index idx_...` naming a different index on each run, or panics with
`index out of range [1] with length 1` on one run but not the other. To confirm by re-running the differing files.
