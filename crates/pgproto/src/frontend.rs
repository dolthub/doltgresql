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

use crate::codec::{DecodeError, Reader, Writer};
use crate::{CANCEL_REQUEST_CODE, GSSENC_REQUEST_CODE, SSL_REQUEST_CODE};

/// PasswordKind selects how a 'p' message is decoded, since its meaning depends on the authentication exchange.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum PasswordKind {
    /// A PasswordMessage holding a cleartext or hashed password.
    Password,
    /// A SASLInitialResponse that starts a SASL exchange.
    SASLInitialResponse,
    /// A SASLResponse that continues a SASL exchange.
    SASLResponse,
    /// A GSSResponse holding GSSAPI or SSPI data.
    GSSResponse,
}

/// FrontendMessage is a message sent from the client to the server.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum FrontendMessage {
    /// Bind creates a portal from a prepared statement, where None parameters are NULL.
    Bind {
        destination_portal: String,
        prepared_statement: String,
        parameter_format_codes: Vec<i16>,
        parameters: Vec<Option<Vec<u8>>>,
        result_format_codes: Vec<i16>,
    },
    /// CancelRequest asks the server to cancel the query running on another connection.
    CancelRequest { process_id: u32, secret_key: Vec<u8> },
    /// Close closes a prepared statement ('S') or a portal ('P').
    Close { object_type: u8, name: String },
    /// CopyData carries copy data.
    CopyData { data: Vec<u8> },
    /// CopyDone ends a copy.
    CopyDone,
    /// CopyFail aborts a copy with the given reason.
    CopyFail { message: String },
    /// Describe describes a prepared statement ('S') or a portal ('P').
    Describe { object_type: u8, name: String },
    /// Execute runs a portal, returning at most max_rows rows when it is non-zero.
    Execute { portal: String, max_rows: u32 },
    /// Flush asks the server to send all pending output.
    Flush,
    /// FunctionCall calls a function by OID, where None arguments are NULL.
    FunctionCall {
        function: u32,
        argument_format_codes: Vec<u16>,
        arguments: Vec<Option<Vec<u8>>>,
        result_format_code: u16,
    },
    /// GSSEncRequest asks to start GSSAPI encryption.
    GSSEncRequest,
    /// GSSResponse carries GSSAPI or SSPI data.
    GSSResponse { data: Vec<u8> },
    /// Parse creates a prepared statement, where zero parameter OIDs are left unspecified.
    Parse { name: String, query: String, parameter_oids: Vec<u32> },
    /// PasswordMessage carries a password.
    PasswordMessage { password: String },
    /// Query runs a simple query.
    Query { query: String },
    /// SASLInitialResponse picks a SASL mechanism, where None data means the client sent no initial response.
    SASLInitialResponse { auth_mechanism: String, data: Option<Vec<u8>> },
    /// SASLResponse carries SASL response data.
    SASLResponse { data: Vec<u8> },
    /// SSLRequest asks to start TLS.
    SSLRequest,
    /// StartupMessage starts a session with the given protocol version and parameters, in the order sent.
    StartupMessage { protocol_version: u32, parameters: Vec<(String, String)> },
    /// Sync ends an extended query cycle.
    Sync,
    /// Terminate closes the session.
    Terminate,
}

