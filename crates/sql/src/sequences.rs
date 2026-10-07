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

//! Sequences: root objects that hand out numbers, outside of transactions, as Postgres' sequences do.

use doltdb::database::Database;
use doltdb::root::Root;
use objects::Sequence;
use pg_query::protobuf::{AlterSeqStmt, CreateSeqStmt, DefElem, DropBehavior, DropStmt};
use pg_query::{Node, NodeEnum};
use store::Hash;

use crate::Outcome;
use crate::catalog::id::{self, SECTION_SEQUENCE, SECTION_TABLE, SECTION_TYPE};
use crate::error::{PgError, Result, code};
use crate::expr::{node_name, position};
use crate::oid::{BOOL, INT8, REGCLASS, TEXT};
use crate::query::Ctx;
use crate::types::Value;

/// COLLECTION is the position of the sequence collection among a root value's root object collections.
pub const COLLECTION: usize = 0;

/// sequence_id returns the ID of a sequence.
pub fn sequence_id(schema: &str, name: &str) -> Vec<u8> {
    id::new(SECTION_SEQUENCE, &[schema, name])
}

/// schema_and_name returns a sequence's schema and name.
pub fn schema_and_name(sequence: &Sequence) -> (String, String) {
    let mut parts = id::segments(&sequence.id).into_iter();
    (parts.next().unwrap_or_default(), parts.next().unwrap_or_default())
}

/// all returns every sequence of a root value.
pub fn all(db: &mut Database, root: &Root) -> Result<Vec<Sequence>> {
    let mut sequences = Vec::new();
    for (_, address) in root.objects(db, COLLECTION)? {
        sequences.push(Sequence::deserialize(&prolly::read_blob(db, &address)?)?);
    }
    Ok(sequences)
}

/// find returns a sequence of a root value by schema and name.
pub fn find(db: &mut Database, root: &Root, schema: &str, name: &str) -> Result<Option<Sequence>> {
    let wanted = sequence_id(schema, name);
    for (key, address) in root.objects(db, COLLECTION)? {
        if key == wanted {
            return Ok(Some(Sequence::deserialize(&prolly::read_blob(db, &address)?)?));
        }
    }
    Ok(None)
}

/// store writes a sequence into a root value.
pub fn store(db: &mut Database, root: &mut Root, sequence: &Sequence) -> Result<()> {
    let data = sequence.serialize();
    let mut sink = |_: Hash, bytes: &[u8]| {
        db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
    };
    let (address, _) = prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty sequence"))?;
    root.put_object(db, COLLECTION, &sequence.id, Some(address))?;
    Ok(())
}

/// type_range returns the smallest and largest values of a sequence's data type.
fn type_range(data_type: &str) -> (i64, i64) {
    match data_type {
        "int2" => (i16::MIN as i64, i16::MAX as i64),
        "int4" => (i32::MIN as i64, i32::MAX as i64),
        _ => (i64::MIN, i64::MAX),
    }
}

/// data_type_name returns the pg_catalog name of a sequence's data type.
fn data_type_name(sequence: &Sequence) -> String {
    id::segments(&sequence.data_type_id).pop().unwrap_or_else(|| "int8".into())
}

/// invalid returns Postgres' error for a sequence option it rejects.
fn invalid(message: String) -> PgError {
    PgError::new(code::INVALID_PARAMETER_VALUE, message)
}

/// type_display_name returns the name Postgres shows for a sequence data type.
fn type_display_name(data_type: &str) -> &'static str {
    match data_type {
        "int2" => "smallint",
        "int4" => "integer",
        _ => "bigint",
    }
}

/// new_sequence returns a sequence of a data type with Postgres' defaults, to which options then apply.
pub fn new_sequence(schema: &str, name: &str, data_type: &str) -> Sequence {
    let (_, max) = type_range(data_type);
    Sequence {
        id: sequence_id(schema, name),
        data_type_id: id::new(SECTION_TYPE, &["pg_catalog", data_type]),
        persistence: 0,
        start: 1,
        current: 1,
        increment: 1,
        minimum: 1,
        maximum: max,
        cache: 1,
        cycle: false,
        is_at_end: false,
        has_been_called: false,
        owner_table: Vec::new(),
        owner_column: Vec::new(),
    }
}

