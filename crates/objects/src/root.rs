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

use store::{Error, Result};

use crate::codec::Reader;
use crate::show::{Fields, list, string_map, strings};
use crate::types::SerializedType;

/// corrupt returns a Corrupt error.
fn corrupt(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}

/// unsupported returns the error for a newer serialization version.
fn unsupported(version: u64, what: &str) -> Error {
    corrupt(format!("version {version} of {what} is not supported, please upgrade the server"))
}

/// finish fails when data remains after an object.
fn finish(r: &Reader<'_>, what: &str) -> Result<()> {
    if r.is_empty() { Ok(()) } else { Err(corrupt(format!("extra data found while deserializing {what}"))) }
}

/// Operation is a compiled PL/pgSQL interpreter operation.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Operation {
    pub op_code: u16,
    pub primary_data: Vec<u8>,
    pub secondary_data: Vec<Vec<u8>>,
    pub target: Vec<u8>,
    pub index: i32,
    pub options: BTreeMap<Vec<u8>, Vec<u8>>,
}

impl Operation {
    fn read_all(r: &mut Reader<'_>) -> Result<Vec<Operation>> {
        let count = r.variable_uint()?;
        (0..count)
            .map(|_| {
                Ok(Operation {
                    op_code: r.uint16()?,
                    primary_data: r.string()?,
                    secondary_data: r.string_slice()?,
                    target: r.string()?,
                    index: r.int32()?,
                    options: r.string_map()?,
                })
            })
            .collect()
    }

    fn show(&self) -> String {
        Fields::new()
            .field("OpCode", self.op_code)
            .string("PrimaryData", &self.primary_data)
            .field("SecondaryData", strings(&self.secondary_data))
            .string("Target", &self.target)
            .field("Index", self.index)
            .field("Options", string_map(&self.options))
            .finish()
    }
}

/// Parameter is a routine parameter.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Parameter {
    pub mode: u8,
    pub name: Vec<u8>,
    pub type_id: Vec<u8>,
    pub default: Vec<u8>,
}

impl Parameter {
    fn show(&self) -> String {
        Fields::new()
            .field("Mode", self.mode)
            .string("Name", &self.name)
            .string("Type", &self.type_id)
            .string("Default", &self.default)
            .finish()
    }
}

/// Sequence is a sequence.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Sequence {
    pub id: Vec<u8>,
    pub data_type_id: Vec<u8>,
    pub persistence: u8,
    pub start: i64,
    pub current: i64,
    pub increment: i64,
    pub minimum: i64,
    pub maximum: i64,
    pub cache: i64,
    pub cycle: bool,
    pub is_at_end: bool,
    pub has_been_called: bool,
    pub owner_table: Vec<u8>,
    pub owner_column: Vec<u8>,
}

impl Sequence {
    pub fn deserialize(data: &[u8]) -> Result<Sequence> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version > 1 {
            return Err(unsupported(version, "sequences"));
        }
        let mut s = Sequence {
            id: r.string()?,
            data_type_id: r.string()?,
            persistence: r.uint8()?,
            start: r.int64()?,
            current: r.int64()?,
            increment: r.int64()?,
            minimum: r.int64()?,
            maximum: r.int64()?,
            cache: r.int64()?,
            cycle: r.bool()?,
            is_at_end: r.bool()?,
            ..Sequence::default()
        };
        if version >= 1 {
            s.has_been_called = r.bool()?;
        }
        s.owner_table = r.string()?;
        s.owner_column = r.string()?;
        finish(&r, "a sequence")?;
        Ok(s)
    }

    fn show(&self) -> String {
        let state = Fields::new()
            .string("Id", &self.id)
            .field("Start", self.start)
            .field("Current", self.current)
            .field("Increment", self.increment)
            .field("Minimum", self.minimum)
            .field("Maximum", self.maximum)
            .field("Cache", self.cache)
            .field("Cycle", self.cycle)
            .field("IsAtEnd", self.is_at_end)
            .field("HasBeenCalled", self.has_been_called)
            .finish();
        Fields::new()
            .string("DataTypeID", &self.data_type_id)
            .field("Persistence", self.persistence)
            .field("SequenceState", state)
            .string("OwnerTable", &self.owner_table)
            .string("OwnerColumn", &self.owner_column)
            .finish()
    }
}

/// Function is a user-defined or extension function.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Function {
    pub id: Vec<u8>,
    pub return_type: Vec<u8>,
    pub all_params: Vec<Parameter>,
    pub variadic: bool,
    pub is_non_deterministic: bool,
    pub strict: bool,
    pub definition: Vec<u8>,
    pub extension_name: Vec<u8>,
    pub extension_symbol: Vec<u8>,
    pub operations: Vec<Operation>,
    pub sql_definition: Vec<u8>,
    pub set_of: bool,
}

