// Copyright 2026 Dolthub, Inc.
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

//! Rewrites the MySQL statements that Dolt ran into the Postgres statements that ask Doltgres the same thing, where
//! MySQL settings and catalogs become their Postgres counterparts: FOREIGN_KEY_CHECKS session_replication_role, and
//! SHOW TRIGGERS and the MySQL forms of information_schema.statistics and key_column_usage queries of the Postgres
//! catalogs.

use pg_query::protobuf::{ScanToken, Token};

/// Catalog runs a query on the server, returning its rows as text, or none when it fails.
pub type Catalog<'c> = dyn FnMut(&str) -> Vec<Vec<String>> + 'c;

/// Translated is what a MySQL statement becomes.
#[derive(Debug, PartialEq)]
pub enum Translated {
    /// Statements that Doltgres runs in its place, whose last result stands for the MySQL statement's.
    Statements(Vec<String>),
    /// Turning autocommit off or on, which Postgres does with explicit transactions.
    Autocommit(bool),
}

/// STATISTICS stands for MySQL's information_schema.statistics: a row for each column of each index of the public
/// schema's tables, a primary key's named PRIMARY as MySQL names it.
const STATISTICS: &str = "(SELECT n.nspname::text AS table_schema, c.relname::text AS table_name, \
     CASE WHEN i.indisprimary THEN 'PRIMARY' ELSE ic.relname::text END AS index_name, (NOT i.indisunique)::int AS non_unique, \
     k.n::int AS seq_in_index, a.attname::text AS column_name FROM pg_index i JOIN pg_class c ON c.oid = i.indrelid \
     JOIN pg_namespace n ON n.oid = c.relnamespace JOIN pg_class ic ON ic.oid = i.indexrelid \
     CROSS JOIN LATERAL unnest(i.indkey::int2[]) WITH ORDINALITY AS k(attnum, n) \
     JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = k.attnum WHERE n.nspname = 'public')";

/// KEY_COLUMN_USAGE stands for MySQL's information_schema.key_column_usage, whose rows of foreign keys also name the
/// columns they refer to: a row for each column of each primary key, unique, and foreign key constraint of the public
/// schema's tables.
const KEY_COLUMN_USAGE: &str = "(SELECT con.conname::text AS constraint_name, n.nspname::text AS table_schema, \
     c.relname::text AS table_name, a.attname::text AS column_name, k.n::int AS ordinal_position, \
     fn.nspname::text AS referenced_table_schema, fc.relname::text AS referenced_table_name, \
     fa.attname::text AS referenced_column_name FROM pg_constraint con JOIN pg_class c ON c.oid = con.conrelid \
     JOIN pg_namespace n ON n.oid = c.relnamespace CROSS JOIN LATERAL unnest(con.conkey) WITH ORDINALITY AS k(attnum, n) \
     JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = k.attnum LEFT JOIN pg_class fc ON fc.oid = con.confrelid \
     LEFT JOIN pg_namespace fn ON fn.oid = fc.relnamespace \
     LEFT JOIN pg_attribute fa ON fa.attrelid = con.confrelid AND fa.attnum = con.confkey[k.n] \
     WHERE con.contype IN ('p', 'u', 'f') AND n.nspname = 'public')";

/// CHECK_CONSTRAINTS stands for MySQL's information_schema.check_constraints, which lists only the CHECK constraints
/// of the database's own tables, where Postgres' also lists the system catalogs' constraints and NOT NULL constraints.
const CHECK_CONSTRAINTS: &str = "(SELECT * FROM information_schema.check_constraints \
     WHERE constraint_schema = 'public' AND check_clause NOT LIKE '% IS NOT NULL')";