/// int_option reads an integer sequence option.
fn int_option(def: &DefElem) -> Result<Option<i64>> {
    match def.arg.as_deref().and_then(|a| a.node.as_ref()) {
        None => Ok(None),
        Some(NodeEnum::Integer(i)) => Ok(Some(i.ival as i64)),
        Some(NodeEnum::Float(f)) => f.fval.parse::<i64>().map(Some).map_err(|_| {
            PgError::new(
                code::NUMERIC_VALUE_OUT_OF_RANGE,
                format!("value \"{}\" is out of range for type bigint", f.fval),
            )
        }),
        _ => Err(PgError::new(code::SYNTAX_ERROR, format!("{} requires a numeric value", def.defname))),
    }
}

/// apply_options applies CREATE SEQUENCE options to a new sequence, or ALTER SEQUENCE options to an existing one, and
/// checks them as Postgres' init_params does, returning the OWNED BY names.
fn apply_options(sequence: &mut Sequence, options: &[Node], init: bool) -> Result<Option<Vec<String>>> {
    let old_type = data_type_name(sequence);
    let mut data_type = old_type.clone();
    let mut increment = None;
    let (mut minimum, mut maximum, mut start, mut restart, mut cache) = (None, None, None, None, None);
    let mut owned_by = None;
    for option in options {
        let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { continue };
        match def.defname.as_str() {
            "as" => {
                let Some(NodeEnum::TypeName(t)) = def.arg.as_deref().and_then(|a| a.node.as_ref()) else { continue };
                let ty = crate::expr::resolve_type_name(t)?;
                data_type = match ty.oid {
                    crate::oid::INT2 => "int2",
                    crate::oid::INT4 => "int4",
                    crate::oid::INT8 => "int8",
                    _ => return Err(invalid("sequence type must be smallint, integer, or bigint".into())),
                }
                .to_string();
            }
            "increment" => increment = int_option(def)?,
            "minvalue" => minimum = Some(int_option(def)?),
            "maxvalue" => maximum = Some(int_option(def)?),
            "start" => start = int_option(def)?,
            "restart" => restart = Some(int_option(def)?),
            "cache" => cache = int_option(def)?,
            "cycle" => {
                sequence.cycle = matches!(
                    def.arg.as_deref().and_then(|a| a.node.as_ref()),
                    Some(NodeEnum::Boolean(b)) if b.boolval
                ) || def.arg.is_none();
            }
            "owned_by" => {
                let Some(NodeEnum::List(list)) = def.arg.as_deref().and_then(|a| a.node.as_ref()) else { continue };
                owned_by = Some(list.items.iter().filter_map(node_name).map(str::to_string).collect());
            }
            _ => {}
        }
    }
    let last = match (sequence.has_been_called, sequence.is_at_end) {
        (true, false) => sequence.current - sequence.increment,
        _ => sequence.current,
    };
    let (old_min, old_max) = type_range(&old_type);
    let changed = data_type != old_type;
    let increment = increment.unwrap_or(if init { 1 } else { sequence.increment });
    if increment == 0 {
        return Err(invalid("INCREMENT must not be zero".into()));
    }
    let (type_min, type_max) = type_range(&data_type);
    let maximum = match maximum {
        Some(Some(m)) => m,
        Some(None) => {
            if increment > 0 {
                type_max
            } else {
                -1
            }
        }
        None if init || changed && sequence.maximum == old_max => {
            if increment > 0 {
                type_max
            } else {
                -1
            }
        }
        None => sequence.maximum,
    };
    let minimum = match minimum {
        Some(Some(m)) => m,
        Some(None) => {
            if increment > 0 {
                1
            } else {
                type_min
            }
        }
        None if init || changed && sequence.minimum == old_min => {
            if increment > 0 {
                1
            } else {
                type_min
            }
        }
        None => sequence.minimum,
    };
    let type_name = type_display_name(&data_type);
    if maximum < type_min || maximum > type_max {
        return Err(invalid(format!("MAXVALUE ({maximum}) is out of range for sequence data type {type_name}")));
    }
    if minimum < type_min || minimum > type_max {
        return Err(invalid(format!("MINVALUE ({minimum}) is out of range for sequence data type {type_name}")));
    }
    if minimum >= maximum {
        return Err(invalid(format!("MINVALUE ({minimum}) must be less than MAXVALUE ({maximum})")));
    }
    let start = match start {
        Some(start) => start,
        None if init => {
            if increment > 0 {
                minimum
            } else {
                maximum
            }
        }
        None => sequence.start,
    };
    if start < minimum {
        return Err(invalid(format!("START value ({start}) cannot be less than MINVALUE ({minimum})")));
    }
    if start > maximum {
        return Err(invalid(format!("START value ({start}) cannot be greater than MAXVALUE ({maximum})")));
    }
    let (last, called) = match restart {
        Some(value) => (value.unwrap_or(start), false),
        None if init => (start, false),
        None => (last, sequence.has_been_called),
    };
    if last < minimum {
        return Err(invalid(format!("RESTART value ({last}) cannot be less than MINVALUE ({minimum})")));
    }
    if last > maximum {
        return Err(invalid(format!("RESTART value ({last}) cannot be greater than MAXVALUE ({maximum})")));
    }
    let cache = cache.unwrap_or(if init { 1 } else { sequence.cache });
    if cache < 1 {
        return Err(invalid(format!("CACHE ({cache}) must be greater than zero")));
    }
    sequence.data_type_id = id::new(SECTION_TYPE, &["pg_catalog", &data_type]);
    sequence.increment = increment;
    sequence.minimum = minimum;
    sequence.maximum = maximum;
    sequence.start = start;
    sequence.cache = cache;
    sequence.current = last;
    sequence.is_at_end = false;
    sequence.has_been_called = called;
    if called {
        advance(sequence)?;
    }
    Ok(owned_by)
}

