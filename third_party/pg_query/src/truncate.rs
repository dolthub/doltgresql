use std::cmp::Ordering;
use std::collections::VecDeque;

use crate::*;

#[derive(Debug)]
enum TruncationAttr {
    TargetList,
    WhereClause,
    ValuesLists,
    CTEQuery,
    Cols,
}

#[derive(Debug)]
struct PossibleTruncation {
    attr: TruncationAttr,
    node: NodeMut,
    depth: i32,
    length: i32,
}

pub fn truncate(protobuf: &protobuf::ParseResult, max_length: usize) -> Result<String> {
    let mut output = protobuf.deparse()?;
    if output.len() <= max_length {
        return Ok(output);
    }

    // SAFETY: within this scope nobody expects to have exclusive access to `protobuf`'s contents, so we can have multiple shared accesses.
    //
    // Raw pointer documentation:
    //
    // https://doc.rust-lang.org/book/ch19-01-unsafe-rust.html#dereferencing-a-raw-pointer
    // https://doc.rust-lang.org/std/primitive.pointer.html
    // https://manishearth.github.io/blog/2015/05/17/the-problem-with-shared-mutability
    // https://ricardomartins.cc/2016/07/11/interior-mutability-behind-the-curtain
    unsafe {
        let mut protobuf = protobuf.clone();
        let mut truncations: VecDeque<PossibleTruncation> = VecDeque::new();
        for (node, depth, _context) in protobuf.nodes_mut().into_iter() {
            match node {
                NodeMut::SelectStmt(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if !s.target_list.is_empty() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::TargetList,
                            node,
                            depth,
                            length: select_target_list_len(s.target_list.clone())?,
                        });
                    }
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                    if !s.values_lists.is_empty() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::ValuesLists,
                            node,
                            depth,
                            length: select_values_lists_len(s.values_lists.clone())?,
                        });
                    }
                }
                NodeMut::UpdateStmt(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if !s.target_list.is_empty() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::TargetList,
                            node,
                            depth,
                            length: update_target_list_len(s.target_list.clone())?,
                        });
                    }
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                }
                NodeMut::DeleteStmt(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                }
                NodeMut::CopyStmt(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                }
                NodeMut::InsertStmt(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if !s.cols.is_empty() {
                        truncations.push_back(PossibleTruncation { attr: TruncationAttr::Cols, node, depth, length: cols_len(s.cols.clone())? });
                    }
                }
                NodeMut::IndexStmt(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                }
                NodeMut::RuleStmt(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                }
                NodeMut::CommonTableExpr(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if let Some(cte) = s.ctequery.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::CTEQuery,
                            node,
                            depth: depth + 1,
                            length: cte.deparse()?.len() as i32,
                        });
                    }
                }
                NodeMut::InferClause(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                }
                NodeMut::OnConflictClause(s) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    if !s.target_list.is_empty() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::TargetList,
                            node,
                            depth,
                            length: update_target_list_len(s.target_list.clone())?,
                        });
                    }
                    if let Some(clause) = s.where_clause.as_ref() {
                        truncations.push_back(PossibleTruncation {
                            attr: TruncationAttr::WhereClause,
                            node,
                            depth,
                            length: where_clause_len((*clause).clone())?,
                        });
                    }
                }
                _ => (),
            }
        }

        truncations.make_contiguous().sort_by(|a, b| match a.depth.cmp(&b.depth).reverse() {
            Ordering::Equal => a.length.cmp(&b.length).reverse(),
            other => other,
        });

        while let Some(truncation) = truncations.pop_front() {
            match (truncation.node, truncation.attr) {
                (NodeMut::SelectStmt(s), TruncationAttr::TargetList) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.target_list = vec![dummy_target()];
                }
                (NodeMut::SelectStmt(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                (NodeMut::SelectStmt(s), TruncationAttr::ValuesLists) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.values_lists = vec![Node { node: Some(NodeEnum::List(protobuf::List { items: vec![*dummy_column()] })) }]
                }
                (NodeMut::UpdateStmt(s), TruncationAttr::TargetList) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.target_list = vec![dummy_target()];
                }
                (NodeMut::UpdateStmt(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                (NodeMut::DeleteStmt(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                (NodeMut::CopyStmt(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                (NodeMut::InsertStmt(s), TruncationAttr::Cols) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.cols = vec![dummy_target()];
                }
                (NodeMut::IndexStmt(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                (NodeMut::RuleStmt(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                (NodeMut::CommonTableExpr(s), TruncationAttr::CTEQuery) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    let old = std::mem::replace(&mut s.ctequery, Some(dummy_select(vec![], Some(dummy_column()), vec![])));
                    if let Some(s) = old {
                        let node = s.node.ok_or(Error::InvalidPointer)?;
                        truncations.retain(|t| t.node.to_enum().unwrap() != node);
                    }
                }
                (NodeMut::InferClause(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                (NodeMut::OnConflictClause(s), TruncationAttr::TargetList) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.target_list = vec![dummy_target()];
                }
                (NodeMut::OnConflictClause(s), TruncationAttr::WhereClause) => {
                    let s = s.as_mut().ok_or(Error::InvalidPointer)?;
                    s.where_clause = Some(dummy_column());
                }
                _ => panic!("unimplemented truncation"),
            }
            output = protobuf.deparse()?;
            output = output.replace("SELECT WHERE \"…\"", "...").replace("\"…\"", "...");
            // the unwanted AS doesn't happen in the Ruby version. I'm not sure where it's coming from
            output = output.replace("SELECT ... AS ...", "SELECT ...");
            if output.len() <= max_length {
                return Ok(output);
            }
        }
    }

    // We couldn't do a proper smart truncation, so we need a hard cut-off
    Ok(format!("{}...", truncate_str(&output, max_length - 3)))
}

// Truncates at character boundaries to prevent panics.
fn truncate_str(string: &str, max_chars: usize) -> &str {
    match string.char_indices().nth(max_chars) {
        None => string,
        Some((idx, _)) => &string[..idx],
    }
}

fn select_target_list_len(nodes: Vec<Node>) -> Result<i32> {
    let fragment = dummy_select(nodes, None, vec![]).deparse()?;
    Ok(fragment.len() as i32 - 7) // "SELECT "
}

fn select_values_lists_len(nodes: Vec<Node>) -> Result<i32> {
    let fragment = dummy_select(vec![], None, nodes).deparse()?;
    Ok(fragment.len() as i32 - 7) // "SELECT "
}

fn update_target_list_len(nodes: Vec<Node>) -> Result<i32> {
    let fragment = dummy_update(nodes).deparse()?;
    Ok(fragment.len() as i32 - 13) // "UPDATE x SET "
}

fn where_clause_len(node: Box<Node>) -> Result<i32> {
    let fragment = dummy_select(vec![], Some(node), vec![]).deparse()?;
    Ok(fragment.len() as i32 - 13) // "SELECT WHERE "
}

fn cols_len(nodes: Vec<Node>) -> Result<i32> {
    let fragment = dummy_insert(nodes).deparse()?;
    Ok(fragment.len() as i32 - 31) // "INSERT INTO x () DEFAULT VALUES"
}

fn dummy_column() -> Box<Node> {
    Box::new(Node {
        node: Some(NodeEnum::ColumnRef(protobuf::ColumnRef {
            location: 0,
            fields: vec![Node { node: Some(NodeEnum::String(protobuf::String { sval: "…".to_string() })) }],
        })),
    })
}

fn dummy_target() -> Node {
    Node {
        node: Some(NodeEnum::ResTarget(Box::new(protobuf::ResTarget {
            name: "…".to_string(),
            location: 0,
            indirection: vec![],
            val: Some(dummy_column()),
        }))),
    }
}

fn dummy_select(target_list: Vec<Node>, where_clause: Option<Box<Node>>, values_lists: Vec<Node>) -> Box<Node> {
    Box::new(Node {
        node: Some(NodeEnum::SelectStmt(Box::new(protobuf::SelectStmt {
            distinct_clause: vec![],
            into_clause: None,
            target_list,
            from_clause: vec![],
            where_clause,
            group_clause: vec![],
            having_clause: None,
            window_clause: vec![],
            values_lists,
            sort_clause: vec![],
            limit_offset: None,
            limit_count: None,
            limit_option: 1,
            locking_clause: vec![],
            with_clause: None,
            op: 1,
            all: false,
            larg: None,
            rarg: None,
            group_distinct: false,
        }))),
    })
}

fn dummy_insert(cols: Vec<Node>) -> Box<Node> {
    Box::new(Node {
        node: Some(NodeEnum::InsertStmt(Box::new(protobuf::InsertStmt {
            relation: Some(protobuf::RangeVar {
                catalogname: "".to_string(),
                schemaname: "".to_string(),
                relname: "x".to_string(),
                inh: true,
                relpersistence: "p".to_string(),
                alias: None,
                location: 0,
            }),
            cols,
            select_stmt: None,
            on_conflict_clause: None,
            returning_list: vec![],
            with_clause: None,
            r#override: 1,
        }))),
    })
}

fn dummy_update(target_list: Vec<Node>) -> Box<Node> {
    Box::new(Node {
        node: Some(NodeEnum::UpdateStmt(Box::new(protobuf::UpdateStmt {
            relation: Some(protobuf::RangeVar {
                catalogname: "".to_string(),
                schemaname: "".to_string(),
                relname: "x".to_string(),
                inh: true,
                relpersistence: "p".to_string(),
                alias: None,
                location: 0,
            }),
            from_clause: vec![],
            target_list,
            where_clause: None,
            returning_list: vec![],
            with_clause: None,
        }))),
    })
}