/// SHOW_TRIGGERS stands for MySQL's SHOW TRIGGERS: each trigger with its event, table, and timing, leaving out its
/// MySQL definition, which Postgres keeps in the trigger's function instead.
const SHOW_TRIGGERS: &str = "SELECT t.tgname::text AS \"Trigger\", CASE WHEN (t.tgtype & 4) <> 0 THEN 'INSERT' \
     WHEN (t.tgtype & 8) <> 0 THEN 'DELETE' ELSE 'UPDATE' END AS \"Event\", c.relname::text AS \"Table\", \
     CASE WHEN (t.tgtype & 2) <> 0 THEN 'BEFORE' ELSE 'AFTER' END AS \"Timing\" \
     FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid WHERE NOT t.tgisinternal ORDER BY 1";

/// BOOL_COLUMNS are the boolean columns of Dolt's system tables, which MySQL compares with integers.
const BOOL_COLUMNS: &[&str] =
    &["staged", "dirty", "ignored", "is_merging", "fast_forward", "data_change", "schema_change"];

/// translate returns the Postgres statements that stand for a MySQL statement, reading the tables they need to know
/// through a catalog.
pub fn translate(mysql: &str, catalog: &mut Catalog<'_>) -> Translated {
    let sql = requote(mysql);
    let Ok(scanned) = pg_query::scan(&sql) else { return Translated::Statements(vec![sql]) };
    let tokens = scanned.tokens;
    let word = |i: usize| tokens.get(i).map(|t| text(&sql, t).to_ascii_lowercase()).unwrap_or_default();
    let after = |i: usize| tokens.get(i).map_or("", |t| &sql[t.end as usize..]);
    let statements = |s: String| Translated::Statements(vec![s]);
    let assignment = tokens.iter().position(|t| matches!(text(&sql, t), "=" | ":="));
    match (word(0).as_str(), word(1).as_str(), word(2).as_str()) {
        ("set", _, _) if assignment.is_some() => {
            let at = assignment.unwrap_or_default();
            let name = setting(&sql[tokens[0].end as usize..tokens[at].start as usize]);
            let value = sql[tokens[at].end as usize..].trim();
            let off = matches!(value.to_ascii_lowercase().as_str(), "0" | "off" | "false");
            match name.as_str() {
                "autocommit" => Translated::Autocommit(!off),
                "foreign_key_checks" => {
                    statements(format!("SET session_replication_role = {}", if off { "replica" } else { "origin" }))
                }
                _ => statements(format!("SET {name} = {}", rewrite(value))),
            }
        }
        ("call", _, _) => statements(format!("SELECT * FROM {}", rewrite(after(0).trim()))),
        ("insert", "ignore", _) => statements(format!("INSERT{} ON CONFLICT DO NOTHING", rewrite(after(1)))),
        ("replace", "into", _) => statements(upsert(&format!("INSERT INTO{}", after(1)), &[], catalog)),
        ("rename", "table", _) => statements(format!("ALTER TABLE {}", rewrite(after(1).trim()))),
        ("show", "triggers", _) => statements(SHOW_TRIGGERS.to_string()),
        ("show", "tables", _) => statements(
            "SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() ORDER BY 1"
                .to_string(),
        ),
        ("drop", "index", _) => {
            let on = (3..tokens.len()).find(|&i| word(i) == "on").unwrap_or(tokens.len());
            statements(sql[..tokens.get(on).map_or(sql.len(), |t| t.start as usize)].trim().to_string())
        }
        ("drop", "trigger", _) => {
            let exists = word(2) == "if";
            let name = text(&sql, &tokens[if exists { 4 } else { 2 }]).trim_matches('"').to_string();
            let query = format!(
                "SELECT c.relname::text FROM pg_trigger t JOIN pg_class c ON c.oid = t.tgrelid WHERE t.tgname = '{name}'"
            );
            match catalog(&query).first().and_then(|row| row.first()) {
                Some(table) => statements(format!("DROP TRIGGER \"{name}\" ON \"{table}\"")),
                None => statements(format!(
                    "DROP TRIGGER {}\"{name}\" ON pg_catalog.pg_class",
                    if exists { "IF EXISTS " } else { "" }
                )),
            }
        }
        ("create", "trigger", _) => Translated::Statements(trigger(&sql, &tokens)),
        ("alter", "table", _) => Translated::Statements(alter_table(&sql, &tokens)),
        _ if word(0) == "insert" && duplicate_key(&tokens, &sql).is_some() => {
            let at = duplicate_key(&tokens, &sql).unwrap_or(tokens.len());
            let assignments = &sql[tokens[at + 3].end as usize..];
            let updates: Vec<String> = split_top(assignments).iter().map(|a| a.trim().to_string()).collect();
            statements(upsert(&sql[..tokens[at].start as usize], &updates, catalog))
        }
        _ => statements(rewrite(&sql)),
    }
}

