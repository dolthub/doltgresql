use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use prost::Message;

use crate::bindings::*;
use crate::error::*;
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
/// # Example
///
/// ```rust
/// use pg_query::{Node, NodeEnum, NodeRef};
///
/// let result = pg_query::parse("SELECT * FROM contacts");
/// assert!(result.is_ok());
/// let result = result.unwrap();
/// assert_eq!(result.tables(), vec!["contacts"]);
/// assert!(matches!(result.protobuf.nodes()[0].0, NodeRef::SelectStmt(_)));
/// ```
pub fn parse(statement: &str) -> Result<ParseResult> {
    let input = CString::new(statement)?;
    let result = unsafe { pg_query_parse_protobuf(input.as_ptr()) };
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
pub fn parse_with_cursor(statement: &str) -> core::result::Result<ParseResult, (Error, i32, String)> {
    let input = CString::new(statement).map_err(|err| (Error::from(err), 0, String::from("42601")))?;
    let result = unsafe { pg_query_parse_protobuf(input.as_ptr()) };
    let parse_result = if !result.error.is_null() {
        let message = unsafe { CStr::from_ptr((*result.error).message) }.to_string_lossy().to_string();
        let cursor = unsafe { (*result.error).cursorpos };
        let code = unsafe { (*result.error).sqlerrcode };
        let state = (0..5).map(|i| char::from(b'0' + ((code >> (6 * i)) & 0x3F) as u8)).collect();
        Err((Error::Parse(message), cursor, state))
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

/// Converts a parsed tree back into a string.
///
/// # Example
///
/// ```rust
/// use pg_query::{Node, NodeEnum, NodeRef};
///
/// let result = pg_query::parse("INSERT INTO other (name) SELECT name FROM contacts");
/// let result = result.unwrap();
/// let insert = result.protobuf.nodes()[0].0;
/// let select = result.protobuf.nodes()[1].0;
/// assert!(matches!(insert, NodeRef::InsertStmt(_)));
/// assert!(matches!(select, NodeRef::SelectStmt(_)));
///
/// // The entire parse result can be deparsed:
/// assert_eq!(result.deparse().unwrap(), "INSERT INTO other (name) SELECT name FROM contacts");
/// // Or an individual node can be deparsed:
/// assert_eq!(insert.deparse().unwrap(), "INSERT INTO other (name) SELECT name FROM contacts");
/// assert_eq!(select.deparse().unwrap(), "SELECT name FROM contacts");
/// ```
///
/// Note that this function will panic if called on a node not defined in `deparseStmt`
pub fn deparse(protobuf: &protobuf::ParseResult) -> Result<String> {
    let buffer = protobuf.encode_to_vec();
    let len = buffer.len();
    let data = buffer.as_ptr() as *const c_char as *mut c_char;
    let protobuf = PgQueryProtobuf { data, len };
    let result = unsafe { pg_query_deparse_protobuf(protobuf) };

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
/// # Example
///
/// ```rust
/// let result = pg_query::fingerprint("SELECT * FROM contacts WHERE name='Paul'");
/// assert!(result.is_ok());
/// let result = result.unwrap();
/// assert_eq!(result.hex, "0e2581a461ece536");
/// ```
pub fn fingerprint(statement: &str) -> Result<Fingerprint> {
    let input = CString::new(statement)?;
    let result = unsafe { pg_query_fingerprint(input.as_ptr()) };
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

/// Split a well-formed query into separate statements.
///
/// # Example
///
/// ```rust
/// let query = r#"select /*;*/ 1; select "2;", (select 3);"#;
/// let statements = pg_query::split_with_parser(query).unwrap();
/// assert_eq!(statements, vec!["select /*;*/ 1", r#" select "2;", (select 3)"#]);
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
