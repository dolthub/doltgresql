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

//! Parsing with Postgres' own grammar, plus the Doltgres-only syntax that it rejects. A query that Postgres' grammar
//! accepts is never reinterpreted, and one that neither accepts fails with Postgres' syntax error.

use std::ops::Range;

use pg_query::protobuf::{ScanToken, Token};
use pg_query::{Node, NodeEnum, NodeRef};

use crate::error::{PgError, Result, code};

/// Statement is one parsed statement.
#[derive(Clone, Debug, PartialEq)]
pub enum Statement {
    /// A statement of Postgres' grammar, with the Doltgres-only clauses that were cut from its text.
    Postgres { node: NodeEnum, extras: Extras },
    /// `USE` switches the session to a database, or to a revision of one written as `database/revision`.
    Use(String),
    /// `SET` of a configuration parameter to an expression, which Postgres' grammar limits to constants.
    SetExpression { name: String, local: bool, value: Node },
    /// `DESCRIBE`, `DESC`, or `EXPLAIN` of a table, with the table's `AS OF` revision in its extras.
    Describe { relation: pg_query::protobuf::RangeVar, extras: Extras },
    /// `SHOW CREATE TABLE` of a table, with the table's `AS OF` revision in its extras.
    ShowCreateTable { relation: pg_query::protobuf::RangeVar, extras: Extras },
    /// `SHOW TABLES`, `SEQUENCES`, `SCHEMAS`, `DATABASES`, or `INDEXES`, with the parts of the name after `FROM`.
    Listing { kind: String, from: Vec<String> },
}

/// LISTINGS are the kinds of objects that a `SHOW` lists.
pub const LISTINGS: [&str; 5] = ["tables", "sequences", "schemas", "databases", "indexes"];

/// Extras are the Doltgres-only clauses of a Postgres statement.
#[derive(Clone, Debug, Default, PartialEq)]
pub struct Extras {
    /// Whether `IF NOT EXISTS` followed `CREATE DATABASE`, `CREATE USER`, or `CREATE ROLE`.
    pub if_not_exists: bool,
    /// The `AS OF` revisions of tables, each with the location of the table it follows.
    pub as_of: Vec<(i32, Node)>,
    /// The statement's source text.
    pub text: String,
}

/// parse parses the statements of a query.
pub fn parse(query: &str) -> Result<Vec<Statement>> {
    match pg_query::parse_with_cursor(query) {
        Ok(result) => Ok(postgres_statements(query, result, Extras::default())),
        Err((err, cursor, state)) => extended(query).ok_or_else(|| syntax_error(err, cursor, &state)),
    }
}

/// postgres_statements returns the statements of a parse result of the query, each with the extras.
fn postgres_statements(query: &str, result: pg_query::ParseResult, extras: Extras) -> Vec<Statement> {
    result
        .protobuf
        .stmts
        .into_iter()
        .filter_map(|raw| {
            let start = (raw.stmt_location.max(0) as usize).min(query.len());
            let end = if raw.stmt_len == 0 { query.len() } else { (start + raw.stmt_len as usize).min(query.len()) };
            let text = query.get(start..end).unwrap_or_default().trim().to_string();
            let node = raw.stmt.and_then(|stmt| stmt.node)?;
            if let NodeEnum::VariableShowStmt(show) = &node
                && LISTINGS.contains(&show.name.as_str())
            {
                return Some(Statement::Listing { kind: show.name.clone(), from: Vec::new() });
            }
            Some(Statement::Postgres { node, extras: Extras { text, ..extras.clone() } })
        })
        .collect()
}

/// expression_node parses the text of one expression, such as a stored default.
pub fn expression_node(text: &str) -> Result<Node> {
    expression(&format!("{}{text}", " ".repeat(7)), 7..7 + text.len())
        .ok_or_else(|| PgError::new(code::SYNTAX_ERROR, format!("invalid expression: {text}")))
}

/// GRAMMAR_CODES are the error codes that Postgres' grammar raises besides syntax errors.
const GRAMMAR_CODES: &[&str] = &[
    code::DUPLICATE_OBJECT,
    code::FEATURE_NOT_SUPPORTED,
    code::INVALID_ESCAPE_SEQUENCE,
    code::INVALID_PARAMETER_VALUE,
    code::NONSTANDARD_USE_OF_ESCAPE_CHARACTER,
    code::RESERVED_NAME,
    code::WINDOWING_ERROR,
];