impl FrontendMessage {
    /// encode appends the message to the buffer.
    pub fn encode(&self, buffer: &mut Vec<u8>) {
        match self {
            FrontendMessage::Bind {
                destination_portal,
                prepared_statement,
                parameter_format_codes,
                parameters,
                result_format_codes,
            } => {
                let mut w = Writer::typed(buffer, b'B');
                w.cstring(destination_portal).cstring(prepared_statement);
                w.u16(parameter_format_codes.len() as u16);
                for code in parameter_format_codes {
                    w.i16(*code);
                }
                w.u16(parameters.len() as u16);
                for parameter in parameters {
                    w.nullable_bytes(parameter);
                }
                w.u16(result_format_codes.len() as u16);
                for code in result_format_codes {
                    w.i16(*code);
                }
                w.finish();
            }
            FrontendMessage::CancelRequest { process_id, secret_key } => {
                let mut w = Writer::untyped(buffer);
                w.i32(CANCEL_REQUEST_CODE).u32(*process_id).bytes(secret_key);
                w.finish();
            }
            FrontendMessage::Close { object_type, name } => {
                let mut w = Writer::typed(buffer, b'C');
                w.u8(*object_type).cstring(name);
                w.finish();
            }
            FrontendMessage::CopyData { data } => {
                let mut w = Writer::typed(buffer, b'd');
                w.bytes(data);
                w.finish();
            }
            FrontendMessage::CopyDone => Writer::typed(buffer, b'c').finish(),
            FrontendMessage::CopyFail { message } => {
                let mut w = Writer::typed(buffer, b'f');
                w.cstring(message);
                w.finish();
            }
            FrontendMessage::Describe { object_type, name } => {
                let mut w = Writer::typed(buffer, b'D');
                w.u8(*object_type).cstring(name);
                w.finish();
            }
            FrontendMessage::Execute { portal, max_rows } => {
                let mut w = Writer::typed(buffer, b'E');
                w.cstring(portal).u32(*max_rows);
                w.finish();
            }
            FrontendMessage::Flush => Writer::typed(buffer, b'H').finish(),
            FrontendMessage::FunctionCall { function, argument_format_codes, arguments, result_format_code } => {
                let mut w = Writer::typed(buffer, b'F');
                w.u32(*function).u16(argument_format_codes.len() as u16);
                for code in argument_format_codes {
                    w.u16(*code);
                }
                w.u16(arguments.len() as u16);
                for argument in arguments {
                    w.nullable_bytes(argument);
                }
                w.u16(*result_format_code);
                w.finish();
            }
            FrontendMessage::GSSEncRequest => {
                let mut w = Writer::untyped(buffer);
                w.i32(GSSENC_REQUEST_CODE);
                w.finish();
            }
            FrontendMessage::GSSResponse { data } => {
                let mut w = Writer::typed(buffer, b'p');
                w.bytes(data);
                w.finish();
            }
            FrontendMessage::Parse { name, query, parameter_oids } => {
                let mut w = Writer::typed(buffer, b'P');
                w.cstring(name).cstring(query).u16(parameter_oids.len() as u16);
                for oid in parameter_oids {
                    w.u32(*oid);
                }
                w.finish();
            }
            FrontendMessage::PasswordMessage { password } => {
                let mut w = Writer::typed(buffer, b'p');
                w.cstring(password);
                w.finish();
            }
            FrontendMessage::Query { query } => {
                let mut w = Writer::typed(buffer, b'Q');
                w.cstring(query);
                w.finish();
            }
            FrontendMessage::SASLInitialResponse { auth_mechanism, data } => {
                let mut w = Writer::typed(buffer, b'p');
                w.cstring(auth_mechanism).nullable_bytes(data);
                w.finish();
            }
            FrontendMessage::SASLResponse { data } => {
                let mut w = Writer::typed(buffer, b'p');
                w.bytes(data);
                w.finish();
            }
            FrontendMessage::SSLRequest => {
                let mut w = Writer::untyped(buffer);
                w.i32(SSL_REQUEST_CODE);
                w.finish();
            }
            FrontendMessage::StartupMessage { protocol_version, parameters } => {
                let mut w = Writer::untyped(buffer);
                w.u32(*protocol_version);
                for (name, value) in parameters {
                    w.cstring(name).cstring(value);
                }
                w.u8(0);
                w.finish();
            }
            FrontendMessage::Sync => Writer::typed(buffer, b'S').finish(),
            FrontendMessage::Terminate => Writer::typed(buffer, b'X').finish(),
        }
    }

