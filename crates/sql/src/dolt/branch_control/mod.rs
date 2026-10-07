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

//! Dolt's branch control: the access table that grants permissions on branches by database, branch, user, and host
//! patterns, the namespace table that limits who may create branches with some names, their system tables, and the
//! `BRCL` file that Go's server reads and writes.

mod weights;

use std::path::PathBuf;

use serial::Builder;
use serial::Table;

use pg_query::NodeEnum;

use crate::error::{PgError, Result, code};
use crate::query::Ctx;
use crate::types::Value;

/// ADMIN grants full control of a branch, including its rows in the branch control tables.
pub const ADMIN: u64 = 1;
/// WRITE allows every change to a branch.
pub const WRITE: u64 = 1 << 1;
/// MERGE allows merging into a branch.
pub const MERGE: u64 = 1 << 2;
/// READ allows reading a branch, as having no permissions does.
pub const READ: u64 = 1 << 3;

/// PERMISSION_NAMES are the names of the permissions in bit order, as the `permissions` column spells them.
const PERMISSION_NAMES: [&str; 4] = ["admin", "write", "merge", "read"];

/// SINGLE_MATCH, ANY_MATCH, and COLUMN_MARKER are the sort orders of `_`, `%`, and the start of a column.
const SINGLE_MATCH: i32 = -1;
const ANY_MATCH: i32 = -2;
const COLUMN_MARKER: i32 = -3;

/// ACCESS_COLUMNS and NAMESPACE_COLUMNS are the columns of dolt_branch_control and dolt_branch_namespace_control.
const ACCESS_COLUMNS: [&str; 5] = ["database", "branch", "user", "host", "permissions"];
const NAMESPACE_COLUMNS: [&str; 4] = ["database", "branch", "user", "host"];

/// FILE_ID is the file identifier of a branch control file.
const FILE_ID: &str = "BRCL";

/// MAX_LENGTH is the longest expression a column holds, in bytes.
const MAX_LENGTH: usize = u16::MAX as usize;

/// BinlogRow is a change to a branch control table, which the file records in order.
#[derive(Clone, Debug)]
struct BinlogRow {
    insert: bool,
    database: String,
    branch: String,
    user: String,
    host: String,
    permissions: u64,
}

/// AccessRow is a row of the access table.
#[derive(Clone, Debug)]
pub struct AccessRow {
    pub database: String,
    pub branch: String,
    pub user: String,
    pub host: String,
    pub permissions: u64,
}

/// Access is the access table, as Dolt's Access keeps it: rows in slots that deletions free for later inserts, and
/// the sort orders of each live expression with its slot.
#[derive(Clone, Debug, Default)]
struct Access {
    rows: Vec<AccessRow>,
    free: Vec<u32>,
    expressions: Vec<(Vec<i32>, u32)>,
    binlog: Vec<BinlogRow>,
}

/// MatchExpression is one column's pattern of a namespace row, as sort orders, with the row's position.
#[derive(Clone, Debug)]
struct MatchExpression {
    index: u32,
    orders: Vec<i32>,
}

/// NamespaceRow is a row of the namespace table.
#[derive(Clone, Debug)]
pub struct NamespaceRow {
    pub database: String,
    pub branch: String,
    pub user: String,
    pub host: String,
}

/// Namespace is the namespace table, as Dolt's Namespace keeps it: a pattern list per column and the rows.
#[derive(Clone, Debug, Default)]
struct Namespace {
    databases: Vec<MatchExpression>,
    branches: Vec<MatchExpression>,
    users: Vec<MatchExpression>,
    hosts: Vec<MatchExpression>,
    values: Vec<NamespaceRow>,
    binlog: Vec<BinlogRow>,
}

/// Controller holds the branch control tables of a server and the file they persist to, if any.
#[derive(Debug, Default)]
pub struct Controller {
    access: Access,
    namespace: Namespace,
    path: Option<PathBuf>,
}

/// fold_expression returns a pattern in its smallest form, with `%_` written `_%` and repeated `%` collapsed, as
/// Dolt's FoldExpression does.
pub fn fold_expression(text: &str) -> String {
    let mut text = text.to_string();
    loop {
        let mut out = String::with_capacity(text.len());
        let (mut skip_next, mut consider_next) = (false, false);
        for c in text.chars() {
            if skip_next {
                skip_next = false;
                out.push(c);
                continue;
            }
            if consider_next {
                consider_next = false;
                match c {
                    '\\' => {
                        out.push('%');
                        out.push(c);
                        skip_next = true;
                    }
                    '_' => {
                        out.push(c);
                        out.push('%');
                    }
                    '%' => out.push(c),
                    _ => {
                        out.push('%');
                        out.push(c);
                    }
                }
                continue;
            }
            match c {
                '\\' => {
                    out.push(c);
                    skip_next = true;
                }
                '%' => consider_next = true,
                _ => out.push(c),
            }
        }
        if consider_next {
            out.push('%');
        }
        if out == text {
            return text;
        }
        text = out;
    }
}

/// Collation is how a column compares characters.
#[derive(Clone, Copy, PartialEq)]
enum Collation {
    /// utf8mb4_0900_ai_ci, which ignores case and accents.
    Insensitive,
    /// utf8mb4_0900_bin, which compares code points.
    Binary,
}

