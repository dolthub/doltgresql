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

use std::collections::HashMap;

use yaml_rust2::Yaml;
use yaml_rust2::parser::{Event, MarkedEventReceiver, Parser};
use yaml_rust2::scanner::{Marker, TScalarStyle};

/// Node is a YAML value that keeps every scalar's text, since Go's yaml.v3 decodes any scalar into a string field as
/// it was written.
#[derive(Clone, Debug, PartialEq)]
pub enum Node {
    Scalar { text: String, plain: bool },
    Sequence(Vec<Node>),
    Mapping(Vec<(String, Node)>),
}

impl Node {
    /// is_null reports whether the node is a plain null, which leaves a Go field unset.
    pub fn is_null(&self) -> bool {
        matches!(self, Node::Scalar { text, plain: true } if matches!(text.as_str(), "" | "~" | "null" | "Null" | "NULL"))
    }

    /// get returns the value of a mapping key.
    pub fn get(&self, key: &str) -> Option<&Node> {
        match self {
            Node::Mapping(entries) => entries.iter().find(|(k, _)| k == key).map(|(_, v)| v),
            _ => None,
        }
    }

    /// string decodes the node into a Go string field, where null is empty.
    pub fn string(&self) -> Result<String, String> {
        match self {
            _ if self.is_null() => Ok(String::new()),
            Node::Scalar { text, .. } => Ok(text.clone()),
            other => Err(format!("cannot decode {other:?} into a string")),
        }
    }

    /// strings decodes the node into a Go []string field.
    pub fn strings(&self) -> Result<Vec<String>, String> {
        match self {
            _ if self.is_null() => Ok(Vec::new()),
            Node::Sequence(items) => items.iter().map(Node::string).collect(),
            other => Err(format!("cannot decode {other:?} into a list of strings")),
        }
    }

    /// int decodes the node into a Go int field.
    pub fn int(&self) -> Result<i64, String> {
        match self {
            _ if self.is_null() => Ok(0),
            Node::Scalar { text, plain: true } => {
                text.parse().map_err(|_| format!("cannot decode {text:?} into an int"))
            }
            other => Err(format!("cannot decode {other:?} into an int")),
        }
    }

    /// sequence returns the items of a sequence, where null is empty.
    pub fn sequence(&self) -> Result<&[Node], String> {
        match self {
            _ if self.is_null() => Ok(&[]),
            Node::Sequence(items) => Ok(items),
            other => Err(format!("expected a list, got {other:?}")),
        }
    }

    /// mapping returns the entries of a mapping, where null is empty.
    pub fn mapping(&self) -> Result<&[(String, Node)], String> {
        match self {
            _ if self.is_null() => Ok(&[]),
            Node::Mapping(entries) => Ok(entries),
            other => Err(format!("expected a mapping, got {other:?}")),
        }
    }

    /// to_yaml converts the node to a typed value, resolving plain scalars the way YAML does.
    pub fn to_yaml(&self) -> Yaml {
        match self {
            Node::Scalar { text, plain: true } => Yaml::from_str(text),
            Node::Scalar { text, plain: false } => Yaml::String(text.clone()),
            Node::Sequence(items) => Yaml::Array(items.iter().map(Node::to_yaml).collect()),
            Node::Mapping(entries) => {
                Yaml::Hash(entries.iter().map(|(k, v)| (Yaml::String(k.clone()), v.to_yaml())).collect())
            }
        }
    }
}

/// Loader builds nodes from parser events.
#[derive(Default)]
struct Loader {
    stack: Vec<(Node, Option<String>, usize)>,
    anchors: HashMap<usize, Node>,
    documents: Vec<Node>,
}

impl Loader {
    fn insert(&mut self, node: Node, anchor: usize) {
        if anchor != 0 {
            self.anchors.insert(anchor, node.clone());
        }
        match self.stack.last_mut() {
            None => self.documents.push(node),
            Some((Node::Sequence(items), _, _)) => items.push(node),
            Some((Node::Mapping(entries), key, _)) => match key.take() {
                None => *key = Some(scalar_key(&node)),
                Some(k) => entries.push((k, node)),
            },
            Some(_) => unreachable!(),
        }
    }
}

/// scalar_key returns the text of a mapping key.
fn scalar_key(node: &Node) -> String {
    match node {
        Node::Scalar { text, .. } => text.clone(),
        other => format!("{other:?}"),
    }
}

impl MarkedEventReceiver for Loader {
    fn on_event(&mut self, event: Event, _: Marker) {
        match event {
            Event::Scalar(text, style, anchor, _) => {
                self.insert(Node::Scalar { text, plain: style == TScalarStyle::Plain }, anchor)
            }
            Event::SequenceStart(anchor, _) => self.stack.push((Node::Sequence(Vec::new()), None, anchor)),
            Event::MappingStart(anchor, _) => self.stack.push((Node::Mapping(Vec::new()), None, anchor)),
            Event::SequenceEnd | Event::MappingEnd => {
                let (node, _, anchor) = self.stack.pop().unwrap();
                self.insert(node, anchor);
            }
            Event::Alias(anchor) => {
                let node =
                    self.anchors.get(&anchor).cloned().unwrap_or(Node::Scalar { text: String::new(), plain: true });
                self.insert(node, 0);
            }
            _ => {}
        }
    }
}

/// parse parses the first document of a YAML text.
pub fn parse(text: &str) -> Result<Node, String> {
    let mut loader = Loader::default();
    Parser::new_from_str(text).load(&mut loader, false).map_err(|e| e.to_string())?;
    Ok(loader.documents.into_iter().next().unwrap_or(Node::Scalar { text: String::new(), plain: true }))
}

/// emit writes a value as a YAML document.
pub fn emit(value: &Yaml) -> Result<String, String> {
    let mut out = String::new();
    yaml_rust2::YamlEmitter::new(&mut out).dump(value).map_err(|e| e.to_string())?;
    Ok(out)
}