    /// decode_startup parses an untyped message sent before the session starts: a StartupMessage, SSLRequest,
    /// GSSEncRequest, or CancelRequest. The body excludes the length.
    pub fn decode_startup(body: &[u8]) -> Result<FrontendMessage, DecodeError> {
        let mut r = Reader::new(body);
        let code = r.i32()?;
        let message = match code {
            SSL_REQUEST_CODE => FrontendMessage::SSLRequest,
            GSSENC_REQUEST_CODE => FrontendMessage::GSSEncRequest,
            CANCEL_REQUEST_CODE => {
                FrontendMessage::CancelRequest { process_id: r.u32()?, secret_key: r.rest().to_vec() }
            }
            _ => {
                let mut parameters = Vec::new();
                loop {
                    let name = r.cstring()?;
                    if name.is_empty() {
                        break;
                    }
                    parameters.push((name, r.cstring()?));
                }
                FrontendMessage::StartupMessage { protocol_version: code as u32, parameters }
            }
        };
        expect_end(&r, "startup")?;
        Ok(message)
    }

    /// decode parses a typed message from its type byte and body. The password kind selects how a 'p' message is
    /// read.
    pub fn decode(tag: u8, body: &[u8], password_kind: PasswordKind) -> Result<FrontendMessage, DecodeError> {
        let mut r = Reader::new(body);
        let message = match tag {
            b'B' => {
                let destination_portal = r.cstring()?;
                let prepared_statement = r.cstring()?;
                let count = r.u16()?;
                let mut parameter_format_codes = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    parameter_format_codes.push(r.i16()?);
                }
                let count = r.u16()?;
                let mut parameters = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    parameters.push(r.nullable_bytes()?);
                }
                let count = r.u16()?;
                let mut result_format_codes = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    result_format_codes.push(r.i16()?);
                }
                FrontendMessage::Bind {
                    destination_portal,
                    prepared_statement,
                    parameter_format_codes,
                    parameters,
                    result_format_codes,
                }
            }
            b'C' => FrontendMessage::Close { object_type: r.u8()?, name: r.cstring()? },
            b'd' => FrontendMessage::CopyData { data: r.rest().to_vec() },
            b'c' => FrontendMessage::CopyDone,
            b'f' => FrontendMessage::CopyFail { message: r.cstring()? },
            b'D' => FrontendMessage::Describe { object_type: r.u8()?, name: r.cstring()? },
            b'E' => FrontendMessage::Execute { portal: r.cstring()?, max_rows: r.u32()? },
            b'H' => FrontendMessage::Flush,
            b'F' => {
                let function = r.u32()?;
                let count = r.u16()?;
                let mut argument_format_codes = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    argument_format_codes.push(r.u16()?);
                }
                let count = r.u16()?;
                let mut arguments = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    arguments.push(r.nullable_bytes()?);
                }
                FrontendMessage::FunctionCall {
                    function,
                    argument_format_codes,
                    arguments,
                    result_format_code: r.u16()?,
                }
            }
            b'P' => {
                let name = r.cstring()?;
                let query = r.cstring()?;
                let count = r.u16()?;
                let mut parameter_oids = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    parameter_oids.push(r.u32()?);
                }
                FrontendMessage::Parse { name, query, parameter_oids }
            }
            b'p' => match password_kind {
                PasswordKind::Password => FrontendMessage::PasswordMessage { password: r.cstring()? },
                PasswordKind::SASLInitialResponse => {
                    FrontendMessage::SASLInitialResponse { auth_mechanism: r.cstring()?, data: r.nullable_bytes()? }
                }
                PasswordKind::SASLResponse => FrontendMessage::SASLResponse { data: r.rest().to_vec() },
                PasswordKind::GSSResponse => FrontendMessage::GSSResponse { data: r.rest().to_vec() },
            },
            b'Q' => FrontendMessage::Query { query: r.cstring()? },
            b'S' => FrontendMessage::Sync,
            b'X' => FrontendMessage::Terminate,
            _ => return Err(DecodeError::new(format!("unknown frontend message type: {}", tag as char))),
        };
        expect_end(&r, &(tag as char).to_string())?;
        Ok(message)
    }
}

/// expect_end returns an error when the reader has unread bytes.
fn expect_end(r: &Reader<'_>, name: &str) -> Result<(), DecodeError> {
    if r.remaining() != 0 {
        return Err(DecodeError::new(format!("{name} message has {} trailing bytes", r.remaining())));
    }
    Ok(())
}
