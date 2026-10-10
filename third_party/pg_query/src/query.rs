use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use prost::Message;

use crate::bindings::*;
use crate::error::*;
use crate::options::{DeparseComment, DeparseOptions, FingerprintOptions, ParserOptions};
use crate::parse_result::ParseResult;
use crate::protobuf;

/// Represents the resulting fingerprint containing both the raw integer form as well as the
/// corresponding 16 character hex value.
pub struct Fingerprint {
    pub value: u64,
    pub hex: String,
}

/// Parses the given SQL statement into the given abstract syntax tree.
///
/// The `parser_options` control how the statement is parsed: pass a
/// [`ParserOptions`] constant, a raw integer, or `0` for the standard
/// top-level SQL grammar.
///
/// # Example
///
/// ```rust
/// use pg_query::{Node, NodeEnum, NodeRef};
///
/// let result = pg_query::parse("SELECT * FROM contacts", 0);
/// assert!(result.is_ok());
/// let result = result.unwrap();
/// assert_eq!(result.tables(), vec!["contacts"]);
/// assert!(matches!(result.protobuf.nodes()[0].0, NodeRef::SelectStmt(_)));
/// ```
pub fn parse(statement: &str, parser_options: impl Into<ParserOptions>) -> Result<ParseResult> {
    let input = CString::new(statement)?;
    let result = unsafe { pg_query_parse_protobuf_opts(input.as_ptr(), parser_options.into().bits()) };
    let parse_result = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Parse(message))
    } else {
        let data = unsafe { std::slice::from_raw_parts(result.parse_tree.data as *const u8, result.parse_tree.len as usize) };
        let stderr = unsafe { CStr::from_ptr(result.stderr_buffer) }.to_string_lossy().to_string();
        protobuf::ParseResult::decode(data).map_err(Error::Decode).map(|result| ParseResult::new(result, stderr))
    };
    unsafe { pg_query_free_protobuf_parse_result(result) };
    parse_result
}

/// Parses like [`parse`], and on a parse error also returns the 1-based character position that the parser reported,
/// which is 0 when it reported none, and the error's SQLSTATE.
///
/// Added for Doltgres, which reports the position and the SQLSTATE to clients.
pub fn parse_with_cursor(statement: &str, parser_options: impl Into<ParserOptions>) -> core::result::Result<ParseResult, (Error, i32, String)> {
    let input = CString::new(statement).map_err(|err| (Error::from(err), 0, String::from("42601")))?;
    let result = unsafe { pg_query_parse_protobuf_opts(input.as_ptr(), parser_options.into().bits()) };
    let parse_result = if !result.error.is_null() {
        Err(cursor_error(unsafe { &*result.error }))
    } else {
        let data = unsafe { std::slice::from_raw_parts(result.parse_tree.data as *const u8, result.parse_tree.len as usize) };
        let stderr = unsafe { CStr::from_ptr(result.stderr_buffer) }.to_string_lossy().to_string();
        protobuf::ParseResult::decode(data)
            .map_err(|err| (Error::Decode(err), 0, String::from("42601")))
            .map(|result| ParseResult::new(result, stderr))
    };
    unsafe { pg_query_free_protobuf_parse_result(result) };
    parse_result
}

/// Returns a parse error with the 1-based character position that the parser reported, which is 0 when it reported
/// none, and the error's SQLSTATE, which is a syntax error's when it reported none.
///
/// Added for Doltgres.
fn cursor_error(error: &PgQueryError) -> (Error, i32, String) {
    let message = unsafe { CStr::from_ptr(error.message) }.to_string_lossy().to_string();
    let state = match error.sqlerrcode {
        0 => String::from("42601"),
        code => (0..5).map(|i| char::from(b'0' + ((code >> (6 * i)) & 0x3F) as u8)).collect(),
    };
    (Error::Parse(message), error.cursorpos, state)
}