/// syntax_error converts a parser error, with the 1-based character position and the SQLSTATE it reported, to
/// Postgres' error.
pub fn syntax_error(err: pg_query::Error, cursor: i32, state: &str) -> PgError {
    let message = match err {
        pg_query::Error::Parse(message) => message,
        other => other.to_string(),
    };
    let code = GRAMMAR_CODES.iter().find(|c| **c == state).copied().unwrap_or(code::SYNTAX_ERROR);
    PgError { position: u32::try_from(cursor).ok().filter(|&p| p > 0), ..PgError::new(code, message) }
}

/// extended parses a query that has Doltgres-only syntax, returning None when some statement is in neither grammar.
fn extended(query: &str) -> Option<Vec<Statement>> {
    let tokens = pg_query::scan(query).ok()?.tokens;
    let mut statements = Vec::new();
    for piece in tokens.split(|t| t.token == Token::Ascii59 as i32).filter(|piece| !piece.is_empty()) {
        statements.extend(extended_statement(query, piece)?);
    }
    Some(statements)
}

/// extended_statement parses one statement of a query from its tokens.
fn extended_statement(query: &str, tokens: &[ScanToken]) -> Option<Vec<Statement>> {
    let range = tokens[0].start as usize..tokens[tokens.len() - 1].end as usize;
    let isolated = isolate(query, range.clone());
    if let Ok(result) = pg_query::parse(&isolated) {
        return Some(postgres_statements(&isolated, result, Extras::default()));
    }
    let words = Words { query, tokens };
    if words.keyword(0, "use") {
        return use_statement(&words).map(|target| vec![Statement::Use(target)]);
    }
    if words.keyword(0, "set") {
        return set_expression(&words).map(|statement| vec![statement]);
    }
    if words.keyword(0, "show")
        && let Some(statement) = listing(&words)
    {
        return Some(vec![statement]);
    }
    if words.keyword(0, "show")
        && words.keyword(1, "create")
        && words.keyword(2, "table")
        && let Some((relation, extras)) = table_reference(query, &words, 3)
    {
        return Some(vec![Statement::ShowCreateTable { relation, extras }]);
    }
    if ["describe", "desc", "explain"].iter().any(|word| words.keyword(0, word))
        && let Some((relation, extras)) = table_reference(query, &words, 1)
    {
        return Some(vec![Statement::Describe { relation, extras }]);
    }
    cut_statement(query, range, &words)
}

/// listing parses `SHOW kind {FROM | IN} name[.name]`.
fn listing(words: &Words<'_>) -> Option<Statement> {
    let kind = words.name(1).filter(|kind| LISTINGS.contains(&kind.as_str()) && kind != "databases")?;
    if !(words.keyword(2, "from") || words.keyword(2, "in")) {
        return None;
    }
    let mut from = vec![words.name(3)?];
    let mut index = 4;
    while words.kind(index) == Token::Ascii46 as i32 {
        from.push(words.name(index + 1)?);
        index += 2;
    }
    (index == words.tokens.len()).then_some(Statement::Listing { kind, from })
}

/// table_reference parses the rest of a statement from a token as a table with an optional `AS OF`, by reading it as
/// `TABLE` of the table, written over the tokens before it and moving the locations back to the query's.
fn table_reference(query: &str, words: &Words<'_>, first: usize) -> Option<(pg_query::protobuf::RangeVar, Extras)> {
    let rest = words.tokens.get(first)?.start as usize;
    let end = words.tokens.last()?.end as usize;
    let shift = 6usize.saturating_sub(rest);
    let mut text = " ".repeat(rest + shift - 6);
    text.push_str("TABLE ");
    text.push_str(&query[rest..end]);
    let tokens = pg_query::scan(&text).ok()?.tokens;
    let statements = extended_statement(&text, &tokens)?;
    let [Statement::Postgres { node: NodeEnum::SelectStmt(select), extras }] = statements.as_slice() else {
        return None;
    };
    let [from] = select.from_clause.as_slice() else { return None };
    let Some(NodeEnum::RangeVar(relation)) = &from.node else { return None };
    let back = |location: i32| location - shift as i32;
    let relation = pg_query::protobuf::RangeVar { location: back(relation.location), ..relation.clone() };
    let as_of = extras.as_of.iter().map(|(location, revision)| (back(*location), revision.clone())).collect();
    let text = query[words.tokens[0].start as usize..end].to_string();
    Some((relation, Extras { as_of, text, ..extras.clone() }))
}