/// advance moves a sequence to its next value and returns the value, as nextval does.
pub fn advance(sequence: &mut Sequence) -> Result<i64> {
    let name = schema_and_name(sequence).1;
    if sequence.is_at_end {
        if !sequence.cycle {
            let (which, limit) =
                if sequence.increment > 0 { ("maximum", sequence.maximum) } else { ("minimum", sequence.minimum) };
            return Err(PgError::new(
                code::SEQUENCE_GENERATOR_LIMIT_EXCEEDED,
                format!("nextval: reached {which} value of sequence \"{name}\" ({limit})"),
            ));
        }
        sequence.is_at_end = false;
        sequence.current = if sequence.increment > 0 { sequence.minimum } else { sequence.maximum };
    }
    let value = sequence.current;
    sequence.has_been_called = true;
    match sequence.current.checked_add(sequence.increment) {
        Some(next) if next <= sequence.maximum && next >= sequence.minimum => sequence.current = next,
        _ => sequence.is_at_end = true,
    }
    Ok(value)
}

/// greater_than reports whether one state of a sequence is further along than another, counting a wrapped sequence
/// as further along, as Doltgres' SequenceState.GreaterThan does.
pub(crate) fn greater_than(a: &Sequence, b: &Sequence) -> bool {
    if a.increment > 0 {
        let (wrapped, other_wrapped) = (a.current < a.start, b.current < a.start);
        if wrapped == other_wrapped { a.current > b.current } else { wrapped }
    } else {
        let (wrapped, other_wrapped) = (a.current > a.start, b.current > a.start);
        if wrapped == other_wrapped { a.current < b.current } else { wrapped }
    }
}

/// parse_qualified_name splits text naming a relation into its identifiers, folding unquoted ones to lower case, as
/// Postgres' stringToQualifiedNameList does.
pub fn parse_qualified_name(text: &str) -> Result<Vec<String>> {
    let bad = || PgError::new(code::INVALID_NAME, "invalid name syntax");
    let mut names = Vec::new();
    let mut chars = text.trim().chars().peekable();
    loop {
        let mut name = String::new();
        if chars.peek() == Some(&'"') {
            chars.next();
            loop {
                match chars.next() {
                    Some('"') if chars.peek() == Some(&'"') => {
                        chars.next();
                        name.push('"');
                    }
                    Some('"') => break,
                    Some(c) => name.push(c),
                    None => return Err(bad()),
                }
            }
            if name.is_empty() {
                return Err(bad());
            }
        } else {
            while let Some(&c) = chars.peek() {
                if c == '.' || c.is_whitespace() {
                    break;
                }
                name.push(c.to_ascii_lowercase());
                chars.next();
            }
            if name.is_empty() {
                return Err(bad());
            }
        }
        names.push(name);
        while chars.peek().is_some_and(|c| c.is_whitespace()) {
            chars.next();
        }
        match chars.next() {
            Some('.') => {
                while chars.peek().is_some_and(|c| c.is_whitespace()) {
                    chars.next();
                }
            }
            None => return Ok(names),
            Some(_) => return Err(bad()),
        }
    }
}