/// setting returns the name of a setting from the text that names it, without MySQL's `@@`, `SESSION`, or `GLOBAL`.
fn setting(text: &str) -> String {
    let name: String = text.split_whitespace().collect::<String>().to_ascii_lowercase();
    let name = name.trim_start_matches('@');
    let name = name.strip_prefix("session").or(name.strip_prefix("global")).unwrap_or(name);
    name.trim_start_matches('.').to_string()
}

/// requote rewrites MySQL's quoting into Postgres': strings in single or double quotes with backslash escapes become
/// single-quoted strings with doubled quotes, and backquoted identifiers become double-quoted ones.
pub fn requote(sql: &str) -> String {
    let mut out = String::with_capacity(sql.len());
    let mut chars = sql.chars().peekable();
    while let Some(c) = chars.next() {
        match c {
            '\'' | '"' => {
                let mut value = String::new();
                while let Some(d) = chars.next() {
                    match d {
                        '\\' => match chars.next() {
                            Some('n') => value.push('\n'),
                            Some('t') => value.push('\t'),
                            Some('0') => value.push('\0'),
                            Some(e) => value.push(e),
                            None => {}
                        },
                        _ if d == c && chars.peek() == Some(&c) => {
                            chars.next();
                            value.push(c);
                        }
                        _ if d == c => break,
                        _ => value.push(d),
                    }
                }
                out.push('\'');
                out.push_str(&value.replace('\'', "''"));
                out.push('\'');
            }
            '`' => {
                out.push('"');
                for d in chars.by_ref() {
                    if d == '`' {
                        break;
                    }
                    out.push(d);
                }
                out.push('"');
            }
            '#' => {
                for d in chars.by_ref() {
                    if d == '\n' {
                        out.push('\n');
                        break;
                    }
                }
            }
            _ => out.push(c),
        }
    }
    out
}

/// duplicate_key returns the position of an insert's `ON DUPLICATE KEY UPDATE`.
fn duplicate_key(tokens: &[ScanToken], sql: &str) -> Option<usize> {
    let word = |i: usize| tokens.get(i).map(|t| text(sql, t).to_ascii_lowercase()).unwrap_or_default();
    (0..tokens.len()).find(|&i| word(i) == "on" && word(i + 1) == "duplicate" && word(i + 2) == "key")
}

/// upsert rewrites an insert into one that updates the row with the same primary key instead, setting the columns
/// that MySQL's assignments name, or every other column without them, knowing the docs table's key since the catalogs
/// leave out the `dolt` schema's tables.
fn upsert(insert: &str, assignments: &[String], catalog: &mut Catalog<'_>) -> String {
    let insert = rewrite(insert);
    let table = pg_query::scan(&insert)
        .ok()
        .and_then(|s| s.tokens.get(2).map(|t| insert[t.start as usize..t.end as usize].to_string()))
        .unwrap_or_default();
    let query = format!(
        "SELECT a.attname::text, (array_position(i.indkey::int2[], a.attnum) IS NOT NULL)::text FROM pg_attribute a \
         LEFT JOIN pg_index i ON i.indrelid = a.attrelid AND i.indisprimary WHERE a.attrelid = '{}'::regclass \
         AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum",
        table.replace('\'', "''")
    );
    let columns = match table.eq_ignore_ascii_case("dolt_docs") {
        true => {
            vec![vec!["doc_name".to_string(), "true".to_string()], vec!["doc_text".to_string(), "false".to_string()]]
        }
        false => catalog(&query),
    };
    let keys: Vec<String> = columns.iter().filter(|c| c[1] == "true").map(|c| format!("\"{}\"", c[0])).collect();
    if keys.is_empty() {
        return insert;
    }
    let sets: Vec<String> = match assignments.is_empty() {
        true => {
            columns.iter().filter(|c| c[1] != "true").map(|c| format!("\"{0}\" = EXCLUDED.\"{0}\"", c[0])).collect()
        }
        false => assignments.iter().map(|a| rewrite(&values_to_excluded(a))).collect(),
    };
    let action = if sets.is_empty() { "NOTHING".to_string() } else { format!("UPDATE SET {}", sets.join(", ")) };
    format!("{insert} ON CONFLICT ({}) DO {action}", keys.join(", "))
}

