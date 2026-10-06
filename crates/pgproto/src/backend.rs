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

use std::collections::BTreeMap;

use crate::codec::{DecodeError, Reader, Writer};

/// FieldDescription describes a single column of a RowDescription.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct FieldDescription {
    /// The column name.
    pub name: String,
    /// The OID of the source table, or zero when the column is not a table column.
    pub table_oid: u32,
    /// The attribute number of the source column, or zero when the column is not a table column.
    pub table_attribute_number: u16,
    /// The OID of the column's type.
    pub data_type_oid: u32,
    /// The type's size as in pg_type.typlen, where negative values mean variable width.
    pub data_type_size: i16,
    /// The type modifier as in pg_attribute.atttypmod.
    pub type_modifier: i32,
    /// The format code, where zero is text and one is binary.
    pub format: i16,
}

/// ErrorFields holds the fields shared by ErrorResponse and NoticeResponse. Empty strings and zero numbers are
/// absent fields.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct ErrorFields {
    /// The localized severity (S).
    pub severity: String,
    /// The severity that is never localized (V).
    pub severity_unlocalized: String,
    /// The SQLSTATE code (C).
    pub code: String,
    /// The primary message (M).
    pub message: String,
    /// The detail message (D).
    pub detail: String,
    /// The hint message (H).
    pub hint: String,
    /// The one-based character position within the query (P).
    pub position: i32,
    /// The one-based character position within the internal query (p).
    pub internal_position: i32,
    /// The text of an internally generated query (q).
    pub internal_query: String,
    /// The context in which the error occurred (W).
    pub where_: String,
    /// The schema name (s).
    pub schema_name: String,
    /// The table name (t).
    pub table_name: String,
    /// The column name (c).
    pub column_name: String,
    /// The data type name (d).
    pub data_type_name: String,
    /// The constraint name (n).
    pub constraint_name: String,
    /// The source file of the error (F).
    pub file: String,
    /// The source line of the error (L).
    pub line: i32,
    /// The source routine of the error (R).
    pub routine: String,
    /// Fields with codes that are not listed above.
    pub unknown_fields: BTreeMap<u8, String>,
}

/// BackendMessage is a message sent from the server to the client.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum BackendMessage {
    /// AuthenticationOk reports that authentication succeeded.
    AuthenticationOk,
    /// AuthenticationCleartextPassword requests a cleartext password.
    AuthenticationCleartextPassword,
    /// AuthenticationMD5Password requests an MD5-hashed password using the salt.
    AuthenticationMD5Password { salt: [u8; 4] },
    /// AuthenticationGSS requests GSSAPI authentication.
    AuthenticationGSS,
    /// AuthenticationGSSContinue carries GSSAPI or SSPI data.
    AuthenticationGSSContinue { data: Vec<u8> },
    /// AuthenticationSASL lists the SASL mechanisms the server accepts.
    AuthenticationSASL { mechanisms: Vec<String> },
    /// AuthenticationSASLContinue carries SASL challenge data.
    AuthenticationSASLContinue { data: Vec<u8> },
    /// AuthenticationSASLFinal carries the final SASL data.
    AuthenticationSASLFinal { data: Vec<u8> },
    /// BackendKeyData carries the key used for CancelRequest, whose secret has a variable length since protocol 3.2.
    BackendKeyData { process_id: u32, secret_key: Vec<u8> },
    /// BindComplete acknowledges a Bind.
    BindComplete,
    /// CloseComplete acknowledges a Close.
    CloseComplete,
    /// CommandComplete reports the command tag of a finished command.
    CommandComplete { command_tag: String },
    /// CopyBothResponse starts a bidirectional copy.
    CopyBothResponse { overall_format: u8, column_format_codes: Vec<u16> },
    /// CopyData carries copy data.
    CopyData { data: Vec<u8> },
    /// CopyDone ends a copy.
    CopyDone,
    /// CopyInResponse starts a copy from the client.
    CopyInResponse { overall_format: u8, column_format_codes: Vec<u16> },
    /// CopyOutResponse starts a copy to the client.
    CopyOutResponse { overall_format: u8, column_format_codes: Vec<u16> },
    /// DataRow carries one result row, where None is NULL.
    DataRow { values: Vec<Option<Vec<u8>>> },
    /// EmptyQueryResponse replaces CommandComplete for an empty query.
    EmptyQueryResponse,
    /// ErrorResponse reports an error.
    ErrorResponse(ErrorFields),
    /// FunctionCallResponse carries the result of a FunctionCall, where None is NULL.
    FunctionCallResponse { result: Option<Vec<u8>> },
    /// NegotiateProtocolVersion reports the newest supported minor version and the unrecognized options.
    NegotiateProtocolVersion { newest_minor_protocol: u32, unrecognized_options: Vec<String> },
    /// NoData reports that a statement or portal returns no rows.
    NoData,
    /// NoticeResponse reports a notice.
    NoticeResponse(ErrorFields),
    /// NotificationResponse delivers a NOTIFY payload.
    NotificationResponse { process_id: u32, channel: String, payload: String },
    /// ParameterDescription lists the parameter types of a statement.
    ParameterDescription { parameter_oids: Vec<u32> },
    /// ParameterStatus reports the value of a runtime parameter.
    ParameterStatus { name: String, value: String },
    /// ParseComplete acknowledges a Parse.
    ParseComplete,
    /// PortalSuspended reports that an Execute reached its row limit.
    PortalSuspended,
    /// ReadyForQuery reports the transaction status: 'I' idle, 'T' in a transaction, or 'E' in a failed one.
    ReadyForQuery { tx_status: u8 },
    /// RowDescription describes the columns of the rows that follow.
    RowDescription { fields: Vec<FieldDescription> },
}