impl Ctx<'_> {
    /// resolve_sequence finds the sequence that text names, searching the session's schemas for an unqualified name.
    pub fn resolve_sequence(&mut self, text: &str, location: i32) -> Result<Sequence> {
        let names = parse_qualified_name(text)?;
        let (schemas, name) = match names.as_slice() {
            [name] => (self.session.search_path(), name.clone()),
            [schema, name] => (vec![schema.clone()], name.clone()),
            [_, schema, name] => (vec![schema.clone()], name.clone()),
            _ => {
                return Err(PgError::new(
                    code::SYNTAX_ERROR,
                    format!("improper relation name (too many dotted names): {text}"),
                ));
            }
        };
        for schema in &schemas {
            if let Some(sequence) = find(self.db, &self.txn.root, schema, &name)? {
                return Ok(sequence);
            }
        }
        Err(PgError {
            position: position(location),
            ..PgError::new(code::UNDEFINED_TABLE, format!("relation \"{text}\" does not exist"))
        })
    }

    /// latest returns a sequence in its latest state across every transaction.
    pub(crate) fn latest(&mut self, sequence: Sequence) -> Result<Sequence> {
        let tracker = self.txn.sequences.lock().map_err(|_| PgError::internal("a lock was poisoned"))?;
        Ok(match tracker.get(&sequence.id) {
            Some(tracked) if greater_than(tracked, &sequence) => Sequence {
                current: tracked.current,
                is_at_end: tracked.is_at_end,
                has_been_called: tracked.has_been_called,
                ..sequence
            },
            _ => sequence,
        })
    }

    /// save records a sequence's new state for every transaction and writes it to the working root.
    fn save(&mut self, sequence: &Sequence) -> Result<()> {
        self.txn
            .sequences
            .lock()
            .map_err(|_| PgError::internal("a lock was poisoned"))?
            .insert(sequence.id.clone(), sequence.clone());
        store(self.db, &mut self.txn.root, sequence)
    }

    /// forget_tracked drops the state that every transaction shares for a sequence, so that a new sequence that takes
    /// a dropped one's name starts over.
    fn forget_tracked(&mut self, id: &[u8]) -> Result<()> {
        self.txn.sequences.lock().map_err(|_| PgError::internal("a lock was poisoned"))?.remove(id);
        Ok(())
    }

    /// next_value advances a sequence, as nextval does.
    pub fn next_value(&mut self, sequence: Sequence) -> Result<i64> {
        let mut sequence = self.latest(sequence)?;
        let value = advance(&mut sequence)?;
        self.save(&sequence)?;
        self.session.sequence_values.insert(sequence.id.clone(), value);
        self.session.last_sequence = Some((sequence.id.clone(), value));
        Ok(value)
    }

    /// create_sequence runs CREATE SEQUENCE.
    pub fn create_sequence(&mut self, stmt: &CreateSeqStmt) -> Result<Outcome> {
        let relation = stmt.sequence.as_ref().ok_or_else(|| PgError::internal("CREATE SEQUENCE without a name"))?;
        let schema = self.target_schema(&relation.schemaname, relation.location)?;
        let name = relation.relname.clone();
        if self.relation_names(&schema)?.contains(&name) {
            let message = format!("relation \"{name}\" already exists");
            if stmt.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE SEQUENCE"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        let mut sequence = new_sequence(&schema, &name, "int8");
        if let Some(owned_by) = apply_options(&mut sequence, &stmt.options, true)? {
            self.set_owner(&mut sequence, &schema, &owned_by)?;
        }
        self.forget_tracked(&sequence.id)?;
        store(self.db, &mut self.txn.root, &sequence)?;
        self.own(crate::auth::Object::Sequence(schema, name))?;
        Ok(Outcome::command("CREATE SEQUENCE"))
    }

    /// set_owner links a sequence to the table column that OWNED BY names, so that dropping the table drops it, or
    /// unlinks it for OWNED BY NONE, failing as Postgres' process_owned_by does.
    fn set_owner(&mut self, sequence: &mut Sequence, schema: &str, owned_by: &[String]) -> Result<()> {
        let (table_schema, table, column) = match owned_by {
            [none] if none == "none" => {
                sequence.owner_table.clear();
                sequence.owner_column.clear();
                return Ok(());
            }
            [table, column] => (String::new(), table.clone(), column.clone()),
            [.., table_schema, table, column] if owned_by.len() <= 4 => {
                (table_schema.clone(), table.clone(), column.clone())
            }
            _ => {
                return Err(PgError {
                    hint: Some("Specify OWNED BY table.column or OWNED BY NONE.".into()),
                    ..PgError::new(code::SYNTAX_ERROR, "invalid OWNED BY option")
                });
            }
        };
        let relation =
            pg_query::protobuf::RangeVar { schemaname: table_schema, relname: table, inh: true, ..Default::default() };
        let def = self.resolve_table(&relation).map_err(|err| PgError { position: None, ..err })?;
        if def.schema != schema {
            return Err(PgError::new(
                code::OBJECT_NOT_IN_PREREQUISITE_STATE,
                "sequence must be in same schema as table it is linked to",
            ));
        }
        if !def.columns.iter().any(|c| c.name == column) {
            return Err(PgError::new(
                code::UNDEFINED_COLUMN,
                format!("column \"{column}\" of relation \"{}\" does not exist", def.name),
            ));
        }
        sequence.owner_table = id::new(SECTION_TABLE, &[schema, &def.name]);
        sequence.owner_column = column.into_bytes();
        Ok(())
    }

    /// find_sequence returns the sequence a relation names, in its schema or else the first schema of the search path
    /// that has it.
    pub(crate) fn find_sequence(&mut self, relation: &pg_query::protobuf::RangeVar) -> Result<Option<Sequence>> {
        let schemas =
            if relation.schemaname.is_empty() { self.session.search_path() } else { vec![relation.schemaname.clone()] };
        for schema in schemas {
            if let Some(sequence) = find(self.db, &self.txn.root, &schema, &relation.relname)? {
                return Ok(Some(sequence));
            }
        }
        Ok(None)
    }

    /// alter_sequence runs ALTER SEQUENCE, changing an existing sequence's options.
    pub fn alter_sequence(&mut self, stmt: &AlterSeqStmt) -> Result<Outcome> {
        let relation = stmt.sequence.as_ref().ok_or_else(|| PgError::internal("ALTER SEQUENCE without a name"))?;
        let Some(sequence) = self.find_sequence(relation)? else {
            let message = format!("relation \"{}\" does not exist", relation.relname);
            if stmt.missing_ok {
                self.session.notice(PgError::notice("00000", format!("{message}, skipping")));
                return Ok(Outcome::command("ALTER SEQUENCE"));
            }
            return Err(PgError::new(code::UNDEFINED_TABLE, message));
        };
        let (schema, name) = schema_and_name(&sequence);
        self.require_owner(&crate::auth::Object::Sequence(schema.clone(), name))?;
        let mut sequence = self.latest(sequence)?;
        if let Some(owned_by) = apply_options(&mut sequence, &stmt.options, false)? {
            self.set_owner(&mut sequence, &schema, &owned_by)?;
        }
        self.save(&sequence)?;
        Ok(Outcome::command("ALTER SEQUENCE"))
    }

    /// drop_sequences runs DROP SEQUENCE, refusing a sequence that a column default uses unless it cascades.
    pub fn drop_sequences(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let cascade = DropBehavior::try_from(drop.behavior) == Ok(DropBehavior::DropCascade);
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let parts: Vec<&str> = list.items.iter().filter_map(node_name).collect();
            let (schemas, name) = match parts.as_slice() {
                [name] => (self.session.search_path(), name.to_string()),
                [.., schema, name] => (vec![schema.to_string()], name.to_string()),
                [] => continue,
            };
            let mut found = None;
            for schema in &schemas {
                if let Some(sequence) = find(self.db, &self.txn.root, schema, &name)? {
                    found = Some(sequence);
                    break;
                }
            }
            match found {
                Some(sequence) => {
                    let (schema, sequence_name) = schema_and_name(&sequence);
                    self.require_owner(&crate::auth::Object::Sequence(schema, sequence_name))?;
                    doomed.push(sequence)
                }
                None => {
                    let schema = if parts.len() > 1 { parts[parts.len() - 2] } else { "" };
                    self.missing_relation(
                        "sequence",
                        schema,
                        &parts.join("."),
                        drop.missing_ok,
                        code::UNDEFINED_TABLE,
                    )?;
                }
            }
        }
        for sequence in doomed {
            let (schema, name) = schema_and_name(&sequence);
            let users = self.default_users(&schema, &name)?;
            if !users.is_empty() && !cascade {
                let detail: Vec<String> = users
                    .iter()
                    .map(|(t, c)| format!("default value for column {c} of table {t} depends on sequence {name}"))
                    .collect();
                return Err(PgError {
                    detail: Some(detail.join("\n")),
                    hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                    ..PgError::new(
                        code::DEPENDENT_OBJECTS_STILL_EXIST,
                        format!("cannot drop sequence {name} because other objects depend on it"),
                    )
                });
            }
            for (table, column) in &users {
                self.session.notice(PgError::notice(
                    "00000",
                    format!("drop cascades to default value for column {column} of table {table}"),
                ));
                self.drop_default(&schema, table, column)?;
            }
            self.txn.root.put_object(self.db, COLLECTION, &sequence.id, None)?;
            let (schema, name) = schema_and_name(&sequence);
            self.forget_object(&crate::auth::Object::Sequence(schema, name))?;
        }
        Ok(Outcome::command("DROP SEQUENCE"))
    }

    /// default_users returns the table and column of each default that calls nextval on a sequence.
    fn default_users(&mut self, schema: &str, name: &str) -> Result<Vec<(String, String)>> {
        let mut users = Vec::new();
        let prefix = doltdb::root::table_key(schema, "");
        let patterns = [format!("'{schema}.{name}'"), format!("'{name}'")];
        for (key, address) in self.txn.root.tables(self.db)? {
            let Some(table) = key.strip_prefix(prefix.as_slice()) else { continue };
            let table =
                crate::catalog::table::TableDef::load(self.db, schema, &String::from_utf8_lossy(table), address)?;
            for column in &table.columns {
                if column.default.contains("nextval(") && patterns.iter().any(|p| column.default.contains(p.as_str())) {
                    users.push((table.name.clone(), column.name.clone()));
                }
            }
        }
        Ok(users)
    }

    /// drop_default removes a column's default.
    fn drop_default(&mut self, schema: &str, table: &str, column: &str) -> Result<()> {
        let Some(mut def) = self.txn.table(self.db, schema, table)? else { return Ok(()) };
        for c in &mut def.columns {
            if c.name == column {
                c.default.clear();
            }
        }
        let mut stored = def.table.clone();
        stored.schema = self.db.write_value(def.schema_message()?)?;
        let address = stored.write(self.db)?;
        Ok(self.txn.root.put_table(self.db, schema, table, Some(address))?)
    }

    /// drop_owned_sequences drops the sequences that a table owns.
    pub fn drop_owned_sequences(&mut self, schema: &str, table: &str) -> Result<()> {
        let owner = id::new(SECTION_TABLE, &[schema, table]);
        for sequence in all(self.db, &self.txn.root)? {
            if sequence.owner_table == owner {
                self.txn.root.put_object(self.db, COLLECTION, &sequence.id, None)?;
                let (schema, name) = schema_and_name(&sequence);
                self.forget_object(&crate::auth::Object::Sequence(schema, name))?;
            }
        }
        Ok(())
    }

    /// create_owned_sequence creates the sequence behind a serial or identity column and returns the column's default.
    pub fn create_owned_sequence(
        &mut self,
        schema: &str,
        table: &str,
        column: &str,
        data_type: &str,
        options: &[Node],
        taken: &mut Vec<String>,
    ) -> Result<String> {
        let named = options.iter().find_map(|option| match option.node.as_ref() {
            Some(NodeEnum::DefElem(def)) if def.defname == "sequence_name" => {
                match def.arg.as_deref()?.node.as_ref()? {
                    NodeEnum::List(list) => list.items.iter().filter_map(node_name).next_back().map(str::to_string),
                    _ => None,
                }
            }
            _ => None,
        });
        let name = match named {
            Some(name) => name,
            None => crate::ddl::choose_relation_name(table, column, "seq", taken),
        };
        taken.push(name.clone());
        let mut sequence = new_sequence(schema, &name, data_type);
        apply_options(&mut sequence, options, true)?;
        sequence.owner_table = id::new(SECTION_TABLE, &[schema, table]);
        sequence.owner_column = column.as_bytes().to_vec();
        self.forget_tracked(&sequence.id)?;
        store(self.db, &mut self.txn.root, &sequence)?;
        self.own(crate::auth::Object::Sequence(schema.to_string(), name.clone()))?;
        let quoted = format!("{}.{}", crate::engine::quote_identifier(schema), crate::engine::quote_identifier(&name));
        Ok(format!("(nextval('{}'))", quoted.replace('\'', "''")))
    }
}

