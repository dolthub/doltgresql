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

//! Merges of root objects, such as functions and sequences, as Doltgres merges them object by object, and the
//! conflicts they leave: the fields that both sides changed, kept as conflict root objects and shown and resolved
//! through `dolt_conflicts_<object>` tables.

use std::collections::BTreeMap;

use doltdb::database::Database;
use doltdb::root::Root;
use objects::{Conflict, Function, Kind, Parameter, Procedure, RootObject, Sequence};
use store::Hash;

use crate::catalog::id;
use crate::dolt::args::error;
use crate::error::{PgError, Result};
use crate::oid::TEXT;
use crate::query::Ctx;
use crate::types::Value;

/// CONFLICTS is the root value's collection of conflict root objects.
pub const CONFLICTS: usize = 5;

/// ROOT_OBJECT is the field of a conflict where one side deleted the whole object.
const ROOT_OBJECT: &str = "root_object";

/// Field is a field value of a root object as conflict tables show it.
#[derive(Clone, Debug, PartialEq)]
enum Field {
    Text(String),
    Bool(bool),
}

impl Field {
    /// output returns the value as Doltgres' types write it.
    fn output(&self) -> String {
        match self {
            Field::Text(text) => text.clone(),
            Field::Bool(b) => if *b { "true" } else { "false" }.to_string(),
        }
    }

    /// text returns the value of a text field.
    fn text(&self) -> String {
        match self {
            Field::Text(text) => text.clone(),
            Field::Bool(b) => b.to_string(),
        }
    }

    /// bool returns the value of a boolean field.
    fn bool(&self) -> bool {
        matches!(self, Field::Bool(true))
    }
}

/// Change is how one side of a merge changed a field.
#[derive(Clone, Copy, Debug, PartialEq)]
enum Change {
    Added,
    Deleted,
    Modified,
    Unchanged,
}

/// Diff is a field of a root object as both sides of a merge left it: Doltgres' RootObjectDiff.
#[derive(Clone, Debug, PartialEq)]
struct Diff {
    from_hash: String,
    field: String,
    ancestor: Option<Field>,
    ours: Option<Field>,
    theirs: Option<Field>,
    our_change: Change,
    their_change: Change,
}

impl Diff {
    /// row returns the diff as a row of its conflict table, as Doltgres' ToRow writes it, which shows a side that
    /// left the field alone as our change when it was their side.
    fn row(&self) -> Vec<Value> {
        let text = |f: &Option<Field>| f.as_ref().map_or(Value::Null, |f| Value::Text(f.output()));
        let change = |c: Change| match c {
            Change::Added => "added",
            Change::Deleted => "deleted",
            Change::Modified => "modified",
            Change::Unchanged => "no_change",
        };
        let mut our_change = Value::Text(change(self.our_change).to_string());
        let mut their_change = Value::Null;
        match self.their_change {
            Change::Unchanged => our_change = Value::Text("no_change".into()),
            other => their_change = Value::Text(change(other).to_string()),
        }
        vec![
            Value::Text(self.from_hash.clone()),
            text(&self.ancestor),
            text(&self.ours),
            our_change,
            text(&self.theirs),
            their_change,
            Value::Text(self.field.clone()),
        ]
    }
}

/// diff_values compares a field's values on both sides of a merge and its ancestor, and reports whether they
/// conflict, leaving the merged value as our value when they do not, as Doltgres' DiffValues does.
fn diff_values(from_hash: &str, field: &str, ours: Field, theirs: Field, ancestor: Option<Field>) -> (bool, Diff) {
    let mut diff = Diff {
        from_hash: from_hash.to_string(),
        field: field.to_string(),
        ancestor: None,
        ours: Some(ours.clone()),
        theirs: Some(theirs.clone()),
        our_change: Change::Unchanged,
        their_change: Change::Unchanged,
    };
    if ours == theirs {
        return (false, diff);
    }
    let Some(ancestor) = ancestor else {
        diff.our_change = Change::Added;
        diff.their_change = Change::Added;
        return (true, diff);
    };
    diff.ancestor = Some(ancestor.clone());
    if ours == ancestor {
        diff.ours = Some(theirs);
        diff.their_change = Change::Modified;
        return (false, diff);
    }
    if theirs == ancestor {
        diff.theirs = Some(ours);
        diff.our_change = Change::Modified;
        return (false, diff);
    }
    diff.our_change = Change::Modified;
    diff.their_change = Change::Modified;
    (true, diff)
}

/// FieldDiffs collects the conflicting fields of a merge of two objects, applying each merged field that does not
/// conflict to ours.
struct FieldDiffs<'a> {
    from_hash: &'a str,
    has_ancestor: bool,
    diffs: Vec<Diff>,
}