/// Converts a parsed tree back into a string.
///
/// `options` accepts a [`DeparseOptions`] struct; pass `Default::default()`
/// for the plain, unformatted output.
///
/// # Example
///
/// ```rust
/// use pg_query::{Node, NodeEnum, NodeRef};
///
/// let result = pg_query::parse("INSERT INTO other (name) SELECT name FROM contacts", 0);
/// let result = result.unwrap();
/// let insert = result.protobuf.nodes()[0].0;
/// let select = result.protobuf.nodes()[1].0;
/// assert!(matches!(insert, NodeRef::InsertStmt(_)));
/// assert!(matches!(select, NodeRef::SelectStmt(_)));
///
/// // The entire parse result can be deparsed:
/// assert_eq!(result.deparse(Default::default()).unwrap(), "INSERT INTO other (name) SELECT name FROM contacts");
/// // Or an individual node can be deparsed:
/// assert_eq!(insert.deparse(Default::default()).unwrap(), "INSERT INTO other (name) SELECT name FROM contacts");
/// assert_eq!(select.deparse(Default::default()).unwrap(), "SELECT name FROM contacts");
/// ```
///
/// Note that this function will panic if called on a node not defined in `deparseStmt`
pub fn deparse(protobuf: &protobuf::ParseResult, options: DeparseOptions) -> Result<String> {
    let buffer = protobuf.encode_to_vec();
    let len = buffer.len();
    let data = buffer.as_ptr() as *const c_char as *mut c_char;
    let protobuf = PgQueryProtobuf { data, len };

    // Rebuild the C comment representation; the C code only borrows these (it
    // copies the pointers, not the strings), so they just need to outlive the call.
    let comment_strings = options.comments.iter().map(|comment| CString::new(comment.text.as_str())).collect::<std::result::Result<Vec<_>, _>>()?;
    let mut comment_structs = options
        .comments
        .iter()
        .zip(comment_strings.iter())
        .map(|(comment, string)| PostgresDeparseComment {
            match_location: comment.match_location,
            newlines_before_comment: comment.newlines_before_comment,
            newlines_after_comment: comment.newlines_after_comment,
            str_: string.as_ptr() as *mut c_char,
        })
        .collect::<Vec<_>>();
    let mut comment_ptrs = comment_structs.iter_mut().map(|comment| comment as *mut _).collect::<Vec<_>>();

    let opts = PostgresDeparseOpts {
        comments: if comment_ptrs.is_empty() { std::ptr::null_mut() } else { comment_ptrs.as_mut_ptr() },
        comment_count: comment_ptrs.len(),
        pretty_print: options.pretty_print,
        indent_size: options.indent_size,
        max_line_length: options.max_line_length,
        trailing_newline: options.trailing_newline,
        commas_start_of_line: options.commas_start_of_line,
    };

    let result = unsafe { pg_query_deparse_protobuf_opts(protobuf, opts) };

    let deparse_result = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Parse(message))
    } else {
        let query = unsafe { CStr::from_ptr(result.query) }.to_string_lossy().to_string();
        Ok(query)
    };

    unsafe { pg_query_free_deparse_result(result) };
    deparse_result
}

/// Extract the comments from a query, along with the metadata needed to
/// re-insert them into deparsed output via [`DeparseOptions::comments`].
///
/// # Example
///
/// ```rust
/// let query = "SELECT 1 -- cast to string\n";
/// let comments = pg_query::deparse_comments_for_query(query).unwrap();
/// assert_eq!(comments.len(), 1);
/// assert_eq!(comments[0].text, "-- cast to string");
///
/// let parsed = pg_query::parse(query, 0).unwrap();
/// let output = pg_query::deparse(
///     &parsed.protobuf,
///     pg_query::DeparseOptions {
///         comments,
///         pretty_print: true,
///         ..Default::default()
///     },
/// )
/// .unwrap();
/// assert!(output.contains("-- cast to string"));
/// ```
pub fn deparse_comments_for_query(query: &str) -> Result<Vec<DeparseComment>> {
    let input = CString::new(query)?;
    let result = unsafe { pg_query_deparse_comments_for_query(input.as_ptr()) };
    let comments = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Scan(message))
    } else {
        let comments = if result.comment_count == 0 {
            Vec::new()
        } else {
            unsafe { std::slice::from_raw_parts(result.comments, result.comment_count) }
                .iter()
                .map(|comment| {
                    let comment = unsafe { &**comment };
                    DeparseComment {
                        match_location: comment.match_location,
                        newlines_before_comment: comment.newlines_before_comment,
                        newlines_after_comment: comment.newlines_after_comment,
                        text: unsafe { CStr::from_ptr(comment.str_) }.to_string_lossy().to_string(),
                    }
                })
                .collect::<Vec<_>>()
        };
        Ok(comments)
    };
    unsafe { pg_query_free_deparse_comments_result(result) };
    comments
}