impl Collation {
    /// weight returns a character's sort order.
    fn weight(self, c: char) -> i32 {
        match self {
            Collation::Insensitive => weights::ai_ci(c),
            Collation::Binary => c as i32,
        }
    }
}

/// COLLATIONS are the collations of the database, branch, user, and host columns.
const COLLATIONS: [Collation; 4] =
    [Collation::Insensitive, Collation::Insensitive, Collation::Binary, Collation::Insensitive];

/// parse_column appends a pattern's sort orders, where `\` escapes the next character.
fn parse_column(orders: &mut Vec<i32>, text: &str, collation: Collation) {
    let mut escaped = false;
    for c in text.chars() {
        if escaped {
            escaped = false;
            orders.push(collation.weight(c));
            continue;
        }
        match c {
            '\\' => escaped = true,
            '%' => orders.push(ANY_MATCH),
            '_' => orders.push(SINGLE_MATCH),
            _ => orders.push(collation.weight(c)),
        }
    }
}

/// clip_bytes returns the longest prefix of a string within a byte length that ends on a character boundary.
fn clip_bytes(text: &str, len: usize) -> &str {
    let mut end = len.min(text.len());
    while !text.is_char_boundary(end) {
        end -= 1;
    }
    &text[..end]
}

/// parse_expression returns the sort orders of the four columns, each after a column marker, as Dolt's MatchNode
/// parses them.
fn parse_expression(columns: [&str; 4]) -> Vec<i32> {
    let mut orders = Vec::new();
    for (text, collation) in columns.into_iter().zip(COLLATIONS) {
        orders.push(COLUMN_MARKER);
        parse_column(&mut orders, clip_bytes(text, MAX_LENGTH), collation);
    }
    orders
}

/// match_lengths returns how specifically a pattern matches an input, once for each way it matches, as Dolt's
/// MatchNode counts it: one per matched character and two for a `%` that ends where its next character matches.
fn match_lengths(pattern: &[i32], input: &[i32]) -> Vec<u32> {
    let mut states = vec![(0usize, 0u32)];
    for &order in input {
        let mut next: Vec<(usize, u32)> = Vec::new();
        for &(position, length) in &states {
            let rest = &pattern[position..];
            let Some(&first) = rest.first() else { continue };
            match first {
                SINGLE_MATCH if order >= SINGLE_MATCH => next.push((position + 1, length + 1)),
                SINGLE_MATCH => {}
                ANY_MATCH => {
                    if rest.len() > 1 && rest[1] == order {
                        next.push((position + 2, length + 2));
                    }
                    if order != COLUMN_MARKER {
                        next.push((position, length));
                    }
                }
                _ if first == order => next.push((position + 1, length + 1)),
                _ => {}
            }
        }
        next.sort_unstable();
        next.dedup();
        states = next;
    }
    states
        .into_iter()
        .filter_map(|(position, length)| match &pattern[position..] {
            [] => Some(length),
            [ANY_MATCH] => Some(length + 1),
            _ => None,
        })
        .collect()
}

/// step advances a pattern of the namespace table by a character's sort order, as Dolt's MatchExpression.Matches
/// does, keeping both paths when a `%` may end.
fn step<'p>(out: &mut Vec<(u32, &'p [i32])>, index: u32, orders: &'p [i32], order: i32) {
    let Some(&head) = orders.first() else { return };
    match head {
        SINGLE_MATCH if order >= SINGLE_MATCH => out.push((index, &orders[1..])),
        SINGLE_MATCH => {}
        ANY_MATCH => {
            out.push((index, orders));
            if orders.len() > 1 && orders[1] == order {
                out.push((index, &orders[2..]));
            }
        }
        _ if head == order => out.push((index, &orders[1..])),
        _ => {}
    }
}

/// expression_matches returns the positions of the patterns that match an input string, as Dolt's Match function
/// does for one column of the namespace table, where an empty input reads as one replacement character as Go's
/// DecodeRuneInString gives it.
fn expression_matches(expressions: &[MatchExpression], text: &str, collation: Collation) -> Vec<u32> {
    let mut chars = text.chars();
    let first = collation.weight(chars.next().unwrap_or('\u{FFFD}'));
    let mut subset: Vec<(u32, &[i32])> = Vec::new();
    for expression in expressions {
        step(&mut subset, expression.index, &expression.orders, first);
    }
    for c in chars {
        let mut next = Vec::new();
        for (index, orders) in subset {
            step(&mut next, index, orders, collation.weight(c));
        }
        subset = next;
    }
    let mut out: Vec<u32> = Vec::new();
    for (index, orders) in subset {
        let at_end = orders.is_empty() || orders == [ANY_MATCH];
        if at_end && out.last() != Some(&index) {
            out.push(index);
        }
    }
    out
}

/// widen adds the permissions that a permission implies: admin implies write, write implies merge, and merge implies
/// read.
fn widen(permissions: u64) -> u64 {
    if permissions & ADMIN != 0 {
        permissions | WRITE | MERGE | READ
    } else if permissions & WRITE != 0 {
        permissions | MERGE | READ
    } else if permissions & MERGE != 0 {
        permissions | READ
    } else {
        permissions
    }
}

/// consolidate returns the strongest permission of a set.
fn consolidate(permissions: u64) -> u64 {
    [ADMIN, WRITE, MERGE].into_iter().find(|p| permissions & p != 0).unwrap_or(READ)
}