impl FieldDiffs<'_> {
    /// field compares one field, returning the merged value when it does not conflict.
    fn field(&mut self, name: &str, ours: Field, theirs: Field, ancestor: Field) -> Option<Field> {
        let ancestor = self.has_ancestor.then_some(ancestor);
        let (conflict, diff) = diff_values(self.from_hash, name, ours, theirs, ancestor);
        if conflict {
            self.diffs.push(diff);
            return None;
        }
        diff.ours
    }
}

/// text returns stored bytes as a text field.
fn text(bytes: &[u8]) -> Field {
    Field::Text(String::from_utf8_lossy(bytes).into_owned())
}

/// parameter_names_and_modes writes a routine's parameter names and modes as Doltgres' ParameterNamesAndModesToString
/// does.
fn parameter_names_and_modes(params: &[Parameter]) -> String {
    params
        .iter()
        .map(|p| {
            let mode = match p.mode {
                1 => "out",
                2 => "inout",
                3 => "variadic",
                _ => "in",
            };
            format!("{} {mode}", String::from_utf8_lossy(&p.name))
        })
        .collect::<Vec<_>>()
        .join(",")
}

/// with_names_and_modes returns parameters with the names and modes that a ParameterNamesAndModesToString string
/// holds, as Doltgres' ParameterNamesAndModesFromString does.
fn with_names_and_modes(params: &[Parameter], text: &str) -> Result<Vec<Parameter>> {
    let mut params = params.to_vec();
    if text.is_empty() {
        return Ok(params);
    }
    let values: Vec<&str> = text.split(',').collect();
    if values.len() != params.len() {
        return Err(error("unhandled procedure parameter count"));
    }
    for (param, value) in params.iter_mut().zip(values) {
        let parts: Vec<&str> = value.split(' ').collect();
        let [name, mode] = parts.as_slice() else { return Err(error("invalid namesAndModes")) };
        param.name = name.as_bytes().to_vec();
        param.mode = match *mode {
            "out" => 1,
            "inout" => 2,
            "variadic" => 3,
            _ => 0,
        };
    }
    Ok(params)
}

/// inner_definition returns the body of a routine's definition, between its dollar quotes or its single quotes, as
/// Doltgres' GetInnerDefinition finds it.
fn inner_definition(definition: &[u8]) -> String {
    let definition = String::from_utf8_lossy(definition);
    if let (Some(start), Some(end)) = (definition.find("$$"), definition.rfind("$$")) {
        return definition.get(start + 2..end).unwrap_or_default().trim().to_string();
    }
    if let (Some(start), Some(end)) = (definition.to_lowercase().find("as '"), definition.rfind('\'')) {
        return definition.get(start + 4..end).unwrap_or_default().trim().to_string();
    }
    definition.into_owned()
}

/// replace_definition returns a routine's definition with a new body, as Doltgres' ReplaceDefinition writes it.
fn replace_definition(definition: &[u8], inner: &str) -> Vec<u8> {
    let current = inner_definition(definition);
    String::from_utf8_lossy(definition).replacen(&current, inner, 1).into_bytes()
}

/// type_name returns the name of a type ID.
fn type_name(type_id: &[u8]) -> String {
    id::segments(type_id).get(1).cloned().unwrap_or_default()
}

/// with_type_name returns a type ID with a new name in the same schema.
fn with_type_name(type_id: &[u8], name: &str) -> Vec<u8> {
    let schema = id::segments(type_id).first().cloned().unwrap_or_default();
    id::new(id::SECTION_TYPE, &[&schema, name])
}

