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

//! COPY, which moves rows between tables and Postgres' text, CSV, and binary copy formats.

use pg_query::NodeEnum;
use pg_query::protobuf::{CopyStmt, RangeVar};

use crate::auth::Object;
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};
use crate::expr::{node_name, position};
use crate::plan::Planner;
use crate::query::{Ctx, column, scan};
use crate::types::Value;
use crate::{Column, Outcome};

/// SIGNATURE starts every file of the binary copy format.
const SIGNATURE: &[u8] = b"PGCOPY\n\xff\r\n\0";

/// MAX_DISPLAY is the most bytes of a line or value that an error's context shows, as Postgres'
/// MAX_COPY_DATA_DISPLAY sets.
const MAX_DISPLAY: usize = 100;

/// Format is a copy format.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Format {
    Text,
    Csv,
    Binary,
}

/// Options are the options of a COPY statement.
#[derive(Clone, Debug)]
pub struct Options {
    pub format: Format,
    pub delimiter: u8,
    pub null: String,
    pub header: bool,
    pub quote: u8,
    pub escape: u8,
    /// Whether CSV output quotes every value that is not NULL.
    pub force_quote_all: bool,
    /// The columns whose values CSV output quotes when they are not NULL.
    pub force_quote: Vec<String>,
    /// The columns whose CSV input reads the NULL string as itself.
    pub force_not_null: Vec<String>,
    /// The columns whose CSV input reads the quoted NULL string as NULL.
    pub force_null: Vec<String>,
}

/// CopyFrom is a COPY FROM STDIN waiting for its data.
#[derive(Clone, Debug)]
pub struct CopyFrom {
    pub table: TableDef,
    pub columns: Vec<usize>,
    pub options: Options,
}

/// context returns the context of an error at a line of the copy data of a table.
pub fn context(table: &str, line: usize) -> String {
    format!("COPY {table}, line {line}")
}

/// with_context sets the context of an error.
pub fn with_context(err: PgError, context: String) -> PgError {
    let mut objects = err.objects.map(|o| *o).unwrap_or_default();
    objects.where_ = Some(context);
    PgError { objects: Some(Box::new(objects)), ..err }
}

/// display returns the text that an error's context shows for some copy data, which is cut short after
/// `MAX_DISPLAY` bytes.
fn display(bytes: &[u8]) -> String {
    let text = String::from_utf8_lossy(bytes);
    if text.len() <= MAX_DISPLAY {
        return text.into_owned();
    }
    let mut end = MAX_DISPLAY;
    while !text.is_char_boundary(end) {
        end -= 1;
    }
    format!("{}...", &text[..end])
}