/// permissions_text returns a permission set as the `permissions` column shows it.
pub fn permissions_text(permissions: u64) -> String {
    let names: Vec<&str> =
        PERMISSION_NAMES.iter().enumerate().filter(|(i, _)| permissions & (1 << i) != 0).map(|(_, n)| *n).collect();
    names.join(",")
}

/// parse_permissions reads a permission set from its comma-separated names, failing as Go's server does for others.
pub fn parse_permissions(text: &str) -> Result<u64> {
    let mut permissions = 0;
    for name in text.split(',').filter(|n| !n.is_empty()) {
        let Some(bit) = PERMISSION_NAMES.iter().position(|p| p.eq_ignore_ascii_case(name.trim_end())) else {
            return Err(PgError::internal("Data truncated for column 'permissions' at row 1"));
        };
        permissions |= 1 << bit;
    }
    Ok(permissions)
}

/// quoted returns a string as Go's `%q` verb writes it.
fn quoted(text: &str) -> String {
    format!("{text:?}")
}

/// normalize folds and lowercases a row's expressions as Dolt stores them, leaving the user's case alone.
fn normalize(database: &str, branch: &str, user: &str, host: &str) -> [String; 4] {
    [
        fold_expression(database).to_lowercase(),
        fold_expression(branch).to_lowercase(),
        fold_expression(user),
        fold_expression(host).to_lowercase(),
    ]
}

impl Access {
    /// insert_default resets the table to Dolt's default row, which lets everyone write to every branch.
    fn insert_default(&mut self) {
        *self = Access::default();
        self.insert(["%", "%", "%", "%"].map(String::from), WRITE);
    }

    /// insert adds a row of expressions already folded, replacing the permissions of an identical expression.
    fn insert(&mut self, row: [String; 4], permissions: u64) {
        let row = row.map(|c| match c.len() > MAX_LENGTH {
            true => format!("{}%", clip_bytes(&c, MAX_LENGTH - 1)),
            false => c,
        });
        let [database, branch, user, host] = row;
        self.binlog.push(BinlogRow {
            insert: true,
            database: database.clone(),
            branch: branch.clone(),
            user: user.clone(),
            host: host.clone(),
            permissions,
        });
        let orders = parse_expression([&database, &branch, &user, &host]);
        let row = AccessRow { database, branch, user, host, permissions };
        let slot = match self.free.pop() {
            Some(slot) => {
                self.rows[slot as usize] = row;
                slot
            }
            None => {
                self.rows.push(row);
                (self.rows.len() - 1) as u32
            }
        };
        match self.expressions.iter_mut().find(|(o, _)| *o == orders) {
            Some(entry) => entry.1 = slot,
            None => self.expressions.push((orders, slot)),
        }
    }

    /// delete removes the row of identical expressions already folded.
    fn delete(&mut self, row: [String; 4]) {
        let row = row.map(|c| match c.len() > MAX_LENGTH {
            true => format!("{}%", clip_bytes(&c, MAX_LENGTH - 1)),
            false => c,
        });
        let orders = parse_expression([&row[0], &row[1], &row[2], &row[3]]);
        let Some(position) = self.expressions.iter().position(|(o, _)| *o == orders) else { return };
        let (_, slot) = self.expressions.remove(position);
        let [database, branch, user, host] = row;
        self.binlog.push(BinlogRow { insert: false, database, branch, user, host, permissions: 0 });
        self.free.push(slot);
    }

    /// matches returns whether any row matches a database, branch, user, and host, and the permissions of the most
    /// specific matches, leaving out one slot when given.
    fn matches(&self, input: [&str; 4], ignored: Option<u32>) -> (bool, u64) {
        let input = parse_expression(input);
        let (mut any, mut longest, mut permissions) = (false, 0u32, 0u64);
        for (orders, slot) in &self.expressions {
            for length in match_lengths(orders, &input) {
                any = true;
                if Some(*slot) == ignored {
                    continue;
                }
                let row_permissions = self.rows[*slot as usize].permissions;
                if length > longest {
                    (longest, permissions) = (length, row_permissions);
                } else if length == longest {
                    permissions |= row_permissions;
                }
            }
        }
        (any, widen(permissions))
    }

    /// exact returns the slot of the row with identical expressions already folded.
    fn exact(&self, row: &[String; 4]) -> Option<u32> {
        let orders = parse_expression([&row[0], &row[1], &row[2], &row[3]]);
        self.expressions.iter().find(|(o, _)| *o == orders).map(|(_, slot)| *slot)
    }

    /// live_rows returns the rows in slot order, leaving out freed slots.
    fn live_rows(&self) -> Vec<&AccessRow> {
        self.rows.iter().enumerate().filter(|(i, _)| !self.free.contains(&(*i as u32))).map(|(_, r)| r).collect()
    }
}