/// function_diffs returns the conflicting fields of two versions of a function and ours with every other field
/// merged, as Doltgres' function DiffRootObjects does, which updates only the definition's text and not its compiled
/// operations.
fn function_diffs(
    from_hash: &str,
    ours: &Function,
    theirs: &Function,
    ancestor: Option<&Function>,
) -> (Vec<Diff>, Function) {
    let default = Function::default();
    let anc = ancestor.unwrap_or(&default);
    let mut d = FieldDiffs { from_hash, has_ancestor: ancestor.is_some(), diffs: Vec::new() };
    let mut merged = ours.clone();
    let params = |f: &Function| Field::Text(parameter_names_and_modes(&f.all_params));
    if let Some(value) = d.field("parameters", params(ours), params(theirs), params(anc)) {
        merged.all_params = with_names_and_modes(&merged.all_params, &value.text()).unwrap_or(merged.all_params);
    }
    if ours.return_type != theirs.return_type {
        let name = |f: &Function| Field::Text(type_name(&f.return_type));
        if let Some(value) = d.field("return_type", name(ours), name(theirs), name(anc)) {
            merged.return_type = with_type_name(&merged.return_type, &value.text());
        }
    }
    if ours.is_non_deterministic != theirs.is_non_deterministic {
        let flag = |f: &Function| Field::Bool(f.is_non_deterministic);
        if let Some(value) = d.field("non_deterministic", flag(ours), flag(theirs), flag(anc)) {
            merged.is_non_deterministic = value.bool();
        }
    }
    if ours.strict != theirs.strict {
        let flag = |f: &Function| Field::Bool(f.strict);
        if let Some(value) = d.field("strict", flag(ours), flag(theirs), flag(anc)) {
            merged.strict = value.bool();
        }
    }
    if ours.definition != theirs.definition {
        let inner = |f: &Function| Field::Text(inner_definition(&f.definition));
        if let Some(value) = d.field("definition", inner(ours), inner(theirs), inner(anc)) {
            merged.definition = replace_definition(&merged.definition, &value.text());
        }
    }
    for (field, get) in [
        ("extension_name", (|f: &Function| &f.extension_name) as fn(&Function) -> &Vec<u8>),
        ("extension_symbol", |f: &Function| &f.extension_symbol),
        ("sql_definition", |f: &Function| &f.sql_definition),
    ] {
        if get(ours) != get(theirs)
            && let Some(value) = d.field(field, text(get(ours)), text(get(theirs)), text(get(anc)))
        {
            let bytes = value.text().into_bytes();
            match field {
                "extension_name" => merged.extension_name = bytes,
                "extension_symbol" => merged.extension_symbol = bytes,
                _ => merged.sql_definition = bytes,
            }
        }
    }
    if ours.set_of != theirs.set_of
        && let Some(value) =
            d.field("set_of", Field::Bool(ours.set_of), Field::Bool(theirs.set_of), Field::Bool(anc.set_of))
    {
        merged.set_of = value.bool();
    }
    (d.diffs, merged)
}

/// procedure_diffs returns the conflicting fields of two versions of a procedure and ours with every other field
/// merged, as Doltgres' procedure DiffRootObjects does.
fn procedure_diffs(
    from_hash: &str,
    ours: &Procedure,
    theirs: &Procedure,
    ancestor: Option<&Procedure>,
) -> (Vec<Diff>, Procedure) {
    let default = Procedure::default();
    let anc = ancestor.unwrap_or(&default);
    let mut d = FieldDiffs { from_hash, has_ancestor: ancestor.is_some(), diffs: Vec::new() };
    let mut merged = ours.clone();
    let params = |p: &Procedure| Field::Text(parameter_names_and_modes(&p.all_params));
    if let Some(value) = d.field("parameters", params(ours), params(theirs), params(anc)) {
        merged.all_params = with_names_and_modes(&merged.all_params, &value.text()).unwrap_or(merged.all_params);
    }
    if ours.definition != theirs.definition {
        let inner = |p: &Procedure| Field::Text(inner_definition(&p.definition));
        if let Some(value) = d.field("definition", inner(ours), inner(theirs), inner(anc)) {
            merged.definition = replace_definition(&merged.definition, &value.text());
        }
    }
    for (field, get) in [
        ("extension_name", (|p: &Procedure| &p.extension_name) as fn(&Procedure) -> &Vec<u8>),
        ("extension_symbol", |p: &Procedure| &p.extension_symbol),
        ("sql_definition", |p: &Procedure| &p.sql_definition),
    ] {
        if get(ours) != get(theirs)
            && let Some(value) = d.field(field, text(get(ours)), text(get(theirs)), text(get(anc)))
        {
            let bytes = value.text().into_bytes();
            match field {
                "extension_name" => merged.extension_name = bytes,
                "extension_symbol" => merged.extension_symbol = bytes,
                _ => merged.sql_definition = bytes,
            }
        }
    }
    (d.diffs, merged)
}

/// conflict_diffs returns the conflicting fields of a conflict and its merged object, or None when the merge
/// removes it, as Doltgres' DiffRootObjects does.
fn conflict_diffs(conflict: &Conflict) -> Result<(Vec<Diff>, Option<RootObject>)> {
    let from_hash = String::from_utf8_lossy(&conflict.from_hash).into_owned();
    let deleted = |ours: bool| Diff {
        from_hash: from_hash.clone(),
        field: ROOT_OBJECT.to_string(),
        ancestor: Some(Field::Text("ancestor".into())),
        ours: ours.then(|| Field::Text("ours".into())),
        theirs: (!ours).then(|| Field::Text("theirs".into())),
        our_change: if ours { Change::Modified } else { Change::Deleted },
        their_change: if ours { Change::Deleted } else { Change::Modified },
    };
    let ancestor = conflict.ancestor.as_deref();
    match (conflict.ours.as_deref(), conflict.theirs.as_deref()) {
        (None, None) => Ok((Vec::new(), None)),
        (None, Some(_)) => Ok((vec![deleted(false)], None)),
        (Some(_), None) => Ok((vec![deleted(true)], None)),
        (Some(RootObject::Function(o)), Some(RootObject::Function(t))) => {
            let a = match ancestor {
                Some(RootObject::Function(a)) => Some(a),
                _ => None,
            };
            let (diffs, merged) = function_diffs(&from_hash, o, t, a);
            Ok((diffs, Some(RootObject::Function(merged))))
        }
        (Some(RootObject::Procedure(o)), Some(RootObject::Procedure(t))) => {
            let a = match ancestor {
                Some(RootObject::Procedure(a)) => Some(a),
                _ => None,
            };
            let (diffs, merged) = procedure_diffs(&from_hash, o, t, a);
            Ok((diffs, Some(RootObject::Procedure(merged))))
        }
        (Some(ours), Some(_)) => Ok((Vec::new(), Some(ours.clone()))),
    }
}

