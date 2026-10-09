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

//! The expected bytes were produced by encoding the same messages with pgx v5.9.2's pgproto3, which the Go
//! version of Doltgres uses for its wire protocol.

use pgproto::{BackendMessage, ErrorFields, FieldDescription, FrameReader, FrontendMessage, PasswordKind};

/// hex renders bytes as lowercase hexadecimal.
fn hex(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

/// check_backend asserts that the message encodes to the expected bytes and decodes back to itself.
fn check_backend(name: &str, message: BackendMessage, expected: &str) {
    let mut encoded = Vec::new();
    message.encode(&mut encoded);
    assert_eq!(hex(&encoded), expected, "{name}: encoding");
    let mut reader = FrameReader::new();
    reader.extend(&encoded);
    let frame = reader.next_frame().unwrap().unwrap();
    assert!(reader.buffered().is_empty(), "{name}: leftover bytes");
    assert_eq!(BackendMessage::decode(frame.tag, &frame.body).unwrap(), message, "{name}: decoding");
}

/// check_frontend asserts that the message encodes to the expected bytes and decodes back to itself.
fn check_frontend(name: &str, message: FrontendMessage, kind: PasswordKind, expected: &str) {
    let mut encoded = Vec::new();
    message.encode(&mut encoded);
    assert_eq!(hex(&encoded), expected, "{name}: encoding");
    let mut reader = FrameReader::new();
    reader.extend(&encoded);
    let decoded = match message {
        FrontendMessage::CancelRequest { .. }
        | FrontendMessage::GSSEncRequest
        | FrontendMessage::SSLRequest
        | FrontendMessage::StartupMessage { .. } => {
            FrontendMessage::decode_startup(&reader.next_untyped_frame().unwrap().unwrap()).unwrap()
        }
        _ => {
            let frame = reader.next_frame().unwrap().unwrap();
            FrontendMessage::decode(frame.tag, &frame.body, kind).unwrap()
        }
    };
    assert!(reader.buffered().is_empty(), "{name}: leftover bytes");
    assert_eq!(decoded, message, "{name}: decoding");
}

#[test]
fn backend_messages_match_pgproto3() {
    check_backend("auth_ok", BackendMessage::AuthenticationOk, "520000000800000000");
    check_backend("auth_cleartext", BackendMessage::AuthenticationCleartextPassword, "520000000800000003");
    check_backend(
        "auth_md5",
        BackendMessage::AuthenticationMD5Password { salt: [1, 2, 3, 4] },
        "520000000c0000000501020304",
    );
    check_backend(
        "auth_sasl",
        BackendMessage::AuthenticationSASL {
            mechanisms: vec!["SCRAM-SHA-256".to_string(), "SCRAM-SHA-256-PLUS".to_string()],
        },
        "520000002a0000000a534352414d2d5348412d32353600534352414d2d5348412d3235362d504c55530000",
    );
    check_backend(
        "auth_sasl_continue",
        BackendMessage::AuthenticationSASLContinue { data: b"r=abc,s=def,i=4096".to_vec() },
        "520000001a0000000b723d6162632c733d6465662c693d34303936",
    );
    check_backend(
        "auth_sasl_final",
        BackendMessage::AuthenticationSASLFinal { data: b"v=xyz".to_vec() },
        "520000000d0000000c763d78797a",
    );
    check_backend(
        "backend_key_data",
        BackendMessage::BackendKeyData { process_id: 1234, secret_key: vec![0xde, 0xad, 0xbe, 0xef] },
        "4b0000000c000004d2deadbeef",
    );
    check_backend("bind_complete", BackendMessage::BindComplete, "3200000004");
    check_backend("close_complete", BackendMessage::CloseComplete, "3300000004");
    check_backend(
        "command_complete",
        BackendMessage::CommandComplete { command_tag: "INSERT 0 1".to_string() },
        "430000000f494e534552542030203100",
    );
    check_backend(
        "copy_in_response",
        BackendMessage::CopyInResponse { overall_format: 0, column_format_codes: vec![0, 0, 1] },
        "470000000d000003000000000001",
    );
    check_backend(
        "copy_out_response",
        BackendMessage::CopyOutResponse { overall_format: 1, column_format_codes: vec![1] },
        "48000000090100010001",
    );
    check_backend(
        "copy_both_response",
        BackendMessage::CopyBothResponse { overall_format: 0, column_format_codes: vec![] },
        "5700000007000000",
    );
    check_backend("copy_data", BackendMessage::CopyData { data: b"1\tabc\n".to_vec() }, "640000000a31096162630a");
    check_backend("copy_done", BackendMessage::CopyDone, "6300000004");
    check_backend(
        "data_row",
        BackendMessage::DataRow { values: vec![Some(b"7".to_vec()), None, Some(vec![])] },
        "440000001300030000000137ffffffff00000000",
    );
    check_backend("empty_query_response", BackendMessage::EmptyQueryResponse, "4900000004");
    check_backend(
        "error_response",
        BackendMessage::ErrorResponse(ErrorFields {
            severity: "ERROR".to_string(),
            severity_unlocalized: "ERROR".to_string(),
            code: "23505".to_string(),
            message: r#"duplicate key value violates unique constraint "t_pkey""#.to_string(),
            detail: "Key (i)=(1) already exists.".to_string(),
            hint: "hint".to_string(),
            position: 15,
            internal_position: 3,
            internal_query: "SELECT 1".to_string(),
            where_: "PL/pgSQL function f()".to_string(),
            schema_name: "public".to_string(),
            table_name: "t".to_string(),
            column_name: "i".to_string(),
            data_type_name: "int4".to_string(),
            constraint_name: "t_pkey".to_string(),
            file: "nbtinsert.c".to_string(),
            line: 666,
            routine: "_bt_check_unique".to_string(),
            ..ErrorFields::default()
        }),
        "45000000de534552524f5200564552524f5200433233353035004d6475706c6963617465206b65792076616c75652076696f6c61746573\
         20756e6971756520636f6e73747261696e742022745f706b65792200444b6579202869293d28312920616c7265616479206578697374\
         732e004868696e7400503135007033007153454c45435420310057504c2f706753514c2066756e6374696f6e2066282900737075626c\
         69630074740063690064696e7434006e745f706b657900466e6274696e736572742e63004c36363600525f62745f636865636b5f756e\
         697175650000",
    );
    check_backend(
        "error_response_minimal",
        BackendMessage::ErrorResponse(ErrorFields {
            severity: "ERROR".to_string(),
            code: "42601".to_string(),
            message: "syntax error".to_string(),
            ..ErrorFields::default()
        }),
        "4500000021534552524f5200433432363031004d73796e746178206572726f720000",
    );
    check_backend(
        "function_call_response",
        BackendMessage::FunctionCallResponse { result: Some(vec![0, 0, 0, 1]) },
        "560000000c0000000400000001",
    );
    check_backend(
        "function_call_response_null",
        BackendMessage::FunctionCallResponse { result: None },
        "5600000008ffffffff",
    );
    check_backend(
        "negotiate_protocol_version",
        BackendMessage::NegotiateProtocolVersion {
            newest_minor_protocol: 0,
            unrecognized_options: vec!["_pq_.opt".to_string()],
        },
        "760000001500000000000000015f70715f2e6f707400",
    );
    check_backend("no_data", BackendMessage::NoData, "6e00000004");
    check_backend(
        "notice_response",
        BackendMessage::NoticeResponse(ErrorFields {
            severity: "WARNING".to_string(),
            severity_unlocalized: "WARNING".to_string(),
            code: "25P01".to_string(),
            message: "there is no transaction in progress".to_string(),
            ..ErrorFields::default()
        }),
        "4e00000043535741524e494e4700565741524e494e4700433235503031004d7468657265206973206e6f207472616e73616374696f6e\
         20696e2070726f67726573730000",
    );
    check_backend(
        "notification_response",
        BackendMessage::NotificationResponse {
            process_id: 42,
            channel: "chan".to_string(),
            payload: "payload".to_string(),
        },
        "41000000150000002a6368616e007061796c6f616400",
    );
    check_backend(
        "parameter_description",
        BackendMessage::ParameterDescription { parameter_oids: vec![23, 25] },
        "740000000e00020000001700000019",
    );
    check_backend(
        "parameter_status",
        BackendMessage::ParameterStatus { name: "server_version".to_string(), value: "15.0".to_string() },
        "53000000187365727665725f76657273696f6e0031352e3000",
    );
    check_backend("parse_complete", BackendMessage::ParseComplete, "3100000004");
    check_backend("portal_suspended", BackendMessage::PortalSuspended, "7300000004");
    check_backend("ready_for_query", BackendMessage::ReadyForQuery { tx_status: b'I' }, "5a0000000549");
    check_backend(
        "row_description",
        BackendMessage::RowDescription {
            fields: vec![
                FieldDescription {
                    name: "pk".to_string(),
                    table_oid: 16384,
                    table_attribute_number: 1,
                    data_type_oid: 23,
                    data_type_size: 4,
                    type_modifier: -1,
                    format: 0,
                },
                FieldDescription {
                    name: "v".to_string(),
                    table_oid: 0,
                    table_attribute_number: 0,
                    data_type_oid: 1043,
                    data_type_size: -1,
                    type_modifier: 14,
                    format: 1,
                },
            ],
        },
        "540000002f0002706b00000040000001000000170004ffffffff0000760000000000000000000413ffff0000000e0001",
    );
}

#[test]
fn frontend_messages_match_pgproto3() {
    let password = PasswordKind::Password;
    check_frontend(
        "bind",
        FrontendMessage::Bind {
            destination_portal: "p".to_string(),
            prepared_statement: "s".to_string(),
            parameter_format_codes: vec![1],
            parameters: vec![Some(vec![0, 0, 0, 1]), None],
            result_format_codes: vec![0, 1],
        },
        password,
        "4200000020700073000001000100020000000400000001ffffffff000200000001",
    );
    check_frontend(
        "cancel_request",
        FrontendMessage::CancelRequest { process_id: 1234, secret_key: vec![0, 0, 0x16, 0x2e] },
        password,
        "0000001004d2162e000004d20000162e",
    );
    check_frontend(
        "close",
        FrontendMessage::Close { object_type: b'S', name: "stmt".to_string() },
        password,
        "430000000a5373746d7400",
    );
    check_frontend(
        "copy_fail",
        FrontendMessage::CopyFail { message: "aborted".to_string() },
        password,
        "660000000c61626f7274656400",
    );
    check_frontend(
        "describe",
        FrontendMessage::Describe { object_type: b'P', name: String::new() },
        password,
        "44000000065000",
    );
    check_frontend(
        "execute",
        FrontendMessage::Execute { portal: String::new(), max_rows: 10 },
        password,
        "4500000009000000000a",
    );
    check_frontend("flush", FrontendMessage::Flush, password, "4800000004");
    check_frontend(
        "function_call",
        FrontendMessage::FunctionCall {
            function: 1598,
            argument_format_codes: vec![1],
            arguments: vec![Some(vec![1]), None],
            result_format_code: 1,
        },
        password,
        "46000000190000063e0001000100020000000101ffffffff0001",
    );
    check_frontend("gss_enc_request", FrontendMessage::GSSEncRequest, password, "0000000804d21630");
    check_frontend(
        "parse",
        FrontendMessage::Parse {
            name: "s".to_string(),
            query: "SELECT $1::int4".to_string(),
            parameter_oids: vec![23],
        },
        password,
        "500000001c730053454c4543542024313a3a696e743400000100000017",
    );
    check_frontend(
        "password_message",
        FrontendMessage::PasswordMessage { password: "password".to_string() },
        password,
        "700000000d70617373776f726400",
    );
    check_frontend(
        "query",
        FrontendMessage::Query { query: "SELECT 1;".to_string() },
        password,
        "510000000e53454c45435420313b00",
    );
    check_frontend(
        "sasl_initial_response",
        FrontendMessage::SASLInitialResponse {
            auth_mechanism: "SCRAM-SHA-256".to_string(),
            data: Some(b"n,,n=,r=abc".to_vec()),
        },
        PasswordKind::SASLInitialResponse,
        "7000000021534352414d2d5348412d323536000000000b6e2c2c6e3d2c723d616263",
    );
    check_frontend(
        "sasl_response",
        FrontendMessage::SASLResponse { data: b"c=biws,r=abc,p=xyz".to_vec() },
        PasswordKind::SASLResponse,
        "7000000016633d626977732c723d6162632c703d78797a",
    );
    check_frontend("ssl_request", FrontendMessage::SSLRequest, password, "0000000804d2162f");
    check_frontend(
        "startup_message",
        FrontendMessage::StartupMessage {
            protocol_version: pgproto::PROTOCOL_VERSION_3 as u32,
            parameters: vec![("user".to_string(), "postgres".to_string())],
        },
        password,
        "00000017000300007573657200706f7374677265730000",
    );
    check_frontend("sync", FrontendMessage::Sync, password, "5300000004");
    check_frontend("terminate", FrontendMessage::Terminate, password, "5800000004");
}

#[test]
fn error_numbers_decode_like_pgproto3() {
    // A message with an embedded NUL leaves its tail to be read as an internal position, which pgproto3 reads as zero
    let body = b"SERROR\0Mroot returned table `fkpart1.\0pk11` but it could not be found\0P99999999999\0L-7\0\0";
    let BackendMessage::ErrorResponse(fields) = BackendMessage::decode(b'E', body).unwrap() else {
        panic!("not an ErrorResponse");
    };
    assert_eq!(fields.message, "root returned table `fkpart1.");
    assert_eq!(fields.internal_position, 0);
    assert_eq!(fields.position, i32::MAX);
    assert_eq!(fields.line, -7);
}