/// values_to_excluded rewrites MySQL's `VALUES(column)` in an assignment into Postgres' `EXCLUDED.column`.
fn values_to_excluded(assignment: &str) -> String {
    let mut out = assignment.to_string();
    while let Some(at) = out.to_ascii_lowercase().find("values(") {
        let Some(close) = out[at..].find(')') else { break };
        let column = out[at + 7..at + close].trim().to_string();
        out.replace_range(at..at + close + 1, &format!("EXCLUDED.{column}"));
    }
    out
}

/// split_top splits text at the commas outside parentheses and quotes.
fn split_top(text: &str) -> Vec<String> {
    let (mut parts, mut current, mut depth, mut quote) = (Vec::new(), String::new(), 0, false);
    for c in text.chars() {
        match c {
            '\'' => quote = !quote,
            '(' if !quote => depth += 1,
            ')' if !quote => depth -= 1,
            ',' if !quote && depth == 0 => {
                parts.push(std::mem::take(&mut current));
                continue;
            }
            _ => {}
        }
        current.push(c);
    }
    parts.push(current);
    parts
}

/// alter_table rewrites MySQL's ALTER TABLE forms that Postgres spells differently: adding and dropping indexes,
/// MODIFY, and CHANGE.
fn alter_table(sql: &str, tokens: &[ScanToken]) -> Vec<String> {
    let word = |i: usize| tokens.get(i).map(|t| text(sql, t).to_ascii_lowercase()).unwrap_or_default();
    let table = tokens.get(2).map(|t| text(sql, t)).unwrap_or_default();
    let rest = |i: usize| tokens.get(i).map_or("", |t| &sql[t.start as usize..]);
    match (word(3).as_str(), word(4).as_str()) {
        ("drop", "index" | "key") => vec![format!("DROP INDEX {}", text(sql, &tokens[5]))],
        ("add", "index" | "key") => {
            vec![format!("CREATE INDEX {} ON {table} {}", text(sql, &tokens[5]), rewrite(rest(6)))]
        }
        ("add", "unique") => {
            let at = if matches!(word(5).as_str(), "index" | "key") { 6 } else { 5 };
            vec![format!("CREATE UNIQUE INDEX {} ON {table} {}", text(sql, &tokens[at]), rewrite(rest(at + 1)))]
        }
        ("modify", _) => {
            let at = if word(4) == "column" { 5 } else { 4 };
            vec![format!("ALTER TABLE {table} ALTER COLUMN {} TYPE {}", text(sql, &tokens[at]), rewrite(rest(at + 1)))]
        }
        ("change", _) => {
            let at = if word(4) == "column" { 5 } else { 4 };
            let (old, new) = (text(sql, &tokens[at]), text(sql, &tokens[at + 1]));
            let mut out = Vec::new();
            if !old.eq_ignore_ascii_case(new) {
                out.push(format!("ALTER TABLE {table} RENAME COLUMN {old} TO {new}"));
            }
            out.push(format!("ALTER TABLE {table} ALTER COLUMN {new} TYPE {}", rewrite(rest(at + 2))));
            out
        }
        _ => vec![rewrite(sql)],
    }
}