/// options reads the options of a COPY statement and checks them as Postgres' ProcessCopyOptions does.
fn options(stmt: &CopyStmt) -> Result<Options> {
    let mut seen: Vec<&str> = Vec::new();
    let mut format = Format::Text;
    let (mut delimiter, mut null, mut header, mut quote, mut escape) = (None, None, false, None, None);
    let (mut force_quote_all, mut force_quote) = (false, None);
    let (mut force_not_null, mut force_null) = (None, None);
    for option in &stmt.options {
        let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { continue };
        let located = |error: &'static str, message: String| PgError {
            position: position(def.location),
            ..PgError::new(error, message)
        };
        if seen.contains(&def.defname.as_str()) {
            return Err(located(code::SYNTAX_ERROR, "conflicting or redundant options".into()));
        }
        seen.push(&def.defname);
        let arg = def.arg.as_deref().and_then(|a| a.node.as_ref());
        let text = match arg {
            Some(NodeEnum::String(s)) => s.sval.clone(),
            Some(NodeEnum::Integer(i)) => i.ival.to_string(),
            Some(NodeEnum::Boolean(b)) => b.boolval.to_string(),
            _ => String::new(),
        };
        let names: Vec<String> = match arg {
            Some(NodeEnum::List(list)) => list.items.iter().filter_map(node_name).map(str::to_string).collect(),
            _ => Vec::new(),
        };
        match def.defname.as_str() {
            "format" => {
                format = match text.as_str() {
                    "text" => Format::Text,
                    "csv" => Format::Csv,
                    "binary" => Format::Binary,
                    other => {
                        let message = format!("COPY format \"{other}\" not recognized");
                        return Err(located(code::INVALID_PARAMETER_VALUE, message));
                    }
                }
            }
            "delimiter" => delimiter = Some(text),
            "null" => null = Some(text),
            "header" => {
                header = match text.to_ascii_lowercase().as_str() {
                    "" | "true" | "on" | "1" | "yes" => true,
                    "false" | "off" | "0" | "no" => false,
                    "match" if stmt.is_from => return Err(PgError::unsupported("HEADER MATCH")),
                    "match" => {
                        return Err(PgError::new(
                            code::FEATURE_NOT_SUPPORTED,
                            "cannot use \"match\" with HEADER in COPY TO",
                        ));
                    }
                    _ => {
                        let message = format!("{} requires a Boolean value or \"match\"", def.defname);
                        return Err(PgError::new(code::SYNTAX_ERROR, message));
                    }
                }
            }
            "quote" => quote = Some(text),
            "escape" => escape = Some(text),
            "force_quote" if matches!(arg, Some(NodeEnum::AStar(_))) => force_quote_all = true,
            "force_quote" => force_quote = Some(names),
            "force_not_null" => force_not_null = Some(names),
            "force_null" => force_null = Some(names),
            "freeze" | "encoding" => {}
            other => return Err(located(code::SYNTAX_ERROR, format!("option \"{other}\" not recognized"))),
        }
    }
    let (binary, csv) = (format == Format::Binary, format == Format::Csv);
    let unsupported = |message: &str| PgError::new(code::FEATURE_NOT_SUPPORTED, message);
    if binary && delimiter.is_some() {
        return Err(PgError::new(code::SYNTAX_ERROR, "cannot specify DELIMITER in BINARY mode"));
    }
    if binary && null.is_some() {
        return Err(PgError::new(code::SYNTAX_ERROR, "cannot specify NULL in BINARY mode"));
    }
    let delimiter = delimiter.unwrap_or_else(|| if csv { ",".into() } else { "\t".into() });
    let null = null.unwrap_or_else(|| if csv { String::new() } else { "\\N".into() });
    if delimiter.len() != 1 {
        return Err(unsupported("COPY delimiter must be a single one-byte character"));
    }
    if delimiter == "\r" || delimiter == "\n" {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "COPY delimiter cannot be newline or carriage return"));
    }
    if null.contains(['\r', '\n']) {
        return Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            "COPY null representation cannot use newline or carriage return",
        ));
    }
    if !csv && "\\.abcdefghijklmnopqrstuvwxyz0123456789".contains(delimiter.as_str()) {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("COPY delimiter cannot be \"{delimiter}\"")));
    }
    if binary && header {
        return Err(unsupported("cannot specify HEADER in BINARY mode"));
    }
    if !csv && quote.is_some() {
        return Err(unsupported("COPY QUOTE requires CSV mode"));
    }
    let quote = quote.unwrap_or_else(|| "\"".into());
    if quote.len() != 1 {
        return Err(unsupported("COPY quote must be a single one-byte character"));
    }
    if csv && delimiter == quote {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "COPY delimiter and quote must be different"));
    }
    if !csv && escape.is_some() {
        return Err(unsupported("COPY ESCAPE requires CSV mode"));
    }
    let escape = escape.unwrap_or_else(|| quote.clone());
    if escape.len() != 1 {
        return Err(unsupported("COPY escape must be a single one-byte character"));
    }
    let forces_quotes = force_quote_all || force_quote.is_some();
    if !csv && forces_quotes {
        return Err(unsupported("COPY FORCE_QUOTE requires CSV mode"));
    }
    if forces_quotes && stmt.is_from {
        return Err(unsupported("COPY FORCE_QUOTE cannot be used with COPY FROM"));
    }
    if !csv && force_not_null.is_some() {
        return Err(unsupported("COPY FORCE_NOT_NULL requires CSV mode"));
    }
    if force_not_null.is_some() && !stmt.is_from {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "COPY FORCE_NOT_NULL cannot be used with COPY TO"));
    }
    if !csv && force_null.is_some() {
        return Err(unsupported("COPY FORCE_NULL requires CSV mode"));
    }
    if force_null.is_some() && !stmt.is_from {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "COPY FORCE_NULL cannot be used with COPY TO"));
    }
    if null.contains(delimiter.as_str()) {
        return Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            "COPY delimiter character must not appear in the NULL specification",
        ));
    }
    if csv && null.contains(quote.as_str()) {
        return Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            "CSV quote character must not appear in the NULL specification",
        ));
    }
    Ok(Options {
        format,
        delimiter: delimiter.as_bytes()[0],
        null,
        header,
        quote: quote.as_bytes()[0],
        escape: escape.as_bytes()[0],
        force_quote_all,
        force_quote: force_quote.unwrap_or_default(),
        force_not_null: force_not_null.unwrap_or_default(),
        force_null: force_null.unwrap_or_default(),
    })
}