impl BackendMessage {
    /// encode appends the message to the buffer.
    pub fn encode(&self, buffer: &mut Vec<u8>) {
        match self {
            BackendMessage::AuthenticationOk => encode_authentication(buffer, 0, |_| {}),
            BackendMessage::AuthenticationCleartextPassword => encode_authentication(buffer, 3, |_| {}),
            BackendMessage::AuthenticationMD5Password { salt } => encode_authentication(buffer, 5, |w| {
                w.bytes(salt);
            }),
            BackendMessage::AuthenticationGSS => encode_authentication(buffer, 7, |_| {}),
            BackendMessage::AuthenticationGSSContinue { data } => encode_authentication(buffer, 8, |w| {
                w.bytes(data);
            }),
            BackendMessage::AuthenticationSASL { mechanisms } => encode_authentication(buffer, 10, |w| {
                for mechanism in mechanisms {
                    w.cstring(mechanism);
                }
                w.u8(0);
            }),
            BackendMessage::AuthenticationSASLContinue { data } => encode_authentication(buffer, 11, |w| {
                w.bytes(data);
            }),
            BackendMessage::AuthenticationSASLFinal { data } => encode_authentication(buffer, 12, |w| {
                w.bytes(data);
            }),
            BackendMessage::BackendKeyData { process_id, secret_key } => {
                let mut w = Writer::typed(buffer, b'K');
                w.u32(*process_id).bytes(secret_key);
                w.finish();
            }
            BackendMessage::BindComplete => Writer::typed(buffer, b'2').finish(),
            BackendMessage::CloseComplete => Writer::typed(buffer, b'3').finish(),
            BackendMessage::CommandComplete { command_tag } => {
                let mut w = Writer::typed(buffer, b'C');
                w.cstring(command_tag);
                w.finish();
            }
            BackendMessage::CopyBothResponse { overall_format, column_format_codes } => {
                encode_copy_response(buffer, b'W', *overall_format, column_format_codes)
            }
            BackendMessage::CopyData { data } => {
                let mut w = Writer::typed(buffer, b'd');
                w.bytes(data);
                w.finish();
            }
            BackendMessage::CopyDone => Writer::typed(buffer, b'c').finish(),
            BackendMessage::CopyInResponse { overall_format, column_format_codes } => {
                encode_copy_response(buffer, b'G', *overall_format, column_format_codes)
            }
            BackendMessage::CopyOutResponse { overall_format, column_format_codes } => {
                encode_copy_response(buffer, b'H', *overall_format, column_format_codes)
            }
            BackendMessage::DataRow { values } => {
                let mut w = Writer::typed(buffer, b'D');
                w.u16(values.len() as u16);
                for value in values {
                    w.nullable_bytes(value);
                }
                w.finish();
            }
            BackendMessage::EmptyQueryResponse => Writer::typed(buffer, b'I').finish(),
            BackendMessage::ErrorResponse(fields) => encode_error_fields(buffer, b'E', fields),
            BackendMessage::FunctionCallResponse { result } => {
                let mut w = Writer::typed(buffer, b'V');
                w.nullable_bytes(result);
                w.finish();
            }
            BackendMessage::NegotiateProtocolVersion { newest_minor_protocol, unrecognized_options } => {
                let mut w = Writer::typed(buffer, b'v');
                w.u32(*newest_minor_protocol).u32(unrecognized_options.len() as u32);
                for option in unrecognized_options {
                    w.cstring(option);
                }
                w.finish();
            }
            BackendMessage::NoData => Writer::typed(buffer, b'n').finish(),
            BackendMessage::NoticeResponse(fields) => encode_error_fields(buffer, b'N', fields),
            BackendMessage::NotificationResponse { process_id, channel, payload } => {
                let mut w = Writer::typed(buffer, b'A');
                w.u32(*process_id).cstring(channel).cstring(payload);
                w.finish();
            }
            BackendMessage::ParameterDescription { parameter_oids } => {
                let mut w = Writer::typed(buffer, b't');
                w.u16(parameter_oids.len() as u16);
                for oid in parameter_oids {
                    w.u32(*oid);
                }
                w.finish();
            }
            BackendMessage::ParameterStatus { name, value } => {
                let mut w = Writer::typed(buffer, b'S');
                w.cstring(name).cstring(value);
                w.finish();
            }
            BackendMessage::ParseComplete => Writer::typed(buffer, b'1').finish(),
            BackendMessage::PortalSuspended => Writer::typed(buffer, b's').finish(),
            BackendMessage::ReadyForQuery { tx_status } => {
                let mut w = Writer::typed(buffer, b'Z');
                w.u8(*tx_status);
                w.finish();
            }
            BackendMessage::RowDescription { fields } => {
                let mut w = Writer::typed(buffer, b'T');
                w.u16(fields.len() as u16);
                for field in fields {
                    w.cstring(&field.name)
                        .u32(field.table_oid)
                        .u16(field.table_attribute_number)
                        .u32(field.data_type_oid)
                        .i16(field.data_type_size)
                        .i32(field.type_modifier)
                        .i16(field.format);
                }
                w.finish();
            }
        }
    }