/// kind returns the root object kind of a collection.
fn kind(collection: usize) -> Result<Kind> {
    Ok(Kind::from_id(collection as i64 + 1)?)
}

/// load_object reads a root object of a collection from the address its collection holds.
fn load_object(db: &mut Database, collection: usize, address: Hash) -> Result<RootObject> {
    Ok(RootObject::deserialize(kind(collection)?, &prolly::read_blob(db, &address)?)?)
}

/// store_object writes a root object into a collection of a root.
fn store_object(db: &mut Database, root: &mut Root, collection: usize, key: &[u8], object: &RootObject) -> Result<()> {
    let data = object.serialize();
    let mut sink = |_: Hash, bytes: &[u8]| {
        db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
    };
    let (address, _) =
        prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty root object"))?;
    root.put_object(db, collection, key, Some(address))?;
    Ok(())
}

/// object_key returns the ID that a root object is stored under.
fn object_key(object: &RootObject) -> Vec<u8> {
    match object {
        RootObject::Sequence(o) => o.id.clone(),
        RootObject::Type(o) => o.id.clone(),
        RootObject::Function(o) => o.id.clone(),
        RootObject::Trigger(o) => o.id.clone(),
        RootObject::Extension(o) => o.ext_name.clone(),
        RootObject::Conflict(o) => o.id.clone(),
        RootObject::Procedure(o) => o.id.clone(),
        RootObject::Cast(o) => o.id.clone(),
        RootObject::Operator(o) => o.id.clone(),
        RootObject::Aggregate(o) => o.id.clone(),
    }
}

/// put_conflict stores a conflict on a root, or the object it merges to once no fields conflict, keeping the merged
/// object in its own collection meanwhile, as Doltgres' PutRootObject does for conflicts.
pub fn put_conflict(db: &mut Database, root: &mut Root, conflict: &Conflict) -> Result<()> {
    root.put_object(db, CONFLICTS, &conflict.id, None)?;
    let collection = conflict.root_object_id as usize - 1;
    let (diffs, merged) = conflict_diffs(conflict)?;
    if !diffs.is_empty() {
        store_object(db, root, CONFLICTS, &conflict.id, &RootObject::Conflict(conflict.clone()))?;
    }
    match merged {
        Some(object) => store_object(db, root, collection, &object_key(&object), &object),
        None => Ok(root.put_object(db, collection, &conflict.id, None)?),
    }
}

/// Sides are the addresses of one root object in ours, theirs, and the merge base, where each may lack it.
type Sides = (Option<Hash>, Option<Hash>, Option<Hash>);

/// Merged is how a merge leaves one root object: as ours, replaced, removed, or in conflict.
enum Merged {
    Keep,
    Put(RootObject),
    Remove,
    Conflict(Conflict),
}

/// conflict_of builds the conflict of two versions of an object, which may merge cleanly, as Doltgres'
/// CreateConflict does.
fn conflict_of(
    from_hash: &str,
    collection: usize,
    ours: Option<RootObject>,
    theirs: Option<RootObject>,
    ancestor: Option<RootObject>,
    key: &[u8],
) -> Result<Merged> {
    let conflict = Conflict {
        id: key.to_vec(),
        from_hash: from_hash.as_bytes().to_vec(),
        root_object_id: collection as i64 + 1,
        ours: ours.clone().map(Box::new),
        theirs: theirs.map(Box::new),
        ancestor: ancestor.map(Box::new),
    };
    let (diffs, merged) = conflict_diffs(&conflict)?;
    if ours.is_none() || conflict.theirs.is_none() {
        return Ok(Merged::Conflict(conflict));
    }
    if !diffs.is_empty() {
        return Ok(Merged::Conflict(Conflict { ours: merged.map(Box::new), ..conflict }));
    }
    match merged {
        Some(merged) if Some(&merged) != ours.as_ref() => Ok(Merged::Put(merged)),
        _ => Ok(Merged::Keep),
    }
}