/// Normalizes the given SQL statement, returning a parametized version.
///
/// # Example
///
/// ```rust
/// let result = pg_query::normalize("SELECT * FROM contacts WHERE name='Paul'");
/// assert!(result.is_ok());
/// let result = result.unwrap();
/// assert_eq!(result, "SELECT * FROM contacts WHERE name=$1");
/// ```
pub fn normalize(statement: &str) -> Result<String> {
    let input = CString::new(statement)?;
    let result = unsafe { pg_query_normalize(input.as_ptr()) };
    let normalized_query = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Parse(message))
    } else {
        let n = unsafe { CStr::from_ptr(result.normalized_query) };
        Ok(n.to_string_lossy().to_string())
    };
    unsafe { pg_query_free_normalize_result(result) };
    normalized_query
}

/// Fingerprints the given SQL statement. Useful for comparing parse trees across different implementations
/// of `libpg_query`.
///
/// `parser_options` and `fingerprint_options` accept [`ParserOptions`] /
/// [`FingerprintOptions`] constants or raw integers; pass `0` to use the
/// defaults for either.
///
/// Fingerprinting is a superset of the Postgres `queryid` mechanism: for a
/// given `queryid` there is only one valid fingerprint (e.g. constants are
/// ignored, so `WHERE id = 123` and `WHERE id = 456` fingerprint the same).
///
/// Fingerprints follow Postgres 18 query ID behavior: in SELECT/DML
/// statements a relation alias (if present) replaces the relation name and
/// schema names are ignored, while sequences of 2+ digits in relation names
/// are ignored by default. Pass [`FingerprintOptions::RANGEVAR_PG17_COMPAT`]
/// for Postgres 17 compatible fingerprints, or use the other
/// [`FingerprintOptions`] flags to tune relation handling. See the
/// "Fingerprinting a query" section of the README for the full rules.
///
/// # Example
///
/// ```rust
/// let result = pg_query::fingerprint("SELECT * FROM contacts WHERE name='Paul'", 0, 0);
/// assert!(result.is_ok());
/// let result = result.unwrap();
/// assert_eq!(result.hex, "0e2581a461ece536");
/// ```
///
/// # Example: with fingerprint options
///
/// ```rust
/// // By default the full relation name is not fingerprinted, so both match...
/// let default = pg_query::fingerprint("SELECT * FROM orders_2024_01", 0, 0).unwrap();
/// assert_eq!(default.hex, pg_query::fingerprint("SELECT * FROM orders_2024_02", 0, 0).unwrap().hex);
///
/// // ...unless PG_QUERY_FINGERPRINT_FULL_RELNAME is set
/// let result = pg_query::fingerprint("SELECT * FROM orders_2024_01", 0, pg_query::FingerprintOptions::FULL_RELNAME).unwrap();
/// assert_eq!(result.hex, "3cc2d1ca3f22c9bf");
/// ```
pub fn fingerprint(
    statement: &str, parser_options: impl Into<ParserOptions>, fingerprint_options: impl Into<FingerprintOptions>,
) -> Result<Fingerprint> {
    let input = CString::new(statement)?;
    let result = unsafe { pg_query_fingerprint_opts(input.as_ptr(), parser_options.into().bits(), fingerprint_options.into().bits()) };
    let fingerprint = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Parse(message))
    } else {
        let hex = unsafe { CStr::from_ptr(result.fingerprint_str) };
        Ok(Fingerprint { value: result.fingerprint, hex: hex.to_string_lossy().to_string() })
    };
    unsafe { pg_query_free_fingerprint_result(result) };
    fingerprint
}

