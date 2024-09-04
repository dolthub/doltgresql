// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestDbsize(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_dbsize)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_dbsize,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT size, pg_size_pretty(size), pg_size_pretty(-1 * size) FROM
    (VALUES (10::bigint), (1000::bigint), (1000000::bigint),
            (1000000000::bigint), (1000000000000::bigint),
            (1000000000000000::bigint)) x(size);`,
				Results: []sql.Row{{10, `10 bytes`, `-10 bytes`}, {1000, `1000 bytes`, `-1000 bytes`}, {1000000, `977 kB`, `-977 kB`}, {1000000000, `954 MB`, `-954 MB`}, {1000000000000, `931 GB`, `-931 GB`}, {1000000000000000, `909 TB`, `-909 TB`}},
			},
			{
				Statement: `SELECT size, pg_size_pretty(size), pg_size_pretty(-1 * size) FROM
    (VALUES (10::numeric), (1000::numeric), (1000000::numeric),
            (1000000000::numeric), (1000000000000::numeric),
            (1000000000000000::numeric),
            (10.5::numeric), (1000.5::numeric), (1000000.5::numeric),
            (1000000000.5::numeric), (1000000000000.5::numeric),
            (1000000000000000.5::numeric)) x(size);`,
				Results: []sql.Row{{10, `10 bytes`, `-10 bytes`}, {1000, `1000 bytes`, `-1000 bytes`}, {1000000, `977 kB`, `-977 kB`}, {1000000000, `954 MB`, `-954 MB`}, {1000000000000, `931 GB`, `-931 GB`}, {1000000000000000, `909 TB`, `-909 TB`}, {10.5, `10.5 bytes`, `-10.5 bytes`}, {1000.5, `1000.5 bytes`, `-1000.5 bytes`}, {1000000.5, `977 kB`, `-977 kB`}, {1000000000.5, `954 MB`, `-954 MB`}, {1000000000000.5, `931 GB`, `-931 GB`}, {1000000000000000.5, `909 TB`, `-909 TB`}},
			},
			{
				Statement: `SELECT size, pg_size_pretty(size), pg_size_pretty(-1 * size) FROM
    (VALUES (10239::bigint), (10240::bigint),
            (10485247::bigint), (10485248::bigint),
            (10736893951::bigint), (10736893952::bigint),
            (10994579406847::bigint), (10994579406848::bigint),
            (11258449312612351::bigint), (11258449312612352::bigint)) x(size);`,
				Results: []sql.Row{{10239, `10239 bytes`, `-10239 bytes`}, {10240, `10 kB`, `-10 kB`}, {10485247, `10239 kB`, `-10239 kB`}, {10485248, `10 MB`, `-10 MB`}, {10736893951, `10239 MB`, `-10239 MB`}, {10736893952, `10 GB`, `-10 GB`}, {10994579406847, `10239 GB`, `-10239 GB`}, {10994579406848, `10 TB`, `-10 TB`}, {11258449312612351, `10239 TB`, `-10239 TB`}, {11258449312612352, `10 PB`, `-10 PB`}},
			},
			{
				Statement: `SELECT size, pg_size_pretty(size), pg_size_pretty(-1 * size) FROM
    (VALUES (10239::numeric), (10240::numeric),
            (10485247::numeric), (10485248::numeric),
            (10736893951::numeric), (10736893952::numeric),
            (10994579406847::numeric), (10994579406848::numeric),
            (11258449312612351::numeric), (11258449312612352::numeric),
            (11528652096115048447::numeric), (11528652096115048448::numeric)) x(size);`,
				Results: []sql.Row{{10239, `10239 bytes`, `-10239 bytes`}, {10240, `10 kB`, `-10 kB`}, {10485247, `10239 kB`, `-10239 kB`}, {10485248, `10 MB`, `-10 MB`}, {10736893951, `10239 MB`, `-10239 MB`}, {10736893952, `10 GB`, `-10 GB`}, {10994579406847, `10239 GB`, `-10239 GB`}, {10994579406848, `10 TB`, `-10 TB`}, {11258449312612351, `10239 TB`, `-10239 TB`}, {11258449312612352, `10 PB`, `-10 PB`}, {11528652096115048447.0, `10239 PB`, `-10239 PB`}, {11528652096115048448.0, `10240 PB`, `-10240 PB`}},
			},
			{
				Statement: `SELECT size, pg_size_bytes(size) FROM
    (VALUES ('1'), ('123bytes'), ('1kB'), ('1MB'), (' 1 GB'), ('1.5 GB '),
            ('1TB'), ('3000 TB'), ('1e6 MB'), ('99 PB')) x(size);`,
				Results: []sql.Row{{1, 1}, {`123bytes`, 123}, {`1kB`, 1024}, {`1MB`, 1048576}, {`1 GB`, 1073741824}, {`1.5 GB`, 1610612736}, {`1TB`, 1099511627776}, {`3000 TB`, 3298534883328000}, {`1e6 MB`, 1048576000000}, {`99 PB`, 111464090777419776}},
			},
			{
				Statement: `SELECT size, pg_size_bytes(size) FROM
    (VALUES ('1'), ('123bYteS'), ('1kb'), ('1mb'), (' 1 Gb'), ('1.5 gB '),
            ('1tb'), ('3000 tb'), ('1e6 mb'), ('99 pb')) x(size);`,
				Results: []sql.Row{{1, 1}, {`123bYteS`, 123}, {`1kb`, 1024}, {`1mb`, 1048576}, {`1 Gb`, 1073741824}, {`1.5 gB`, 1610612736}, {`1tb`, 1099511627776}, {`3000 tb`, 3298534883328000}, {`1e6 mb`, 1048576000000}, {`99 pb`, 111464090777419776}},
			},
			{
				Statement: `SELECT size, pg_size_bytes(size) FROM
    (VALUES ('-1'), ('-123bytes'), ('-1kb'), ('-1mb'), (' -1 Gb'), ('-1.5 gB '),
            ('-1tb'), ('-3000 TB'), ('-10e-1 MB'), ('-99 PB')) x(size);`,
				Results: []sql.Row{{-1, -1}, {`-123bytes`, -123}, {`-1kb`, -1024}, {`-1mb`, -1048576}, {`-1 Gb`, -1073741824}, {`-1.5 gB`, -1610612736}, {`-1tb`, -1099511627776}, {`-3000 TB`, -3298534883328000}, {`-10e-1 MB`, -1048576}, {`-99 PB`, -111464090777419776}},
			},
			{
				Statement: `SELECT size, pg_size_bytes(size) FROM
     (VALUES ('-1.'), ('-1.kb'), ('-1. kb'), ('-0. gb'),
             ('-.1'), ('-.1kb'), ('-.1 kb'), ('-.0 gb')) x(size);`,
				Results: []sql.Row{{-1., -1}, {`-1.kb`, -1024}, {`-1. kb`, -1024}, {`-0. gb`, 0}, {-.1, 0}, {`-.1kb`, -102}, {`-.1 kb`, -102}, {`-.0 gb`, 0}},
			},
			{
				Statement:   `SELECT pg_size_bytes('1 AB');`,
				ErrorString: `invalid size: "1 AB"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('1 AB A');`,
				ErrorString: `invalid size: "1 AB A"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('1 AB A    ');`,
				ErrorString: `invalid size: "1 AB A    "`,
			},
			{
				Statement:   `SELECT pg_size_bytes('9223372036854775807.9');`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT pg_size_bytes('1e100');`,
				ErrorString: `bigint out of range`,
			},
			{
				Statement:   `SELECT pg_size_bytes('1e1000000000000000000');`,
				ErrorString: `value overflows numeric format`,
			},
			{
				Statement:   `SELECT pg_size_bytes('1 byte');  -- the singular "byte" is not supported`,
				ErrorString: `invalid size: "1 byte"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('');`,
				ErrorString: `invalid size: ""`,
			},
			{
				Statement:   `SELECT pg_size_bytes('kb');`,
				ErrorString: `invalid size: "kb"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('..');`,
				ErrorString: `invalid size: ".."`,
			},
			{
				Statement:   `SELECT pg_size_bytes('-.');`,
				ErrorString: `invalid size: "-."`,
			},
			{
				Statement:   `SELECT pg_size_bytes('-.kb');`,
				ErrorString: `invalid size: "-.kb"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('-. kb');`,
				ErrorString: `invalid size: "-. kb"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('.+912');`,
				ErrorString: `invalid size: ".+912"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('+912+ kB');`,
				ErrorString: `invalid size: "+912+ kB"`,
			},
			{
				Statement:   `SELECT pg_size_bytes('++123 kB');`,
				ErrorString: `invalid size: "++123 kB"`,
			},
		},
	})
}