/// isolate returns the query with everything outside the range replaced by spaces, so that the locations the parser
/// reports stay locations in the query.
fn isolate(query: &str, range: Range<usize>) -> String {
    let mut text = " ".repeat(range.start);
    text.push_str(&query[range]);
    text
}

/// Words gives keyword and identifier views of a statement's tokens.
struct Words<'a> {
    query: &'a str,
    tokens: &'a [ScanToken],
}

impl Words<'_> {
    /// text returns the source text of a token.
    fn text(&self, index: usize) -> &str {
        self.tokens.get(index).map_or("", |t| &self.query[t.start as usize..t.end as usize])
    }

    /// kind returns the token kind of a token, or Nul past the end.
    fn kind(&self, index: usize) -> i32 {
        self.tokens.get(index).map_or(Token::Nul as i32, |t| t.token)
    }

    /// keyword reports whether a token is the word, as a keyword or an unquoted identifier.
    fn keyword(&self, index: usize, word: &str) -> bool {
        self.tokens.get(index).is_some_and(|t| t.keyword_kind != 0 || t.token == Token::Ident as i32)
            && self.text(index).eq_ignore_ascii_case(word)
    }

    /// name returns a token as a name: an identifier or keyword folded to lowercase, or a quoted identifier.
    fn name(&self, index: usize) -> Option<String> {
        let text = self.text(index);
        if self.kind(index) == Token::Ident as i32 && text.starts_with('"') {
            return Some(text[1..text.len() - 1].replace("\"\"", "\""));
        }
        let token = self.tokens.get(index)?;
        (token.token == Token::Ident as i32 || token.keyword_kind != 0).then(|| text.to_ascii_lowercase())
    }

    /// string returns the value of a string constant token in standard syntax.
    fn string(&self, index: usize) -> Option<String> {
        let text = self.text(index);
        (self.kind(index) == Token::Sconst as i32 && text.starts_with('\''))
            .then(|| text[1..text.len() - 1].replace("''", "'"))
    }
}

/// use_statement parses `USE target`, where the target is a string, a name, or two names joined by a slash.
fn use_statement(words: &Words<'_>) -> Option<String> {
    match words.tokens.len() {
        2 => words.string(1).or_else(|| words.name(1)),
        4 if words.kind(2) == Token::Ascii47 as i32 => Some(format!("{}/{}", words.name(1)?, words.name(3)?)),
        _ => None,
    }
}

/// set_expression parses `SET [SESSION | LOCAL] name {= | TO} expression`.
fn set_expression(words: &Words<'_>) -> Option<Statement> {
    let mut index = 1;
    let local = words.keyword(index, "local");
    if local || words.keyword(index, "session") {
        index += 1;
    }
    let mut name = words.name(index)?;
    index += 1;
    while words.kind(index) == Token::Ascii46 as i32 {
        name = format!("{name}.{}", words.name(index + 1)?);
        index += 2;
    }
    if !(words.text(index) == "=" || words.keyword(index, "to")) {
        return None;
    }
    let start = words.tokens.get(index + 1)?.start as usize;
    let end = words.tokens[words.tokens.len() - 1].end as usize;
    let value = expression(words.query, start..end)?;
    Some(Statement::SetExpression { name, local, value })
}