impl Function {
    pub fn deserialize(data: &[u8]) -> Result<Function> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version > 4 {
            return Err(unsupported(version, "functions"));
        }
        let mut f = Function { id: r.string()?, return_type: r.string()?, ..Function::default() };
        let names = r.string_slice()?;
        let types = r.string_slice()?;
        f.variadic = r.bool()?;
        f.is_non_deterministic = r.bool()?;
        f.strict = r.bool()?;
        f.definition = r.string()?;
        f.operations = Operation::read_all(&mut r)?;
        if version >= 1 {
            f.extension_name = r.string()?;
            f.extension_symbol = r.string()?;
        }
        if version >= 2 {
            f.sql_definition = r.string()?;
            f.set_of = r.bool()?;
        }
        let defaults = if version >= 3 { Some(r.string_slice()?) } else { None };
        let modes = if version >= 4 {
            let count = r.variable_uint()?;
            Some((0..count).map(|_| r.uint8()).collect::<Result<Vec<_>>>()?)
        } else {
            None
        };
        f.all_params = parameters(names, types, defaults, modes)?;
        finish(&r, "a function")?;
        Ok(f)
    }

    fn show(&self) -> String {
        Fields::new()
            .string("ID", &self.id)
            .string("ReturnType", &self.return_type)
            .field("AllParams", list(self.all_params.iter().map(Parameter::show)))
            .field("Variadic", self.variadic)
            .field("IsNonDeterministic", self.is_non_deterministic)
            .field("Strict", self.strict)
            .string("Definition", &self.definition)
            .string("ExtensionName", &self.extension_name)
            .string("ExtensionSymbol", &self.extension_symbol)
            .field("Operations", list(self.operations.iter().map(Operation::show)))
            .string("SQLDefinition", &self.sql_definition)
            .field("SetOf", self.set_of)
            .finish()
    }
}

/// parameters zips the parallel parameter lists of a routine, which index the names, as Go's deserializers do.
fn parameters(
    names: Vec<Vec<u8>>,
    types: Vec<Vec<u8>>,
    defaults: Option<Vec<Vec<u8>>>,
    modes: Option<Vec<u8>>,
) -> Result<Vec<Parameter>> {
    let short = || corrupt("routine parameter lists have different lengths");
    names
        .into_iter()
        .enumerate()
        .map(|(i, name)| {
            Ok(Parameter {
                name,
                type_id: types.get(i).cloned().ok_or_else(short)?,
                default: match &defaults {
                    Some(defaults) => defaults.get(i).cloned().ok_or_else(short)?,
                    None => Vec::new(),
                },
                mode: match &modes {
                    Some(modes) => *modes.get(i).ok_or_else(short)?,
                    None => 0,
                },
            })
        })
        .collect()
}

/// TriggerEvent is an event that fires a trigger.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct TriggerEvent {
    pub event_type: u8,
    pub column_names: Vec<Vec<u8>>,
}

/// Trigger is a trigger.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Trigger {
    pub id: Vec<u8>,
    pub function: Vec<u8>,
    pub timing: u8,
    pub events: Vec<TriggerEvent>,
    pub for_each_row: bool,
    pub when: Vec<Operation>,
    pub deferrable: u8,
    pub referenced_table_name: Vec<u8>,
    pub constraint: bool,
    pub old_transition_name: Vec<u8>,
    pub new_transition_name: Vec<u8>,
    pub arguments: Vec<Vec<u8>>,
    pub definition: Vec<u8>,
}

