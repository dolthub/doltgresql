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

//! The extensions that Doltgres emulates, whose objects CREATE EXTENSION writes as their installation scripts would and
//! whose library functions Rust implements.

mod index;
mod install;
mod uuid_ossp;
mod vector;

pub use index::{no_default_class, vector_rendering};
pub use install::COLLECTION;

use std::cmp::Ordering;
use std::sync::OnceLock;

use crate::catalog::ColumnType;
use crate::error::Result;
use crate::query::Ctx;
use crate::types::Value;

/// Elements returns the elements of a stored vector.
pub type Elements = fn(&[u8]) -> Vec<f32>;

/// Implementation computes a routine's result from its arguments, given the routine's result type.
pub type Implementation = fn(&mut Ctx<'_>, &[Value], ColumnType) -> Result<Value>;

/// Control holds what an extension's control file declares.
pub struct Control {
    pub default_version: &'static str,
    pub comment: &'static str,
    pub superuser: bool,
    pub trusted: bool,
    pub relocatable: bool,
}

/// BaseType is a base type that an extension provides, whose values hold the bytes that Doltgres stores for them and
/// whose support routines are named after it, as `<name>_in`, `<name>_out`, `<name>_recv`, `<name>_send`,
/// `<name>_typmod_in`, and `<name>_cmp`.
pub struct BaseType {
    pub name: &'static str,
    /// input reads a value from its text format, under a type modifier.
    pub input: fn(&str, i32) -> Result<Vec<u8>>,
    /// output writes a value's text format.
    pub output: fn(&[u8]) -> String,
    /// receive reads a value from its binary format, under a type modifier.
    pub receive: fn(&[u8], i32) -> Result<Vec<u8>>,
    /// send writes a value's binary format.
    pub send: fn(&[u8]) -> Vec<u8>,
    /// typmod_in computes a type modifier from the modifiers written after the type's name.
    pub typmod_in: fn(&[String]) -> Result<i32>,
    /// typmod checks that a value fits a type modifier.
    pub typmod: fn(&[u8], i32) -> Result<()>,
    /// compare orders two values.
    pub compare: fn(&[u8], &[u8]) -> Ordering,
    /// vector returns a value's elements for a vector index, for a type that Dolt's vector indexes can hold.
    pub vector: Option<Elements>,
}

impl std::fmt::Debug for BaseType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.name)
    }
}

impl PartialEq for BaseType {
    fn eq(&self, other: &BaseType) -> bool {
        std::ptr::eq(self, other)
    }
}

/// Routine is a function that an extension provides, where `symbol` names its implementation uniquely within the
/// extension.
pub struct Routine {
    pub name: String,
    pub symbol: String,
    /// The parameters' names and type names.
    pub params: Vec<(&'static str, String)>,
    pub returns: String,
    pub strict: bool,
    pub implementation: Implementation,
}

/// Operator is an operator that an extension provides, whose routine is named by its symbol.
pub struct Operator {
    pub name: &'static str,
    pub left: String,
    pub right: String,
    pub routine: String,
    pub commutator: &'static str,
    pub negator: &'static str,
}

/// Cast is a cast that an extension provides, with Go's numbering of the cast's context in `crate::casts`.
pub struct Cast {
    pub source: String,
    pub target: String,
    pub routine: String,
    pub context: u8,
}

/// Aggregate is an aggregate that an extension provides, whose support routines are named by their symbols.
pub struct Aggregate {
    pub name: &'static str,
    pub params: Vec<String>,
    pub returns: String,
    pub state_type: String,
    pub transition: String,
    pub final_routine: String,
    pub combine: String,
    pub init_cond: Option<&'static str>,
}

/// OperatorClass is an operator class that an extension provides for its index access methods.
pub struct OperatorClass {
    pub name: String,
    pub access_methods: Vec<&'static str>,
    /// The access methods for which this is the type's default operator class.
    pub default_for: Vec<&'static str>,
    pub type_name: &'static str,
    /// The distance of the index that the class builds, or None when Doltgres cannot build its indexes yet.
    pub distance: Option<prolly::Distance>,
    /// The most dimensions an indexed column may have.
    pub max_dimensions: i32,
}

/// AccessMethod is an index access method that an extension provides.
pub struct AccessMethod {
    pub name: &'static str,
    pub handler: &'static str,
    /// The integer storage parameters the method takes, with their least, greatest, and default values.
    pub params: Vec<(&'static str, i64, i64, i64)>,
}

/// Extension is an extension that Doltgres emulates.
pub struct Extension {
    pub name: &'static str,
    pub control: Control,
    pub types: Vec<BaseType>,
    pub routines: Vec<Routine>,
    pub operators: Vec<Operator>,
    pub casts: Vec<Cast>,
    pub aggregates: Vec<Aggregate>,
    pub operator_classes: Vec<OperatorClass>,
    pub access_methods: Vec<AccessMethod>,
}

/// all returns every extension that Doltgres emulates.
pub fn all() -> &'static [Extension] {
    static EXTENSIONS: OnceLock<Vec<Extension>> = OnceLock::new();
    EXTENSIONS.get_or_init(|| vec![uuid_ossp::extension(), vector::extension()])
}

/// get returns the emulated extension of the name.
pub fn get(name: &str) -> Option<&'static Extension> {
    all().iter().find(|e| e.name == name)
}

/// base_type returns the base type whose send routine has the name.
pub fn base_type(send_routine: &str) -> Option<&'static BaseType> {
    all().iter().flat_map(|e| &e.types).find(|t| send_routine.strip_prefix(t.name) == Some("_send"))
}

/// implementation returns the implementation of an extension's routine.
pub fn implementation(extension: &str, symbol: &str) -> Option<Implementation> {
    get(extension)?.routines.iter().find(|r| r.symbol == symbol).map(|r| r.implementation)
}
