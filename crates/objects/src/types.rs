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

use store::{Error, Result};

use crate::codec::{Reader, Writer};
use crate::show::quote;

/// TypeCheck is a domain check constraint of a type.
#[derive(Clone, Debug, PartialEq)]
pub struct TypeCheck {
    pub name: Vec<u8>,
    pub expression: Vec<u8>,
}

/// EnumLabel is a label of an enum type.
#[derive(Clone, Debug, PartialEq)]
pub struct EnumLabel {
    pub id: Vec<u8>,
    pub sort_order: f32,
}

/// CompositeAttribute is an attribute of a composite type.
#[derive(Clone, Debug, PartialEq)]
pub struct CompositeAttribute {
    pub rel_id: Vec<u8>,
    pub name: Vec<u8>,
    pub type_id: Vec<u8>,
    pub num: i16,
    pub collation: Vec<u8>,
}

/// SerializedType is a Doltgres type as serialized, with its function and related type references as internal IDs.
#[derive(Clone, Debug, PartialEq)]
pub struct SerializedType {
    pub version: u8,
    pub id: Vec<u8>,
    pub typ_length: i16,
    pub passed_by_val: bool,
    pub typ_type: Vec<u8>,
    pub typ_category: Vec<u8>,
    pub is_preferred: bool,
    pub is_defined: bool,
    pub delimiter: Vec<u8>,
    pub rel_id: Vec<u8>,
    pub subscript_func: Vec<u8>,
    pub elem: Vec<u8>,
    pub array: Vec<u8>,
    pub input_func: Vec<u8>,
    pub output_func: Vec<u8>,
    pub receive_func: Vec<u8>,
    pub send_func: Vec<u8>,
    pub mod_in_func: Vec<u8>,
    pub mod_out_func: Vec<u8>,
    pub analyze_func: Vec<u8>,
    pub align: Vec<u8>,
    pub storage: Vec<u8>,
    pub not_null: bool,
    pub base_type: Vec<u8>,
    pub typ_mod: i32,
    pub n_dims: i32,
    pub typ_collation: Vec<u8>,
    pub default_bin: Vec<u8>,
    pub default: Vec<u8>,
    pub acl: Vec<Vec<u8>>,
    pub checks: Vec<TypeCheck>,
    pub att_typ_mod: i32,
    pub compare_func: Vec<u8>,
    /// The enum labels, sorted by ID as they are serialized.
    pub enum_labels: Vec<EnumLabel>,
    pub composite_attrs: Vec<CompositeAttribute>,
    pub internal_name: Vec<u8>,
}

impl SerializedType {
    /// deserialize decodes a serialized type.
    pub fn deserialize(data: &[u8]) -> Result<SerializedType> {
        if data.is_empty() {
            return Err(Error::Corrupt("deserializing empty type data".to_string()));
        }
        let mut r = Reader::new(data);
        let version = r.variable_uint()?;
        if version > 1 {
            return Err(Error::Corrupt(format!(
                "version {version} of types is not supported, please upgrade the server"
            )));
        }
        let mut t = SerializedType {
            version: version as u8,
            id: r.string()?,
            typ_length: r.int16()?,
            passed_by_val: r.bool()?,
            typ_type: r.string()?,
            typ_category: r.string()?,
            is_preferred: r.bool()?,
            is_defined: r.bool()?,
            delimiter: r.string()?,
            rel_id: r.string()?,
            subscript_func: r.string()?,
            elem: r.string()?,
            array: r.string()?,
            input_func: r.string()?,
            output_func: r.string()?,
            receive_func: r.string()?,
            send_func: r.string()?,
            mod_in_func: r.string()?,
            mod_out_func: r.string()?,
            analyze_func: r.string()?,
            align: r.string()?,
            storage: r.string()?,
            not_null: r.bool()?,
            base_type: r.string()?,
            typ_mod: r.int32()?,
            n_dims: r.int32()?,
            typ_collation: r.string()?,
            default_bin: r.string()?,
            default: r.string()?,
            acl: Vec::new(),
            checks: Vec::new(),
            att_typ_mod: 0,
            compare_func: Vec::new(),
            enum_labels: Vec::new(),
            composite_attrs: Vec::new(),
            internal_name: Vec::new(),
        };
        for _ in 0..r.variable_uint()? {
            t.acl.push(r.string()?);
        }
        for _ in 0..r.variable_uint()? {
            t.checks.push(TypeCheck { name: r.string()?, expression: r.string()? });
        }
        t.att_typ_mod = r.int32()?;
        t.compare_func = r.string()?;
        for _ in 0..r.variable_uint()? {
            t.enum_labels.push(EnumLabel { id: r.string()?, sort_order: r.float32()? });
        }
        for _ in 0..r.variable_uint()? {
            t.composite_attrs.push(CompositeAttribute {
                rel_id: r.string()?,
                name: r.string()?,
                type_id: r.string()?,
                num: r.int16()?,
                collation: r.string()?,
            });
        }
        t.internal_name = r.string()?;
        if !r.is_empty() {
            return Err(Error::Corrupt("extra data found while deserializing type".to_string()));
        }
        Ok(t)
    }

