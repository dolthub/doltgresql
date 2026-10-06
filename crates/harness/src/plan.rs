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

//! Plan facts: what an EXPLAIN shows about how a query runs, such as which index answers it, independent of how a
//! server formats its plans. Tests assert facts instead of plan text, since plan text is an implementation detail.

/// PlanFact is something a plan must show.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum PlanFact {
    /// The table is read through an index on these columns, constrained to these ranges when they are not empty.
    IndexScan {
        /// The table.
        table: &'static str,
        /// The index columns, in order.
        columns: &'static [&'static str],
        /// The ranges in go-mysql-server's notation, or empty to accept any ranges.
        ranges: &'static str,
    },
    /// The table is read through an index in descending order.
    ReverseScan {
        /// The table.
        table: &'static str,
    },
    /// The table is read in full, without an index.
    FullScan {
        /// The table.
        table: &'static str,
    },
    /// The plan joins two inputs with the given algorithm, whose first tables are these.
    Join {
        /// The algorithm, such as LookupJoin, HashJoin, MergeJoin, or InnerJoin for a nested loop.
        kind: &'static str,
        /// The first table of the left input.
        left: &'static str,
        /// The first table of the right input.
        right: &'static str,
    },
    /// The plan sorts.
    Sort,
    /// The plan does not sort, since an index already provides the order.
    NoSort,
}

/// Node is a parsed plan node.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Node {
    /// The node's label, such as IndexedTableAccess(t).
    pub label: String,
    /// The node's properties, such as ("index", "[t.v1]").
    pub properties: Vec<(String, String)>,
    /// The child nodes.
    pub children: Vec<Node>,
}

/// NODE_PREFIXES are the labels of go-mysql-server plan nodes, which distinguish nodes from expression lines.
const NODE_PREFIXES: &[&str] = &[
    "Project",
    "Filter",
    "Sort",
    "TopN",
    "Table",
    "IndexedTableAccess",
    "TableAlias",
    "LookupJoin",
    "HashJoin",
    "MergeJoin",
    "InnerJoin",
    "CrossJoin",
    "LeftOuterJoin",
    "LeftOuterLookupJoin",
    "LeftOuterHashJoin",
    "LeftOuterMergeJoin",
    "RightOuterJoin",
    "SemiJoin",
    "SemiLookupJoin",
    "AntiJoin",
    "AntiLookupJoin",
    "HashLookup",
    "Limit",
    "Offset",
    "GroupBy",
    "Distinct",
    "SubqueryAlias",
    "Window",
    "Union",
    "Values",
    "EmptyTable",
    "Having",
    "RecursiveCTE",
    "Concat",
];

/// parse_gms parses go-mysql-server's plan tree text, given one line per row, into its root nodes.
pub fn parse_gms(lines: &[String]) -> Vec<Node> {
    let mut entries = Vec::new();
    for line in lines {
        let column = line.find("├─ ").or_else(|| line.find("└─ "));
        let (depth, text) = match column {
            Some(index) => {
                (line[..index].chars().count().div_ceil(4), line[index..].chars().skip(3).collect::<String>())
            }
            None => (0, line.trim().to_string()),
        };
        entries.push((depth, text.trim_end().to_string()));
    }
    let mut index = 0;
    let mut roots = Vec::new();
    while index < entries.len() {
        let depth = entries[index].0;
        roots.push(parse_node(&entries, &mut index, depth));
    }
    roots
}

/// parse_node parses the node at the index and its descendants.
fn parse_node(entries: &[(usize, String)], index: &mut usize, depth: usize) -> Node {
    let mut node = Node { label: entries[*index].1.clone(), ..Node::default() };
    *index += 1;
    while *index < entries.len() && entries[*index].0 > depth {
        let (child_depth, text) = &entries[*index];
        if NODE_PREFIXES.iter().any(|prefix| {
            text == prefix || text.starts_with(&format!("{prefix}(")) || text.starts_with(&format!("{prefix} "))
        }) {
            node.children.push(parse_node(entries, index, *child_depth));
        } else {
            match text.split_once(": ") {
                Some((key, value)) if key.chars().all(|c| c.is_ascii_lowercase() || c == '-' || c == '_') => {
                    node.properties.push((key.to_string(), value.to_string()));
                }
                _ => node.properties.push(("expression".to_string(), text.clone())),
            }
            *index += 1;
        }
    }
    node
}