impl Trigger {
    pub fn deserialize(data: &[u8]) -> Result<Trigger> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version != 0 {
            return Err(unsupported(version, "triggers"));
        }
        let mut t = Trigger {
            id: r.string()?,
            function: r.string()?,
            timing: r.uint8()?,
            for_each_row: r.bool()?,
            deferrable: r.uint8()?,
            referenced_table_name: r.string()?,
            constraint: r.bool()?,
            old_transition_name: r.string()?,
            new_transition_name: r.string()?,
            arguments: r.string_slice()?,
            definition: r.string()?,
            ..Trigger::default()
        };
        t.when = Operation::read_all(&mut r)?;
        let count = r.variable_uint()?;
        for _ in 0..count {
            t.events.push(TriggerEvent { event_type: r.uint8()?, column_names: r.string_slice()? });
        }
        finish(&r, "a trigger")?;
        Ok(t)
    }

    fn show(&self) -> String {
        let events = self
            .events
            .iter()
            .map(|e| Fields::new().field("Type", e.event_type).field("ColumnNames", strings(&e.column_names)).finish());
        Fields::new()
            .string("ID", &self.id)
            .string("Function", &self.function)
            .field("Timing", self.timing)
            .field("Events", list(events))
            .field("ForEachRow", self.for_each_row)
            .field("When", list(self.when.iter().map(Operation::show)))
            .field("Deferrable", self.deferrable)
            .string("ReferencedTableName", &self.referenced_table_name)
            .field("Constraint", self.constraint)
            .string("OldTransitionName", &self.old_transition_name)
            .string("NewTransitionName", &self.new_transition_name)
            .field("Arguments", strings(&self.arguments))
            .string("Definition", &self.definition)
            .finish()
    }
}

/// Procedure is a user-defined or extension procedure.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Procedure {
    pub id: Vec<u8>,
    pub all_params: Vec<Parameter>,
    pub definition: Vec<u8>,
    pub extension_name: Vec<u8>,
    pub extension_symbol: Vec<u8>,
    pub operations: Vec<Operation>,
    pub sql_definition: Vec<u8>,
}

impl Procedure {
    pub fn deserialize(data: &[u8]) -> Result<Procedure> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version > 1 {
            return Err(unsupported(version, "procedures"));
        }
        let mut p = Procedure { id: r.string()?, ..Procedure::default() };
        let names = r.string_slice()?;
        let types = r.string_slice()?;
        p.definition = r.string()?;
        p.extension_name = r.string()?;
        p.extension_symbol = r.string()?;
        p.sql_definition = r.string()?;
        let count = r.variable_uint()?;
        let modes = (0..count).map(|_| r.uint8()).collect::<Result<Vec<_>>>()?;
        p.operations = Operation::read_all(&mut r)?;
        let defaults = if version >= 1 { Some(r.string_slice()?) } else { None };
        p.all_params = parameters(names, types, defaults, Some(modes))?;
        finish(&r, "a procedure")?;
        Ok(p)
    }

    fn show(&self) -> String {
        Fields::new()
            .string("ID", &self.id)
            .field("AllParams", list(self.all_params.iter().map(Parameter::show)))
            .string("Definition", &self.definition)
            .string("ExtensionName", &self.extension_name)
            .string("ExtensionSymbol", &self.extension_symbol)
            .field("Operations", list(self.operations.iter().map(Operation::show)))
            .string("SQLDefinition", &self.sql_definition)
            .finish()
    }
}

/// Extension is an installed extension.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Extension {
    pub ext_name: Vec<u8>,
    pub namespace: Vec<u8>,
    pub relocatable: bool,
    pub version: Vec<u8>,
}

impl Extension {
    pub fn deserialize(data: &[u8]) -> Result<Extension> {
        let mut r = Reader::new(data);
        match r.variable_uint()? {
            0 => {
                return Err(corrupt(
                    "extensions have been completely revamped, please reimport your database using a newer version",
                ));
            }
            1 => {}
            version => {
                return Err(corrupt(format!(
                    "version {version} of extensions are not supported, please upgrade the server"
                )));
            }
        }
        let e =
            Extension { ext_name: r.string()?, namespace: r.string()?, relocatable: r.bool()?, version: r.string()? };
        finish(&r, "an extension")?;
        Ok(e)
    }

    fn show(&self) -> String {
        Fields::new()
            .string("ExtName", &self.ext_name)
            .string("Namespace", &self.namespace)
            .field("Relocatable", self.relocatable)
            .string("Version", &self.version)
            .finish()
    }
}

/// Cast is a user-defined cast.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Cast {
    pub id: Vec<u8>,
    pub cast_type: u8,
    pub function: Vec<u8>,
    pub use_in_out: bool,
}

impl Cast {
    pub fn deserialize(data: &[u8]) -> Result<Cast> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version != 0 {
            return Err(unsupported(version, "casts"));
        }
        let c = Cast { id: r.string()?, cast_type: r.uint8()?, function: r.string()?, use_in_out: r.bool()? };
        finish(&r, "a cast")?;
        Ok(c)
    }

    fn show(&self) -> String {
        Fields::new()
            .string("ID", &self.id)
            .field("CastType", self.cast_type)
            .string("Function", &self.function)
            .field("BuiltIn", "nil")
            .field("UseInOut", self.use_in_out)
            .field("request", 0)
            .finish()
    }
}