/// resolve picks a merged field value: the side that changed it from the ancestor, the shared value, or else what a
/// rule picks from both, as Doltgres' ResolveMergeValues does.
fn resolve<T: PartialEq + Clone>(ours: &T, theirs: &T, ancestor: Option<&T>, rule: impl Fn(&T, &T) -> T) -> T {
    if ancestor == Some(ours) {
        return theirs.clone();
    }
    if ancestor == Some(theirs) || ours == theirs {
        return ours.clone();
    }
    rule(ours, theirs)
}

/// merge_sequence merges both sides' changes to a sequence field by field, as Doltgres' sequence HandleMerge does.
fn merge_sequence(ours: &Sequence, theirs: &Sequence, ancestor: Option<&Sequence>) -> Sequence {
    let up = ours.increment >= 0 && theirs.increment >= 0;
    let down = ours.increment < 0 && theirs.increment < 0;
    let directed = |o: i64, t: i64, when_up: fn(i64, i64) -> i64, when_down: fn(i64, i64) -> i64| match (up, down) {
        (true, _) => when_up(o, t),
        (_, true) => when_down(o, t),
        _ => o,
    };
    let mut merged = ours.clone();
    merged.minimum = resolve(&ours.minimum, &theirs.minimum, ancestor.map(|a| &a.minimum), |o, t| *o.min(t));
    merged.maximum = resolve(&ours.maximum, &theirs.maximum, ancestor.map(|a| &a.maximum), |o, t| *o.max(t));
    merged.cache = resolve(&ours.cache, &theirs.cache, ancestor.map(|a| &a.cache), |o, t| *o.min(t));
    merged.cycle = resolve(&ours.cycle, &theirs.cycle, ancestor.map(|a| &a.cycle), |o, t| *o || *t);
    merged.data_type_id =
        resolve(&ours.data_type_id, &theirs.data_type_id, ancestor.map(|a| &a.data_type_id), |o, t| {
            let (o_name, t_name) = (type_name(o), type_name(t));
            let widens =
                (o_name == "int2" && (t_name == "int4" || t_name == "int8")) || (o_name == "int4" && t_name == "int8");
            if widens { t.clone() } else { o.clone() }
        });
    merged.increment = resolve(&ours.increment, &theirs.increment, ancestor.map(|a| &a.increment), |o, t| {
        directed(*o, *t, i64::min, i64::max)
    });
    merged.start =
        resolve(&ours.start, &theirs.start, ancestor.map(|a| &a.start), |o, t| directed(*o, *t, i64::min, i64::max));
    merged.current = resolve(&ours.current, &theirs.current, ancestor.map(|a| &a.current), |o, t| {
        directed(*o, *t, i64::max, i64::min)
    });
    merged.has_been_called =
        resolve(&ours.has_been_called, &theirs.has_been_called, ancestor.map(|a| &a.has_been_called), |o, t| *o || *t);
    merged
}

/// merge_both merges an object that both sides changed, as each collection's HandleMerge does.
fn merge_both(
    from_hash: &str,
    collection: usize,
    key: &[u8],
    ours: RootObject,
    theirs: RootObject,
    ancestor: Option<RootObject>,
) -> Result<Merged> {
    let name = || crate::dolt::diff::full_name(&crate::dolt::diff::object_name(collection, key));
    match (&ours, &theirs) {
        (RootObject::Function(_), RootObject::Function(_)) | (RootObject::Procedure(_), RootObject::Procedure(_)) => {
            conflict_of(from_hash, collection, Some(ours), Some(theirs), ancestor, key)
        }
        (RootObject::Sequence(o), RootObject::Sequence(t)) => {
            let a = match &ancestor {
                Some(RootObject::Sequence(a)) => Some(a),
                _ => None,
            };
            Ok(Merged::Put(RootObject::Sequence(merge_sequence(o, t, a))))
        }
        (RootObject::Extension(o), RootObject::Extension(t)) => {
            Ok(if o.version >= t.version { Merged::Keep } else { Merged::Put(theirs.clone()) })
        }
        (RootObject::Type(o), RootObject::Type(t)) => {
            if o.typ_type != t.typ_type {
                return Err(error(format!(
                    "cannot merge type \"{}\" because type types do not match: '{}' and '{}'\"",
                    type_name(&t.id),
                    String::from_utf8_lossy(&o.typ_type),
                    String::from_utf8_lossy(&t.typ_type)
                )));
            }
            if t.typ_type != b"d" {
                return Err(error(format!("cannot merge `{}` due to unsupported type", type_name(&o.id))));
            }
            if o.base_type != t.base_type {
                return Err(error(format!("base types of domain type \"{}\" do not match", type_name(&t.id))));
            }
            let a = match &ancestor {
                Some(RootObject::Type(a)) => Some(a),
                _ => None,
            };
            let mut merged = o.clone();
            let resolve_default = || -> Result<Vec<u8>> {
                if a.is_some_and(|a| o.default == a.default) {
                    return Ok(t.default.clone());
                }
                if a.is_some_and(|a| t.default == a.default) || o.default == t.default {
                    return Ok(o.default.clone());
                }
                if o.default.is_empty() {
                    return Ok(t.default.clone());
                }
                if !t.default.is_empty() {
                    return Err(error(format!("default values of domain type \"{}\" do not match", type_name(&t.id))));
                }
                Ok(o.default.clone())
            };
            merged.default = resolve_default()?;
            merged.not_null = match a {
                Some(a) if o.not_null == a.not_null => t.not_null,
                Some(a) if t.not_null == a.not_null => o.not_null,
                _ => o.not_null || t.not_null,
            };
            Ok(Merged::Put(RootObject::Type(merged)))
        }
        _ => Err(error(format!("unable to merge `{}`", name()))),
    }
}