/// Record is one record of text or CSV copy data: the number of the line it ends on and its bytes.
type Record<'a> = (usize, &'a [u8]);

/// records splits the text or CSV copy data of a table into records, which end at newlines outside CSV quotes,
/// stopping at the end-of-data marker of text data, which CSV data reads as a value, as Postgres 18 does.
fn records<'a>(data: &'a [u8], options: &Options, table: &str) -> Result<Vec<Record<'a>>> {
    let csv = options.format == Format::Csv;
    let mut records = Vec::new();
    let (mut start, mut line, mut in_quotes) = (0, 1, false);
    let mut i = 0;
    while i < data.len() {
        let b = data[i];
        if csv
            && in_quotes
            && b == options.escape
            && b != options.quote
            && data.get(i + 1).is_some_and(|&next| next == options.quote || next == options.escape)
        {
            i += 2;
            continue;
        }
        if csv && b == options.quote {
            in_quotes = !in_quotes;
        } else if !csv && b == b'\\' {
            i += 1;
        } else if b == b'\n' && !in_quotes {
            let record = &data[start..i];
            let record = record.strip_suffix(b"\r").unwrap_or(record);
            if !csv && record == b"\\." {
                return Ok(records);
            }
            records.push((line, record));
            start = i + 1;
        }
        if data.get(i) == Some(&b'\n') {
            line += 1;
        }
        i += 1;
    }
    let rest = &data[start.min(data.len())..];
    if in_quotes {
        let err = PgError::new(code::BAD_COPY_FILE_FORMAT, "unterminated CSV quoted field");
        return Err(with_context(err, format!("{}: \"{}\"", context(table, line), display(rest))));
    }
    if !rest.is_empty() && (csv || rest != b"\\.") {
        records.push((line, rest));
    }
    Ok(records)
}

/// unescape replaces the backslash escapes of a text field, as Postgres' CopyReadAttributesText does.
fn unescape(raw: &[u8]) -> Vec<u8> {
    let digits = |i: &mut usize, radix: u32, count: usize, first: u32| {
        let mut value = first;
        for _ in 0..count {
            match raw.get(*i).and_then(|d| (*d as char).to_digit(radix)) {
                Some(d) => {
                    value = value * radix + d;
                    *i += 1;
                }
                None => break,
            }
        }
        value as u8
    };
    let mut out = Vec::with_capacity(raw.len());
    let mut i = 0;
    while i < raw.len() {
        let b = raw[i];
        i += 1;
        if b != b'\\' || i == raw.len() {
            out.push(b);
            continue;
        }
        let c = raw[i];
        i += 1;
        out.push(match c {
            b'0'..=b'7' => digits(&mut i, 8, 2, (c - b'0') as u32),
            b'x' if raw.get(i).is_some_and(u8::is_ascii_hexdigit) => digits(&mut i, 16, 2, 0),
            b'b' => 0x08,
            b'f' => 0x0c,
            b'n' => b'\n',
            b'r' => b'\r',
            b't' => b'\t',
            b'v' => 0x0b,
            other => other,
        });
    }
    out
}