/// expression parses the text in the range as one expression, keeping its locations in the query.
fn expression(query: &str, range: Range<usize>) -> Option<Node> {
    let mut text = " ".repeat(range.start.checked_sub(7)?);
    text.push_str("SELECT ");
    text.push_str(&query[range]);
    let mut result = pg_query::parse(&text).ok()?;
    if result.protobuf.stmts.len() != 1 {
        return None;
    }
    let Some(NodeEnum::SelectStmt(select)) = result.protobuf.stmts.pop()?.stmt?.node else { return None };
    let only_targets = pg_query::protobuf::SelectStmt {
        target_list: select.target_list.clone(),
        limit_option: select.limit_option,
        op: select.op,
        ..Default::default()
    };
    if *select != only_targets || select.target_list.len() != 1 {
        return None;
    }
    let Some(NodeEnum::ResTarget(target)) = select.target_list.into_iter().next()?.node else { return None };
    if !target.name.is_empty() {
        return None;
    }
    target.val.map(|value| *value)
}

/// cut_statement parses a statement after cutting the Doltgres-only clauses out of its text, and quoting a role name
/// written as a string.
fn cut_statement(query: &str, range: Range<usize>, words: &Words<'_>) -> Option<Vec<Statement>> {
    let mut text = isolate(query, range);
    let blank = |text: &mut String, span: Range<usize>| text.replace_range(span.clone(), &" ".repeat(span.len()));
    let mut extras = Extras::default();
    let mut cuts = Vec::new();
    if words.keyword(0, "create")
        && ["database", "user", "role"].iter().any(|kind| words.keyword(1, kind))
        && words.keyword(2, "if")
        && words.keyword(3, "not")
        && words.keyword(4, "exists")
    {
        blank(&mut text, words.tokens[2].start as usize..words.tokens[4].end as usize);
        extras.if_not_exists = true;
        if !words.keyword(1, "database")
            && let Some(name) = words.string(5)
            && !name.contains(['"', '\''])
        {
            let token = &words.tokens[5];
            text.replace_range(token.start as usize..token.end as usize, &format!("\"{name}\""));
        }
    }
    let mut index = 0;
    while index + 1 < words.tokens.len() {
        if !(words.keyword(index, "as") && words.keyword(index + 1, "of")) {
            index += 1;
            continue;
        }
        let mut first = index + 2;
        if words.keyword(first, "system") && words.keyword(first + 1, "time") {
            first += 2;
        }
        let last = as_of_end(words, first)?;
        let span = words.tokens[first].start as usize..words.tokens[last].end as usize;
        cuts.push((words.tokens[index].start, expression(query, span)?));
        blank(&mut text, words.tokens[index].start as usize..words.tokens[last].end as usize);
        index = last + 1;
    }
    if !extras.if_not_exists && cuts.is_empty() {
        return None;
    }
    let result = pg_query::parse(&text).ok()?;
    let tables: Vec<i32> = result
        .protobuf
        .nodes()
        .into_iter()
        .filter_map(|(node, ..)| match node {
            NodeRef::RangeVar(table) => Some(table.location),
            _ => None,
        })
        .collect();
    for (start, revision) in cuts {
        let table = tables.iter().copied().filter(|&location| location < start).max()?;
        extras.as_of.push((table, revision));
    }
    Some(postgres_statements(&text, result, extras))
}