/// trigger rewrites MySQL's `CREATE TRIGGER ... FOR EACH ROW` with a statement or a BEGIN ... END block into a
/// PL/pgSQL trigger function that runs it, and a Postgres trigger that calls it for each row.
fn trigger(sql: &str, tokens: &[ScanToken]) -> Vec<String> {
    let word = |i: usize| tokens.get(i).map(|t| text(sql, t).to_ascii_lowercase()).unwrap_or_default();
    let find = |name: &str, from: usize| (from..tokens.len()).find(|&i| word(i) == name);
    let (Some(on), Some(each)) = (find("on", 3), find("each", 3)) else { return vec![rewrite(sql)] };
    let name = text(sql, &tokens[2]).trim_matches('"').to_string();
    let event = sql[tokens[2].end as usize..tokens[on].start as usize].trim();
    let table = text(sql, &tokens[on + 1]);
    let mut body = sql[tokens[each + 1].end as usize..].trim().trim_end_matches(';').to_string();
    if body.to_ascii_lowercase().starts_with("begin") && body.to_ascii_lowercase().ends_with("end") {
        body = body[5..body.len() - 3].to_string();
    }
    let statements: Vec<String> = pg_query::split_with_scanner(&body)
        .unwrap_or_else(|_| vec![body.as_str()])
        .into_iter()
        .map(|s| rewrite(s.trim()))
        .filter(|s| !s.is_empty())
        .map(|s| {
            let starts = |prefix: &str| s.get(..prefix.len()).is_some_and(|p| p.eq_ignore_ascii_case(prefix));
            match () {
                _ if starts("select") => format!("PERFORM{}", &s[6..]),
                _ if starts("insert ignore ") => format!("INSERT{} ON CONFLICT DO NOTHING", &s[13..]),
                _ if starts("set ") => s[4..].replacen('=', ":=", 1),
                _ => s,
            }
        })
        .collect();
    let function = format!("\"{name}_oracle_trigger\"");
    vec![
        format!(
            "CREATE OR REPLACE FUNCTION {function}() RETURNS trigger LANGUAGE plpgsql AS $oracle$ BEGIN {}; \
             IF TG_OP = 'DELETE' THEN RETURN OLD; END IF; RETURN NEW; END $oracle$",
            statements.join("; ")
        ),
        format!("CREATE TRIGGER \"{name}\" {event} ON {table} FOR EACH ROW EXECUTE FUNCTION {function}()"),
    ]
}

