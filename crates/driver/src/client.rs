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

use harness::gostd;
use harness::pgx::{Arg, Conn, ConnConfig};

/// BINARY_RESULTS lists the types whose results database/sql asks pgx for in binary.
const BINARY_RESULTS: [u32; 13] = [16, 17, 29, 1082, 700, 701, 21, 23, 20, 26, 1114, 1184, 28];

/// CONNECT_ATTEMPTS is how many times a connection is tried while a server starts.
const CONNECT_ATTEMPTS: usize = 50;

/// Db is a database/sql connection made through pgx's stdlib driver.
pub struct Db {
    conn: Conn,
}

/// QueryResult is a query's column names and its rows as database/sql scans them into NullString, with NULL
/// printed as "NULL".
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct QueryResult {
    pub columns: Vec<String>,
    pub rows: Vec<Vec<String>>,
}

/// url_escape escapes a URL component like Go's url.QueryEscape.
fn url_escape(text: &str) -> String {
    text.bytes()
        .map(|b| match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => (b as char).to_string(),
            b' ' => "+".to_string(),
            _ => format!("%{b:02X}"),
        })
        .collect()
}

/// dsn builds a connection URL like the Go driver's GetDSN, with sslmode=prefer unless the parameters set it.
pub fn dsn(user: &str, password: &str, database: &str, host: &str, port: u16, params: &[(String, String)]) -> String {
    let mut values = std::collections::BTreeMap::new();
    values.insert("sslmode".to_string(), "prefer".to_string());
    for (k, v) in params {
        values.insert(k.clone(), v.clone());
    }
    let query: Vec<String> = values.iter().map(|(k, v)| format!("{}={}", url_escape(k), url_escape(v))).collect();
    format!(
        "postgres://{}:{}@{host}:{port}/{}?{}",
        url_escape(user),
        url_escape(password),
        url_escape(database),
        query.join("&")
    )
}

impl Db {
    /// connect connects and pings, retrying while the server starts.
    pub fn connect(
        user: &str,
        password: &str,
        database: &str,
        host: &str,
        port: u16,
        params: &[(String, String)],
    ) -> Result<Db, String> {
        let config =
            ConnConfig::parse(&dsn(user, password, database, host, port, params)).map_err(|e| e.to_string())?;
        let mut last = String::new();
        for _ in 0..CONNECT_ATTEMPTS {
            match Conn::connect(config.clone()).and_then(|mut conn| conn.ping().map(|_| conn)) {
                Ok(conn) => return Ok(Db { conn }),
                Err(err) => last = err.to_string(),
            }
            crate::process::wait_briefly();
        }
        Err(last)
    }

    /// exec runs a statement, over the simple protocol when it has no arguments.
    pub fn exec(&mut self, sql: &str, args: &[String]) -> Result<(), String> {
        let args: Vec<Arg> = args.iter().map(|a| Arg::Str(a.clone())).collect();
        self.conn.exec(sql, &args).map(|_| ()).map_err(|e| e.to_string())
    }

    /// query runs a query and scans every value into a string.
    pub fn query(&mut self, sql: &str, args: &[String]) -> Result<QueryResult, String> {
        let args: Vec<Arg> = args.iter().map(|a| Arg::Str(a.clone())).collect();
        let result =
            self.conn.query_with_result_formats_by_oid(sql, &args, &BINARY_RESULTS).map_err(|e| e.to_string())?;
        if let Some(err) = result.error {
            return Err(err.to_string());
        }
        let mut rows = Vec::with_capacity(result.rows.len());
        for row in &result.rows {
            let mut values = Vec::with_capacity(row.len());
            for (field, value) in result.fields.iter().zip(row) {
                values.push(match value {
                    None => "NULL".to_string(),
                    Some(bytes) => scan_string(field.data_type_oid, field.format, bytes)
                        .map_err(|e| format!("convert field failed: {e}"))?,
                });
            }
            rows.push(values);
        }
        Ok(QueryResult { columns: result.fields.iter().map(|f| f.name.clone()).collect(), rows })
    }

    /// close closes the connection.
    pub fn close(mut self) {
        self.conn.close();
    }
}

/// POSTGRES_EPOCH_DAYS is the number of days from the Unix epoch to 2000-01-01.
const POSTGRES_EPOCH_DAYS: i64 = 10957;

/// rfc3339_utc formats a UTC time like Go's time.RFC3339Nano.
fn rfc3339_utc(micros_since_2000: i64) -> String {
    let micros = micros_since_2000 + POSTGRES_EPOCH_DAYS * 86_400_000_000;
    let seconds = micros.div_euclid(1_000_000);
    let fraction = micros.rem_euclid(1_000_000);
    let (year, month, day) = gostd::civil_from_days(seconds.div_euclid(86400));
    let time = seconds.rem_euclid(86400);
    let mut out = format!("{year:04}-{month:02}-{day:02}T{:02}:{:02}:{:02}", time / 3600, time / 60 % 60, time % 60);
    if fraction != 0 {
        out.push('.');
        out.push_str(format!("{fraction:06}").trim_end_matches('0'));
    }
    out.push('Z');
    out
}

/// fixed returns the bytes as an array when the length matches.
fn fixed<const N: usize>(bytes: &[u8], what: &str) -> Result<[u8; N], String> {
    bytes.try_into().map_err(|_| format!("invalid length for {what}: {}", bytes.len()))
}

/// scan_string converts a value to a string like pgx's stdlib driver and database/sql do, formatting times in UTC.
fn scan_string(oid: u32, format: i16, bytes: &[u8]) -> Result<String, String> {
    let text = || String::from_utf8_lossy(bytes).into_owned();
    if format == 0 {
        return Ok(match oid {
            16 => (bytes == b"t").to_string(),
            700 | 701 => gostd::format_v64(gostd::parse_float64(&text()).ok_or_else(text)?),
            17 => {
                let hex = bytes.strip_prefix(b"\\x").ok_or("invalid hex format")?;
                let decoded: Result<Vec<u8>, _> = hex
                    .chunks(2)
                    .map(|pair| u8::from_str_radix(std::str::from_utf8(pair).unwrap_or("zz"), 16))
                    .collect();
                String::from_utf8_lossy(&decoded.map_err(|e| e.to_string())?).into_owned()
            }
            _ => text(),
        });
    }
    Ok(match oid {
        16 => (fixed::<1>(bytes, "bool")?[0] == 1).to_string(),
        17 => text(),
        21 => i16::from_be_bytes(fixed(bytes, "int2")?).to_string(),
        23 => i32::from_be_bytes(fixed(bytes, "int4")?).to_string(),
        20 => i64::from_be_bytes(fixed(bytes, "int8")?).to_string(),
        26 | 28 | 29 => u32::from_be_bytes(fixed(bytes, "uint32")?).to_string(),
        700 => gostd::format_v64(f32::from_be_bytes(fixed(bytes, "float4")?) as f64),
        701 => gostd::format_v64(f64::from_be_bytes(fixed(bytes, "float8")?)),
        1082 => match i32::from_be_bytes(fixed(bytes, "date")?) {
            i32::MAX => "infinity".to_string(),
            i32::MIN => "-infinity".to_string(),
            days => rfc3339_utc(days as i64 * 86_400_000_000),
        },
        1114 | 1184 => match i64::from_be_bytes(fixed(bytes, "timestamp")?) {
            i64::MAX => "infinity".to_string(),
            i64::MIN => "-infinity".to_string(),
            micros => rfc3339_utc(micros),
        },
        _ => text(),
    })
}
