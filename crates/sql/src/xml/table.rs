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

//! XMLTABLE: the rows that an XPath expression finds in an xml value, with columns that more expressions compute.

use std::collections::HashMap;

use super::xpath::{self, Document, Value as XValue};
use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::expr::Expr;
use crate::query::Ctx;
use crate::types::Value;

/// XmlColumn is a column of XMLTABLE: its name, its type, its path expression or None for the ordinality column, its
/// default, and whether it may not be NULL.
#[derive(Clone, Debug, PartialEq)]
pub struct XmlColumn {
    pub name: String,
    pub ty: ColumnType,
    pub path: Option<Expr>,
    pub default: Option<Expr>,
    pub not_null: bool,
}

/// XmlTable is a planned XMLTABLE: the document, the row expression, the namespaces with their URI expressions,
/// and the columns.
#[derive(Clone, Debug, PartialEq)]
pub struct XmlTable {
    pub document: Expr,
    pub row: Expr,
    pub namespaces: Vec<(String, Expr)>,
    pub columns: Vec<XmlColumn>,
}

/// text_of evaluates an expression to text, failing with an error when it is NULL.
fn text_of(ctx: &mut Ctx<'_>, expr: &Expr, null_error: impl FnOnce() -> PgError) -> Result<String> {
    match expr.eval(ctx, &[])? {
        Value::Null => Err(null_error()),
        Value::Text(s) | Value::Xml(s) => Ok(s),
        other => Ok(other.output().unwrap_or_default()),
    }
}

/// column_text returns the text an XPath result gives a column of a type, or None for NULL, as Postgres'
/// XmlTableGetValue does.
fn column_text(result: XValue, document: &Document, ty: ColumnType) -> Result<Option<String>> {
    let xml = ty.oid == crate::oid::XML;
    Ok(match result {
        XValue::Nodes(nodes) if nodes.is_empty() => None,
        XValue::Nodes(nodes) if xml => Some(nodes.iter().map(|&n| super::node_text(document, n)).collect()),
        XValue::Nodes(nodes) => {
            if nodes.len() > 1 {
                return Err(PgError::new(
                    code::CARDINALITY_VIOLATION,
                    "more than one value returned by column XPath expression",
                ));
            }
            Some(document.string_value(nodes[0]))
        }
        XValue::String(s) => Some(if xml { super::escape(&s) } else { s }),
        XValue::Bool(b) => {
            let numeric = crate::catalog::builtin_type(ty.oid).is_some_and(|t| t.definition.typ_category == b"N");
            Some(if numeric { (b as u8).to_string() } else { b.to_string() })
        }
        XValue::Number(n) => Some(xpath::number_text(n)),
    })
}

/// rows returns the rows of an XMLTABLE, as Postgres' table function execution with the XML routine computes them.
pub fn rows(ctx: &mut Ctx<'_>, table: &XmlTable) -> Result<Vec<Vec<Value>>> {
    let text = match table.document.eval(ctx, &[])? {
        Value::Null => return Ok(Vec::new()),
        Value::Xml(s) | Value::Text(s) => s,
        other => other.output().unwrap_or_default(),
    };
    let document = super::parse_for_xpath(&text)?;
    let mut namespaces = HashMap::new();
    for (prefix, uri) in &table.namespaces {
        let uri = text_of(ctx, uri, || PgError::new(code::NULL_VALUE_NOT_ALLOWED, "namespace URI must not be null"))?;
        namespaces.insert(prefix.clone(), uri);
    }
    let row_path = text_of(ctx, &table.row, || {
        PgError::new(code::NULL_VALUE_NOT_ALLOWED, "row filter expression must not be null")
    })?;
    if row_path.is_empty() {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_XQUERY, "row path filter must not be empty string"));
    }
    let row_query = super::Query { expr: super::compile(&row_path)?, namespaces: namespaces.clone() };
    let mut column_queries = Vec::with_capacity(table.columns.len());
    for column in &table.columns {
        let Some(path) = &column.path else {
            column_queries.push(None);
            continue;
        };
        let path = text_of(ctx, path, || PgError {
            detail: Some(format!("Filter for column \"{}\" is null.", column.name)),
            ..PgError::new(code::NULL_VALUE_NOT_ALLOWED, "column filter expression must not be null")
        })?;
        if path.is_empty() {
            return Err(PgError::new(code::INVALID_ARGUMENT_FOR_XQUERY, "column path filter must not be empty string"));
        }
        column_queries.push(Some(super::compile(&path)?));
    }
    let row_nodes = match super::evaluate(&row_query, &document)? {
        XValue::Nodes(nodes) => nodes,
        _ => Vec::new(),
    };
    let mut out = Vec::with_capacity(row_nodes.len());
    for (ordinal, &node) in row_nodes.iter().enumerate() {
        let mut row = Vec::with_capacity(table.columns.len());
        for (column, query) in table.columns.iter().zip(&column_queries) {
            let Some(query) = query else {
                row.push(Value::Int4(ordinal as i32 + 1));
                continue;
            };
            let result = xpath::evaluate(query, &document, node, &namespaces).map_err(|err| PgError {
                detail: Some(err.0),
                ..PgError::new(code::INVALID_ARGUMENT_FOR_XQUERY, "could not create XPath object")
            })?;
            let mut value = match column_text(result, &document, column.ty)? {
                Some(text) => crate::cast::cast_value(Value::Text(text), column.ty, false)?,
                None => Value::Null,
            };
            if value.is_null()
                && let Some(default) = &column.default
            {
                value = default.eval(ctx, &[])?;
            }
            if value.is_null() && column.not_null {
                return Err(PgError::new(
                    code::NULL_VALUE_NOT_ALLOWED,
                    format!("null is not allowed in column \"{}\"", column.name),
                ));
            }
            row.push(value);
        }
        out.push(row);
    }
    Ok(out)
}