/// merge_object merges one root object of a collection, given each side's address of it, as Dolt's
/// MaybeShortCircuit and then Doltgres' HandleMerge do.
fn merge_object(
    db: &mut Database,
    from_hash: &str,
    collection: usize,
    key: &[u8],
    (ours, theirs, ancestor): Sides,
) -> Result<Merged> {
    if ours == theirs || (ancestor.is_some() && theirs == ancestor) {
        return Ok(Merged::Keep);
    }
    if ancestor.is_none() && ours.is_none() {
        return Ok(Merged::Put(load_object(db, collection, theirs.unwrap_or_default())?));
    }
    if ancestor.is_none() && theirs.is_none() {
        return Ok(Merged::Keep);
    }
    if ours == ancestor {
        return Ok(match theirs {
            Some(theirs) => Merged::Put(load_object(db, collection, theirs)?),
            None => Merged::Remove,
        });
    }
    let mut load = |address: Option<Hash>| address.map(|a| load_object(db, collection, a)).transpose();
    let (o, t, a) = (load(ours)?, load(theirs)?, load(ancestor)?);
    match (o, t) {
        (Some(o), Some(t)) => merge_both(from_hash, collection, key, o, t, a),
        (None, Some(t)) => match t.serialize() == a.as_ref().map(RootObject::serialize).unwrap_or_default() {
            true => Ok(Merged::Keep),
            false => conflict_of(from_hash, collection, None, Some(t), a, key),
        },
        (Some(o), None) => match o.serialize() == a.as_ref().map(RootObject::serialize).unwrap_or_default() {
            true => Ok(Merged::Remove),
            false => conflict_of(from_hash, collection, Some(o), None, a, key),
        },
        (None, None) => Ok(Merged::Keep),
    }
}

/// merge_collections merges every root object of the collections other than conflicts into a merged root that starts
/// as ours, storing the conflicts it finds, and reports how many objects conflict.
pub fn merge_collections(
    ctx: &mut Ctx<'_>,
    merged: &mut Root,
    roots: (&Root, &Root, &Root),
    from: Hash,
) -> Result<usize> {
    let (ours, theirs, base) = roots;
    let from_hash = from.to_string();
    let mut conflicts = 0;
    for collection in 0..merged.root_objects.len() {
        if collection == CONFLICTS
            || (ours.root_objects[collection] == theirs.root_objects[collection])
            || theirs.root_objects[collection] == base.root_objects[collection]
        {
            continue;
        }
        let mut objects: BTreeMap<Vec<u8>, Sides> = BTreeMap::new();
        for (key, address) in ours.objects(ctx.db, collection)? {
            objects.entry(key).or_default().0 = Some(address);
        }
        for (key, address) in theirs.objects(ctx.db, collection)? {
            objects.entry(key).or_default().1 = Some(address);
        }
        for (key, address) in base.objects(ctx.db, collection)? {
            objects.entry(key).or_default().2 = Some(address);
        }
        for (key, sides) in objects {
            match merge_object(ctx.db, &from_hash, collection, &key, sides)? {
                Merged::Keep => {}
                Merged::Put(object) => store_object(ctx.db, merged, collection, &key, &object)?,
                Merged::Remove => merged.put_object(ctx.db, collection, &key, None)?,
                Merged::Conflict(conflict) => {
                    conflicts += 1;
                    put_conflict(ctx.db, merged, &conflict)?;
                }
            }
        }
    }
    Ok(conflicts)
}