/// rewrite returns a statement or expression with MySQL's functions, types, and clauses rewritten into Postgres.
pub fn rewrite(sql: &str) -> String {
    let Ok(scanned) = pg_query::scan(sql) else { return sql.to_string() };
    let tokens = scanned.tokens;
    let word = |i: usize| tokens.get(i).map(|t| text(sql, t).to_ascii_lowercase()).unwrap_or_default();
    let is = |i: usize, kind: Token| tokens.get(i).is_some_and(|t| t.token == kind as i32);
    let definition = matches!(word(0).as_str(), "create" | "alter");
    let ignore_insert = word(0) == "insert" && word(1) == "into" && word(2) == "dolt_ignore";
    let lower = sql.to_ascii_lowercase();
    let system_tables = lower.contains("dolt_") && !lower.contains("information_schema");
    let (mut booleans, mut tables) = (Vec::new(), Vec::new());
    let key_lists: Vec<(usize, usize)> = (0..tokens.len())
        .filter(|&i| definition && matches!(word(i).as_str(), "unique" | "key") && is(i + 1, Token::Ascii40))
        .filter_map(|i| close_paren(&tokens, i + 1).map(|close| (i + 1, close)))
        .collect();
    let depth = |i: usize| {
        let opened = tokens[..i].iter().filter(|t| t.token == Token::Ascii40 as i32).count();
        opened - tokens[..i].iter().filter(|t| t.token == Token::Ascii41 as i32).count().min(opened)
    };
    let mut out = String::new();
    let mut copied = 0;
    let mut i = 0;
    while i < tokens.len() {
        let start = tokens[i].start as usize;
        let next_paren = is(i + 1, Token::Ascii40);
        let previous = word(i.saturating_sub(1));
        let replacement: Option<(String, usize)> = match word(i).as_str() {
            "char" if next_paren && !definition && previous != "as" => Some(("chr".into(), i)),
            "char" | "signed" | "unsigned" if previous == "as" && !next_paren => {
                let integer = word(i) != "char" && matches!(word(i + 1).as_str(), "integer" | "int");
                Some((if word(i) == "char" { "text" } else { "bigint" }.into(), if integer { i + 1 } else { i }))
            }
            "as" if definition && next_paren && previous != "always" && (word(0) == "alter" || depth(i) > 0) => {
                Some(("GENERATED ALWAYS AS".into(), i))
            }
            "ifnull" | "coalesce" | "group_concat" | "if" if next_paren => close_paren(&tokens, i + 1).map(|close| {
                let args: Vec<String> = split_top(&sql[tokens[i + 1].end as usize..tokens[close].start as usize])
                    .iter()
                    .map(|a| rewrite(a.trim()))
                    .collect();
                let rewritten = match word(i).as_str() {
                    "if" => format!(
                        "CASE WHEN {} THEN {} ELSE {} END",
                        args.first().cloned().unwrap_or_default(),
                        args.get(1).cloned().unwrap_or_default(),
                        args.get(2).cloned().unwrap_or_default()
                    ),
                    "group_concat" => group_concat(&args.join(", ")),
                    _ if args.iter().any(|a| a.starts_with('\'')) => {
                        let args: Vec<String> = args.iter().map(|a| format!("({a})::text")).collect();
                        format!("coalesce({})", args.join(", "))
                    }
                    _ => format!("coalesce({})", args.join(", ")),
                };
                (rewritten, close)
            }),
            "database" | "schema" if next_paren && is(i + 2, Token::Ascii41) => {
                Some(("current_schema()".into(), i + 2))
            }
            "information_schema"
                if word(i + 1) == "." && word(i + 2) == "key_column_usage" && lower.contains("referenced_") =>
            {
                let aliased = is(i + 3, Token::Ident) || word(i + 3) == "as";
                Some((
                    if aliased { KEY_COLUMN_USAGE.into() } else { format!("{KEY_COLUMN_USAGE} AS key_column_usage") },
                    i + 2,
                ))
            }
            "information_schema" if word(i + 1) == "." && word(i + 2) == "check_constraints" => {
                let aliased = is(i + 3, Token::Ident) || word(i + 3) == "as";
                let table = CHECK_CONSTRAINTS;
                Some((if aliased { table.into() } else { format!("{table} AS check_constraints") }, i + 2))
            }
            "information_schema" if word(i + 1) == "." && word(i + 2) == "statistics" => {
                let aliased = is(i + 3, Token::Ident) || word(i + 3) == "as";
                Some((if aliased { STATISTICS.into() } else { format!("{STATISTICS} AS statistics") }, i + 2))
            }
            "json_extract" | "date_format" | "timestampdiff" if next_paren => {
                close_paren(&tokens, i + 1).map(|close| {
                    let args: Vec<String> = split_top(&sql[tokens[i + 1].end as usize..tokens[close].start as usize])
                        .iter()
                        .map(|a| rewrite(a.trim()))
                        .collect();
                    let arg = |n: usize| args.get(n).cloned().unwrap_or_default();
                    let rewritten = match word(i).as_str() {
                        "json_extract" => {
                            format!("jsonb_path_query_first(({})::jsonb, ({})::jsonpath)", arg(0), arg(1))
                        }
                        "date_format" => format!("to_char({}, {})", arg(0), date_format(&arg(1))),
                        _ => {
                            let seconds =
                                format!("extract(epoch FROM ({})::timestamp - ({})::timestamp)", arg(2), arg(1));
                            match arg(0).to_ascii_lowercase().as_str() {
                                "minute" => format!("floor({seconds} / 60)::bigint"),
                                "hour" => format!("floor({seconds} / 3600)::bigint"),
                                "day" => format!("floor({seconds} / 86400)::bigint"),
                                _ => format!("floor({seconds})::bigint"),
                            }
                        }
                    };
                    (rewritten, close)
                })
            }
            "hex" if next_paren => close_paren(&tokens, i + 1).map(|close| {
                let argument = rewrite(&sql[tokens[i + 1].end as usize..tokens[close].start as usize]);
                (format!("upper(encode(({argument})::text::bytea, 'hex'))"), close)
            }),
            "dual" if previous == "from" => Some((String::new(), i)),
            "from" if word(i + 1) == "dual" => Some((String::new(), i + 1)),
            "limit" if is(i + 1, Token::Iconst) && word(i + 2) == "," && is(i + 3, Token::Iconst) => {
                Some((format!("LIMIT {} OFFSET {}", text(sql, &tokens[i + 3]), text(sql, &tokens[i + 1])), i + 3))
            }
            "tinyint" | "int" | "integer" | "bigint" | "smallint" | "mediumint" if definition => {
                let display = next_paren && is(i + 2, Token::Iconst) && is(i + 3, Token::Ascii41);
                let boolean = word(i) == "tinyint" && display && word(i + 2) == "1";
                let current = word(i);
                let name = match current.as_str() {
                    _ if boolean => "boolean",
                    "tinyint" | "smallint" => "smallint",
                    "mediumint" => "integer",
                    other => other,
                };
                let last = if display { i + 3 } else { i };
                let last = if word(last + 1) == "unsigned" { last + 1 } else { last };
                Some((name.to_string(), last))
            }
            "unsigned" | "zerofill" if definition => Some((String::new(), i)),
            "datetime" if definition || previous == "as" => Some(("timestamp".into(), i)),
            "double" if (definition || previous == "as") && word(i + 1) != "precision" => {
                Some(("double precision".into(), i))
            }
            "blob" | "longblob" | "mediumblob" | "tinyblob" | "varbinary" | "binary"
                if definition || previous == "as" =>
            {
                let last = if next_paren { close_paren(&tokens, i + 1).unwrap_or(i) } else { i };
                Some(("bytea".into(), last))
            }
            "longtext" | "mediumtext" | "tinytext" if definition => Some(("text".into(), i)),
            "enum" if definition && next_paren => close_paren(&tokens, i + 1).map(|close| ("text".to_string(), close)),
            "auto_increment" if definition && previous != "=" => Some(("GENERATED BY DEFAULT AS IDENTITY".into(), i)),
            "engine" | "auto_increment" | "charset" | "collate" if definition && word(i + 1) == "=" => {
                Some((String::new(), i + 2))
            }
            "default" if definition && word(i + 1) == "charset" => Some((String::new(), i + 2)),
            "collate" if definition => Some((String::new(), i + 1)),
            "character" if definition && word(i + 1) == "set" => Some((String::new(), i + 2)),
            name if BOOL_COLUMNS.contains(&name) => {
                let last = match word(i + 1).as_str() {
                    "=" | "<>" | "!=" => i + 2,
                    "in" if is(i + 2, Token::Ascii40) => close_paren(&tokens, i + 2).unwrap_or(i),
                    _ => i,
                };
                booleans.extend((i + 1..=last).filter(|&j| is(j, Token::Iconst)));
                None
            }
            "table_name" | "from_table_name" | "to_table_name" if system_tables => {
                let last = match word(i + 1).as_str() {
                    "=" | "<>" | "!=" => i + 2,
                    "in" if is(i + 2, Token::Ascii40) => close_paren(&tokens, i + 2).unwrap_or(i),
                    _ => i,
                };
                tables.extend((i + 1..=last).filter(|&j| is(j, Token::Sconst) && !text(sql, &tokens[j]).contains('.')));
                None
            }
            _ if tables.contains(&i) => Some((format!("'public.{}", &text(sql, &tokens[i])[1..]), i)),
            "0" | "1" if (ignore_insert || booleans.contains(&i)) && is(i, Token::Iconst) => {
                Some((if word(i) == "1" { "true" } else { "false" }.into(), i))
            }
            "<=>" => Some(("IS NOT DISTINCT FROM".into(), i)),
            "force" | "use" | "ignore"
                if matches!(word(i + 1).as_str(), "index" | "key") && is(i + 2, Token::Ascii40) =>
            {
                close_paren(&tokens, i + 2).map(|close| (String::new(), close))
            }
            _ if is(i, Token::Xconst) => {
                let text = text(sql, &tokens[i]);
                Some((format!("'\\x{}'::bytea", &text[2..text.len() - 1]), i))
            }
            "asc" | "desc" if key_lists.iter().any(|&(open, close)| open < i && i < close) => Some((String::new(), i)),
            "column_type" => Some(("data_type".into(), i)),
            _ if text(sql, &tokens[i]).starts_with("@@") => {
                Some((format!("current_setting('{}')", setting(&word(i))), i))
            }
            _ => None,
        };
        match replacement {
            Some((rewritten, last)) => {
                out.push_str(&sql[copied..start]);
                out.push_str(&rewritten);
                copied = tokens[last].end as usize;
                i = last + 1;
            }
            None => i += 1,
        }
    }
    out.push_str(&sql[copied..]);
    out
}