    /// decode parses a message from its type byte and body.
    pub fn decode(tag: u8, body: &[u8]) -> Result<BackendMessage, DecodeError> {
        let mut r = Reader::new(body);
        let message = match tag {
            b'R' => decode_authentication(&mut r)?,
            b'K' => BackendMessage::BackendKeyData { process_id: r.u32()?, secret_key: r.rest().to_vec() },
            b'2' => BackendMessage::BindComplete,
            b'3' => BackendMessage::CloseComplete,
            b'C' => BackendMessage::CommandComplete { command_tag: r.cstring()? },
            b'W' => {
                let (overall_format, column_format_codes) = decode_copy_response(&mut r)?;
                BackendMessage::CopyBothResponse { overall_format, column_format_codes }
            }
            b'd' => BackendMessage::CopyData { data: r.rest().to_vec() },
            b'c' => BackendMessage::CopyDone,
            b'G' => {
                let (overall_format, column_format_codes) = decode_copy_response(&mut r)?;
                BackendMessage::CopyInResponse { overall_format, column_format_codes }
            }
            b'H' => {
                let (overall_format, column_format_codes) = decode_copy_response(&mut r)?;
                BackendMessage::CopyOutResponse { overall_format, column_format_codes }
            }
            b'D' => {
                let count = r.u16()?;
                let mut values = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    values.push(r.nullable_bytes()?);
                }
                BackendMessage::DataRow { values }
            }
            b'I' => BackendMessage::EmptyQueryResponse,
            b'E' => BackendMessage::ErrorResponse(decode_error_fields(&mut r)?),
            b'V' => BackendMessage::FunctionCallResponse { result: r.nullable_bytes()? },
            b'v' => {
                let newest_minor_protocol = r.u32()?;
                let count = r.u32()?;
                let mut unrecognized_options = Vec::new();
                for _ in 0..count {
                    unrecognized_options.push(r.cstring()?);
                }
                BackendMessage::NegotiateProtocolVersion { newest_minor_protocol, unrecognized_options }
            }
            b'n' => BackendMessage::NoData,
            b'N' => BackendMessage::NoticeResponse(decode_error_fields(&mut r)?),
            b'A' => BackendMessage::NotificationResponse {
                process_id: r.u32()?,
                channel: r.cstring()?,
                payload: r.cstring()?,
            },
            b't' => {
                let count = r.u16()?;
                let mut parameter_oids = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    parameter_oids.push(r.u32()?);
                }
                BackendMessage::ParameterDescription { parameter_oids }
            }
            b'S' => BackendMessage::ParameterStatus { name: r.cstring()?, value: r.cstring()? },
            b'1' => BackendMessage::ParseComplete,
            b's' => BackendMessage::PortalSuspended,
            b'Z' => BackendMessage::ReadyForQuery { tx_status: r.u8()? },
            b'T' => {
                let count = r.u16()?;
                let mut fields = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    fields.push(FieldDescription {
                        name: r.cstring()?,
                        table_oid: r.u32()?,
                        table_attribute_number: r.u16()?,
                        data_type_oid: r.u32()?,
                        data_type_size: r.i16()?,
                        type_modifier: r.i32()?,
                        format: r.i16()?,
                    });
                }
                BackendMessage::RowDescription { fields }
            }
            _ => return Err(DecodeError::new(format!("unknown backend message type: {}", tag as char))),
        };
        if r.remaining() != 0 {
            return Err(DecodeError::new(format!(
                "backend message type {} has {} trailing bytes",
                tag as char,
                r.remaining()
            )));
        }
        Ok(message)
    }
}