/// FUNCTIONS are the sequence functions.
pub const FUNCTIONS: &[crate::functions::Function] = &[
    f("nextval", &[TEXT], nextval),
    f("currval", &[TEXT], currval),
    f("lastval", &[], lastval),
    f("setval", &[TEXT, INT8], setval),
    f("setval", &[TEXT, INT8, BOOL], setval),
    f("nextval", &[REGCLASS], nextval),
    f("currval", &[REGCLASS], currval),
    f("setval", &[REGCLASS, INT8], setval),
    f("setval", &[REGCLASS, INT8, BOOL], setval),
];

/// f declares a strict sequence function returning bigint.
const fn f(
    name: &'static str,
    args: &'static [u32],
    implementation: crate::functions::Implementation,
) -> crate::functions::Function {
    crate::functions::Function { name, args, ret: INT8, strict: true, variadic: false, implementation }
}

/// text_arg returns a text argument.
fn text_arg(value: &Value) -> String {
    value.output().unwrap_or_default()
}

/// nextval advances a sequence and returns its value.
pub fn nextval(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let sequence = ctx.resolve_sequence(&text_arg(&args[0]), -1)?;
    let (schema, name) = schema_and_name(&sequence);
    ctx.require_sequence(&schema, &name, &["U", "w"])?;
    Ok(Value::Int8(ctx.next_value(sequence)?))
}