/// ObjectConflictTable is the table of one root object's conflicting fields: the object's schema and name as diffs
/// show it, and the conflict's ID.
#[derive(Clone, Debug, PartialEq)]
pub struct ObjectConflictTable {
    pub schema: String,
    pub name: String,
    pub id: Vec<u8>,
}

/// COLUMNS are the columns of a root object's conflicts table.
const COLUMNS: &[&str] = &[
    "from_root_ish",
    "base_value",
    "our_value",
    "our_diff_type",
    "their_value",
    "their_diff_type",
    "dolt_conflict_id",
];

/// conflicts returns the conflicts of a root.
fn conflicts(db: &mut Database, root: &Root) -> Result<Vec<Conflict>> {
    let mut out = Vec::new();
    for (_, address) in root.objects(db, CONFLICTS)? {
        if let RootObject::Conflict(conflict) = load_object(db, CONFLICTS, address)? {
            out.push(conflict);
        }
    }
    Ok(out)
}

/// conflict_name returns the schema and name of the object that a conflict holds, as diffs show it.
fn conflict_name(conflict: &Conflict) -> (String, String) {
    crate::dolt::diff::object_name(conflict.root_object_id as usize - 1, &conflict.id)
}

/// summary_rows returns the working root's root objects with conflicts and how many fields conflict, which the
/// dolt_conflicts table shows after the tables.
pub fn summary_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let mut rows = Vec::new();
    for conflict in conflicts(ctx.db, &ctx.txn.root.clone())? {
        let count = conflict_diffs(&conflict)?.0.len() as i64;
        rows.push(vec![
            Value::Text(crate::dolt::diff::full_name(&conflict_name(&conflict))),
            Value::Numeric(crate::numeric::Numeric::from_i64(count)),
        ]);
    }
    Ok(rows)
}

/// conflict_names returns the schemas and names of a root's root objects with conflicts.
pub fn conflict_names(db: &mut Database, root: &Root) -> Result<Vec<(String, String)>> {
    Ok(conflicts(db, root)?.iter().map(conflict_name).collect())
}

impl ObjectConflictTable {
    /// lookup returns the conflicts table of the root object that a name after `dolt_conflicts_` names, searching
    /// the session's schemas for an unqualified name.
    pub fn lookup(ctx: &mut Ctx<'_>, schema: &str, name: &str) -> Result<Option<ObjectConflictTable>> {
        let Some(object) = name.strip_prefix("dolt_conflicts_") else { return Ok(None) };
        let schemas = if schema.is_empty() { ctx.session.search_path() } else { vec![schema.to_string()] };
        let found = conflicts(ctx.db, &ctx.txn.root.clone())?;
        for schema in schemas {
            if let Some(conflict) = found.iter().find(|c| conflict_name(c) == (schema.clone(), object.to_string())) {
                return Ok(Some(ObjectConflictTable { schema, name: object.to_string(), id: conflict.id.clone() }));
            }
        }
        Ok(None)
    }

    /// columns returns the table's column names and types.
    pub fn columns(&self) -> Vec<(String, crate::catalog::ColumnType)> {
        COLUMNS.iter().map(|c| (c.to_string(), crate::expr::typ(TEXT))).collect()
    }

    /// conflict loads the table's conflict from the working root.
    fn conflict(&self, ctx: &mut Ctx<'_>) -> Result<Option<Conflict>> {
        Ok(conflicts(ctx.db, &ctx.txn.root.clone())?.into_iter().find(|c| c.id == self.id))
    }

    /// rows returns a row for each conflicting field.
    pub fn rows(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        let Some(conflict) = self.conflict(ctx)? else { return Ok(Vec::new()) };
        Ok(conflict_diffs(&conflict)?.0.iter().map(Diff::row).collect())
    }

    /// delete resolves the conflicting fields whose rows were selected by keeping our values, as Doltgres'
    /// RemoveDiffs does, and returns how many it resolved.
    pub fn delete(&self, ctx: &mut Ctx<'_>, selected: &[Vec<Value>]) -> Result<usize> {
        let Some(mut conflict) = self.conflict(ctx)? else { return Ok(0) };
        let (diffs, _) = conflict_diffs(&conflict)?;
        let chosen: Vec<&Diff> = diffs.iter().filter(|d| selected.contains(&d.row())).collect();
        if chosen.is_empty() {
            return Ok(0);
        }
        if let [diff] = chosen.as_slice()
            && diff.field == ROOT_OBJECT
        {
            conflict.theirs = conflict.ours.clone();
            conflict.ancestor = None;
        } else {
            for diff in &chosen {
                let value = diff.ours.clone();
                conflict.theirs = update_field(ctx, conflict.theirs.take(), &diff.field, value)?;
            }
        }
        let mut root = ctx.txn.root.clone();
        put_conflict(ctx.db, &mut root, &conflict)?;
        ctx.txn.root = root;
        Ok(chosen.len())
    }