/// encode_authentication writes an Authentication message with the given subtype.
fn encode_authentication(buffer: &mut Vec<u8>, subtype: i32, body: impl FnOnce(&mut Writer<'_>)) {
    let mut w = Writer::typed(buffer, b'R');
    w.i32(subtype);
    body(&mut w);
    w.finish();
}

/// decode_authentication parses the body of an Authentication message.
fn decode_authentication(r: &mut Reader<'_>) -> Result<BackendMessage, DecodeError> {
    Ok(match r.i32()? {
        0 => BackendMessage::AuthenticationOk,
        3 => BackendMessage::AuthenticationCleartextPassword,
        5 => {
            let salt = r.take(4)?;
            BackendMessage::AuthenticationMD5Password { salt: [salt[0], salt[1], salt[2], salt[3]] }
        }
        7 => BackendMessage::AuthenticationGSS,
        8 => BackendMessage::AuthenticationGSSContinue { data: r.rest().to_vec() },
        10 => {
            let mut mechanisms = Vec::new();
            loop {
                let mechanism = r.cstring()?;
                if mechanism.is_empty() {
                    break;
                }
                mechanisms.push(mechanism);
            }
            BackendMessage::AuthenticationSASL { mechanisms }
        }
        11 => BackendMessage::AuthenticationSASLContinue { data: r.rest().to_vec() },
        12 => BackendMessage::AuthenticationSASLFinal { data: r.rest().to_vec() },
        subtype => return Err(DecodeError::new(format!("unknown authentication type: {subtype}"))),
    })
}

/// encode_copy_response writes a CopyInResponse, CopyOutResponse, or CopyBothResponse.
fn encode_copy_response(buffer: &mut Vec<u8>, tag: u8, overall_format: u8, column_format_codes: &[u16]) {
    let mut w = Writer::typed(buffer, tag);
    w.u8(overall_format).u16(column_format_codes.len() as u16);
    for code in column_format_codes {
        w.u16(*code);
    }
    w.finish();
}

/// decode_copy_response parses the body of a CopyInResponse, CopyOutResponse, or CopyBothResponse.
fn decode_copy_response(r: &mut Reader<'_>) -> Result<(u8, Vec<u16>), DecodeError> {
    let overall_format = r.u8()?;
    let count = r.u16()?;
    let mut codes = Vec::with_capacity(count as usize);
    for _ in 0..count {
        codes.push(r.u16()?);
    }
    Ok((overall_format, codes))
}

/// encode_error_fields writes an ErrorResponse or NoticeResponse. Fields are written in a fixed order and absent
/// fields are omitted.
fn encode_error_fields(buffer: &mut Vec<u8>, tag: u8, fields: &ErrorFields) {
    let mut w = Writer::typed(buffer, tag);
    let strings = [
        (b'S', &fields.severity),
        (b'V', &fields.severity_unlocalized),
        (b'C', &fields.code),
        (b'M', &fields.message),
        (b'D', &fields.detail),
        (b'H', &fields.hint),
    ];
    for (code, value) in strings {
        if !value.is_empty() {
            w.u8(code).cstring(value);
        }
    }
    if fields.position != 0 {
        w.u8(b'P').cstring(&fields.position.to_string());
    }
    if fields.internal_position != 0 {
        w.u8(b'p').cstring(&fields.internal_position.to_string());
    }
    let strings = [
        (b'q', &fields.internal_query),
        (b'W', &fields.where_),
        (b's', &fields.schema_name),
        (b't', &fields.table_name),
        (b'c', &fields.column_name),
        (b'd', &fields.data_type_name),
        (b'n', &fields.constraint_name),
        (b'F', &fields.file),
    ];
    for (code, value) in strings {
        if !value.is_empty() {
            w.u8(code).cstring(value);
        }
    }
    if fields.line != 0 {
        w.u8(b'L').cstring(&fields.line.to_string());
    }
    if !fields.routine.is_empty() {
        w.u8(b'R').cstring(&fields.routine);
    }
    for (code, value) in &fields.unknown_fields {
        w.u8(*code).cstring(value);
    }
    w.u8(0);
    w.finish();
}

/// decode_error_fields parses the body of an ErrorResponse or NoticeResponse.
fn decode_error_fields(r: &mut Reader<'_>) -> Result<ErrorFields, DecodeError> {
    let mut fields = ErrorFields::default();
    loop {
        let code = r.u8()?;
        if code == 0 {
            return Ok(fields);
        }
        let value = r.cstring()?;
        match code {
            b'S' => fields.severity = value,
            b'V' => fields.severity_unlocalized = value,
            b'C' => fields.code = value,
            b'M' => fields.message = value,
            b'D' => fields.detail = value,
            b'H' => fields.hint = value,
            b'P' => fields.position = parse_error_number(&value),
            b'p' => fields.internal_position = parse_error_number(&value),
            b'q' => fields.internal_query = value,
            b'W' => fields.where_ = value,
            b's' => fields.schema_name = value,
            b't' => fields.table_name = value,
            b'c' => fields.column_name = value,
            b'd' => fields.data_type_name = value,
            b'n' => fields.constraint_name = value,
            b'F' => fields.file = value,
            b'L' => fields.line = parse_error_number(&value),
            b'R' => fields.routine = value,
            _ => {
                fields.unknown_fields.insert(code, value);
            }
        }
    }
}

/// parse_error_number parses a numeric error field like pgproto3, which reads text that is not a number as zero and
/// clamps numbers outside of the int32 range.
fn parse_error_number(value: &str) -> i32 {
    let digits = value.strip_prefix(['+', '-']).unwrap_or(value);
    if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
        return 0;
    }
    value.parse::<i32>().unwrap_or(if value.starts_with('-') { i32::MIN } else { i32::MAX })
}