/// Operator is a user-defined operator.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Operator {
    pub id: Vec<u8>,
    pub function: Vec<u8>,
    pub return_type: Vec<u8>,
    pub commutator: Vec<u8>,
    pub negator: Vec<u8>,
    pub hashes: bool,
    pub merges: bool,
}

impl Operator {
    pub fn deserialize(data: &[u8]) -> Result<Operator> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version != 0 {
            return Err(unsupported(version, "operators"));
        }
        let o = Operator {
            id: r.string()?,
            function: r.string()?,
            return_type: r.string()?,
            commutator: r.string()?,
            negator: r.string()?,
            hashes: r.bool()?,
            merges: r.bool()?,
        };
        finish(&r, "an operator")?;
        Ok(o)
    }

    fn show(&self) -> String {
        Fields::new()
            .string("ID", &self.id)
            .string("Function", &self.function)
            .string("ReturnType", &self.return_type)
            .string("Commutator", &self.commutator)
            .string("Negator", &self.negator)
            .field("Hashes", self.hashes)
            .field("Merges", self.merges)
            .finish()
    }
}

/// Aggregate is a user-defined aggregate.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Aggregate {
    pub id: Vec<u8>,
    pub return_type: Vec<u8>,
    pub s_func: Vec<u8>,
    pub s_type: Vec<u8>,
    pub final_func: Vec<u8>,
    pub combine_func: Vec<u8>,
    pub init_cond: Vec<u8>,
    pub has_init_cond: bool,
}

impl Aggregate {
    pub fn deserialize(data: &[u8]) -> Result<Aggregate> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version != 0 {
            return Err(unsupported(version, "aggregates"));
        }
        let a = Aggregate {
            id: r.string()?,
            return_type: r.string()?,
            s_func: r.string()?,
            s_type: r.string()?,
            final_func: r.string()?,
            combine_func: r.string()?,
            init_cond: r.string()?,
            has_init_cond: r.bool()?,
        };
        finish(&r, "an aggregate")?;
        Ok(a)
    }

    fn show(&self) -> String {
        Fields::new()
            .string("ID", &self.id)
            .string("ReturnType", &self.return_type)
            .string("SFunc", &self.s_func)
            .string("SType", &self.s_type)
            .string("FinalFunc", &self.final_func)
            .string("CombineFunc", &self.combine_func)
            .string("InitCond", &self.init_cond)
            .field("HasInitCond", self.has_init_cond)
            .finish()
    }
}

/// Conflict is a merge conflict between versions of a root object.
#[derive(Clone, Debug, PartialEq, Default)]
pub struct Conflict {
    pub id: Vec<u8>,
    pub from_hash: Vec<u8>,
    pub root_object_id: i64,
    pub ours: Option<Box<RootObject>>,
    pub theirs: Option<Box<RootObject>>,
    pub ancestor: Option<Box<RootObject>>,
}

impl Conflict {
    pub fn deserialize(data: &[u8]) -> Result<Conflict> {
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version > 0 {
            return Err(unsupported(version, "conflicts"));
        }
        let mut c =
            Conflict { id: r.string()?, from_hash: r.string()?, root_object_id: r.int64()?, ..Conflict::default() };
        let (has_ours, has_theirs, has_ancestor) = (r.bool()?, r.bool()?, r.bool()?);
        let (ours, theirs, ancestor) = (r.bytes()?, r.bytes()?, r.bytes()?);
        let kind = Kind::from_id(c.root_object_id)?;
        let load = |present: bool, data: &[u8]| -> Result<Option<Box<RootObject>>> {
            if present { RootObject::deserialize(kind, data).map(|o| Some(Box::new(o))) } else { Ok(None) }
        };
        c.ours = load(has_ours, &ours)?;
        c.theirs = load(has_theirs, &theirs)?;
        c.ancestor = load(has_ancestor, &ancestor)?;
        finish(&r, "a conflict")?;
        Ok(c)
    }

    fn show(&self) -> String {
        let object = |o: &Option<Box<RootObject>>| match o {
            Some(o) => o.show_interface(),
            None => "nil".to_string(),
        };
        Fields::new()
            .string("ID", &self.id)
            .string("FromHash", &self.from_hash)
            .field("RootObjectID", self.root_object_id)
            .field("Ours", object(&self.ours))
            .field("Theirs", object(&self.theirs))
            .field("Ancestor", object(&self.ancestor))
            .finish()
    }
}

/// Kind is a root object collection.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    Sequences,
    Types,
    Functions,
    Triggers,
    Extensions,
    Conflicts,
    Procedures,
    Casts,
    Operators,
    Aggregates,
}

