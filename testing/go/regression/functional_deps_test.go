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

func TestFunctionalDeps(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_functional_deps)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_functional_deps,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TEMP TABLE articles (
    id int CONSTRAINT articles_pkey PRIMARY KEY,
    keywords text,
    title text UNIQUE NOT NULL,
    body text UNIQUE,
    created date
);`,
			},
			{
				Statement: `CREATE TEMP TABLE articles_in_category (
    article_id int,
    category_id int,
    changed date,
    PRIMARY KEY (article_id, category_id)
);`,
			},
			{
				Statement: `SELECT id, keywords, title, body, created
FROM articles
GROUP BY id;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT id, keywords, title, body, created
FROM articles
GROUP BY title;`,
				ErrorString: `column "articles.id" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT id, keywords, title, body, created
FROM articles
GROUP BY body;`,
				ErrorString: `column "articles.id" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT id, keywords, title, body, created
FROM articles
GROUP BY keywords;`,
				ErrorString: `column "articles.id" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT a.id, a.keywords, a.title, a.body, a.created
FROM articles AS a, articles_in_category AS aic
WHERE a.id = aic.article_id AND aic.category_id in (14,62,70,53,138)
GROUP BY a.id;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT a.id, a.keywords, a.title, a.body, a.created
FROM articles AS a, articles_in_category AS aic
WHERE a.id = aic.article_id AND aic.category_id in (14,62,70,53,138)
GROUP BY aic.article_id, aic.category_id;`,
				ErrorString: `column "a.id" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT a.id, a.keywords, a.title, a.body, a.created
FROM articles AS a JOIN articles_in_category AS aic ON a.id = aic.article_id
WHERE aic.category_id in (14,62,70,53,138)
GROUP BY a.id;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT a.id, a.keywords, a.title, a.body, a.created
FROM articles AS a JOIN articles_in_category AS aic ON a.id = aic.article_id
WHERE aic.category_id in (14,62,70,53,138)
GROUP BY aic.article_id, aic.category_id;`,
				ErrorString: `column "a.id" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `SELECT aic.changed
FROM articles AS a JOIN articles_in_category AS aic ON a.id = aic.article_id
WHERE aic.category_id in (14,62,70,53,138)
GROUP BY aic.category_id, aic.article_id;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT aic.changed
FROM articles AS a JOIN articles_in_category AS aic ON a.id = aic.article_id
WHERE aic.category_id in (14,62,70,53,138)
GROUP BY aic.article_id;`,
				ErrorString: `column "aic.changed" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `CREATE TEMP TABLE products (product_id int, name text, price numeric);`,
			},
			{
				Statement: `CREATE TEMP TABLE sales (product_id int, units int);`,
			},
			{
				Statement: `SELECT product_id, p.name, (sum(s.units) * p.price) AS sales
    FROM products p LEFT JOIN sales s USING (product_id)
    GROUP BY product_id, p.name, p.price;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT product_id, p.name, (sum(s.units) * p.price) AS sales
    FROM products p LEFT JOIN sales s USING (product_id)
    GROUP BY product_id;`,
				ErrorString: `column "p.name" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `ALTER TABLE products ADD PRIMARY KEY (product_id);`,
			},
			{
				Statement: `SELECT product_id, p.name, (sum(s.units) * p.price) AS sales
    FROM products p LEFT JOIN sales s USING (product_id)
    GROUP BY product_id;`,
				Results: []sql.Row{},
			},
			{
				Statement: `CREATE TEMP TABLE node (
    nid SERIAL,
    vid integer NOT NULL default '0',
    type varchar(32) NOT NULL default '',
    title varchar(128) NOT NULL default '',
    uid integer NOT NULL default '0',
    status integer NOT NULL default '1',
    created integer NOT NULL default '0',
    -- snip
    PRIMARY KEY (nid, vid)
);`,
			},
			{
				Statement: `CREATE TEMP TABLE users (
    uid integer NOT NULL default '0',
    name varchar(60) NOT NULL default '',
    pass varchar(32) NOT NULL default '',
    -- snip
    PRIMARY KEY (uid),
    UNIQUE (name)
);`,
			},
			{
				Statement: `SELECT u.uid, u.name FROM node n
INNER JOIN users u ON u.uid = n.uid
WHERE n.type = 'blog' AND n.status = 1
GROUP BY u.uid, u.name;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT u.uid, u.name FROM node n
INNER JOIN users u ON u.uid = n.uid
WHERE n.type = 'blog' AND n.status = 1
GROUP BY u.uid;`,
				Results: []sql.Row{},
			},
			{
				Statement: `CREATE TEMP VIEW fdv1 AS
SELECT id, keywords, title, body, created
FROM articles
GROUP BY body;`,
				ErrorString: `column "articles.id" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
			{
				Statement: `CREATE TEMP VIEW fdv1 AS
SELECT id, keywords, title, body, created
FROM articles
GROUP BY id;`,
			},
			{
				Statement:   `ALTER TABLE articles DROP CONSTRAINT articles_pkey RESTRICT;`,
				ErrorString: `cannot drop constraint articles_pkey on table articles because other objects depend on it`,
			},
			{
				Statement: `DROP VIEW fdv1;`,
			},
			{
				Statement: `CREATE TEMP VIEW fdv2 AS
SELECT a.id, a.keywords, a.title, aic.category_id, aic.changed
FROM articles AS a JOIN articles_in_category AS aic ON a.id = aic.article_id
WHERE aic.category_id in (14,62,70,53,138)
GROUP BY a.id, aic.category_id, aic.article_id;`,
			},
			{
				Statement:   `ALTER TABLE articles DROP CONSTRAINT articles_pkey RESTRICT; -- fail`,
				ErrorString: `cannot drop constraint articles_pkey on table articles because other objects depend on it`,
			},
			{
				Statement: `ALTER TABLE articles_in_category DROP CONSTRAINT articles_in_category_pkey RESTRICT; --fail
ERROR:  cannot drop constraint articles_in_category_pkey on table articles_in_category because other objects depend on it
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP VIEW fdv2;`,
			},
			{
				Statement: `CREATE TEMP VIEW fdv3 AS
SELECT id, keywords, title, body, created
FROM articles
GROUP BY id
UNION
SELECT id, keywords, title, body, created
FROM articles
GROUP BY id;`,
			},
			{
				Statement:   `ALTER TABLE articles DROP CONSTRAINT articles_pkey RESTRICT; -- fail`,
				ErrorString: `cannot drop constraint articles_pkey on table articles because other objects depend on it`,
			},
			{
				Statement: `DROP VIEW fdv3;`,
			},
			{
				Statement: `CREATE TEMP VIEW fdv4 AS
SELECT * FROM articles WHERE title IN (SELECT title FROM articles GROUP BY id);`,
			},
			{
				Statement:   `ALTER TABLE articles DROP CONSTRAINT articles_pkey RESTRICT; -- fail`,
				ErrorString: `cannot drop constraint articles_pkey on table articles because other objects depend on it`,
			},
			{
				Statement: `DROP VIEW fdv4;`,
			},
			{
				Statement: `PREPARE foo AS
  SELECT id, keywords, title, body, created
  FROM articles
  GROUP BY id;`,
			},
			{
				Statement: `EXECUTE foo;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ALTER TABLE articles DROP CONSTRAINT articles_pkey RESTRICT;`,
			},
			{
				Statement:   `EXECUTE foo;  -- fail`,
				ErrorString: `column "articles.keywords" must appear in the GROUP BY clause or be used in an aggregate function`,
			},
		},
	})
}