    /// update sets our values of conflicting fields, given each changed row's field and new value, as Doltgres'
    /// conflict UpdateField does.
    pub fn update(&self, ctx: &mut Ctx<'_>, changes: &[(String, Option<String>)]) -> Result<()> {
        if changes.is_empty() {
            return Ok(());
        }
        let Some(mut conflict) = self.conflict(ctx)? else { return Ok(()) };
        for (field, value) in changes {
            if field == ROOT_OBJECT {
                match value.as_deref() {
                    Some("ours") => conflict.theirs = conflict.ours.clone(),
                    Some("theirs") => conflict.ours = conflict.theirs.clone(),
                    Some("ancestor") => {
                        conflict.ours = conflict.ancestor.clone();
                        conflict.theirs = conflict.ancestor.clone();
                    }
                    other => {
                        return Err(error(format!("cannot replace the object with `{}`", other.unwrap_or_default())));
                    }
                }
                conflict.ancestor = None;
                continue;
            }
            let kind = field_kind(conflict.root_object_id, field)
                .ok_or_else(|| error(format!("cannot find a field named `{field}`")))?;
            let value = value.as_ref().map(|v| match kind {
                FieldKind::Bool => Field::Bool(matches!(v.as_str(), "true" | "t" | "yes" | "on" | "1")),
                FieldKind::Text => Field::Text(v.clone()),
            });
            conflict.ours = update_field(ctx, conflict.ours.take(), field, value)?;
        }
        let mut root = ctx.txn.root.clone();
        put_conflict(ctx.db, &mut root, &conflict)?;
        ctx.txn.root = root;
        Ok(())
    }
}

/// FieldKind is the type of a conflict field's values.
enum FieldKind {
    Text,
    Bool,
}

/// field_kind returns the type of a field of a kind of root object, as Doltgres' GetFieldType does.
fn field_kind(root_object_id: i64, field: &str) -> Option<FieldKind> {
    match (root_object_id, field) {
        (3, "non_deterministic" | "strict" | "set_of") => Some(FieldKind::Bool),
        (3, "parameters" | "return_type" | "definition" | "extension_name" | "extension_symbol" | "sql_definition") => {
            Some(FieldKind::Text)
        }
        (7, "parameters" | "definition" | "extension_name" | "extension_symbol" | "sql_definition") => {
            Some(FieldKind::Text)
        }
        _ => None,
    }
}

/// update_field returns an object with one field changed, recompiling a function's or procedure's body when its
/// definition changes, as Doltgres' UpdateField does.
fn update_field(
    ctx: &mut Ctx<'_>,
    object: Option<Box<RootObject>>,
    field: &str,
    value: Option<Field>,
) -> Result<Option<Box<RootObject>>> {
    let Some(mut object) = object else { return Err(error("cannot update a field of a deleted object")) };
    let value = value.unwrap_or(Field::Text(String::new()));
    match object.as_mut() {
        RootObject::Function(f) => match field {
            "parameters" => f.all_params = with_names_and_modes(&f.all_params, &value.text())?,
            "return_type" => f.return_type = with_type_name(&f.return_type, &value.text()),
            "non_deterministic" => f.is_non_deterministic = value.bool(),
            "strict" => f.strict = value.bool(),
            "definition" => {
                f.definition = replace_definition(&f.definition, &value.text());
                let text = String::from_utf8_lossy(&f.definition).into_owned();
                f.operations = crate::plpgsql::compile(ctx, &text, &value.text())?;
            }
            "extension_name" => f.extension_name = value.text().into_bytes(),
            "extension_symbol" => f.extension_symbol = value.text().into_bytes(),
            "sql_definition" => f.sql_definition = value.text().into_bytes(),
            "set_of" => f.set_of = value.bool(),
            _ => return Err(error(format!("cannot find a field named `{field}`"))),
        },
        RootObject::Procedure(p) => match field {
            "parameters" => p.all_params = with_names_and_modes(&p.all_params, &value.text())?,
            "definition" => {
                p.definition = replace_definition(&p.definition, &value.text());
                let text = String::from_utf8_lossy(&p.definition).into_owned();
                p.operations = crate::plpgsql::compile(ctx, &text, &value.text())?;
            }
            "extension_name" => p.extension_name = value.text().into_bytes(),
            "extension_symbol" => p.extension_symbol = value.text().into_bytes(),
            "sql_definition" => p.sql_definition = value.text().into_bytes(),
            _ => return Err(error(format!("cannot find a field named `{field}`"))),
        },
        _ => return Err(error(format!("cannot find a field named `{field}`"))),
    }
    Ok(Some(object))
}