/// An experimental API which parses a PLPGSQL function. This currently returns the raw JSON structure.
///
/// # Example
///
/// ```rust
/// let result = pg_query::parse_plpgsql("
///     CREATE OR REPLACE FUNCTION cs_fmt_browser_version(v_name varchar, v_version varchar)
///     RETURNS varchar AS $$
///     BEGIN
///         IF v_version IS NULL THEN
///             RETURN v_name;
///         END IF;
///         RETURN v_name || '/' || v_version;
///     END;
///     $$ LANGUAGE plpgsql;
/// ");
/// assert!(result.is_ok());
/// ```
pub fn parse_plpgsql(stmt: &str) -> Result<serde_json::Value> {
    let input = CString::new(stmt)?;
    let result = unsafe { pg_query_parse_plpgsql(input.as_ptr()) };
    let structure = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Parse(message))
    } else {
        let raw = unsafe { CStr::from_ptr(result.plpgsql_funcs) };
        serde_json::from_str(&raw.to_string_lossy()).map_err(|e| Error::InvalidJson(e.to_string()))
    };
    unsafe { pg_query_free_plpgsql_parse_result(result) };
    structure
}

/// Parses like [`parse_plpgsql`], and on a parse error also returns the position and SQLSTATE as
/// [`parse_with_cursor`] does.
///
/// Added for Doltgres, which reports the position and the SQLSTATE to clients.
pub fn parse_plpgsql_with_cursor(stmt: &str) -> core::result::Result<serde_json::Value, (Error, i32, String)> {
    let input = CString::new(stmt).map_err(|err| (Error::from(err), 0, String::from("42601")))?;
    let result = unsafe { pg_query_parse_plpgsql(input.as_ptr()) };
    let structure = if !result.error.is_null() {
        Err(cursor_error(unsafe { &*result.error }))
    } else {
        let raw = unsafe { CStr::from_ptr(result.plpgsql_funcs) };
        serde_json::from_str(&raw.to_string_lossy()).map_err(|e| (Error::InvalidJson(e.to_string()), 0, String::from("42601")))
    };
    unsafe { pg_query_free_plpgsql_parse_result(result) };
    structure
}

/// Split a well-formed query into separate statements.
///
/// # Example
///
/// ```rust
/// let query = r#"select /*;*/ 1; select "2;", (select 3);"#;
/// let statements = pg_query::split_with_parser(query).unwrap();
/// assert_eq!(statements, vec!["select /*;*/ 1", r#"select "2;", (select 3)"#]);
/// ```
///
/// However, `split_with_parser` will fail on malformed statements
///
/// ```rust
/// let query = "select 1; this statement is not sql; select 2;";
/// let result = pg_query::split_with_parser(query);
/// let err = r#"syntax error at or near "this""#;
/// assert_eq!(result, Err(pg_query::Error::Split(err.to_string())));
/// ```
pub fn split_with_parser(query: &str) -> Result<Vec<&str>> {
    let input = CString::new(query)?;
    let result = unsafe { pg_query_split_with_parser(input.as_ptr()) };
    let split_result = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Split(message))
    } else {
        let n_stmts = result.n_stmts as usize;
        let mut statements = Vec::with_capacity(n_stmts);
        for offset in 0..n_stmts {
            let split_stmt = unsafe { *result.stmts.add(offset).read() };
            let start = split_stmt.stmt_location as usize;
            let end = start + split_stmt.stmt_len as usize;
            statements.push(&query[start..end]);
            // not sure the start..end slice'll hold up for non-utf8 charsets
        }
        Ok(statements)
    };
    unsafe { pg_query_free_split_result(result) };
    split_result
}

/// Scan a sql query into a its component of tokens.
///
/// # Example
///
/// ```rust
/// use pg_query::protobuf::*;
/// let sql = "SELECT update AS left /* comment */ FROM between";
/// let result = pg_query::scan(sql).unwrap();
/// let tokens: Vec<std::string::String> = result.tokens.iter().map(|token| {
///     format!("{:?}", token)
/// }).collect();
/// assert_eq!(
///     tokens,
///     vec![
///         "ScanToken { start: 0, end: 6, token: Select, keyword_kind: ReservedKeyword }",
///         "ScanToken { start: 7, end: 13, token: Update, keyword_kind: UnreservedKeyword }",
///         "ScanToken { start: 14, end: 16, token: As, keyword_kind: ReservedKeyword }",
///         "ScanToken { start: 17, end: 21, token: Left, keyword_kind: TypeFuncNameKeyword }",
///         "ScanToken { start: 22, end: 35, token: CComment, keyword_kind: NoKeyword }",
///         "ScanToken { start: 36, end: 40, token: From, keyword_kind: ReservedKeyword }",
///         "ScanToken { start: 41, end: 48, token: Between, keyword_kind: ColNameKeyword }"
///     ]);
/// ```
pub fn scan(sql: &str) -> Result<protobuf::ScanResult> {
    let input = CString::new(sql)?;
    let result = unsafe { pg_query_scan(input.as_ptr()) };
    let scan_result = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Scan(message))
    } else {
        let data = unsafe { std::slice::from_raw_parts(result.pbuf.data as *const u8, result.pbuf.len as usize) };
        protobuf::ScanResult::decode(data).map_err(Error::Decode)
    };
    unsafe { pg_query_free_scan_result(result) };
    scan_result
}

