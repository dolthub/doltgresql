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

## Arrays of enums can't be read after a restart (confirmed)

Creating an enum, a table with a column of the enum's array type, and a row, then restarting the server and selecting
the column panics with a nil pointer dereference in `DoltgresType.CallReceive`, reached from `deserializeArray`. The
element type that the column's stored array type refers to is not resolved after the restart, so it has neither a
deserialization function nor a receive function. The same session that created the type reads the column fine.

Impact: any database with an array-of-enum column, or an array of any user-defined type (confirmed for pgvector's
`vector[]` too), is unreadable by Go after a restart.

Hard to change in Go: needs element types of stored array types resolved through the type collection on load.

Rust: user-defined types are registered by OID from the stored definitions, so element types always resolve.

## Serial defaults name mixed-case sequences unquoted (confirmed)

A serial column's default is stored as `nextval('schema.name')` without quoting, and Go's `nextval` never folds case,
so `"Id" SERIAL` on table `regions` stores `nextval('public.regions_Id_seq')`. Postgres reads that text as
`public.regions_id_seq`, which does not exist.

Impact: Rust, which resolves the name as Postgres does, cannot run the default of a mixed-case serial column that Go
created. Go reads the quoted defaults that Rust writes, since it trims the quotes.

Hard to change in Go: fixing the stored text breaks nothing, but every existing database keeps the unquoted form.

Rust: writes `nextval('public."regions_Id_seq"')`.

## Vector index searches ignore Postgres' NULL and zero-vector rules (confirmed)

go-mysql-server answers every `ORDER BY v <op> q LIMIT n` with a vector index when one matches, however small the
table, and Dolt's proximity map leaves out NULL vectors, so `LIMIT 5` over five rows with two NULLs returns three.
Postgres costs the plan and scans a small table sequentially, returning all five. Dolt's cosine distance of a zero
vector is 0, where pgvector's is NaN, so the zero vector sorts first instead of last.

Hard to change in Go: the first needs a cost model for vector indexes, the second changes stored index order.

Rust: matches Go, with the Postgres expectations kept as skipped assertions.

## DOLT_PATCH writes MySQL statements (confirmed)

Dolt's sqlfmt writes the patch's string literals with MySQL's backslash escapes (`'it\'s'`, `'{\"a\": 1}'`), booleans
as `'1'`, and schema changes as MySQL statements (`RENAME TABLE`, `MODIFY COLUMN`, `DROP PRIMARY KEY`, `ADD INDEX`,
`DROP FOREIGN KEY`), none of which Postgres runs as written.

Hard to change in Go: the statements come from Dolt's shared formatter, which Doltgres only partly overrides.

Rust: writes Postgres statements (`''` quoting, `ALTER TABLE ... RENAME TO`, `ALTER COLUMN ... TYPE`, `DROP CONSTRAINT`,
`CREATE INDEX`), keeping Go's text where Go's is already valid Postgres; the one test that showed the escapes expects
the Postgres form.

## DOLT_VERIFY_CONSTRAINTS without table names checks nothing (confirmed)

With no table arguments, Dolt's parseTablesToCheck lists the tables of the default schema `""` (its TODO notes the
missing search path), which holds no Doltgres tables, so the procedure returns 0 even when `public` tables have
constraint violations. Violations are still recorded, since the merge records them for every table.

Small fix in Go, but in Dolt's shared procedure rather than Doltgres.

Rust: checks the `public` tables and returns 1 when one has violations.

## Logical replication drops large transactions and builds broken statements (confirmed by reading)

The replicator asks pgoutput for `streaming 'true'`, so Postgres sends any transaction larger than
logical_decoding_work_mem as stream segments before it commits. Those changes carry no Begin message, so
`processMessages` is false and every one is logged as stale and dropped, and StreamCommit only logs. Values are spliced
into SQL with no escaping (a quote in a string breaks the statement), key conditions for multi-column keys are joined
without AND (and the separator goes into the SET list), an update that changes a key looks the row up by its new key,
an unchanged TOAST column becomes `col = ` with no value, TRUNCATE is ignored, and the failure counter never resets,
so ten errors over any span of time stop replication.

Small fixes in Go, apart from streaming, which needs per-transaction buffering or dropping the option.

Rust: requests no streaming (Postgres then sends large transactions whole at commit), quotes every value as an
escaped literal that the replica casts, joins key conditions with AND, finds updated rows by the old key when
Postgres sends it, leaves unchanged TOAST columns out, applies TRUNCATE, and resets the counter after each message.

## OSS remotes cannot be pushed to (confirmed against a fake OSS server)

Dolt's OSS factory opens its store with NewBSStore, whose persister writes each table file as its records and tail
and then concatenates them, but OSSBlobstore.Concatenate always fails ("Conjoin is not implemented for
OSSBlobstore"), so every push to an oss:// remote fails with "unknown push error". Its CheckAndPutManifest also has
no condition: it passes the expected version as a versionId parameter, which a write ignores, so concurrent pushes
can overwrite each other's manifests.

Small fix in Go: open the store with NewNoConjoinBSStore, as the S3 and OCI factories do.

Rust: writes OSS table files whole, as the no-conjoin persister does, so pushes work; the manifest write keeps Go's
missing condition.