/// text_fields splits a record of the text format into its fields, where None is NULL.
fn text_fields(record: &[u8], options: &Options) -> Vec<Option<Vec<u8>>> {
    let mut fields = Vec::new();
    let (mut start, mut i) = (0, 0);
    loop {
        if i >= record.len() || record[i] == options.delimiter {
            let raw = &record[start..i.min(record.len())];
            fields.push(if raw == options.null.as_bytes() { None } else { Some(unescape(raw)) });
            if i >= record.len() {
                return fields;
            }
            start = i + 1;
        } else if record[i] == b'\\' {
            i += 1;
        }
        i += 1;
    }
}

/// csv_fields splits a record of the CSV format into its fields, where None is NULL, as FORCE_NOT_NULL and
/// FORCE_NULL adjust for each field's column.
fn csv_fields(record: &[u8], options: &Options, names: &[String]) -> Vec<Option<Vec<u8>>> {
    let mut fields = Vec::new();
    let mut i = 0;
    loop {
        let start = i;
        let (mut field, mut quoted, mut in_quotes) = (Vec::new(), false, false);
        while i < record.len() {
            let b = record[i];
            if in_quotes {
                if b == options.escape
                    && let Some(&next) = record.get(i + 1)
                    && (next == options.quote || next == options.escape)
                {
                    field.push(next);
                    i += 2;
                    continue;
                }
                if b == options.quote {
                    in_quotes = false;
                } else {
                    field.push(b);
                }
            } else if b == options.delimiter {
                break;
            } else if b == options.quote {
                (in_quotes, quoted) = (true, true);
            } else {
                field.push(b);
            }
            i += 1;
        }
        let column = names.get(fields.len()).map(String::as_str).unwrap_or_default();
        let null = if quoted {
            field == options.null.as_bytes() && options.force_null.iter().any(|c| c == column)
        } else {
            &record[start..i] == options.null.as_bytes() && !options.force_not_null.iter().any(|c| c == column)
        };
        fields.push(if null { None } else { Some(field) });
        if i >= record.len() {
            return fields;
        }
        i += 1;
    }
}

/// Tuple is one tuple of binary copy data: its number and each field's bytes, where None is NULL.
type Tuple = (usize, Vec<Option<Vec<u8>>>);

/// tuples reads the tuples of the binary copy data of a table with a number of fields, which may end without the
/// trailer as Postgres allows.
fn tuples(data: &[u8], table: &str, expected: usize) -> Result<Vec<Tuple>> {
    let bad = |message: &str, line: usize| {
        with_context(PgError::new(code::BAD_COPY_FILE_FORMAT, message), context(table, line))
    };
    if !data.starts_with(SIGNATURE) {
        return Err(bad("COPY file signature not recognized", 1));
    }
    let int32 = |at: usize| data.get(at..at + 4).map(|b| i32::from_be_bytes([b[0], b[1], b[2], b[3]]));
    let mut i = SIGNATURE.len();
    let flags = int32(i).ok_or_else(|| bad("invalid COPY file header (missing flags)", 1))?;
    if flags & (1 << 16) != 0 {
        return Err(bad("invalid COPY file header (WITH OIDS)", 1));
    }
    if flags >> 16 != 0 {
        return Err(bad("unrecognized critical flags in COPY file header", 1));
    }
    let extension = int32(i + 4).ok_or_else(|| bad("invalid COPY file header (missing length)", 1))?;
    if extension < 0 || data.len() < i + 8 + extension as usize {
        return Err(bad("invalid COPY file header (wrong length)", 1));
    }
    i += 8 + extension as usize;
    let mut tuples = Vec::new();
    while i < data.len() {
        let line = tuples.len() + 1;
        let eof = || bad("unexpected EOF in COPY data", line);
        let count = data.get(i..i + 2).map(|b| i16::from_be_bytes([b[0], b[1]])).ok_or_else(eof)?;
        i += 2;
        if count == -1 {
            break;
        }
        if usize::try_from(count) != Ok(expected) {
            return Err(bad(&format!("row field count is {count}, expected {expected}"), line));
        }
        let mut fields = Vec::with_capacity(expected);
        for _ in 0..count {
            let length = int32(i).ok_or_else(eof)?;
            i += 4;
            if length == -1 {
                fields.push(None);
                continue;
            }
            if length < 0 {
                return Err(bad("invalid field size", line));
            }
            fields.push(Some(data.get(i..i + length as usize).ok_or_else(eof)?.to_vec()));
            i += length as usize;
        }
        tuples.push((line, fields));
    }
    Ok(tuples)
}

