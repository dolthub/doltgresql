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

//! XML functions.

use std::collections::HashMap;

use super::Function;
use crate::array::Array;
use crate::error::{PgError, Result, code};
use crate::oid::{BOOL, TEXT, TEXT_ARRAY, XML, XML_ARRAY};
use crate::query::Ctx;
use crate::types::Value;
use crate::xml::{self as x, Query};

/// f declares a strict XML function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the XML functions.
pub const FUNCTIONS: &[Function] = &[
    f("xpath", &[TEXT, XML], XML_ARRAY, xpath),
    f("xpath", &[TEXT, XML, TEXT_ARRAY], XML_ARRAY, xpath),
    f("xpath_exists", &[TEXT, XML], BOOL, xpath_exists),
    f("xpath_exists", &[TEXT, XML, TEXT_ARRAY], BOOL, xpath_exists),
    f("xmlexists", &[TEXT, XML], BOOL, xpath_exists),
    f("xml_is_well_formed", &[TEXT], BOOL, xml_is_well_formed),
    f("xml_is_well_formed_document", &[TEXT], BOOL, xml_is_well_formed_document),
    f("xml_is_well_formed_content", &[TEXT], BOOL, xml_is_well_formed_content),
    f("xmlcomment", &[TEXT], XML, xmlcomment),
    f("xml", &[TEXT], XML, xml),
];

/// text returns the text of a string argument.
fn text(value: &Value) -> &str {
    match value {
        Value::Text(s) | Value::Xml(s) => s,
        _ => "",
    }
}

/// namespaces reads the prefix-to-URI pairs of an xpath namespace array, which must be two-dimensional with pairs.
pub fn namespaces(value: Option<&Value>) -> Result<HashMap<String, String>> {
    let Some(Value::Array(array)) = value else { return Ok(HashMap::new()) };
    if array.values.is_empty() {
        return Ok(HashMap::new());
    }
    if array.dims.len() != 2 || array.dims[1].0 != 2 {
        return Err(PgError {
            detail: Some("The array must be two-dimensional with length of the second axis equal to 2.".into()),
            ..PgError::new(code::DATA_EXCEPTION, "invalid array for XML namespace mapping")
        });
    }
    let mut out = HashMap::new();
    for pair in array.values.chunks(2) {
        let (Value::Text(prefix), Value::Text(uri)) = (&pair[0], &pair[1]) else {
            return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "neither namespace name nor URI may be null"));
        };
        out.insert(prefix.clone(), uri.clone());
    }
    Ok(out)
}

/// query compiles the expression and namespaces of an xpath call.
fn query(args: &[Value]) -> Result<Query> {
    Ok(Query { expr: x::compile(text(&args[0]))?, namespaces: namespaces(args.get(2))? })
}

/// xpath evaluates an XPath expression over an xml value, returning the xml values it finds.
fn xpath(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let found = x::xpath(&query(args)?, text(&args[1]))?;
    Ok(Value::Array(Box::new(Array::one_dimensional(XML, found.into_iter().map(Value::Xml).collect()))))
}

/// xpath_exists reports whether an XPath expression over an xml value finds anything.
fn xpath_exists(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(x::xpath_exists(&query(args)?, text(&args[1]))?))
}

/// xml_is_well_formed reports whether text is well-formed XML as the xmloption setting reads it.
fn xml_is_well_formed(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(x::is_well_formed(text(&args[0]), x::document_option())))
}

/// xml_is_well_formed_document reports whether text is a well-formed XML document.
fn xml_is_well_formed_document(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(x::is_well_formed(text(&args[0]), true)))
}

/// xml_is_well_formed_content reports whether text is well-formed XML content.
fn xml_is_well_formed_content(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(x::is_well_formed(text(&args[0]), false)))
}

/// xmlcomment returns an XML comment of text, which must not hold `--` or end with `-`.
fn xmlcomment(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let comment = text(&args[0]);
    if comment.contains("--") || comment.ends_with('-') {
        return Err(PgError::new(code::INVALID_XML_COMMENT, "invalid XML comment"));
    }
    Ok(Value::Xml(format!("<!--{comment}-->")))
}

/// xml parses text as XML content, or as a document when the xmloption setting says so.
fn xml(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::cast::input(text(&args[0]), XML)
}