/// date_format rewrites a MySQL DATE_FORMAT pattern literal into the to_char pattern that writes the same text.
fn date_format(pattern: &str) -> String {
    let mut out = String::new();
    let mut chars = pattern.chars();
    while let Some(c) = chars.next() {
        if c != '%' {
            out.push(c);
            continue;
        }
        out.push_str(match chars.next() {
            Some('Y') => "YYYY",
            Some('m') => "MM",
            Some('d') => "DD",
            Some('H') => "HH24",
            Some('i') => "MI",
            Some('s' | 'S') => "SS",
            Some('f') => "US",
            Some(other) => {
                out.push(other);
                continue;
            }
            None => "",
        });
    }
    out
}

/// group_concat rewrites the arguments of MySQL's group_concat(), with its ORDER BY and SEPARATOR clauses, into
/// Postgres' string_agg().
fn group_concat(arguments: &str) -> String {
    let lower = arguments.to_ascii_lowercase();
    let (body, separator) = match lower.rfind(" separator ") {
        Some(at) => (&arguments[..at], arguments[at + 11..].trim().to_string()),
        None => (arguments, "','".to_string()),
    };
    let lower = body.to_ascii_lowercase();
    let (value, order) = match lower.find(" order by ") {
        Some(at) => (&body[..at], format!(" {}", &body[at + 1..])),
        None => (body, String::new()),
    };
    let value = value.trim();
    match value.get(..9).filter(|p| p.eq_ignore_ascii_case("distinct ")) {
        Some(_) => format!("string_agg(DISTINCT ({})::text, {separator}{order})", &value[9..]),
        None => format!("string_agg(({value})::text, {separator}{order})"),
    }
}

/// text returns a token's text.
fn text<'s>(sql: &'s str, token: &ScanToken) -> &'s str {
    &sql[token.start as usize..token.end as usize]
}

/// close_paren returns the position of the token that closes the parenthesis at a position.
fn close_paren(tokens: &[ScanToken], open: usize) -> Option<usize> {
    let mut depth = 0;
    for (i, token) in tokens.iter().enumerate().skip(open) {
        if token.token == Token::Ascii40 as i32 {
            depth += 1;
        } else if token.token == Token::Ascii41 as i32 {
            depth -= 1;
            if depth == 0 {
                return Some(i);
            }
        }
    }
    None
}