impl Namespace {
    /// can_create reports whether a user at a host may create a branch with a name in a database, as Dolt's
    /// CanCreate decides: when no expression claims the name, or the longest claiming expressions name the user.
    fn can_create(&self, database: &str, branch: &str, user: &str, host: &str) -> bool {
        let pick = |list: &[MatchExpression], indexes: &[u32]| -> Vec<MatchExpression> {
            indexes.iter().map(|&i| list[i as usize].clone()).collect()
        };
        let matched = expression_matches(&self.databases, database, Collation::Insensitive);
        if matched.is_empty() {
            return true;
        }
        let matched = expression_matches(&pick(&self.branches, &matched), branch, Collation::Insensitive);
        if matched.is_empty() {
            return true;
        }
        let longest = matched.iter().map(|&i| self.values[i as usize].branch.len()).max().unwrap_or(0);
        let longest: Vec<u32> =
            matched.into_iter().filter(|&i| self.values[i as usize].branch.len() == longest).collect();
        let matched = expression_matches(&pick(&self.users, &longest), user, Collation::Binary);
        !expression_matches(&pick(&self.hosts, &matched), host, Collation::Insensitive).is_empty()
    }

    /// position returns the position of the row with identical expressions.
    fn position(&self, row: &[String; 4]) -> Option<usize> {
        self.values
            .iter()
            .position(|v| v.database == row[0] && v.branch == row[1] && v.user == row[2] && v.host == row[3])
    }

    /// insert adds a row of folded expressions.
    fn insert(&mut self, row: [String; 4]) {
        let [database, branch, user, host] = row;
        self.binlog.push(BinlogRow {
            insert: true,
            database: database.clone(),
            branch: branch.clone(),
            user: user.clone(),
            host: host.clone(),
            permissions: 0,
        });
        let index = self.values.len() as u32;
        let expression = |text: &str, collation: Collation| {
            let mut orders = Vec::new();
            if text.len() <= MAX_LENGTH {
                parse_column(&mut orders, text, collation);
            }
            MatchExpression { index, orders }
        };
        self.databases.push(expression(&database, Collation::Insensitive));
        self.branches.push(expression(&branch, Collation::Insensitive));
        self.users.push(expression(&user, Collation::Binary));
        self.hosts.push(expression(&host, Collation::Insensitive));
        self.values.push(NamespaceRow { database, branch, user, host });
    }

    /// delete removes the row with identical expressions, moving the last row into its place.
    fn delete(&mut self, row: [String; 4]) {
        let Some(position) = self.position(&row) else { return };
        let [database, branch, user, host] = row;
        self.binlog.push(BinlogRow { insert: false, database, branch, user, host, permissions: 0 });
        for list in [&mut self.databases, &mut self.branches, &mut self.users, &mut self.hosts] {
            list.swap_remove(position);
            if let Some(moved) = list.get_mut(position) {
                moved.index = position as u32;
            }
        }
        self.values.swap_remove(position);
    }
}

/// write_binlog writes a binlog table.
fn write_binlog(b: &mut Builder, rows: &[BinlogRow]) -> u32 {
    let mut offsets = Vec::with_capacity(rows.len());
    for row in rows {
        let strings = [&row.database, &row.branch, &row.user, &row.host].map(|s| b.create_string(s.as_bytes()));
        b.start_object(6);
        b.add_u64(5, row.permissions, 0);
        for (field, offset) in strings.into_iter().enumerate().rev() {
            b.add_offset(field + 1, offset);
        }
        b.add_bool(0, row.insert, false);
        offsets.push(b.end_object());
    }
    let rows = b.create_vector_of_tables(&offsets);
    b.start_object(2);
    b.add_u32(1, 1, 0);
    b.add_offset(0, rows);
    b.end_object()
}

/// write_expressions writes a vector of match expressions.
fn write_expressions(b: &mut Builder, expressions: &[MatchExpression]) -> u32 {
    let mut offsets = Vec::with_capacity(expressions.len());
    for expression in expressions {
        b.start_vector(4, expression.orders.len(), 4);
        for &order in expression.orders.iter().rev() {
            b.prepend_u32(order as u32);
        }
        let orders = b.end_vector(expression.orders.len());
        b.start_object(2);
        b.add_offset(1, orders);
        b.add_u32(0, expression.index, 0);
        offsets.push(b.end_object());
    }
    b.create_vector_of_tables(&offsets)
}

/// read_binlog reads a binlog table, migrating version 0's read permission bit.
fn read_binlog(table: Option<Table<'_>>) -> Result<Vec<BinlogRow>> {
    let Some(table) = table else { return Ok(Vec::new()) };
    let version = table.u32(1, 0)?;
    let Some(rows) = table.vector(0, 4)? else { return Ok(Vec::new()) };
    let mut out = Vec::with_capacity(rows.len());
    for i in 0..rows.len() {
        let row = rows.table(i)?;
        let text = |field: usize| -> Result<String> {
            Ok(String::from_utf8_lossy(row.string(field)?.unwrap_or_default()).into_owned())
        };
        let mut permissions = row.u64(5, 0)?;
        if version == 0 && permissions & 4 != 0 {
            permissions = (permissions & !4) | 8;
        }
        out.push(BinlogRow {
            insert: row.bool(0, false)?,
            database: text(1)?,
            branch: text(2)?,
            user: text(3)?,
            host: text(4)?,
            permissions,
        });
    }
    Ok(out)
}

/// read_expressions reads a vector of match expressions.
fn read_expressions(table: &Table<'_>, field: usize) -> Result<Vec<MatchExpression>> {
    let Some(vector) = table.vector(field, 4)? else { return Ok(Vec::new()) };
    let mut out = Vec::with_capacity(vector.len());
    for i in 0..vector.len() {
        let expression = vector.table(i)?;
        let orders = match expression.vector(1, 4)? {
            Some(orders) => orders.bytes().as_chunks::<4>().0.iter().map(|c| i32::from_le_bytes(*c)).collect(),
            None => Vec::new(),
        };
        out.push(MatchExpression { index: expression.u32(0, 0)?, orders });
    }
    Ok(out)
}