/// as_of_end returns the index of the last token of an `AS OF` revision that starts at the token: a string, a typed
/// string such as `TIMESTAMP '...'`, or a function call.
fn as_of_end(words: &Words<'_>, first: usize) -> Option<usize> {
    if words.kind(first) == Token::Sconst as i32 {
        return Some(first);
    }
    words.name(first)?;
    if words.kind(first + 1) == Token::Sconst as i32 {
        return Some(first + 1);
    }
    let mut index = first + 1;
    while words.kind(index) == Token::Ascii46 as i32 && words.name(index + 1).is_some() {
        index += 2;
    }
    if words.kind(index) != Token::Ascii40 as i32 {
        return None;
    }
    let mut depth = 0;
    for (offset, token) in words.tokens[index..].iter().enumerate() {
        if token.token == Token::Ascii40 as i32 {
            depth += 1;
        } else if token.token == Token::Ascii41 as i32 {
            depth -= 1;
            if depth == 0 {
                return Some(index + offset);
            }
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    /// error returns the error that parsing the query fails with.
    fn error(query: &str) -> PgError {
        parse(query).expect_err(query)
    }

    #[test]
    fn syntax_errors_report_postgres_positions() {
        for (query, message, position) in [
            ("SELECT 1 +", "syntax error at end of input", 11),
            ("SELECT 1 +   ", "syntax error at end of input", 14),
            ("SELECT 'é' FROM WHERE", r#"syntax error at or near "WHERE""#, 17),
            ("SELECT 1; SELEC 2", r#"syntax error at or near "SELEC""#, 11),
            ("SELECT 1 NOT xy", r#"syntax error at or near "xy""#, 14),
            ("SELECT \"abc", r#"unterminated quoted identifier at or near ""abc""#, 8),
            ("SELECT \"\"", r#"zero-length delimited identifier at or near """""#, 8),
            ("DESC t1", r#"syntax error at or near "DESC""#, 1),
            ("USE a b", r#"syntax error at or near "USE""#, 1),
            ("SET CONSTRAINTS a.b.c.d IMMEDIATE;", "improper qualified name (too many dotted names): a.b.c.d", 17),
        ] {
            let err = error(query);
            assert_eq!((err.code, err.message.as_str(), err.position), (code::SYNTAX_ERROR, message, Some(position)));
        }
        let err = error("CREATE TABLE b6 (x INTEGER, CHECK (x > 0) DEFERRABLE);");
        assert_eq!((err.message.as_str(), err.position), ("CHECK constraints cannot be marked DEFERRABLE", None));
    }

    #[test]
    fn use_takes_a_database_or_revision() {
        for (query, target) in [
            ("USE test", "test"),
            ("use Test;", "test"),
            ("USE test/b1", "test/b1"),
            ("USE \"test/main\"", "test/main"),
            ("USE 'test/tag1'", "test/tag1"),
            ("use \"mydb/main~\"", "mydb/main~"),
        ] {
            assert_eq!(parse(query).unwrap(), vec![Statement::Use(target.into())], "{query}");
        }
    }

    #[test]
    fn if_not_exists_is_cut_from_create_database_and_roles() {
        let statements = parse("SELECT 1; create database if not exists mydb").unwrap();
        let Statement::Postgres { node: NodeEnum::CreatedbStmt(create), extras } = &statements[1] else {
            panic!("{statements:?}")
        };
        assert_eq!((create.dbname.as_str(), extras.if_not_exists), ("mydb", true));
        let statements = parse("create user if not exists 'auth_test' with superuser password 'p';").unwrap();
        let Statement::Postgres { node: NodeEnum::CreateRoleStmt(create), extras } = &statements[0] else {
            panic!("{statements:?}")
        };
        assert_eq!((create.role.as_str(), create.options.len(), extras.if_not_exists), ("auth_test", 2, true));
    }

    #[test]
    fn as_of_attaches_to_the_table_it_follows() {
        let query = "SELECT * FROM test AS OF 'HEAD~3' t1 join s.test2 AS OF GREATEST('a', 'b') AS t2 on t1.a = t2.b";
        let statements = parse(query).unwrap();
        let Statement::Postgres { extras, .. } = &statements[0] else { panic!("{statements:?}") };
        let tables: Vec<i32> = extras.as_of.iter().map(|(table, _)| *table).collect();
        assert_eq!(tables, vec![14, 42]);
        let Some(NodeEnum::AConst(revision)) = &extras.as_of[0].1.node else { panic!("{extras:?}") };
        assert_eq!(revision.location, 25);
        let Some(NodeEnum::MinMaxExpr(revision)) = &extras.as_of[1].1.node else { panic!("{extras:?}") };
        assert_eq!(revision.location, 56);
        assert!(parse("SELECT * FROM test AS OF SYSTEM TIME 'HEAD~3' join test2 AS t2 on test.a = t2.b").is_ok());
    }

    #[test]
    fn set_takes_an_expression() {
        let statements = parse("SET doltgres_enginetest.Commit1 = (select hashof('HEAD'))").unwrap();
        let Statement::SetExpression { name, local, value } = &statements[0] else { panic!("{statements:?}") };
        assert_eq!((name.as_str(), *local), ("doltgres_enginetest.commit1", false));
        assert!(matches!(value.node, Some(NodeEnum::SubLink(_))));
        assert!(parse("set myvar.var_value to (select 'a')").is_ok());
        assert!(parse("SET x = 1 FROM t").is_err());
    }
}