impl Kind {
    /// from_id returns the collection with Go's RootObjectID.
    pub fn from_id(id: i64) -> Result<Kind> {
        const KINDS: [Kind; 10] = [
            Kind::Sequences,
            Kind::Types,
            Kind::Functions,
            Kind::Triggers,
            Kind::Extensions,
            Kind::Conflicts,
            Kind::Procedures,
            Kind::Casts,
            Kind::Operators,
            Kind::Aggregates,
        ];
        match id {
            1..=10 => Ok(KINDS[id as usize - 1]),
            0 => Err(corrupt(format!("invalid root object ID: {id}"))),
            _ => Err(corrupt("unsupported object found, please upgrade the server")),
        }
    }

    /// from_field returns the collection stored in the root value field with the name.
    pub fn from_field(name: &str) -> Option<Kind> {
        Some(match name {
            "sequences" => Kind::Sequences,
            "types" => Kind::Types,
            "functions" => Kind::Functions,
            "triggers" => Kind::Triggers,
            "extensions" => Kind::Extensions,
            "conflicts" => Kind::Conflicts,
            "procedures" => Kind::Procedures,
            "casts" => Kind::Casts,
            "operators" => Kind::Operators,
            "aggregates" => Kind::Aggregates,
            _ => return None,
        })
    }
}

/// RootObject is a deserialized root object.
#[derive(Clone, Debug, PartialEq)]
pub enum RootObject {
    Sequence(Sequence),
    Type(Box<SerializedType>),
    Function(Function),
    Trigger(Trigger),
    Extension(Extension),
    Conflict(Conflict),
    Procedure(Procedure),
    Cast(Cast),
    Operator(Operator),
    Aggregate(Aggregate),
}

impl RootObject {
    /// deserialize decodes a root object of the collection.
    pub fn deserialize(kind: Kind, data: &[u8]) -> Result<RootObject> {
        Ok(match kind {
            Kind::Sequences => RootObject::Sequence(Sequence::deserialize(data)?),
            Kind::Types => RootObject::Type(Box::new(SerializedType::deserialize(data)?)),
            Kind::Functions => RootObject::Function(Function::deserialize(data)?),
            Kind::Triggers => RootObject::Trigger(Trigger::deserialize(data)?),
            Kind::Extensions => RootObject::Extension(Extension::deserialize(data)?),
            Kind::Conflicts => RootObject::Conflict(Conflict::deserialize(data)?),
            Kind::Procedures => RootObject::Procedure(Procedure::deserialize(data)?),
            Kind::Casts => RootObject::Cast(Cast::deserialize(data)?),
            Kind::Operators => RootObject::Operator(Operator::deserialize(data)?),
            Kind::Aggregates => RootObject::Aggregate(Aggregate::deserialize(data)?),
        })
    }

    /// show renders the object as the Go graph oracle renders the value its deserializer returns.
    pub fn show(&self) -> String {
        match self {
            RootObject::Sequence(o) => o.show(),
            RootObject::Type(o) => o.show(),
            RootObject::Function(o) => o.show(),
            RootObject::Trigger(o) => o.show(),
            RootObject::Extension(o) => o.show(),
            RootObject::Conflict(o) => o.show(),
            RootObject::Procedure(o) => o.show(),
            RootObject::Cast(o) => o.show(),
            RootObject::Operator(o) => o.show(),
            RootObject::Aggregate(o) => o.show(),
        }
    }

    /// show_interface renders the object as a Go interface value holding what a collection's DeserializeRootObject
    /// returns, prefixed by its Go type.
    fn show_interface(&self) -> String {
        match self {
            RootObject::Sequence(o) => format!("*sequences.Sequence{}", o.show()),
            RootObject::Type(o) => format!("typecollection.TypeWrapper{{Type:{}}}", o.show()),
            RootObject::Function(o) => format!("functions.Function{}", o.show()),
            RootObject::Trigger(o) => format!("triggers.Trigger{}", o.show()),
            RootObject::Extension(o) => format!("extensions.Extension{}", o.show()),
            RootObject::Conflict(o) => format!("conflicts.Conflict{}", o.show()),
            RootObject::Procedure(o) => format!("procedures.Procedure{}", o.show()),
            RootObject::Cast(o) => format!("casts.Cast{}", o.show()),
            RootObject::Operator(o) => format!("operators.Operator{}", o.show()),
            RootObject::Aggregate(o) => format!("aggregates.Aggregate{}", o.show()),
        }
    }
}