impl Controller {
    /// load reads the branch control file at a path, or starts with Dolt's default row when the path is None or the
    /// file is missing or empty.
    pub fn load(path: Option<PathBuf>) -> Result<Controller> {
        let mut controller = Controller { path, ..Controller::default() };
        let data = match &controller.path {
            Some(path) => match std::fs::read(path) {
                Ok(data) => data,
                Err(err) if err.kind() == std::io::ErrorKind::NotFound => Vec::new(),
                Err(err) => return Err(PgError::internal(err.to_string())),
            },
            None => Vec::new(),
        };
        if data.is_empty() {
            controller.access.insert_default();
            return Ok(controller);
        }
        let failed = |err: &dyn std::fmt::Display| {
            PgError::internal(format!("failed to deserialize branch control config: {err}"))
        };
        if data.get(8..12) != Some(FILE_ID.as_bytes()) {
            return Err(failed(&"unknown file ID"));
        }
        let root = Table::root(&data, 4).map_err(|e| failed(&e))?;
        let access = root.table(0).map_err(|e| failed(&e))?;
        let binlog = match &access {
            Some(access) => read_binlog(access.table(0).map_err(|e| failed(&e))?)?,
            None => Vec::new(),
        };
        for row in binlog {
            let columns = [row.database, row.branch, row.user, row.host];
            match row.insert {
                true => controller.access.insert(columns, row.permissions),
                false => controller.access.delete(columns),
            }
        }
        if let Some(namespace) = root.table(1).map_err(|e| failed(&e))? {
            let lists: Vec<Vec<MatchExpression>> =
                (1..5).map(|field| read_expressions(&namespace, field)).collect::<Result<_>>()?;
            let values = match namespace.vector(5, 4).map_err(|e| failed(&e))? {
                Some(values) => (0..values.len())
                    .map(|i| {
                        let value = values.table(i)?;
                        let text = |field: usize| -> Result<String> {
                            Ok(String::from_utf8_lossy(value.string(field)?.unwrap_or_default()).into_owned())
                        };
                        Ok(NamespaceRow { database: text(0)?, branch: text(1)?, user: text(2)?, host: text(3)? })
                    })
                    .collect::<Result<Vec<_>>>()?,
                None => Vec::new(),
            };
            if lists.iter().any(|l| l.len() != values.len()) {
                return Err(failed(&"cannot deserialize a namespace table with differing field lengths"));
            }
            let mut lists = lists.into_iter();
            controller.namespace = Namespace {
                databases: lists.next().unwrap_or_default(),
                branches: lists.next().unwrap_or_default(),
                users: lists.next().unwrap_or_default(),
                hosts: lists.next().unwrap_or_default(),
                values,
                binlog: Vec::new(),
            };
        }
        Ok(controller)
    }

    /// save writes the tables to the branch control file, when the server has one.
    pub fn save(&self) -> Result<()> {
        let Some(path) = &self.path else { return Ok(()) };
        let mut b = Builder::new(1024);
        let binlog = write_binlog(&mut b, &self.access.binlog);
        b.start_object(6);
        b.add_offset(0, binlog);
        let access = b.end_object();
        let namespace = &self.namespace;
        let binlog = write_binlog(&mut b, &namespace.binlog);
        let lists = [&namespace.databases, &namespace.branches, &namespace.users, &namespace.hosts]
            .map(|list| write_expressions(&mut b, list));
        let mut values = Vec::with_capacity(namespace.values.len());
        for value in &namespace.values {
            let strings =
                [&value.database, &value.branch, &value.user, &value.host].map(|s| b.create_string(s.as_bytes()));
            b.start_object(4);
            for (field, offset) in strings.into_iter().enumerate().rev() {
                b.add_offset(field, offset);
            }
            values.push(b.end_object());
        }
        let values = b.create_vector_of_tables(&values);
        b.start_object(6);
        b.add_offset(5, values);
        for (field, offset) in lists.into_iter().enumerate().rev() {
            b.add_offset(field + 1, offset);
        }
        b.add_offset(0, binlog);
        let namespace = b.end_object();
        b.start_object(2);
        b.add_offset(1, namespace);
        b.add_offset(0, access);
        let root = b.end_object();
        let mut data = b.finish_message(root, FILE_ID);
        data[..4].fill(0);
        if let Some(dir) = path.parent().filter(|d| !d.as_os_str().is_empty()) {
            std::fs::create_dir_all(dir).map_err(PgError::internal)?;
        }
        std::fs::write(path, data).map_err(PgError::internal)
    }

    /// access_rows returns the rows of the access table.
    pub fn access_rows(&self) -> Vec<AccessRow> {
        self.access.live_rows().into_iter().cloned().collect()
    }

    /// namespace_rows returns the rows of the namespace table.
    pub fn namespace_rows(&self) -> Vec<NamespaceRow> {
        self.namespace.values.clone()
    }

