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

#![forbid(unsafe_code)]

//! Postgres wire protocol (version 3.0) messages for both the frontend and the backend. This crate only
//! encodes and decodes bytes, and never performs IO.

mod backend;
mod codec;
mod frontend;

pub use backend::{BackendMessage, ErrorFields, FieldDescription};
pub use codec::{DecodeError, Frame, FrameReader};
pub use frontend::{FrontendMessage, PasswordKind};

/// The protocol version number sent in a StartupMessage for protocol 3.0.
pub const PROTOCOL_VERSION_3: i32 = 196608;
/// The magic number that identifies an SSLRequest.
pub const SSL_REQUEST_CODE: i32 = 80877103;
/// The magic number that identifies a GSSENCRequest.
pub const GSSENC_REQUEST_CODE: i32 = 80877104;
/// The magic number that identifies a CancelRequest.
pub const CANCEL_REQUEST_CODE: i32 = 80877102;