/// escape_text writes a value in the text format, escaping backslashes, control characters, and the delimiter, as
/// Postgres' CopyAttributeOutText does.
fn escape_text(text: &str, delimiter: u8) -> String {
    let mut out = String::with_capacity(text.len());
    for c in text.chars() {
        match c {
            '\\' => out.push_str("\\\\"),
            '\x08' => out.push_str("\\b"),
            '\x0c' => out.push_str("\\f"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            '\x0b' => out.push_str("\\v"),
            c if c as u32 == delimiter as u32 => {
                out.push('\\');
                out.push(c);
            }
            c => out.push(c),
        }
    }
    out
}

/// quote_csv writes a value in the CSV format, quoting it when it needs quotes or when the caller forces them, as
/// Postgres' CopyAttributeOutCSV does.
fn quote_csv(text: &str, options: &Options, force: bool, single: bool) -> String {
    let (delimiter, quote, escape) = (options.delimiter as char, options.quote as char, options.escape as char);
    let needs = force
        || text == options.null
        || (single && text == "\\.")
        || text.chars().any(|c| c == delimiter || c == quote || c == '\n' || c == '\r');
    if !needs {
        return text.to_string();
    }
    let mut out = String::with_capacity(text.len() + 2);
    out.push(quote);
    for c in text.chars() {
        if c == quote || c == escape {
            out.push(escape);
        }
        out.push(c);
    }
    out.push(quote);
    out
}

/// format_rows writes rows in a copy format as one chunk of data per row, as Postgres sends them, with the binary
/// header in the first chunk and the binary trailer in a chunk of its own.
fn format_rows(columns: &[Column], rows: &[Vec<Value>], options: &Options) -> Vec<Vec<u8>> {
    let mut chunks = Vec::with_capacity(rows.len() + 1);
    if options.format == Format::Binary {
        let mut header = [SIGNATURE, &0i32.to_be_bytes(), &0i32.to_be_bytes()].concat();
        for row in rows {
            let mut chunk = std::mem::take(&mut header);
            chunk.extend_from_slice(&(row.len() as i16).to_be_bytes());
            for (value, column) in row.iter().zip(columns) {
                match value.encode(column.type_oid, 1) {
                    Some(bytes) => {
                        chunk.extend_from_slice(&(bytes.len() as i32).to_be_bytes());
                        chunk.extend_from_slice(&bytes);
                    }
                    None => chunk.extend_from_slice(&(-1i32).to_be_bytes()),
                }
            }
            chunks.push(chunk);
        }
        header.extend_from_slice(&(-1i16).to_be_bytes());
        chunks.push(header);
        return chunks;
    }
    let csv = options.format == Format::Csv;
    let delimiter = (options.delimiter as char).to_string();
    let single = columns.len() == 1;
    if options.header {
        let names: Vec<String> =
            columns
                .iter()
                .map(|c| {
                    if csv {
                        quote_csv(&c.name, options, false, single)
                    } else {
                        escape_text(&c.name, options.delimiter)
                    }
                })
                .collect();
        chunks.push(format!("{}\n", names.join(&delimiter)).into_bytes());
    }
    for row in rows {
        let fields: Vec<String> = row
            .iter()
            .zip(columns)
            .map(|(value, column)| match value.output() {
                None => options.null.clone(),
                Some(text) if csv => {
                    let force = options.force_quote_all || options.force_quote.contains(&column.name);
                    quote_csv(&text, options, force, single)
                }
                Some(text) => escape_text(&text, options.delimiter),
            })
            .collect();
        chunks.push(format!("{}\n", fields.join(&delimiter)).into_bytes());
    }
    chunks
}

/// io_reason describes a failed file operation as the C library does.
fn io_reason(err: &std::io::Error) -> String {
    match err.kind() {
        std::io::ErrorKind::NotFound => "No such file or directory".into(),
        std::io::ErrorKind::PermissionDenied => "Permission denied".into(),
        std::io::ErrorKind::IsADirectory => "Is a directory".into(),
        _ => err.to_string(),
    }
}

/// file_error returns Postgres' error for a file that COPY cannot open, in the direction it reads or writes.
fn file_error(path: &str, err: &std::io::Error, from: bool) -> PgError {
    let (verb, direction, process) = if from { ("reading", "FROM", "read") } else { ("writing", "TO", "write") };
    PgError {
        hint: Some(format!(
            "COPY {direction} instructs the PostgreSQL server process to {process} a file. You may want a \
             client-side facility such as psql's \\copy."
        )),
        ..PgError::new(code::UNDEFINED_FILE, format!("could not open file \"{path}\" for {verb}: {}", io_reason(err)))
    }
}

/// copy_columns returns the table columns that a COPY names, or every column that is not generated without a list.
fn copy_columns(table: &TableDef, names: &[pg_query::Node]) -> Result<Vec<usize>> {
    if names.is_empty() {
        return Ok((0..table.columns.len()).filter(|&i| !table.columns[i].generated).collect());
    }
    let mut columns = Vec::with_capacity(names.len());
    for name in names.iter().filter_map(node_name) {
        let index = table.columns.iter().position(|c| c.name == name).ok_or_else(|| {
            PgError::new(
                code::UNDEFINED_COLUMN,
                format!("column \"{name}\" of relation \"{}\" does not exist", table.name),
            )
        })?;
        if table.columns[index].generated {
            return Err(PgError {
                detail: Some("Generated columns cannot be used in COPY.".into()),
                ..PgError::new(code::INVALID_COLUMN_REFERENCE, format!("column \"{name}\" is a generated column"))
            });
        }
        if columns.contains(&index) {
            return Err(PgError::new(code::DUPLICATE_COLUMN, format!("column \"{name}\" specified more than once")));
        }
        columns.push(index);
    }
    Ok(columns)
}

impl Ctx<'_> {
    /// copy runs COPY: COPY TO returns the formatted rows, COPY FROM a file inserts its rows, and COPY FROM STDIN
    /// leaves the copy for the session to finish once the client sends the data.
    pub fn copy(&mut self, stmt: &CopyStmt) -> Result<Outcome> {
        if stmt.is_program {
            return Err(PgError::unsupported("COPY with PROGRAM"));
        }
        if stmt.where_clause.is_some() {
            return Err(PgError::unsupported("COPY FROM with WHERE"));
        }
        let options = options(stmt)?;
        if !stmt.is_from {
            return self.copy_to(stmt, options);
        }
        let relation = stmt.relation.as_ref().ok_or_else(|| PgError::internal("COPY FROM without a table"))?;
        let table = self.copy_table(relation, "a")?;
        let columns = copy_columns(&table, &stmt.attlist)?;
        let copy = CopyFrom { table, columns, options };
        if stmt.filename.is_empty() {
            let outcome =
                Outcome::CopyIn { binary: copy.options.format == Format::Binary, columns: copy.columns.len() };
            self.session.pending_copy = Some(Box::new(copy));
            return Ok(outcome);
        }
        let data = std::fs::read(&stmt.filename).map_err(|err| file_error(&stmt.filename, &err, true))?;
        self.copy_rows(&copy, &data)
    }

    /// copy_table resolves the table of a COPY, checking the privilege that the copy's direction needs.
    fn copy_table(&mut self, relation: &RangeVar, privilege: &str) -> Result<TableDef> {
        let table = match self.resolve_table(relation) {
            Ok(table) => table,
            Err(err) => {
                if self.find_view(&relation.schemaname, &relation.relname)?.is_none() {
                    return Err(PgError { position: None, ..err });
                }
                let (direction, hint) = if privilege == "a" {
                    ("to", "To enable copying to a view, provide an INSTEAD OF INSERT trigger.")
                } else {
                    ("from", "Try the COPY (SELECT ...) TO variant.")
                };
                return Err(PgError {
                    hint: Some(hint.into()),
                    ..PgError::new(
                        code::WRONG_OBJECT_TYPE,
                        format!("cannot copy {direction} view \"{}\"", relation.relname),
                    )
                });
            }
        };
        self.require(&Object::Table(table.schema.clone(), table.name.clone()), privilege, -1)?;
        Ok(table)
    }

    /// copy_rows inserts the rows of copy data into the table of a COPY FROM.
    pub fn copy_rows(&mut self, copy: &CopyFrom, data: &[u8]) -> Result<Outcome> {
        let table = &copy.table;
        let names: Vec<String> = copy.columns.iter().map(|&c| table.columns[c].name.clone()).collect();
        let bad = |message: String| PgError::new(code::BAD_COPY_FILE_FORMAT, message);
        let mut rows = Vec::new();
        if copy.options.format == Format::Binary {
            for (line, fields) in tuples(data, &table.name, names.len())? {
                let mut row = Vec::with_capacity(fields.len());
                for ((field, &column), name) in fields.iter().zip(&copy.columns).zip(&names) {
                    let value = Value::decode(table.columns[column].ty.oid, 1, field.as_deref())
                        .map_err(|err| with_context(err, format!("{}, column {name}", context(&table.name, line))))?;
                    row.push(value);
                }
                rows.push(row);
            }
        } else {
            let mut records = records(data, &copy.options, &table.name)?;
            if copy.options.header && records.first().is_some_and(|(line, _)| *line == 1) {
                records.remove(0);
            }
            for (line, record) in records {
                let line_context = || format!("{}: \"{}\"", context(&table.name, line), display(record));
                let fields = match copy.options.format {
                    Format::Csv => csv_fields(record, &copy.options, &names),
                    _ => text_fields(record, &copy.options),
                };
                if fields.len() > names.len() {
                    return Err(with_context(bad("extra data after last expected column".into()), line_context()));
                }
                let mut fields = fields.into_iter();
                let mut row = Vec::with_capacity(names.len());
                for (&column, name) in copy.columns.iter().zip(&names) {
                    let Some(field) = fields.next() else {
                        let message = format!("missing data for column \"{name}\"");
                        return Err(with_context(bad(message), line_context()));
                    };
                    let Some(bytes) = field else {
                        row.push(Value::Null);
                        continue;
                    };
                    let text =
                        crate::encodings::UTF8.decode(&bytes).map_err(|err| with_context(err, line_context()))?;
                    let value_context =
                        format!("{}, column {name}: \"{}\"", context(&table.name, line), display(&bytes));
                    let value = crate::cast::cast_value(Value::Text(text), table.columns[column].ty, false)
                        .map_err(|err| with_context(err, value_context))?;
                    row.push(value);
                }
                rows.push(row);
            }
        }
        let count = rows.len();
        self.plan_copy(table.clone(), copy.columns.clone(), rows)?.run(self)?;
        Ok(Outcome::command(format!("COPY {count}")))
    }

    /// copy_to runs COPY TO, formatting the rows of a table or a query.
    fn copy_to(&mut self, stmt: &CopyStmt, options: Options) -> Result<Outcome> {
        let (columns, rows) = match (&stmt.query, &stmt.relation) {
            (Some(query), _) => {
                let Some(NodeEnum::SelectStmt(select)) = query.node.as_ref() else {
                    return Err(PgError::unsupported("COPY of this statement"));
                };
                let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
                let rows = query.plan.run(self)?;
                (query.columns, rows)
            }
            (None, Some(relation)) => {
                let table = self.copy_table(relation, "r")?;
                let selected = copy_columns(&table, &stmt.attlist)?;
                let columns: Vec<Column> =
                    selected.iter().map(|&c| column(table.columns[c].name.clone(), table.columns[c].ty)).collect();
                let rows = scan(self.db, &table)?
                    .into_iter()
                    .map(|row| selected.iter().map(|&c| row[c].clone()).collect())
                    .collect();
                (columns, rows)
            }
            (None, None) => return Err(PgError::internal("COPY TO without a table or a query")),
        };
        let tag = format!("COPY {}", rows.len());
        let chunks = format_rows(&columns, &rows, &options);
        if !stmt.filename.is_empty() {
            std::fs::write(&stmt.filename, chunks.concat()).map_err(|err| file_error(&stmt.filename, &err, false))?;
            return Ok(Outcome::command(tag));
        }
        Ok(Outcome::CopyOut { binary: options.format == Format::Binary, columns: columns.len(), chunks, tag })
    }
}