    /// serialize encodes the type.
    pub fn serialize(&self) -> Vec<u8> {
        let mut w = Writer::new();
        w.variable_uint(self.version as u64);
        w.string(&self.id);
        w.int16(self.typ_length);
        w.bool(self.passed_by_val);
        w.string(&self.typ_type);
        w.string(&self.typ_category);
        w.bool(self.is_preferred);
        w.bool(self.is_defined);
        w.string(&self.delimiter);
        w.string(&self.rel_id);
        for func in [&self.subscript_func, &self.elem, &self.array, &self.input_func, &self.output_func] {
            w.string(func);
        }
        for func in [&self.receive_func, &self.send_func, &self.mod_in_func, &self.mod_out_func, &self.analyze_func] {
            w.string(func);
        }
        w.string(&self.align);
        w.string(&self.storage);
        w.bool(self.not_null);
        w.string(&self.base_type);
        w.int32(self.typ_mod);
        w.int32(self.n_dims);
        w.string(&self.typ_collation);
        w.string(&self.default_bin);
        w.string(&self.default);
        w.string_slice(&self.acl);
        w.variable_uint(self.checks.len() as u64);
        for check in &self.checks {
            w.string(&check.name);
            w.string(&check.expression);
        }
        w.int32(self.att_typ_mod);
        w.string(&self.compare_func);
        w.variable_uint(self.enum_labels.len() as u64);
        let mut labels: Vec<&EnumLabel> = self.enum_labels.iter().collect();
        labels.sort_by(|a, b| a.id.cmp(&b.id));
        for label in labels {
            w.string(&label.id);
            w.float32(label.sort_order);
        }
        w.variable_uint(self.composite_attrs.len() as u64);
        for attr in &self.composite_attrs {
            w.string(&attr.rel_id);
            w.string(&attr.name);
            w.string(&attr.type_id);
            w.int16(attr.num);
            w.string(&attr.collation);
        }
        w.string(&self.internal_name);
        w.data()
    }

    /// show renders the type as the Go graph oracle does.
    pub fn show(&self) -> String {
        let mut labels: Vec<String> =
            self.enum_labels.iter().map(|l| format!("{}:{}", quote(&l.id), crate::go_float32(l.sort_order))).collect();
        labels.sort();
        let checks: Vec<String> =
            self.checks.iter().map(|c| format!("{}:{}", quote(&c.name), quote(&c.expression))).collect();
        let attrs: Vec<String> = self
            .composite_attrs
            .iter()
            .map(|a| {
                format!(
                    "{{{} {} {} {} {}}}",
                    quote(&a.rel_id),
                    quote(&a.name),
                    quote(&a.type_id),
                    a.num,
                    quote(&a.collation)
                )
            })
            .collect();
        let serialized: String = self.serialize().iter().map(|b| format!("{b:02x}")).collect();
        format!(
            "type{{ID:{} TypType:{} TypCategory:{} TypLength:{} PassedByVal:{} IsPreferred:{} IsDefined:{} \
             Delimiter:{} RelID:{} Elem:{} Array:{} Align:{} Storage:{} NotNull:{} BaseTypeType:{} TypMod:{} \
             NDims:{} TypCollation:{} DefaulBin:{} Default:{} Acl:[{}] Checks:[{}] EnumLabels:[{}] \
             CompositeAttrs:[{}] InternalName:{} Serialized:{serialized}}}",
            quote(&self.id),
            quote(&self.typ_type),
            quote(&self.typ_category),
            self.typ_length,
            self.passed_by_val,
            self.is_preferred,
            self.is_defined,
            quote(&self.delimiter),
            quote(&self.rel_id),
            quote(&self.elem),
            quote(&self.array),
            quote(&self.align),
            quote(&self.storage),
            self.not_null,
            quote(&self.base_type),
            self.typ_mod,
            self.n_dims,
            quote(&self.typ_collation),
            quote(&self.default_bin),
            quote(&self.default),
            self.acl.iter().map(|a| quote(a)).collect::<Vec<_>>().join(" "),
            checks.join(" "),
            labels.join(" "),
            attrs.join(" "),
            quote(&self.internal_name),
        )
    }
}
