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

func TestRegex(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_regex)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_regex,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `set standard_conforming_strings = on;`,
			},
			{
				Statement: `select 'bbbbb' ~ '^([bc])\1*$' as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'ccc' ~ '^([bc])\1*$' as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'xxx' ~ '^([bc])\1*$' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'bbc' ~ '^([bc])\1*$' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'b' ~ '^([bc])\1*$' as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'abc abc abc' ~ '^(\w+)( \1)+$' as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'abc abd abc' ~ '^(\w+)( \1)+$' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'abc abc abd' ~ '^(\w+)( \1)+$' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'abc abc abc' ~ '^(.+)( \1)+$' as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'abc abd abc' ~ '^(.+)( \1)+$' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'abc abc abd' ~ '^(.+)( \1)+$' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select substring('asd TO foo' from ' TO (([a-z0-9._]+|"([^"]+|"")+")+)');`,
				Results:   []sql.Row{{`foo`}},
			},
			{
				Statement: `select substring('a' from '((a))+');`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select substring('a' from '((a)+)');`,
				Results:   []sql.Row{{`a`}},
			},
			{
				Statement: `select regexp_match('abc', '');`,
				Results:   []sql.Row{{`{""}`}},
			},
			{
				Statement: `select regexp_match('abc', 'bc');`,
				Results:   []sql.Row{{`{bc}`}},
			},
			{
				Statement: `select regexp_match('abc', 'd') is null;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select regexp_match('abc', '(B)(c)', 'i');`,
				Results:   []sql.Row{{`{b,c}`}},
			},
			{
				Statement:   `select regexp_match('abc', 'Bd', 'ig'); -- error`,
				ErrorString: `regexp_match() does not support the "global" option`,
			},
			{
				Statement: `select regexp_matches('ab', 'a(?=b)b*');`,
				Results:   []sql.Row{{`{ab}`}},
			},
			{
				Statement: `select regexp_matches('a', 'a(?=b)b*');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('abc', 'a(?=b)b*(?=c)c*');`,
				Results:   []sql.Row{{`{abc}`}},
			},
			{
				Statement: `select regexp_matches('ab', 'a(?=b)b*(?=c)c*');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('ab', 'a(?!b)b*');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('a', 'a(?!b)b*');`,
				Results:   []sql.Row{{`{a}`}},
			},
			{
				Statement: `select regexp_matches('b', '(?=b)b');`,
				Results:   []sql.Row{{`{b}`}},
			},
			{
				Statement: `select regexp_matches('a', '(?=b)b');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('abb', '(?<=a)b*');`,
				Results:   []sql.Row{{`{bb}`}},
			},
			{
				Statement: `select regexp_matches('a', 'a(?<=a)b*');`,
				Results:   []sql.Row{{`{a}`}},
			},
			{
				Statement: `select regexp_matches('abc', 'a(?<=a)b*(?<=b)c*');`,
				Results:   []sql.Row{{`{abc}`}},
			},
			{
				Statement: `select regexp_matches('ab', 'a(?<=a)b*(?<=b)c*');`,
				Results:   []sql.Row{{`{ab}`}},
			},
			{
				Statement: `select regexp_matches('ab', 'a*(?<!a)b*');`,
				Results:   []sql.Row{{`{""}`}},
			},
			{
				Statement: `select regexp_matches('ab', 'a*(?<!a)b+');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('b', 'a*(?<!a)b+');`,
				Results:   []sql.Row{{`{b}`}},
			},
			{
				Statement: `select regexp_matches('a', 'a(?<!a)b*');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('b', '(?<=b)b');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('foobar', '(?<=f)b+');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select regexp_matches('foobar', '(?<=foo)b+');`,
				Results:   []sql.Row{{`{b}`}},
			},
			{
				Statement: `select regexp_matches('foobar', '(?<=oo)b+');`,
				Results:   []sql.Row{{`{b}`}},
			},
			{
				Statement: `select 'xz' ~ 'x(?=[xy])';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'xy' ~ 'x(?=[xy])';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'xz' ~ 'x(?![xy])';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'xy' ~ 'x(?![xy])';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'x'  ~ 'x(?![xy])';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'xyy' ~ '(?<=[xy])yy+';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'zyy' ~ '(?<=[xy])yy+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'xyy' ~ '(?<![xy])yy+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'zyy' ~ '(?<![xy])yy+';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ 'abc';`,
				Results:   []sql.Row{{`Seq Scan on pg_proc`}, {`Filter: (proname ~ 'abc'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^abc';`,
				Results:   []sql.Row{{`Index Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: ((proname >= 'abc'::text) AND (proname < 'abd'::text))`}, {`Filter: (proname ~ '^abc'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^abc$';`,
				Results:   []sql.Row{{`Index Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: (proname = 'abc'::text)`}, {`Filter: (proname ~ '^abc$'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^abcd*e';`,
				Results:   []sql.Row{{`Index Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: ((proname >= 'abc'::text) AND (proname < 'abd'::text))`}, {`Filter: (proname ~ '^abcd*e'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^abc+d';`,
				Results:   []sql.Row{{`Index Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: ((proname >= 'abc'::text) AND (proname < 'abd'::text))`}, {`Filter: (proname ~ '^abc+d'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^(abc)(def)';`,
				Results:   []sql.Row{{`Index Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: ((proname >= 'abcdef'::text) AND (proname < 'abcdeg'::text))`}, {`Filter: (proname ~ '^(abc)(def)'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^(abc)$';`,
				Results:   []sql.Row{{`Index Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: (proname = 'abc'::text)`}, {`Filter: (proname ~ '^(abc)$'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^(abc)?d';`,
				Results:   []sql.Row{{`Seq Scan on pg_proc`}, {`Filter: (proname ~ '^(abc)?d'::text)`}},
			},
			{
				Statement: `explain (costs off) select * from pg_proc where proname ~ '^abcd(x|(?=\w\w)q)';`,
				Results:   []sql.Row{{`Index Scan using pg_proc_proname_args_nsp_index on pg_proc`}, {`Index Cond: ((proname >= 'abcd'::text) AND (proname < 'abce'::text))`}, {`Filter: (proname ~ '^abcd(x|(?=\w\w)q)'::text)`}},
			},
			{
				Statement: `select 'a' ~ '($|^)*';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'a' ~ '(^)+^';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'a' ~ '$($$)+';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'a' ~ '($^)+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'a' ~ '(^$)*';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'aa bb cc' ~ '(^(?!aa))+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'aa x' ~ '(^(?!aa)(?!bb)(?!cc))+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'bb x' ~ '(^(?!aa)(?!bb)(?!cc))+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'cc x' ~ '(^(?!aa)(?!bb)(?!cc))+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'dd x' ~ '(^(?!aa)(?!bb)(?!cc))+';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'a' ~ '((((((a)*)*)*)*)*)*';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'a' ~ '((((((a+|)+|)+|)+|)+|)+|)';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'x' ~ 'abcd(\m)+xyz';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'a' ~ '^abcd*(((((^(a c(e?d)a+|)+|)+|)+|)+|a)+|)';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'x' ~ 'a^(^)bcd*xy(((((($a+|)+|)+|)+$|)+|)+|)^$';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'x' ~ 'xyz(\Y\Y)+';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'x' ~ 'x|(?:\M)+';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement:   `select 'x' ~ repeat('x*y*z*', 1000);`,
				ErrorString: `invalid regular expression: regular expression is too complex`,
			},
			{
				Statement: `select 'Programmer' ~ '(\w).*?\1' as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select regexp_matches('Programmer', '(\w)(.*?\1)', 'g');`,
				Results:   []sql.Row{{`{r,ogr}`}, {`{m,m}`}},
			},
			{
				Statement: `select regexp_matches('foo/bar/baz',
                      '^([^/]+?)(?:/([^/]+?))(?:/([^/]+?))?$', '');`,
				Results: []sql.Row{{`{foo,bar,baz}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*)(.*)(f*)$');`,
				Results:   []sql.Row{{`{ll,mmmfff,""}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*){1,1}(.*)(f*)$');`,
				Results:   []sql.Row{{`{ll,mmmfff,""}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*){1,1}?(.*)(f*)$');`,
				Results:   []sql.Row{{`{"",llmmmfff,""}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*){1,1}?(.*){1,1}?(f*)$');`,
				Results:   []sql.Row{{`{"",llmmm,fff}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*?)(.*)(f*)$');`,
				Results:   []sql.Row{{`{"",llmmmfff,""}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*?){1,1}(.*)(f*)$');`,
				Results:   []sql.Row{{`{ll,mmmfff,""}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*?){1,1}?(.*)(f*)$');`,
				Results:   []sql.Row{{`{"",llmmmfff,""}`}},
			},
			{
				Statement: `select regexp_matches('llmmmfff', '^(l*?){1,1}?(.*){1,1}?(f*)$');`,
				Results:   []sql.Row{{`{"",llmmm,fff}`}},
			},
			{
				Statement: `select 'a' ~ '$()|^\1';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'a' ~ '.. ()|\1';`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'a' ~ '()*\1';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'a' ~ '()+\1';`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'xxx' ~ '(.){0}(\1)' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'xxx' ~ '((.)){0}(\2)' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'xyz' ~ '((.)){0}(\2){0}' as t;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `select 'abcdef' ~ '^(.)\1|\1.' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select 'abadef' ~ '^((.)\2|..)\2' as f;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `select regexp_match('xy', '.|...');`,
				Results:   []sql.Row{{`{x}`}},
			},
			{
				Statement: `select regexp_match('xyz', '.|...');`,
				Results:   []sql.Row{{`{xyz}`}},
			},
			{
				Statement: `select regexp_match('xy', '.*');`,
				Results:   []sql.Row{{`{xy}`}},
			},
			{
				Statement: `select regexp_match('fooba', '(?:..)*');`,
				Results:   []sql.Row{{`{foob}`}},
			},
			{
				Statement: `select regexp_match('xyz', repeat('.', 260));`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `select regexp_match('foo', '(?:.|){99}');`,
				Results:   []sql.Row{{`{foo}`}},
			},
			{
				Statement:   `select 'xyz' ~ 'x(\w)(?=\1)';  -- no backrefs in LACONs`,
				ErrorString: `invalid regular expression: invalid backreference number`,
			},
			{
				Statement:   `select 'xyz' ~ 'x(\w)(?=(\1))';`,
				ErrorString: `invalid regular expression: invalid backreference number`,
			},
			{
				Statement:   `select 'a' ~ '\x7fffffff';  -- invalid chr code`,
				ErrorString: `invalid regular expression: invalid escape \ sequence`,
			},
		},
	})
}