/// Scan a sql query into its component tokens, without protobuf serialization.
///
/// Unlike [`scan`], which round-trips the result through protobuf, this
/// returns the tokens directly from the C scanner. The `token` and
/// `keyword_kind` values are identical to the protobuf `Token` and
/// `KeywordKind` enums, so the result is equivalent to `scan(sql)?.tokens`.
///
/// # Example
///
/// ```rust
/// let sql = "SELECT update AS left /* comment */ FROM between";
/// let tokens = pg_query::scan_tokens(sql).unwrap();
/// assert_eq!(tokens, pg_query::scan(sql).unwrap().tokens);
/// ```
pub fn scan_tokens(sql: &str) -> Result<Vec<protobuf::ScanToken>> {
    let input = CString::new(sql)?;
    let result = unsafe { pg_query_scan_tokens(input.as_ptr()) };
    let tokens = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Scan(message))
    } else {
        Ok(unsafe { std::slice::from_raw_parts(result.tokens, result.n_tokens as usize) }
            .iter()
            .map(|token| protobuf::ScanToken {
                start: token.start,
                end: token.end,
                token: token.token as i32,
                keyword_kind: token.keyword_kind as i32,
            })
            .collect::<Vec<_>>())
    };
    unsafe { pg_query_free_scan_tokens_result(result) };
    tokens
}

/// Determine whether each statement in a query is a utility statement.
///
/// Utility statements are statements that are not SELECT/INSERT/UPDATE/DELETE
/// (e.g. `SHOW`, `SET`, `CREATE TABLE`, ...). For a multi-statement query a
/// boolean is returned per statement, in order.
///
/// # Example
///
/// ```rust
/// let result = pg_query::is_utility_stmt("SELECT 1; SHOW fsync;").unwrap();
/// assert_eq!(result, vec![false, true]);
/// ```
pub fn is_utility_stmt(query: &str) -> Result<Vec<bool>> {
    let input = CString::new(query)?;
    let result = unsafe { pg_query_is_utility_stmt(input.as_ptr()) };
    let utilities = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::IsUtility(message))
    } else {
        Ok(unsafe { std::slice::from_raw_parts(result.items, result.length as usize) }.to_vec())
    };
    unsafe { pg_query_free_is_utility_result(result) };
    utilities
}

/// Split a potentially-malformed query into separate statements. Note that
/// invalid tokens will be skipped
/// ```rust
/// let query = r#"select /*;*/ 1; asdf; select "2;", (select 3); asdf"#;
/// let statements = pg_query::split_with_scanner(query).unwrap();
/// assert_eq!(statements, vec![
///     "select /*;*/ 1",
///     // skipped " asdf" since it was an invalid token
///     r#" select "2;", (select 3)"#,
/// ]);
/// ```
pub fn split_with_scanner(query: &str) -> Result<Vec<&str>> {
    let input = CString::new(query)?;
    let result = unsafe { pg_query_split_with_scanner(input.as_ptr()) };
    let split_result = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        Err(Error::Split(message))
    } else {
        // don't use result.stderr_buffer since it appears unused unless
        // libpg_query is compiled with DEBUG defined.
        let n_stmts = result.n_stmts as usize;
        let mut start: usize;
        let mut end: usize;
        let mut statements = Vec::with_capacity(n_stmts);
        for offset in 0..n_stmts {
            let split_stmt = unsafe { *result.stmts.add(offset).read() };
            start = split_stmt.stmt_location as usize;
            // TODO: consider comparing the new value of start to the old value
            // of end to see if any region larger than a statement-separator got skipped
            end = start + split_stmt.stmt_len as usize;
            statements.push(&query[start..end]);
        }
        Ok(statements)
    };
    unsafe { pg_query_free_split_result(result) };
    split_result
}