/// label_argument returns the text in a label's parentheses, such as t for IndexedTableAccess(t).
fn label_argument(label: &str) -> Option<&str> {
    label.split_once('(').map(|(_, rest)| rest.trim_end_matches(')'))
}

/// property returns a node's property.
fn property<'a>(node: &'a Node, key: &str) -> Option<&'a str> {
    node.properties.iter().find(|(k, _)| k == key).map(|(_, v)| v.as_str())
}

/// first_table returns the first table a subtree reads.
fn first_table(node: &Node) -> Option<String> {
    if node.label.starts_with("IndexedTableAccess(") {
        return label_argument(&node.label).map(str::to_string);
    }
    if node.label == "Table" {
        return property(node, "name").map(str::to_string);
    }
    node.children.iter().find_map(first_table)
}

/// Fact is an owned PlanFact, as found in a parsed plan.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Fact {
    /// An index scan of the table on the columns with the ranges.
    IndexScan(String, Vec<String>, String),
    /// A descending index scan of the table.
    ReverseScan(String),
    /// A full scan of the table.
    FullScan(String),
    /// A join of the kind between inputs whose first tables are these.
    Join(String, String, String),
    /// A sort.
    Sort,
}

/// facts returns every fact a parsed plan shows.
pub fn facts(nodes: &[Node]) -> Vec<Fact> {
    let mut found = Vec::new();
    for node in nodes {
        collect(node, &mut found);
    }
    found
}

/// collect adds the facts of a node and its descendants.
fn collect(node: &Node, found: &mut Vec<Fact>) {
    if node.label.starts_with("IndexedTableAccess(") {
        let table = label_argument(&node.label).unwrap_or_default().to_string();
        let columns = property(node, "index")
            .map(|index| {
                index
                    .trim_matches(['[', ']'])
                    .split(',')
                    .map(|c| c.trim().rsplit('.').next().unwrap_or_default().to_string())
                    .collect()
            })
            .unwrap_or_default();
        let ranges = property(node, "filters").unwrap_or_default().to_string();
        if property(node, "reverse") == Some("true") {
            found.push(Fact::ReverseScan(table.clone()));
        }
        found.push(Fact::IndexScan(table, columns, ranges));
    } else if node.label == "Table" {
        if let Some(name) = property(node, "name") {
            found.push(Fact::FullScan(name.to_string()));
        }
    } else if node.label == "Sort" || node.label.starts_with("Sort(") || node.label.starts_with("TopN") {
        found.push(Fact::Sort);
    } else if node.label.contains("Join") && node.children.len() >= 2 {
        let kind = node.label.split(['(', ' ']).next().unwrap_or_default().to_string();
        let left = first_table(&node.children[0]).unwrap_or_default();
        let right = first_table(&node.children[1]).unwrap_or_default();
        found.push(Fact::Join(kind, left, right));
    }
    for child in &node.children {
        collect(child, found);
    }
}

/// check_facts compares the expected facts against the facts of a plan, returning what is missing.
pub fn check_facts(expected: &[PlanFact], actual: &[Fact]) -> Vec<String> {
    let mut problems = Vec::new();
    for fact in expected {
        let present = match fact {
            PlanFact::IndexScan { table, columns, ranges } => actual.iter().any(|a| {
                matches!(a, Fact::IndexScan(t, c, r)
                    if t == table && c.iter().map(String::as_str).eq(columns.iter().copied()) && (ranges.is_empty() || r == ranges))
            }),
            PlanFact::ReverseScan { table } => actual.iter().any(|a| matches!(a, Fact::ReverseScan(t) if t == table)),
            PlanFact::FullScan { table } => actual.iter().any(|a| matches!(a, Fact::FullScan(t) if t == table)),
            PlanFact::Join { kind, left, right } => {
                actual.iter().any(|a| matches!(a, Fact::Join(k, l, r) if k == kind && l == left && r == right))
            }
            PlanFact::Sort => actual.contains(&Fact::Sort),
            PlanFact::NoSort => !actual.contains(&Fact::Sort),
        };
        if !present {
            problems.push(format!("the plan does not show {fact:?}; it shows {actual:?}"));
        }
    }
    problems
}