/// currval returns the value nextval last returned for a sequence in this session.
pub fn currval(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let sequence = ctx.resolve_sequence(&text_arg(&args[0]), -1)?;
    let (schema, name) = schema_and_name(&sequence);
    ctx.require_sequence(&schema, &name, &["U", "r"])?;
    match ctx.session.sequence_values.get(&sequence.id) {
        Some(value) => Ok(Value::Int8(*value)),
        None => Err(PgError::new(
            code::OBJECT_NOT_IN_PREREQUISITE_STATE,
            format!("currval of sequence \"{}\" is not yet defined in this session", schema_and_name(&sequence).1),
        )),
    }
}

/// lastval returns the value of the session's most recent nextval.
pub fn lastval(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    let not_defined =
        || PgError::new(code::OBJECT_NOT_IN_PREREQUISITE_STATE, "lastval is not yet defined in this session");
    let Some((sequence_id, value)) = ctx.session.last_sequence.clone() else { return Err(not_defined()) };
    let mut parts = id::segments(&sequence_id).into_iter();
    let (schema, name) = (parts.next().unwrap_or_default(), parts.next().unwrap_or_default());
    if find(ctx.db, &ctx.txn.root, &schema, &name)?.is_none() {
        return Err(not_defined());
    }
    Ok(Value::Int8(value))
}

/// setval sets a sequence's value, and whether nextval has handed it out, defaulting to true.
pub fn setval(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let sequence = ctx.resolve_sequence(&text_arg(&args[0]), -1)?;
    let (schema, name) = schema_and_name(&sequence);
    ctx.require_sequence(&schema, &name, &["w"])?;
    let Value::Int8(value) = args[1] else { return Ok(Value::Null) };
    let called = !matches!(args.get(2), Some(Value::Bool(false)));
    let mut sequence = ctx.latest(sequence)?;
    if value < sequence.minimum || value > sequence.maximum {
        return Err(PgError::new(
            code::NUMERIC_VALUE_OUT_OF_RANGE,
            format!(
                "setval: value {value} is out of bounds for sequence \"{}\" ({}..{})",
                schema_and_name(&sequence).1,
                sequence.minimum,
                sequence.maximum
            ),
        ));
    }
    sequence.current = value;
    sequence.is_at_end = false;
    sequence.has_been_called = called;
    if called {
        advance(&mut sequence)?;
    }
    ctx.save(&sequence)?;
    Ok(Value::Int8(value))
}
