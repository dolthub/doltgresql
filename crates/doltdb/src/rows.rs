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

use prolly::{Node, Tuple, read_blob, walk_leaves};
use serial::{Index, Message, TableSchema};
use sha2::{Digest, Sha256};
use store::{ChunkReader, Hash, Result};

use crate::graph::hex;

/// Field encodings that hold the address of a blob tree.
const ADDRESS_ENCODINGS: [u8; 5] = [21, 23, 24, 26, 27];
/// Field encodings that hold an adaptive value: inline after a 0 byte, or a size and an address.
const ADAPTIVE_ENCODINGS: [u8; 5] = [135, 136, 137, 138, 139];

/// resolve returns a field's value, reading out-of-band values from their blob trees.
fn resolve(reader: &dyn ChunkReader, encoding: u8, field: &[u8]) -> Result<Vec<u8>> {
    if ADDRESS_ENCODINGS.contains(&encoding) {
        return read_blob(reader, &serial::hash(field)?);
    }
    if ADAPTIVE_ENCODINGS.contains(&encoding) {
        if field[0] == 0 {
            return Ok(field[1..].to_vec());
        }
        let Some(address) = field.len().checked_sub(Hash::LEN) else {
            return Err(store::Error::Corrupt("adaptive value is too short for an address".to_string()));
        };
        return read_blob(reader, &serial::hash(&field[address..])?);
    }
    Ok(field.to_vec())
}

/// render renders a tuple's fields with their encodings, with NULL as "-" and other values in hex.
fn render(reader: &dyn ChunkReader, tuple: Tuple<'_>, encodings: &[u8]) -> Result<String> {
    let mut fields = Vec::with_capacity(encodings.len());
    for (i, &encoding) in encodings.iter().enumerate() {
        fields.push(match tuple.field(i)? {
            Some(field) => hex(&resolve(reader, encoding, field)?),
            None => "-".to_string(),
        });
    }
    Ok(fields.join(","))
}

/// digest returns the item count and the SHA-256 of the rendered items of the tree at the root node.
fn digest(reader: &dyn ChunkReader, root: &Node, keys: &[u8], values: &[u8]) -> Result<(usize, String)> {
    let mut hasher = Sha256::new();
    let mut count = 0;
    walk_leaves(reader, root, &mut |key, value| {
        let line = format!("{} {}\n", render(reader, Tuple(key), keys)?, render(reader, Tuple(value), values)?);
        if std::env::var_os("DOLTDB_ROWS").is_some() {
            eprint!("{line}");
        }
        hasher.update(line.as_bytes());
        count += 1;
        Ok(())
    })?;
    Ok((count, hex(&hasher.finalize())))
}

/// encodings returns the encodings of the columns at the indexes.
fn encodings(columns: &[serial::Column<'_>], indexes: &[u16]) -> Vec<u8> {
    indexes.iter().map(|&i| columns.get(i as usize).map_or(0, |c| c.encoding)).collect()
}

/// describe_index renders an index of a schema.
fn describe_index(schema: Hash, kind: &str, index: &Index<'_>) -> String {
    let list = |values: &[u16]| values.iter().map(u16::to_string).collect::<Vec<_>>().join(",");
    let bools = |values: &[bool]| values.iter().map(bool::to_string).collect::<Vec<_>>().join(",");
    format!(
        "index {schema} {kind} name={} idx={} keys={} values={} pk={} unique={} system={} prefix={} predicate={} \
         desc={} nullslast={} opclasses={}",
        hex(index.name),
        list(&index.index_columns),
        list(&index.key_columns),
        list(&index.value_columns),
        index.primary_key,
        index.unique_key,
        index.system_defined,
        list(&index.prefix_lengths),
        hex(index.predicate),
        bools(&index.descending),
        bools(&index.nulls_last),
        index.op_classes.iter().map(|c| hex(c)).collect::<Vec<_>>().join(","),
    )
}

/// describe_table renders a table's schema and the digests of its primary and secondary index rows.
pub(crate) fn describe_table(
    reader: &dyn ChunkReader,
    address: Hash,
    schema_address: Hash,
    primary: &[u8],
    secondary: &[(Vec<u8>, Hash)],
    lines: &mut Vec<String>,
) -> Result<()> {
    let chunk = reader.require(&schema_address)?;
    let schema = TableSchema::new(Message(&chunk.data))?;
    let columns = schema.columns()?;
    lines.push(format!(
        "schema {schema_address} columns={} collation={} comment={} rowsize={}",
        columns.len(),
        schema.collation()?,
        hex(schema.comment()?),
        schema.target_row_size()?
    ));
    for (i, c) in columns.iter().enumerate() {
        lines.push(format!(
            "column {schema_address} {i} name={} type={} default={} comment={} order={} tag={} enc={} pk={} null={} \
             ai={} hidden={} generated={} virtual={} onupdate={} adaptive={} hiddensys={} breaking={}",
            hex(c.name),
            hex(c.sql_type),
            hex(c.default_value),
            hex(c.comment),
            c.display_order,
            c.tag,
            c.encoding,
            c.primary_key,
            c.nullable,
            c.auto_increment,
            c.hidden,
            c.generated,
            c.is_virtual,
            hex(c.on_update_value),
            c.uses_adaptive_encoding,
            c.hidden_system,
            c.adaptive_encoding_breaking_change,
        ));
    }
    let clustered = schema.clustered_index()?;
    lines.push(describe_index(schema_address, "clustered", &clustered));
    let indexes = schema.secondary_indexes()?;
    for index in &indexes {
        lines.push(describe_index(schema_address, "secondary", index));
    }
    for check in schema.checks()? {
        lines.push(format!(
            "check {schema_address} name={} expr={} enforced={} notvalid={}",
            hex(check.name),
            hex(check.expression),
            check.enforced,
            check.is_not_valid
        ));
    }
    // A keyless table's key is a content hash and its value starts with the row's cardinality.
    let (key_encodings, value_encodings) = if clustered.key_columns.is_empty() {
        let mut values = vec![10];
        values.extend(encodings(&columns, &clustered.value_columns));
        (vec![14], values)
    } else {
        (encodings(&columns, &clustered.key_columns), encodings(&columns, &clustered.value_columns))
    };
    let (count, sum) = digest(reader, &Node::decode(primary.to_vec())?, &key_encodings, &value_encodings)?;
    lines.push(format!("rows {address} primary count={count} digest={sum}"));
    for (name, root) in secondary {
        let Some(index) = indexes.iter().find(|index| index.name == name.as_slice()) else {
            lines.push(format!("rows {address} index {} missing", hex(name)));
            continue;
        };
        let keys = encodings(&columns, &index.key_columns);
        let (count, sum) = digest(reader, &Node::load(reader, root)?, &keys, &[])?;
        lines.push(format!("rows {address} index {} count={count} digest={sum}", hex(name)));
    }
    Ok(())
}