    /// permissions returns what a user at a host may do on a branch of a database.
    pub fn permissions(&self, database: &str, branch: &str, user: &str, host: &str) -> u64 {
        let database = database.split('/').next().unwrap_or_default();
        self.access.matches([database, branch, user, host], None).1
    }
}

/// Identity is who changes the branch control tables: the session's user and host, and whether its role is a
/// superuser, which Go's privilege layer treats as holding every database privilege.
struct Identity {
    user: String,
    host: String,
    superuser: bool,
}

impl Ctx<'_> {
    /// branch_controller returns the server's branch control tables.
    fn branch_controller(&self) -> Result<std::sync::MutexGuard<'_, Controller>> {
        self.session.engine.branch_control().lock().map_err(|_| PgError::internal("a lock was poisoned"))
    }

    /// identity returns who the session is for branch control.
    fn identity(&self) -> Identity {
        Identity { user: self.session.user.clone(), host: self.session.host.clone(), superuser: self.is_superuser() }
    }

    /// check_branch_access fails as Dolt's CheckAccess does unless the session holds permissions on its branch.
    pub fn check_branch_access(&mut self, wanted: u64) -> Result<()> {
        let (database, branch) = (self.session.database.clone(), self.session.branch.clone());
        let identity = self.identity();
        let permissions = self.branch_controller()?.permissions(&database, &branch, &identity.user, &identity.host);
        if permissions & wanted == wanted {
            return Ok(());
        }
        Err(PgError::internal(format!(
            "`{}`@`{}` does not have the correct permissions on branch `{branch}`",
            identity.user, identity.host
        )))
    }

    /// check_branch_write fails as Dolt's CheckAccessForDb does unless the session may write to its branch.
    pub fn check_branch_write(&mut self) -> Result<()> {
        let (database, branch) = (self.session.database.clone(), self.session.branch.clone());
        let identity = self.identity();
        let permissions = self.branch_controller()?.permissions(&database, &branch, &identity.user, &identity.host);
        if permissions & (WRITE | ADMIN) != 0 {
            return Ok(());
        }
        Err(PgError::internal(format!(
            "`{}`@`{}` does not have the correct permissions on branch `{branch}`",
            identity.user, identity.host
        )))
    }

    /// can_create_branch fails as Dolt's CanCreateBranch does when the namespace table reserves a branch name for
    /// others.
    pub fn can_create_branch(&mut self, branch: &str) -> Result<()> {
        let database = self.session.database.split('/').next().unwrap_or_default().to_string();
        let identity = self.identity();
        if self.branch_controller()?.namespace.can_create(&database, branch, &identity.user, &identity.host) {
            return Ok(());
        }
        Err(PgError::internal(format!(
            "`{}`@`{}` cannot create a branch named `{branch}`",
            identity.user, identity.host
        )))
    }

    /// can_delete_branch fails as Dolt's CanDeleteBranch does unless the session may write to a branch.
    pub fn can_delete_branch(&mut self, branch: &str) -> Result<()> {
        let database = self.session.database.clone();
        let identity = self.identity();
        let permissions = self.branch_controller()?.permissions(&database, branch, &identity.user, &identity.host);
        if permissions & (WRITE | ADMIN) != 0 {
            return Ok(());
        }
        Err(PgError::internal(format!("`{}`@`{}` cannot delete the branch `{branch}`", identity.user, identity.host)))
    }

    /// add_branch_admin makes the session an admin of a branch it created, as Dolt's AddAdminForContext does.
    pub fn add_branch_admin(&mut self, branch: &str) -> Result<()> {
        let database = self.session.database.split('/').next().unwrap_or_default().to_string();
        let identity = self.identity();
        let mut controller = self.branch_controller()?;
        if controller.permissions(&database, branch, &identity.user, &identity.host) & ADMIN != 0 {
            return Ok(());
        }
        let row = [database, branch.to_string(), identity.user, identity.host];
        controller.access.insert(normalize(&row[0], &row[1], &row[2], &row[3]), ADMIN);
        controller.save()
    }

    /// branch_control_rows returns the rows of dolt_branch_control.
    pub fn branch_control_rows(&self) -> Result<Vec<Vec<Value>>> {
        let rows = self.branch_controller()?.access_rows();
        Ok(rows
            .into_iter()
            .map(|r| {
                vec![
                    Value::Text(r.database),
                    Value::Text(r.branch),
                    Value::Text(r.user),
                    Value::Text(r.host),
                    Value::Text(permissions_text(r.permissions)),
                ]
            })
            .collect())
    }

    /// branch_namespace_rows returns the rows of dolt_branch_namespace_control.
    pub fn branch_namespace_rows(&self) -> Result<Vec<Vec<Value>>> {
        let rows = self.branch_controller()?.namespace_rows();
        Ok(rows
            .into_iter()
            .map(|r| vec![Value::Text(r.database), Value::Text(r.branch), Value::Text(r.user), Value::Text(r.host)])
            .collect())
    }

    /// branch_control_table reports whether a DML statement's target is dolt_branch_control, or with the flag set
    /// dolt_branch_namespace_control, rather than a user table.
    fn branch_control_table(&mut self, relation: Option<&pg_query::protobuf::RangeVar>) -> Result<Option<bool>> {
        let Some(relation) = relation else { return Ok(None) };
        if self.resolve_table(relation).is_ok() {
            return Ok(None);
        }
        Ok(match crate::dolt::tables::lookup(&relation.schemaname, &relation.relname) {
            Some(crate::dolt::tables::SystemTable::BranchControl) => Some(false),
            Some(crate::dolt::tables::SystemTable::BranchNamespaceControl) => Some(true),
            _ => None,
        })
    }

    /// is_branch_control_dml reports whether an INSERT, UPDATE, or DELETE targets a branch control table.
    pub fn is_branch_control_dml(&mut self, node: &NodeEnum) -> Result<bool> {
        let relation = match node {
            NodeEnum::InsertStmt(s) => s.relation.as_ref(),
            NodeEnum::UpdateStmt(s) => s.relation.as_ref(),
            NodeEnum::DeleteStmt(s) => s.relation.as_ref(),
            _ => None,
        };
        Ok(self.branch_control_table(relation)?.is_some())
    }

    /// branch_control_dml runs an INSERT, UPDATE, or DELETE of a branch control table, or returns None when the
    /// statement targets another table.
    pub fn branch_control_dml(&mut self, node: &NodeEnum) -> Result<Option<crate::Outcome>> {
        use crate::dolt::conflicts::star;
        let column_ref = |name: &str| pg_query::Node {
            node: Some(NodeEnum::ColumnRef(pg_query::protobuf::ColumnRef {
                fields: vec![pg_query::Node {
                    node: Some(NodeEnum::String(pg_query::protobuf::String { sval: name.to_string() })),
                }],
                location: -1,
            })),
        };
        let (namespace, changes, tag) = match node {
            NodeEnum::InsertStmt(insert) => {
                let Some(namespace) = self.branch_control_table(insert.relation.as_ref())? else { return Ok(None) };
                let names: &[&str] = if namespace { &NAMESPACE_COLUMNS } else { &ACCESS_COLUMNS };
                let targets: Vec<usize> = match insert.cols.is_empty() {
                    true => (0..names.len()).collect(),
                    false => insert
                        .cols
                        .iter()
                        .map(|c| match c.node.as_ref() {
                            Some(NodeEnum::ResTarget(t)) => names.iter().position(|n| *n == t.name).ok_or_else(|| {
                                PgError::new(code::UNDEFINED_COLUMN, format!("column \"{}\" does not exist", t.name))
                            }),
                            _ => Err(PgError::internal("an INSERT column")),
                        })
                        .collect::<Result<_>>()?,
                };
                let Some(NodeEnum::SelectStmt(select)) = insert.select_stmt.as_ref().and_then(|s| s.node.as_ref())
                else {
                    return Err(PgError::unsupported("INSERT DEFAULT VALUES into a branch control table"));
                };
                let query = crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
                let mut changes = Vec::new();
                for source in query.plan.run(self)? {
                    let mut row = vec![Value::Text(String::new()); names.len()];
                    for (value, &target) in source.into_iter().zip(&targets) {
                        row[target] = value;
                    }
                    changes.push((None, Some(row)));
                }
                let tag = format!("INSERT 0 {}", changes.len());
                (namespace, changes, tag)
            }
            NodeEnum::UpdateStmt(update) => {
                let Some(namespace) = self.branch_control_table(update.relation.as_ref())? else { return Ok(None) };
                let relation = update.relation.as_ref().ok_or_else(|| PgError::internal("UPDATE without a table"))?;
                let names: &[&str] = if namespace { &NAMESPACE_COLUMNS } else { &ACCESS_COLUMNS };
                let mut targets: Vec<pg_query::Node> = names.iter().map(|n| column_ref(n)).collect();
                for item in &update.target_list {
                    let Some(NodeEnum::ResTarget(target)) = item.node.as_ref() else { continue };
                    let position = names.iter().position(|n| *n == target.name).ok_or_else(|| {
                        PgError::new(code::UNDEFINED_COLUMN, format!("column \"{}\" does not exist", target.name))
                    })?;
                    targets[position] = target.val.as_deref().cloned().unwrap_or_else(|| column_ref(names[position]));
                }
                let old = self.select_rows(relation, vec![star()], update.where_clause.clone())?;
                let targets = targets
                    .into_iter()
                    .map(|val| pg_query::Node {
                        node: Some(NodeEnum::ResTarget(Box::new(pg_query::protobuf::ResTarget {
                            val: Some(Box::new(val)),
                            ..Default::default()
                        }))),
                    })
                    .collect();
                let new = self.select_rows(relation, targets, update.where_clause.clone())?;
                let changes: Vec<_> = old.into_iter().zip(new).map(|(o, n)| (Some(o), Some(n))).collect();
                let tag = format!("UPDATE {}", changes.len());
                (namespace, changes, tag)
            }
            NodeEnum::DeleteStmt(delete) => {
                let Some(namespace) = self.branch_control_table(delete.relation.as_ref())? else { return Ok(None) };
                let relation = delete.relation.as_ref().ok_or_else(|| PgError::internal("DELETE without a table"))?;
                let old = self.select_rows(relation, vec![star()], delete.where_clause.clone())?;
                let changes: Vec<_> = old.into_iter().map(|o| (Some(o), None)).collect();
                let tag = format!("DELETE {}", changes.len());
                (namespace, changes, tag)
            }
            _ => return Ok(None),
        };
        self.change_branch_control(namespace, &changes)?;
        Ok(Some(crate::Outcome::command(tag)))
    }

    /// change_branch_control applies inserted, updated, and deleted rows to dolt_branch_control or, when `namespace`
    /// is set, dolt_branch_namespace_control, checking each as Dolt's system tables do, and saves the file.
    pub fn change_branch_control(&mut self, namespace: bool, changes: &[crate::foreign::Change]) -> Result<()> {
        let identity = self.identity();
        let mut controller = self.branch_controller()?;
        let text = |row: &[Value], i: usize| row.get(i).and_then(Value::output).unwrap_or_default();
        let columns = |row: &[Value]| normalize(&text(row, 0), &text(row, 1), &text(row, 2), &text(row, 3));
        let is_admin = |controller: &Controller, database: &str, branch: &str| {
            identity.superuser
                || controller.access.matches([database, branch, &identity.user, &identity.host], None).1 & ADMIN != 0
        };
        let too_long = |c: &[String; 4]| {
            PgError::internal(format!(
                "expressions are too long [{}, {}, {}, {}]",
                quoted(&c[0]),
                quoted(&c[1]),
                quoted(&c[2]),
                quoted(&c[3])
            ))
        };
        let duplicate = |items: Vec<String>| {
            let shown: Vec<String> = items.iter().map(|i| quoted(i)).collect();
            PgError::new(code::UNIQUE_VIOLATION, format!("duplicate primary key given: [{}]", shown.join(", ")))
        };
        let row_text = |c: &[String; 4]| c.iter().map(|i| quoted(i)).collect::<Vec<_>>().join(", ");
        for (old, new) in changes {
            let old = old.as_deref().map(columns);
            let new_columns = new.as_deref().map(columns);
            if let Some(c) = new_columns.as_ref().filter(|c| c.iter().any(|v| v.len() > MAX_LENGTH)) {
                return Err(too_long(c));
            }
            match (old, new_columns) {
                (None, Some(c)) if namespace => {
                    if !is_admin(&controller, &c[0], &c[1]) {
                        return Err(PgError::internal(format!(
                            "`{}`@`{}` cannot add the row [{}]",
                            identity.user,
                            identity.host,
                            row_text(&c)
                        )));
                    }
                    if controller.namespace.position(&c).is_some() {
                        return Err(duplicate(c.to_vec()));
                    }
                    controller.namespace.insert(c);
                }
                (None, Some(c)) => {
                    let permissions = parse_permissions(&text(new.as_deref().unwrap_or_default(), 4))?;
                    let shown = permissions_text(permissions);
                    if !is_admin(&controller, &c[0], &c[1]) {
                        return Err(PgError::internal(format!(
                            "`{}`@`{}` cannot add the row [{}, {}]",
                            identity.user,
                            identity.host,
                            row_text(&c),
                            quoted(&shown)
                        )));
                    }
                    let (found, existing) = controller.access.matches([&c[0], &c[1], &c[2], &c[3]], None);
                    if (found && consolidate(permissions) == consolidate(existing))
                        || controller.access.exact(&c).is_some()
                    {
                        let mut items = c.to_vec();
                        items.push(shown);
                        return Err(duplicate(items));
                    }
                    controller.access.insert(c, permissions);
                }
                (Some(o), Some(c)) => {
                    let changed = o != c;
                    if namespace {
                        if changed && controller.namespace.position(&c).is_some() {
                            return Err(duplicate(c.to_vec()));
                        }
                    } else {
                        let permissions = parse_permissions(&text(new.as_deref().unwrap_or_default(), 4))?;
                        let ignored = if changed { None } else { controller.access.exact(&c) };
                        let (found, existing) = controller.access.matches([&c[0], &c[1], &c[2], &c[3]], ignored);
                        let exact = changed && controller.access.exact(&c).is_some();
                        if (found && consolidate(permissions) == consolidate(existing)) || exact {
                            let mut items = c.to_vec();
                            items.push(permissions_text(existing));
                            return Err(duplicate(items));
                        }
                    }
                    if !is_admin(&controller, &o[0], &o[1]) {
                        return Err(PgError::internal(format!(
                            "`{}`@`{}` cannot update the row [{}]",
                            identity.user,
                            identity.host,
                            row_text(&o)
                        )));
                    }
                    if !is_admin(&controller, &c[0], &c[1]) {
                        return Err(PgError::internal(format!(
                            "`{}`@`{}` cannot update the row [{}] to the new branch expression [{}, {}]",
                            identity.user,
                            identity.host,
                            row_text(&o),
                            quoted(&c[0]),
                            quoted(&c[1])
                        )));
                    }
                    if namespace {
                        controller.namespace.delete(o);
                        controller.namespace.insert(c);
                    } else {
                        let permissions = parse_permissions(&text(new.as_deref().unwrap_or_default(), 4))?;
                        controller.access.delete(o);
                        controller.access.insert(c, permissions);
                    }
                }
                (Some(o), None) => {
                    if !is_admin(&controller, &o[0], &o[1]) {
                        return Err(PgError::internal(format!(
                            "`{}`@`{}` cannot delete the row [{}]",
                            identity.user,
                            identity.host,
                            row_text(&o)
                        )));
                    }
                    match namespace {
                        true => controller.namespace.delete(o),
                        false => controller.access.delete(o),
                    }
                }
                (None, None) => {}
            }
        }
        controller.save()
    }
}
