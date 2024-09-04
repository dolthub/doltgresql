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
)

func TestCopydml(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_copydml)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_copydml,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `create table copydml_test (id serial, t text);`,
			},
			{
				Statement: `insert into copydml_test (t) values ('a');`,
			},
			{
				Statement: `insert into copydml_test (t) values ('b');`,
			},
			{
				Statement: `insert into copydml_test (t) values ('c');`,
			},
			{
				Statement: `insert into copydml_test (t) values ('d');`,
			},
			{
				Statement: `insert into copydml_test (t) values ('e');`,
			},
			{
				Statement: `copy (insert into copydml_test (t) values ('f') returning id) to stdout;`,
			},
			{
				Statement: `6
copy (update copydml_test set t = 'g' where t = 'f' returning id) to stdout;`,
			},
			{
				Statement: `6
copy (delete from copydml_test where t = 'g' returning id) to stdout;`,
			},
			{
				Statement: `6
\copy (insert into copydml_test (t) values ('f') returning id) to stdout;`,
			},
			{
				Statement: `7
\copy (update copydml_test set t = 'g' where t = 'f' returning id) to stdout;`,
			},
			{
				Statement: `7
\copy (delete from copydml_test where t = 'g' returning id) to stdout;`,
			},
			{
				Statement: `7
copy (insert into copydml_test default values) to stdout;`,
				ErrorString: `COPY query must have a RETURNING clause`,
			},
			{
				Statement:   `copy (update copydml_test set t = 'g') to stdout;`,
				ErrorString: `COPY query must have a RETURNING clause`,
			},
			{
				Statement:   `copy (delete from copydml_test) to stdout;`,
				ErrorString: `COPY query must have a RETURNING clause`,
			},
			{
				Statement: `create rule qqq as on insert to copydml_test do instead nothing;`,
			},
			{
				Statement:   `copy (insert into copydml_test default values) to stdout;`,
				ErrorString: `DO INSTEAD NOTHING rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on insert to copydml_test do also delete from copydml_test;`,
			},
			{
				Statement:   `copy (insert into copydml_test default values) to stdout;`,
				ErrorString: `DO ALSO rules are not supported for the COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on insert to copydml_test do instead (delete from copydml_test; delete from copydml_test);`,
			},
			{
				Statement:   `copy (insert into copydml_test default values) to stdout;`,
				ErrorString: `multi-statement DO INSTEAD rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on insert to copydml_test where new.t <> 'f' do instead delete from copydml_test;`,
			},
			{
				Statement:   `copy (insert into copydml_test default values) to stdout;`,
				ErrorString: `conditional DO INSTEAD rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on update to copydml_test do instead nothing;`,
			},
			{
				Statement:   `copy (update copydml_test set t = 'f') to stdout;`,
				ErrorString: `DO INSTEAD NOTHING rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on update to copydml_test do also delete from copydml_test;`,
			},
			{
				Statement:   `copy (update copydml_test set t = 'f') to stdout;`,
				ErrorString: `DO ALSO rules are not supported for the COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on update to copydml_test do instead (delete from copydml_test; delete from copydml_test);`,
			},
			{
				Statement:   `copy (update copydml_test set t = 'f') to stdout;`,
				ErrorString: `multi-statement DO INSTEAD rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on update to copydml_test where new.t <> 'f' do instead delete from copydml_test;`,
			},
			{
				Statement:   `copy (update copydml_test set t = 'f') to stdout;`,
				ErrorString: `conditional DO INSTEAD rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on delete to copydml_test do instead nothing;`,
			},
			{
				Statement:   `copy (delete from copydml_test) to stdout;`,
				ErrorString: `DO INSTEAD NOTHING rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on delete to copydml_test do also insert into copydml_test default values;`,
			},
			{
				Statement:   `copy (delete from copydml_test) to stdout;`,
				ErrorString: `DO ALSO rules are not supported for the COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on delete to copydml_test do instead (insert into copydml_test default values; insert into copydml_test default values);`,
			},
			{
				Statement:   `copy (delete from copydml_test) to stdout;`,
				ErrorString: `multi-statement DO INSTEAD rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create rule qqq as on delete to copydml_test where old.t <> 'f' do instead insert into copydml_test default values;`,
			},
			{
				Statement:   `copy (delete from copydml_test) to stdout;`,
				ErrorString: `conditional DO INSTEAD rules are not supported for COPY`,
			},
			{
				Statement: `drop rule qqq on copydml_test;`,
			},
			{
				Statement: `create function qqq_trig() returns trigger as $$
begin
if tg_op in ('INSERT', 'UPDATE') then
    raise notice '% % %', tg_when, tg_op, new.id;`,
			},
			{
				Statement: `    return new;`,
			},
			{
				Statement: `else
    raise notice '% % %', tg_when, tg_op, old.id;`,
			},
			{
				Statement: `    return old;`,
			},
			{
				Statement: `end if;`,
			},
			{
				Statement: `end
$$ language plpgsql;`,
			},
			{
				Statement: `create trigger qqqbef before insert or update or delete on copydml_test
    for each row execute procedure qqq_trig();`,
			},
			{
				Statement: `create trigger qqqaf after insert or update or delete on copydml_test
    for each row execute procedure qqq_trig();`,
			},
			{
				Statement: `copy (insert into copydml_test (t) values ('f') returning id) to stdout;`,
			},
			{
				Statement: `8
copy (update copydml_test set t = 'g' where t = 'f' returning id) to stdout;`,
			},
			{
				Statement: `8
copy (delete from copydml_test where t = 'g' returning id) to stdout;`,
			},
			{
				Statement: `8
drop table copydml_test;`,
			},
			{
				Statement: `drop function qqq_trig();`,
			},
		},
	})
}
