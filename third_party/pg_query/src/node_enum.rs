use crate::*;

pub use protobuf::node::Node as NodeEnum;
use std::collections::VecDeque;

#[derive(Copy, Clone, Debug, Eq, PartialEq, Hash, serde::Serialize)]
pub enum Context {
    None,
    Select,
    DML,
    DDL,
    Call,
}

impl NodeEnum {
    pub fn deparse(&self) -> Result<String> {
        crate::deparse(&protobuf::ParseResult {
            version: crate::bindings::PG_VERSION_NUM as i32,
            stmts: vec![protobuf::RawStmt { stmt: Some(Box::new(Node { node: Some(self.clone()) })), stmt_location: 0, stmt_len: 0 }],
        })
    }

    pub fn nodes(&self) -> Vec<(NodeRef<'_>, i32, Context, bool)> {
        let mut iter = VecDeque::new();
        iter.push_front((self.to_ref(), 0, Context::None, false));
        let mut nodes = Vec::new();
        while let Some((node, depth, context, has_filter_columns)) = iter.pop_front() {
            let depth = depth + 1;
            match node {
                //
                // The following statement types do not modify tables
                //
                NodeRef::SelectStmt(s) => {
                    s.target_list.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    });
                    if let Some(n) = &s.where_clause {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, true));
                        }
                    }
                    s.sort_clause.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    });
                    s.group_clause.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    });
                    if let Some(n) = &s.having_clause {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    }
                    if let Some(clause) = &s.with_clause {
                        clause.ctes.iter().for_each(|n| {
                            if let Some(n) = n.node.as_ref() {
                                iter.push_back((n.to_ref(), depth, Context::Select, false));
                            }
                        });
                    }
                    match protobuf::SetOperation::try_from(s.op) {
                        Ok(protobuf::SetOperation::SetopNone) => {
                            s.from_clause.iter().for_each(|n| {
                                if let Some(n) = n.node.as_ref() {
                                    iter.push_back((n.to_ref(), depth, Context::Select, false));
                                }
                            });
                        }
                        Ok(protobuf::SetOperation::SetopUnion) => {
                            if let Some(left) = s.larg.as_ref() {
                                iter.push_back((left.to_ref(), depth, Context::Select, false));
                            }
                            if let Some(right) = s.rarg.as_ref() {
                                iter.push_back((right.to_ref(), depth, Context::Select, false));
                            }
                        }
                        Ok(protobuf::SetOperation::SetopExcept) => {
                            if let Some(left) = s.larg.as_ref() {
                                iter.push_back((left.to_ref(), depth, Context::Select, false));
                            }
                            if let Some(right) = s.rarg.as_ref() {
                                iter.push_back((right.to_ref(), depth, Context::Select, false));
                            }
                        }
                        Ok(protobuf::SetOperation::SetopIntersect) => {
                            if let Some(left) = s.larg.as_ref() {
                                iter.push_back((left.to_ref(), depth, Context::Select, false));
                            }
                            if let Some(right) = s.rarg.as_ref() {
                                iter.push_back((right.to_ref(), depth, Context::Select, false));
                            }
                        }
                        Ok(protobuf::SetOperation::Undefined) | Err(_) => (),
                    }
                }
                NodeRef::InsertStmt(s) => {
                    if let Some(n) = &s.select_stmt {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DML, false));
                        }
                    }
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DML, false));
                    }
                    if let Some(clause) = &s.with_clause {
                        clause.ctes.iter().for_each(|n| {
                            if let Some(n) = n.node.as_ref() {
                                iter.push_back((n.to_ref(), depth, Context::DML, false));
                            }
                        });
                    }
                    if let Some(n) = &s.on_conflict_clause {
                        iter.push_back((n.to_ref(), depth, Context::DML, false));
                    }
                }
                NodeRef::UpdateStmt(s) => {
                    s.target_list.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DML, false));
                        }
                    });
                    s.where_clause.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DML, true));
                        }
                    });
                    s.from_clause.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    });
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DML, false));
                    }
                    if let Some(clause) = &s.with_clause {
                        clause.ctes.iter().for_each(|n| {
                            if let Some(n) = n.node.as_ref() {
                                iter.push_back((n.to_ref(), depth, Context::DML, false));
                            }
                        });
                    }
                }
                NodeRef::DeleteStmt(s) => {
                    if let Some(n) = &s.where_clause {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DML, true));
                        }
                    }
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DML, false));
                    }
                    if let Some(clause) = &s.with_clause {
                        clause.ctes.iter().for_each(|n| {
                            if let Some(n) = n.node.as_ref() {
                                iter.push_back((n.to_ref(), depth, Context::DML, false));
                            }
                        });
                    }
                    s.using_clause.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    });
                }
                NodeRef::MergeStmt(m) => {
                    if let Some(t) = m.relation.as_ref() {
                        iter.push_back((t.to_ref(), depth, Context::DML, false));
                    }

                    if let Some(clause) = &m.with_clause {
                        clause.ctes.iter().for_each(|n| {
                            if let Some(n) = n.node.as_ref() {
                                iter.push_back((n.to_ref(), depth, Context::DML, false));
                            }
                        });
                    }

                    m.source_relation.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    });
                    m.merge_when_clauses.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DML, true));
                        }
                    });
                    m.join_condition.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::Select, false));
                        }
                    });
                }
                NodeRef::CommonTableExpr(s) => {
                    if let Some(n) = &s.ctequery {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, false));
                        }
                    }
                }
                NodeRef::CopyStmt(s) => {
                    if let Some(n) = &s.query {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DML, false));
                        }
                    }
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DML, false));
                    }
                }
                NodeRef::CallStmt(s) => {
                    if let Some(n) = s.funccall.as_ref() {
                        iter.push_back((n.to_ref(), depth, Context::Call, false));
                    }
                }
                //
                // The following statement types are DDL (changing table structure)
                //
                NodeRef::AlterTableStmt(s) => {
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                    }
                }
                NodeRef::CreateStmt(s) => {
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                    }
                }
                NodeRef::CreateTableAsStmt(s) => {
                    if let Some(n) = &s.query {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DDL, false));
                        }
                    }
                    if let Some(n) = &s.into {
                        if let Some(rel) = n.rel.as_ref() {
                            iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                        }
                    }
                }
                NodeRef::TruncateStmt(s) => {
                    s.relations.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DDL, false));
                        }
                    });
                }
                NodeRef::ViewStmt(s) => {
                    if let Some(n) = &s.query {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DDL, false));
                        }
                    }
                    if let Some(rel) = s.view.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                    }
                }
                NodeRef::IndexStmt(s) => {
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                    }
                    s.index_params.iter().for_each(|n| {
                        if let Some(NodeEnum::IndexElem(n)) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DDL, false));

                            if let Some(n) = n.expr.as_ref().and_then(|n| n.node.as_ref()) {
                                iter.push_back((n.to_ref(), depth, Context::DDL, false));
                            }
                        }
                    });
                    if let Some(n) = s.where_clause.as_ref() {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DDL, true));
                        }
                    }
                }
                NodeRef::CreateTrigStmt(s) => {
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                    }
                }
                NodeRef::RuleStmt(s) => {
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                    }
                }
                NodeRef::VacuumStmt(s) => {
                    for node in &s.rels {
                        if let Some(NodeEnum::VacuumRelation(r)) = &node.node {
                            if let Some(rel) = r.relation.as_ref() {
                                iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                            }
                        }
                    }
                }
                NodeRef::RefreshMatViewStmt(s) => {
                    if let Some(rel) = s.relation.as_ref() {
                        iter.push_back((rel.to_ref(), depth, Context::DDL, false));
                    }
                }
                NodeRef::GrantStmt(s) => {
                    if let Ok(protobuf::ObjectType::ObjectTable) = protobuf::ObjectType::try_from(s.objtype) {
                        s.objects.iter().for_each(|n| {
                            if let Some(n) = n.node.as_ref() {
                                iter.push_back((n.to_ref(), depth, Context::DDL, false));
                            }
                        });
                    }
                }
                NodeRef::LockStmt(s) => {
                    s.relations.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, Context::DDL, false));
                        }
                    });
                }
                NodeRef::ExplainStmt(s) => {
                    if let Some(n) = &s.query {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, false));
                        }
                    }
                }
                //
                // Subselect items
                //
                NodeRef::AExpr(e) => {
                    if let Some(n) = &e.lexpr {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                    if let Some(n) = &e.rexpr {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::BoolExpr(e) => {
                    e.args.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                }
                NodeRef::BooleanTest(e) => {
                    if let Some(n) = &e.arg {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::CoalesceExpr(e) => {
                    e.args.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                }
                NodeRef::MinMaxExpr(e) => {
                    e.args.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                }
                NodeRef::NullTest(e) => {
                    if let Some(n) = &e.arg {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::ResTarget(t) => {
                    if let Some(n) = &t.val {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::SubLink(l) => {
                    if let Some(n) = &l.subselect {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::FuncCall(c) => {
                    c.args.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                }
                NodeRef::CaseExpr(c) => {
                    c.args.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                    if let Some(n) = &c.defresult {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::CaseWhen(w) => {
                    if let Some(n) = &w.expr {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                    if let Some(n) = &w.result {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::SortBy(n) => {
                    if let Some(n) = &n.node {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::TypeCast(n) => {
                    if let Some(n) = &n.arg {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                //
                // from-clause items
                //
                NodeRef::List(l) => {
                    l.items.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                }
                NodeRef::JoinExpr(e) => {
                    [&e.larg, &e.rarg, &e.quals].iter().for_each(|n| {
                        if let Some(n) = n {
                            if let Some(n) = n.node.as_ref() {
                                iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                            }
                        }
                    });
                }
                NodeRef::RowExpr(e) => {
                    e.args.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                }
                NodeRef::RangeSubselect(s) => {
                    if let Some(n) = &s.subquery {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    }
                }
                NodeRef::RangeFunction(f) => {
                    f.functions.iter().for_each(|n| {
                        if let Some(n) = n.node.as_ref() {
                            iter.push_back((n.to_ref(), depth, context, has_filter_columns));
                        }
                    });
                }
                _ => (),
            }
            nodes.push((node, depth, context, has_filter_columns));
        }
        nodes
    }

    /// Returns a mutable reference to nested nodes.
    ///
    /// # Safety
    ///
    /// The caller may have to deal with dangling pointers, and passing an
    /// invalid tree back to libpg_query may cause it to panic.
    pub unsafe fn nodes_mut(&mut self) -> Vec<(NodeMut, i32, Context)> {
        let mut iter = VecDeque::new();
        iter.push_front((self.to_mut(), 0, Context::None));
        let mut nodes = Vec::new();
        while let Some((node, depth, context)) = iter.pop_front() {
            let depth = depth + 1;
            match node {
                //
                // The following statement types do not modify tables
                //
                NodeMut::SelectStmt(s) => {
                    let s = s.as_mut().unwrap();
                    s.target_list.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::Select));
                        }
                    });
                    if let Some(n) = s.where_clause.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::Select));
                        }
                    }
                    s.sort_clause.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::Select));
                        }
                    });
                    s.group_clause.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::Select));
                        }
                    });
                    if let Some(n) = s.having_clause.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::Select));
                        }
                    }
                    if let Some(clause) = s.with_clause.as_mut() {
                        clause.ctes.iter_mut().for_each(|n| {
                            if let Some(n) = n.node.as_mut() {
                                iter.push_back((n.to_mut(), depth, Context::Select));
                            }
                        });
                    }
                    match protobuf::SetOperation::try_from(s.op) {
                        Ok(protobuf::SetOperation::SetopNone) => {
                            s.from_clause.iter_mut().for_each(|n| {
                                if let Some(n) = n.node.as_mut() {
                                    iter.push_back((n.to_mut(), depth, Context::Select));
                                }
                            });
                        }
                        Ok(protobuf::SetOperation::SetopUnion) => {
                            if let Some(left) = s.larg.as_mut() {
                                iter.push_back((left.to_mut(), depth, Context::Select));
                            }
                            if let Some(right) = s.rarg.as_mut() {
                                iter.push_back((right.to_mut(), depth, Context::Select));
                            }
                        }
                        Ok(protobuf::SetOperation::SetopExcept) => {
                            if let Some(left) = s.larg.as_mut() {
                                iter.push_back((left.to_mut(), depth, Context::Select));
                            }
                            if let Some(right) = s.rarg.as_mut() {
                                iter.push_back((right.to_mut(), depth, Context::Select));
                            }
                        }
                        Ok(protobuf::SetOperation::SetopIntersect) => {
                            if let Some(left) = s.larg.as_mut() {
                                iter.push_back((left.to_mut(), depth, Context::Select));
                            }
                            if let Some(right) = s.rarg.as_mut() {
                                iter.push_back((right.to_mut(), depth, Context::Select));
                            }
                        }
                        Ok(protobuf::SetOperation::Undefined) | Err(_) => (),
                    }
                }
                NodeMut::InsertStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.select_stmt.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DML));
                        }
                    }
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DML));
                    }
                    if let Some(clause) = s.with_clause.as_mut() {
                        clause.ctes.iter_mut().for_each(|n| {
                            if let Some(n) = n.node.as_mut() {
                                iter.push_back((n.to_mut(), depth, Context::DML));
                            }
                        });
                    }
                    if let Some(n) = s.on_conflict_clause.as_mut() {
                        iter.push_back((n.to_mut(), depth, Context::DML));
                    }
                }
                NodeMut::UpdateStmt(s) => {
                    let s = s.as_mut().unwrap();
                    s.target_list.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DML));
                        }
                    });
                    s.where_clause.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DML));
                        }
                    });
                    s.from_clause.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::Select));
                        }
                    });
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DML));
                    }
                    if let Some(clause) = s.with_clause.as_mut() {
                        clause.ctes.iter_mut().for_each(|n| {
                            if let Some(n) = n.node.as_mut() {
                                iter.push_back((n.to_mut(), depth, Context::DML));
                            }
                        });
                    }
                }
                NodeMut::DeleteStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.where_clause.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DML));
                        }
                    }
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DML));
                    }
                    if let Some(clause) = s.with_clause.as_mut() {
                        clause.ctes.iter_mut().for_each(|n| {
                            if let Some(n) = n.node.as_mut() {
                                iter.push_back((n.to_mut(), depth, Context::DML));
                            }
                        });
                    }
                    s.using_clause.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::Select));
                        }
                    });
                }
                NodeMut::CommonTableExpr(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.ctequery.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::CopyStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.query.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DML));
                        }
                    }
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DML));
                    }
                }
                NodeMut::CallStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.funccall.as_mut() {
                        iter.push_back((n.to_mut(), depth, Context::Call));
                    }
                }
                //
                // The following statement types are DDL (changing table structure)
                //
                NodeMut::AlterTableStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DDL));
                    }
                }
                NodeMut::CreateStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DDL));
                    }
                }
                NodeMut::CreateTableAsStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.query.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DDL));
                        }
                    }
                    if let Some(n) = s.into.as_mut() {
                        if let Some(rel) = n.rel.as_mut() {
                            iter.push_back((rel.to_mut(), depth, Context::DDL));
                        }
                    }
                }
                NodeMut::TruncateStmt(s) => {
                    let s = s.as_mut().unwrap();
                    s.relations.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DDL));
                        }
                    });
                }
                NodeMut::ViewStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.query.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DDL));
                        }
                    }
                    if let Some(rel) = s.view.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DDL));
                    }
                }
                NodeMut::IndexStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DDL));
                    }
                    s.index_params.iter_mut().for_each(|n| {
                        if let Some(NodeEnum::IndexElem(n)) = n.node.as_mut() {
                            if let Some(n) = n.expr.as_mut().and_then(|n| n.node.as_mut()) {
                                iter.push_back((n.to_mut(), depth, Context::DDL));
                            }
                        }
                    });
                }
                NodeMut::CreateTrigStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DDL));
                    }
                }
                NodeMut::RuleStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DDL));
                    }
                }
                NodeMut::VacuumStmt(s) => {
                    let s = s.as_mut().unwrap();
                    for node in s.rels.iter_mut() {
                        if let Some(NodeEnum::VacuumRelation(r)) = node.node.as_mut() {
                            if let Some(rel) = r.relation.as_mut() {
                                iter.push_back((rel.to_mut(), depth, Context::DDL));
                            }
                        }
                    }
                }
                NodeMut::RefreshMatViewStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(rel) = s.relation.as_mut() {
                        iter.push_back((rel.to_mut(), depth, Context::DDL));
                    }
                }
                NodeMut::GrantStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Ok(protobuf::ObjectType::ObjectTable) = protobuf::ObjectType::try_from(s.objtype) {
                        s.objects.iter_mut().for_each(|n| {
                            if let Some(n) = n.node.as_mut() {
                                iter.push_back((n.to_mut(), depth, Context::DDL));
                            }
                        });
                    }
                }
                NodeMut::LockStmt(s) => {
                    let s = s.as_mut().unwrap();
                    s.relations.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, Context::DDL));
                        }
                    });
                }
                NodeMut::ExplainStmt(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.query.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                //
                // Subselect items
                //
                NodeMut::AExpr(e) => {
                    let e = e.as_mut().unwrap();
                    if let Some(n) = e.lexpr.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                    if let Some(n) = e.rexpr.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::BoolExpr(e) => {
                    let e = e.as_mut().unwrap();
                    e.args.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                }
                NodeMut::CoalesceExpr(e) => {
                    let e = e.as_mut().unwrap();
                    e.args.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                }
                NodeMut::MinMaxExpr(e) => {
                    let e = e.as_mut().unwrap();
                    e.args.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                }
                NodeMut::NullTest(e) => {
                    let e = e.as_mut().unwrap();
                    if let Some(n) = e.arg.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::ResTarget(t) => {
                    let t = t.as_mut().unwrap();
                    if let Some(n) = t.val.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::SubLink(l) => {
                    let l = l.as_mut().unwrap();
                    if let Some(n) = l.subselect.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::FuncCall(c) => {
                    let c = c.as_mut().unwrap();
                    c.args.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                }
                NodeMut::CaseExpr(c) => {
                    let c = c.as_mut().unwrap();
                    c.args.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                    if let Some(n) = c.defresult.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::CaseWhen(w) => {
                    let w = w.as_mut().unwrap();
                    if let Some(n) = w.expr.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                    if let Some(n) = w.result.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::SortBy(n) => {
                    let n = n.as_mut().unwrap();
                    if let Some(n) = n.node.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::TypeCast(t) => {
                    let t = t.as_mut().unwrap();
                    if let Some(n) = t.arg.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                //
                // from-clause items
                //
                NodeMut::List(l) => {
                    let l = l.as_mut().unwrap();
                    l.items.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                }
                NodeMut::JoinExpr(e) => {
                    let e = e.as_mut().unwrap();
                    if let Some(n) = e.larg.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                    if let Some(n) = e.rarg.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                    if let Some(n) = e.quals.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::RowExpr(e) => {
                    let e = e.as_mut().unwrap();
                    e.args.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                }
                NodeMut::RangeSubselect(s) => {
                    let s = s.as_mut().unwrap();
                    if let Some(n) = s.subquery.as_mut() {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    }
                }
                NodeMut::RangeFunction(f) => {
                    let f = f.as_mut().unwrap();
                    f.functions.iter_mut().for_each(|n| {
                        if let Some(n) = n.node.as_mut() {
                            iter.push_back((n.to_mut(), depth, context));
                        }
                    });
                }
                _ => (),
            }
            nodes.push((node, depth, context));
        }
        nodes
    }

    pub fn to_ref(&self) -> NodeRef<'_> {
        match self {
            NodeEnum::Alias(n) => NodeRef::Alias(n),
            NodeEnum::RangeVar(n) => NodeRef::RangeVar(n),
            NodeEnum::TableFunc(n) => NodeRef::TableFunc(n),
            NodeEnum::Var(n) => NodeRef::Var(n),
            NodeEnum::Param(n) => NodeRef::Param(n),
            NodeEnum::Aggref(n) => NodeRef::Aggref(n),
            NodeEnum::GroupingFunc(n) => NodeRef::GroupingFunc(n),
            NodeEnum::WindowFunc(n) => NodeRef::WindowFunc(n),
            NodeEnum::SubscriptingRef(n) => NodeRef::SubscriptingRef(n),
            NodeEnum::FuncExpr(n) => NodeRef::FuncExpr(n),
            NodeEnum::NamedArgExpr(n) => NodeRef::NamedArgExpr(n),
            NodeEnum::OpExpr(n) => NodeRef::OpExpr(n),
            NodeEnum::DistinctExpr(n) => NodeRef::DistinctExpr(n),
            NodeEnum::NullIfExpr(n) => NodeRef::NullIfExpr(n),
            NodeEnum::ScalarArrayOpExpr(n) => NodeRef::ScalarArrayOpExpr(n),
            NodeEnum::BoolExpr(n) => NodeRef::BoolExpr(n),
            NodeEnum::SubLink(n) => NodeRef::SubLink(n),
            NodeEnum::SubPlan(n) => NodeRef::SubPlan(n),
            NodeEnum::AlternativeSubPlan(n) => NodeRef::AlternativeSubPlan(n),
            NodeEnum::FieldSelect(n) => NodeRef::FieldSelect(n),
            NodeEnum::FieldStore(n) => NodeRef::FieldStore(n),
            NodeEnum::RelabelType(n) => NodeRef::RelabelType(n),
            NodeEnum::CoerceViaIo(n) => NodeRef::CoerceViaIo(n),
            NodeEnum::ArrayCoerceExpr(n) => NodeRef::ArrayCoerceExpr(n),
            NodeEnum::ConvertRowtypeExpr(n) => NodeRef::ConvertRowtypeExpr(n),
            NodeEnum::CollateExpr(n) => NodeRef::CollateExpr(n),
            NodeEnum::CaseExpr(n) => NodeRef::CaseExpr(n),
            NodeEnum::CaseWhen(n) => NodeRef::CaseWhen(n),
            NodeEnum::CaseTestExpr(n) => NodeRef::CaseTestExpr(n),
            NodeEnum::ArrayExpr(n) => NodeRef::ArrayExpr(n),
            NodeEnum::RowExpr(n) => NodeRef::RowExpr(n),
            NodeEnum::RowCompareExpr(n) => NodeRef::RowCompareExpr(n),
            NodeEnum::CoalesceExpr(n) => NodeRef::CoalesceExpr(n),
            NodeEnum::MinMaxExpr(n) => NodeRef::MinMaxExpr(n),
            NodeEnum::SqlvalueFunction(n) => NodeRef::SqlvalueFunction(n),
            NodeEnum::XmlExpr(n) => NodeRef::XmlExpr(n),
            NodeEnum::NullTest(n) => NodeRef::NullTest(n),
            NodeEnum::BooleanTest(n) => NodeRef::BooleanTest(n),
            NodeEnum::CoerceToDomain(n) => NodeRef::CoerceToDomain(n),
            NodeEnum::CoerceToDomainValue(n) => NodeRef::CoerceToDomainValue(n),
            NodeEnum::SetToDefault(n) => NodeRef::SetToDefault(n),
            NodeEnum::CurrentOfExpr(n) => NodeRef::CurrentOfExpr(n),
            NodeEnum::NextValueExpr(n) => NodeRef::NextValueExpr(n),
            NodeEnum::InferenceElem(n) => NodeRef::InferenceElem(n),
            NodeEnum::TargetEntry(n) => NodeRef::TargetEntry(n),
            NodeEnum::RangeTblRef(n) => NodeRef::RangeTblRef(n),
            NodeEnum::JoinExpr(n) => NodeRef::JoinExpr(n),
            NodeEnum::FromExpr(n) => NodeRef::FromExpr(n),
            NodeEnum::OnConflictExpr(n) => NodeRef::OnConflictExpr(n),
            NodeEnum::IntoClause(n) => NodeRef::IntoClause(n),
            NodeEnum::RawStmt(n) => NodeRef::RawStmt(n),
            NodeEnum::Query(n) => NodeRef::Query(n),
            NodeEnum::InsertStmt(n) => NodeRef::InsertStmt(n),
            NodeEnum::DeleteStmt(n) => NodeRef::DeleteStmt(n),
            NodeEnum::UpdateStmt(n) => NodeRef::UpdateStmt(n),
            NodeEnum::SelectStmt(n) => NodeRef::SelectStmt(n),
            NodeEnum::AlterTableStmt(n) => NodeRef::AlterTableStmt(n),
            NodeEnum::AlterTableCmd(n) => NodeRef::AlterTableCmd(n),
            NodeEnum::AlterDomainStmt(n) => NodeRef::AlterDomainStmt(n),
            NodeEnum::SetOperationStmt(n) => NodeRef::SetOperationStmt(n),
            NodeEnum::GrantStmt(n) => NodeRef::GrantStmt(n),
            NodeEnum::GrantRoleStmt(n) => NodeRef::GrantRoleStmt(n),
            NodeEnum::AlterDefaultPrivilegesStmt(n) => NodeRef::AlterDefaultPrivilegesStmt(n),
            NodeEnum::ClosePortalStmt(n) => NodeRef::ClosePortalStmt(n),
            NodeEnum::ClusterStmt(n) => NodeRef::ClusterStmt(n),
            NodeEnum::CopyStmt(n) => NodeRef::CopyStmt(n),
            NodeEnum::CreateStmt(n) => NodeRef::CreateStmt(n),
            NodeEnum::DefineStmt(n) => NodeRef::DefineStmt(n),
            NodeEnum::DropStmt(n) => NodeRef::DropStmt(n),
            NodeEnum::TruncateStmt(n) => NodeRef::TruncateStmt(n),
            NodeEnum::CommentStmt(n) => NodeRef::CommentStmt(n),
            NodeEnum::FetchStmt(n) => NodeRef::FetchStmt(n),
            NodeEnum::IndexStmt(n) => NodeRef::IndexStmt(n),
            NodeEnum::CreateFunctionStmt(n) => NodeRef::CreateFunctionStmt(n),
            NodeEnum::AlterFunctionStmt(n) => NodeRef::AlterFunctionStmt(n),
            NodeEnum::DoStmt(n) => NodeRef::DoStmt(n),
            NodeEnum::RenameStmt(n) => NodeRef::RenameStmt(n),
            NodeEnum::RuleStmt(n) => NodeRef::RuleStmt(n),
            NodeEnum::NotifyStmt(n) => NodeRef::NotifyStmt(n),
            NodeEnum::ListenStmt(n) => NodeRef::ListenStmt(n),
            NodeEnum::UnlistenStmt(n) => NodeRef::UnlistenStmt(n),
            NodeEnum::TransactionStmt(n) => NodeRef::TransactionStmt(n),
            NodeEnum::ViewStmt(n) => NodeRef::ViewStmt(n),
            NodeEnum::LoadStmt(n) => NodeRef::LoadStmt(n),
            NodeEnum::CreateDomainStmt(n) => NodeRef::CreateDomainStmt(n),
            NodeEnum::CreatedbStmt(n) => NodeRef::CreatedbStmt(n),
            NodeEnum::DropdbStmt(n) => NodeRef::DropdbStmt(n),
            NodeEnum::VacuumStmt(n) => NodeRef::VacuumStmt(n),
            NodeEnum::ExplainStmt(n) => NodeRef::ExplainStmt(n),
            NodeEnum::CreateTableAsStmt(n) => NodeRef::CreateTableAsStmt(n),
            NodeEnum::CreateSeqStmt(n) => NodeRef::CreateSeqStmt(n),
            NodeEnum::AlterSeqStmt(n) => NodeRef::AlterSeqStmt(n),
            NodeEnum::VariableSetStmt(n) => NodeRef::VariableSetStmt(n),
            NodeEnum::VariableShowStmt(n) => NodeRef::VariableShowStmt(n),
            NodeEnum::DiscardStmt(n) => NodeRef::DiscardStmt(n),
            NodeEnum::CreateTrigStmt(n) => NodeRef::CreateTrigStmt(n),
            NodeEnum::CreatePlangStmt(n) => NodeRef::CreatePlangStmt(n),
            NodeEnum::CreateRoleStmt(n) => NodeRef::CreateRoleStmt(n),
            NodeEnum::AlterRoleStmt(n) => NodeRef::AlterRoleStmt(n),
            NodeEnum::DropRoleStmt(n) => NodeRef::DropRoleStmt(n),
            NodeEnum::LockStmt(n) => NodeRef::LockStmt(n),
            NodeEnum::ConstraintsSetStmt(n) => NodeRef::ConstraintsSetStmt(n),
            NodeEnum::ReindexStmt(n) => NodeRef::ReindexStmt(n),
            NodeEnum::CheckPointStmt(n) => NodeRef::CheckPointStmt(n),
            NodeEnum::CreateSchemaStmt(n) => NodeRef::CreateSchemaStmt(n),
            NodeEnum::AlterDatabaseStmt(n) => NodeRef::AlterDatabaseStmt(n),
            NodeEnum::AlterDatabaseSetStmt(n) => NodeRef::AlterDatabaseSetStmt(n),
            NodeEnum::AlterRoleSetStmt(n) => NodeRef::AlterRoleSetStmt(n),
            NodeEnum::CreateConversionStmt(n) => NodeRef::CreateConversionStmt(n),
            NodeEnum::CreateCastStmt(n) => NodeRef::CreateCastStmt(n),
            NodeEnum::CreateOpClassStmt(n) => NodeRef::CreateOpClassStmt(n),
            NodeEnum::CreateOpFamilyStmt(n) => NodeRef::CreateOpFamilyStmt(n),
            NodeEnum::AlterOpFamilyStmt(n) => NodeRef::AlterOpFamilyStmt(n),
            NodeEnum::PrepareStmt(n) => NodeRef::PrepareStmt(n),
            NodeEnum::ExecuteStmt(n) => NodeRef::ExecuteStmt(n),
            NodeEnum::DeallocateStmt(n) => NodeRef::DeallocateStmt(n),
            NodeEnum::DeclareCursorStmt(n) => NodeRef::DeclareCursorStmt(n),
            NodeEnum::CreateTableSpaceStmt(n) => NodeRef::CreateTableSpaceStmt(n),
            NodeEnum::DropTableSpaceStmt(n) => NodeRef::DropTableSpaceStmt(n),
            NodeEnum::AlterObjectDependsStmt(n) => NodeRef::AlterObjectDependsStmt(n),
            NodeEnum::AlterObjectSchemaStmt(n) => NodeRef::AlterObjectSchemaStmt(n),
            NodeEnum::AlterOwnerStmt(n) => NodeRef::AlterOwnerStmt(n),
            NodeEnum::AlterOperatorStmt(n) => NodeRef::AlterOperatorStmt(n),
            NodeEnum::AlterTypeStmt(n) => NodeRef::AlterTypeStmt(n),
            NodeEnum::DropOwnedStmt(n) => NodeRef::DropOwnedStmt(n),
            NodeEnum::ReassignOwnedStmt(n) => NodeRef::ReassignOwnedStmt(n),
            NodeEnum::CompositeTypeStmt(n) => NodeRef::CompositeTypeStmt(n),
            NodeEnum::CreateEnumStmt(n) => NodeRef::CreateEnumStmt(n),
            NodeEnum::CreateRangeStmt(n) => NodeRef::CreateRangeStmt(n),
            NodeEnum::AlterEnumStmt(n) => NodeRef::AlterEnumStmt(n),
            NodeEnum::AlterTsdictionaryStmt(n) => NodeRef::AlterTsdictionaryStmt(n),
            NodeEnum::AlterTsconfigurationStmt(n) => NodeRef::AlterTsconfigurationStmt(n),
            NodeEnum::CreateFdwStmt(n) => NodeRef::CreateFdwStmt(n),
            NodeEnum::AlterFdwStmt(n) => NodeRef::AlterFdwStmt(n),
            NodeEnum::CreateForeignServerStmt(n) => NodeRef::CreateForeignServerStmt(n),
            NodeEnum::AlterForeignServerStmt(n) => NodeRef::AlterForeignServerStmt(n),
            NodeEnum::CreateUserMappingStmt(n) => NodeRef::CreateUserMappingStmt(n),
            NodeEnum::AlterUserMappingStmt(n) => NodeRef::AlterUserMappingStmt(n),
            NodeEnum::DropUserMappingStmt(n) => NodeRef::DropUserMappingStmt(n),
            NodeEnum::AlterTableSpaceOptionsStmt(n) => NodeRef::AlterTableSpaceOptionsStmt(n),
            NodeEnum::AlterTableMoveAllStmt(n) => NodeRef::AlterTableMoveAllStmt(n),
            NodeEnum::SecLabelStmt(n) => NodeRef::SecLabelStmt(n),
            NodeEnum::CreateForeignTableStmt(n) => NodeRef::CreateForeignTableStmt(n),
            NodeEnum::ImportForeignSchemaStmt(n) => NodeRef::ImportForeignSchemaStmt(n),
            NodeEnum::CreateExtensionStmt(n) => NodeRef::CreateExtensionStmt(n),
            NodeEnum::AlterExtensionStmt(n) => NodeRef::AlterExtensionStmt(n),
            NodeEnum::AlterExtensionContentsStmt(n) => NodeRef::AlterExtensionContentsStmt(n),
            NodeEnum::CreateEventTrigStmt(n) => NodeRef::CreateEventTrigStmt(n),
            NodeEnum::AlterEventTrigStmt(n) => NodeRef::AlterEventTrigStmt(n),
            NodeEnum::RefreshMatViewStmt(n) => NodeRef::RefreshMatViewStmt(n),
            NodeEnum::ReplicaIdentityStmt(n) => NodeRef::ReplicaIdentityStmt(n),
            NodeEnum::AlterSystemStmt(n) => NodeRef::AlterSystemStmt(n),
            NodeEnum::CreatePolicyStmt(n) => NodeRef::CreatePolicyStmt(n),
            NodeEnum::AlterPolicyStmt(n) => NodeRef::AlterPolicyStmt(n),
            NodeEnum::CreateTransformStmt(n) => NodeRef::CreateTransformStmt(n),
            NodeEnum::CreateAmStmt(n) => NodeRef::CreateAmStmt(n),
            NodeEnum::CreatePublicationStmt(n) => NodeRef::CreatePublicationStmt(n),
            NodeEnum::AlterPublicationStmt(n) => NodeRef::AlterPublicationStmt(n),
            NodeEnum::CreateSubscriptionStmt(n) => NodeRef::CreateSubscriptionStmt(n),
            NodeEnum::AlterSubscriptionStmt(n) => NodeRef::AlterSubscriptionStmt(n),
            NodeEnum::DropSubscriptionStmt(n) => NodeRef::DropSubscriptionStmt(n),
            NodeEnum::CreateStatsStmt(n) => NodeRef::CreateStatsStmt(n),
            NodeEnum::AlterCollationStmt(n) => NodeRef::AlterCollationStmt(n),
            NodeEnum::CallStmt(n) => NodeRef::CallStmt(n),
            NodeEnum::AlterStatsStmt(n) => NodeRef::AlterStatsStmt(n),
            NodeEnum::AExpr(n) => NodeRef::AExpr(n),
            NodeEnum::ColumnRef(n) => NodeRef::ColumnRef(n),
            NodeEnum::ParamRef(n) => NodeRef::ParamRef(n),
            NodeEnum::AConst(n) => NodeRef::AConst(n),
            NodeEnum::FuncCall(n) => NodeRef::FuncCall(n),
            NodeEnum::AStar(n) => NodeRef::AStar(n),
            NodeEnum::AIndices(n) => NodeRef::AIndices(n),
            NodeEnum::AIndirection(n) => NodeRef::AIndirection(n),
            NodeEnum::AArrayExpr(n) => NodeRef::AArrayExpr(n),
            NodeEnum::ResTarget(n) => NodeRef::ResTarget(n),
            NodeEnum::MultiAssignRef(n) => NodeRef::MultiAssignRef(n),
            NodeEnum::TypeCast(n) => NodeRef::TypeCast(n),
            NodeEnum::CollateClause(n) => NodeRef::CollateClause(n),
            NodeEnum::SortBy(n) => NodeRef::SortBy(n),
            NodeEnum::WindowDef(n) => NodeRef::WindowDef(n),
            NodeEnum::RangeSubselect(n) => NodeRef::RangeSubselect(n),
            NodeEnum::RangeFunction(n) => NodeRef::RangeFunction(n),
            NodeEnum::RangeTableSample(n) => NodeRef::RangeTableSample(n),
            NodeEnum::RangeTableFunc(n) => NodeRef::RangeTableFunc(n),
            NodeEnum::RangeTableFuncCol(n) => NodeRef::RangeTableFuncCol(n),
            NodeEnum::TypeName(n) => NodeRef::TypeName(n),
            NodeEnum::ColumnDef(n) => NodeRef::ColumnDef(n),
            NodeEnum::IndexElem(n) => NodeRef::IndexElem(n),
            NodeEnum::Constraint(n) => NodeRef::Constraint(n),
            NodeEnum::DefElem(n) => NodeRef::DefElem(n),
            NodeEnum::RangeTblEntry(n) => NodeRef::RangeTblEntry(n),
            NodeEnum::RangeTblFunction(n) => NodeRef::RangeTblFunction(n),
            NodeEnum::TableSampleClause(n) => NodeRef::TableSampleClause(n),
            NodeEnum::WithCheckOption(n) => NodeRef::WithCheckOption(n),
            NodeEnum::SortGroupClause(n) => NodeRef::SortGroupClause(n),
            NodeEnum::GroupingSet(n) => NodeRef::GroupingSet(n),
            NodeEnum::WindowClause(n) => NodeRef::WindowClause(n),
            NodeEnum::ObjectWithArgs(n) => NodeRef::ObjectWithArgs(n),
            NodeEnum::AccessPriv(n) => NodeRef::AccessPriv(n),
            NodeEnum::CreateOpClassItem(n) => NodeRef::CreateOpClassItem(n),
            NodeEnum::TableLikeClause(n) => NodeRef::TableLikeClause(n),
            NodeEnum::FunctionParameter(n) => NodeRef::FunctionParameter(n),
            NodeEnum::LockingClause(n) => NodeRef::LockingClause(n),
            NodeEnum::RowMarkClause(n) => NodeRef::RowMarkClause(n),
            NodeEnum::XmlSerialize(n) => NodeRef::XmlSerialize(n),
            NodeEnum::WithClause(n) => NodeRef::WithClause(n),
            NodeEnum::InferClause(n) => NodeRef::InferClause(n),
            NodeEnum::OnConflictClause(n) => NodeRef::OnConflictClause(n),
            NodeEnum::CommonTableExpr(n) => NodeRef::CommonTableExpr(n),
            NodeEnum::RoleSpec(n) => NodeRef::RoleSpec(n),
            NodeEnum::TriggerTransition(n) => NodeRef::TriggerTransition(n),
            NodeEnum::PartitionElem(n) => NodeRef::PartitionElem(n),
            NodeEnum::PartitionSpec(n) => NodeRef::PartitionSpec(n),
            NodeEnum::PartitionBoundSpec(n) => NodeRef::PartitionBoundSpec(n),
            NodeEnum::PartitionRangeDatum(n) => NodeRef::PartitionRangeDatum(n),
            NodeEnum::PartitionCmd(n) => NodeRef::PartitionCmd(n),
            NodeEnum::VacuumRelation(n) => NodeRef::VacuumRelation(n),
            NodeEnum::InlineCodeBlock(n) => NodeRef::InlineCodeBlock(n),
            NodeEnum::CallContext(n) => NodeRef::CallContext(n),
            NodeEnum::Integer(n) => NodeRef::Integer(n),
            NodeEnum::Float(n) => NodeRef::Float(n),
            NodeEnum::Boolean(n) => NodeRef::Boolean(n),
            NodeEnum::String(n) => NodeRef::String(n),
            NodeEnum::BitString(n) => NodeRef::BitString(n),
            NodeEnum::List(n) => NodeRef::List(n),
            NodeEnum::IntList(n) => NodeRef::IntList(n),
            NodeEnum::OidList(n) => NodeRef::OidList(n),
            NodeEnum::MergeStmt(n) => NodeRef::MergeStmt(n),
            NodeEnum::MergeAction(n) => NodeRef::MergeAction(n),
            NodeEnum::AlterDatabaseRefreshCollStmt(n) => NodeRef::AlterDatabaseRefreshCollStmt(n),
            NodeEnum::ReturnStmt(n) => NodeRef::ReturnStmt(n),
            NodeEnum::PlassignStmt(n) => NodeRef::PlassignStmt(n),
            NodeEnum::StatsElem(n) => NodeRef::StatsElem(n),
            NodeEnum::CtesearchClause(n) => NodeRef::CtesearchClause(n),
            NodeEnum::CtecycleClause(n) => NodeRef::CtecycleClause(n),
            NodeEnum::MergeWhenClause(n) => NodeRef::MergeWhenClause(n),
            NodeEnum::PublicationObjSpec(n) => NodeRef::PublicationObjSpec(n),
            NodeEnum::PublicationTable(n) => NodeRef::PublicationTable(n),
            NodeEnum::JsonFormat(n) => NodeRef::JsonFormat(n),
            NodeEnum::JsonReturning(n) => NodeRef::JsonReturning(n),
            NodeEnum::JsonValueExpr(n) => NodeRef::JsonValueExpr(n),
            NodeEnum::JsonConstructorExpr(n) => NodeRef::JsonConstructorExpr(n),
            NodeEnum::JsonIsPredicate(n) => NodeRef::JsonIsPredicate(n),
            NodeEnum::JsonOutput(n) => NodeRef::JsonOutput(n),
            NodeEnum::JsonKeyValue(n) => NodeRef::JsonKeyValue(n),
            NodeEnum::JsonObjectConstructor(n) => NodeRef::JsonObjectConstructor(n),
            NodeEnum::JsonArrayConstructor(n) => NodeRef::JsonArrayConstructor(n),
            NodeEnum::JsonArrayQueryConstructor(n) => NodeRef::JsonArrayQueryConstructor(n),
            NodeEnum::JsonAggConstructor(n) => NodeRef::JsonAggConstructor(n),
            NodeEnum::JsonObjectAgg(n) => NodeRef::JsonObjectAgg(n),
            NodeEnum::JsonArrayAgg(n) => NodeRef::JsonArrayAgg(n),
            NodeEnum::RtepermissionInfo(n) => NodeRef::RtepermissionInfo(n),
            NodeEnum::WindowFuncRunCondition(n) => NodeRef::WindowFuncRunCondition(n),
            NodeEnum::MergeSupportFunc(n) => NodeRef::MergeSupportFunc(n),
            NodeEnum::JsonBehavior(n) => NodeRef::JsonBehavior(n),
            NodeEnum::JsonExpr(n) => NodeRef::JsonExpr(n),
            NodeEnum::JsonTablePath(n) => NodeRef::JsonTablePath(n),
            NodeEnum::JsonTablePathScan(n) => NodeRef::JsonTablePathScan(n),
            NodeEnum::JsonTableSiblingJoin(n) => NodeRef::JsonTableSiblingJoin(n),
            NodeEnum::SinglePartitionSpec(n) => NodeRef::SinglePartitionSpec(n),
            NodeEnum::JsonArgument(n) => NodeRef::JsonArgument(n),
            NodeEnum::JsonFuncExpr(n) => NodeRef::JsonFuncExpr(n),
            NodeEnum::JsonTablePathSpec(n) => NodeRef::JsonTablePathSpec(n),
            NodeEnum::JsonTable(n) => NodeRef::JsonTable(n),
            NodeEnum::JsonTableColumn(n) => NodeRef::JsonTableColumn(n),
            NodeEnum::JsonParseExpr(n) => NodeRef::JsonParseExpr(n),
            NodeEnum::JsonScalarExpr(n) => NodeRef::JsonScalarExpr(n),
            NodeEnum::JsonSerializeExpr(n) => NodeRef::JsonSerializeExpr(n),
        }
    }

    pub fn to_mut(&mut self) -> NodeMut {
        match self {
            NodeEnum::Alias(n) => NodeMut::Alias(n as *mut _),
            NodeEnum::RangeVar(n) => NodeMut::RangeVar(n as *mut _),
            NodeEnum::TableFunc(n) => NodeMut::TableFunc(&mut **n as *mut _),
            NodeEnum::Var(n) => NodeMut::Var(&mut **n as *mut _),
            NodeEnum::Param(n) => NodeMut::Param(&mut **n as *mut _),
            NodeEnum::Aggref(n) => NodeMut::Aggref(&mut **n as *mut _),
            NodeEnum::GroupingFunc(n) => NodeMut::GroupingFunc(&mut **n as *mut _),
            NodeEnum::WindowFunc(n) => NodeMut::WindowFunc(&mut **n as *mut _),
            NodeEnum::SubscriptingRef(n) => NodeMut::SubscriptingRef(&mut **n as *mut _),
            NodeEnum::FuncExpr(n) => NodeMut::FuncExpr(&mut **n as *mut _),
            NodeEnum::NamedArgExpr(n) => NodeMut::NamedArgExpr(&mut **n as *mut _),
            NodeEnum::OpExpr(n) => NodeMut::OpExpr(&mut **n as *mut _),
            NodeEnum::DistinctExpr(n) => NodeMut::DistinctExpr(&mut **n as *mut _),
            NodeEnum::NullIfExpr(n) => NodeMut::NullIfExpr(&mut **n as *mut _),
            NodeEnum::ScalarArrayOpExpr(n) => NodeMut::ScalarArrayOpExpr(&mut **n as *mut _),
            NodeEnum::BoolExpr(n) => NodeMut::BoolExpr(&mut **n as *mut _),
            NodeEnum::SubLink(n) => NodeMut::SubLink(&mut **n as *mut _),
            NodeEnum::SubPlan(n) => NodeMut::SubPlan(&mut **n as *mut _),
            NodeEnum::AlternativeSubPlan(n) => NodeMut::AlternativeSubPlan(&mut **n as *mut _),
            NodeEnum::FieldSelect(n) => NodeMut::FieldSelect(&mut **n as *mut _),
            NodeEnum::FieldStore(n) => NodeMut::FieldStore(&mut **n as *mut _),
            NodeEnum::RelabelType(n) => NodeMut::RelabelType(&mut **n as *mut _),
            NodeEnum::CoerceViaIo(n) => NodeMut::CoerceViaIo(&mut **n as *mut _),
            NodeEnum::ArrayCoerceExpr(n) => NodeMut::ArrayCoerceExpr(&mut **n as *mut _),
            NodeEnum::ConvertRowtypeExpr(n) => NodeMut::ConvertRowtypeExpr(&mut **n as *mut _),
            NodeEnum::CollateExpr(n) => NodeMut::CollateExpr(&mut **n as *mut _),
            NodeEnum::CaseExpr(n) => NodeMut::CaseExpr(&mut **n as *mut _),
            NodeEnum::CaseWhen(n) => NodeMut::CaseWhen(&mut **n as *mut _),
            NodeEnum::CaseTestExpr(n) => NodeMut::CaseTestExpr(&mut **n as *mut _),
            NodeEnum::ArrayExpr(n) => NodeMut::ArrayExpr(&mut **n as *mut _),
            NodeEnum::RowExpr(n) => NodeMut::RowExpr(&mut **n as *mut _),
            NodeEnum::RowCompareExpr(n) => NodeMut::RowCompareExpr(&mut **n as *mut _),
            NodeEnum::CoalesceExpr(n) => NodeMut::CoalesceExpr(&mut **n as *mut _),
            NodeEnum::MinMaxExpr(n) => NodeMut::MinMaxExpr(&mut **n as *mut _),
            NodeEnum::SqlvalueFunction(n) => NodeMut::SqlvalueFunction(&mut **n as *mut _),
            NodeEnum::XmlExpr(n) => NodeMut::XmlExpr(&mut **n as *mut _),
            NodeEnum::NullTest(n) => NodeMut::NullTest(&mut **n as *mut _),
            NodeEnum::BooleanTest(n) => NodeMut::BooleanTest(&mut **n as *mut _),
            NodeEnum::CoerceToDomain(n) => NodeMut::CoerceToDomain(&mut **n as *mut _),
            NodeEnum::CoerceToDomainValue(n) => NodeMut::CoerceToDomainValue(&mut **n as *mut _),
            NodeEnum::SetToDefault(n) => NodeMut::SetToDefault(&mut **n as *mut _),
            NodeEnum::CurrentOfExpr(n) => NodeMut::CurrentOfExpr(&mut **n as *mut _),
            NodeEnum::NextValueExpr(n) => NodeMut::NextValueExpr(&mut **n as *mut _),
            NodeEnum::InferenceElem(n) => NodeMut::InferenceElem(&mut **n as *mut _),
            NodeEnum::TargetEntry(n) => NodeMut::TargetEntry(&mut **n as *mut _),
            NodeEnum::RangeTblRef(n) => NodeMut::RangeTblRef(n as *mut _),
            NodeEnum::JoinExpr(n) => NodeMut::JoinExpr(&mut **n as *mut _),
            NodeEnum::FromExpr(n) => NodeMut::FromExpr(&mut **n as *mut _),
            NodeEnum::OnConflictExpr(n) => NodeMut::OnConflictExpr(&mut **n as *mut _),
            NodeEnum::IntoClause(n) => NodeMut::IntoClause(&mut **n as *mut _),
            NodeEnum::RawStmt(n) => NodeMut::RawStmt(&mut **n as *mut _),
            NodeEnum::Query(n) => NodeMut::Query(&mut **n as *mut _),
            NodeEnum::InsertStmt(n) => NodeMut::InsertStmt(&mut **n as *mut _),
            NodeEnum::DeleteStmt(n) => NodeMut::DeleteStmt(&mut **n as *mut _),
            NodeEnum::UpdateStmt(n) => NodeMut::UpdateStmt(&mut **n as *mut _),
            NodeEnum::SelectStmt(n) => NodeMut::SelectStmt(&mut **n as *mut _),
            NodeEnum::AlterTableStmt(n) => NodeMut::AlterTableStmt(n as *mut _),
            NodeEnum::AlterTableCmd(n) => NodeMut::AlterTableCmd(&mut **n as *mut _),
            NodeEnum::AlterDomainStmt(n) => NodeMut::AlterDomainStmt(&mut **n as *mut _),
            NodeEnum::SetOperationStmt(n) => NodeMut::SetOperationStmt(&mut **n as *mut _),
            NodeEnum::GrantStmt(n) => NodeMut::GrantStmt(n as *mut _),
            NodeEnum::GrantRoleStmt(n) => NodeMut::GrantRoleStmt(n as *mut _),
            NodeEnum::AlterDefaultPrivilegesStmt(n) => NodeMut::AlterDefaultPrivilegesStmt(n as *mut _),
            NodeEnum::ClosePortalStmt(n) => NodeMut::ClosePortalStmt(n as *mut _),
            NodeEnum::ClusterStmt(n) => NodeMut::ClusterStmt(n as *mut _),
            NodeEnum::CopyStmt(n) => NodeMut::CopyStmt(&mut **n as *mut _),
            NodeEnum::CreateStmt(n) => NodeMut::CreateStmt(n as *mut _),
            NodeEnum::DefineStmt(n) => NodeMut::DefineStmt(n as *mut _),
            NodeEnum::DropStmt(n) => NodeMut::DropStmt(n as *mut _),
            NodeEnum::TruncateStmt(n) => NodeMut::TruncateStmt(n as *mut _),
            NodeEnum::CommentStmt(n) => NodeMut::CommentStmt(&mut **n as *mut _),
            NodeEnum::FetchStmt(n) => NodeMut::FetchStmt(n as *mut _),
            NodeEnum::IndexStmt(n) => NodeMut::IndexStmt(&mut **n as *mut _),
            NodeEnum::CreateFunctionStmt(n) => NodeMut::CreateFunctionStmt(&mut **n as *mut _),
            NodeEnum::AlterFunctionStmt(n) => NodeMut::AlterFunctionStmt(n as *mut _),
            NodeEnum::DoStmt(n) => NodeMut::DoStmt(n as *mut _),
            NodeEnum::RenameStmt(n) => NodeMut::RenameStmt(&mut **n as *mut _),
            NodeEnum::RuleStmt(n) => NodeMut::RuleStmt(&mut **n as *mut _),
            NodeEnum::NotifyStmt(n) => NodeMut::NotifyStmt(n as *mut _),
            NodeEnum::ListenStmt(n) => NodeMut::ListenStmt(n as *mut _),
            NodeEnum::UnlistenStmt(n) => NodeMut::UnlistenStmt(n as *mut _),
            NodeEnum::TransactionStmt(n) => NodeMut::TransactionStmt(n as *mut _),
            NodeEnum::ViewStmt(n) => NodeMut::ViewStmt(&mut **n as *mut _),
            NodeEnum::LoadStmt(n) => NodeMut::LoadStmt(n as *mut _),
            NodeEnum::CreateDomainStmt(n) => NodeMut::CreateDomainStmt(&mut **n as *mut _),
            NodeEnum::CreatedbStmt(n) => NodeMut::CreatedbStmt(n as *mut _),
            NodeEnum::DropdbStmt(n) => NodeMut::DropdbStmt(n as *mut _),
            NodeEnum::VacuumStmt(n) => NodeMut::VacuumStmt(n as *mut _),
            NodeEnum::ExplainStmt(n) => NodeMut::ExplainStmt(&mut **n as *mut _),
            NodeEnum::CreateTableAsStmt(n) => NodeMut::CreateTableAsStmt(&mut **n as *mut _),
            NodeEnum::CreateSeqStmt(n) => NodeMut::CreateSeqStmt(n as *mut _),
            NodeEnum::AlterSeqStmt(n) => NodeMut::AlterSeqStmt(n as *mut _),
            NodeEnum::VariableSetStmt(n) => NodeMut::VariableSetStmt(n as *mut _),
            NodeEnum::VariableShowStmt(n) => NodeMut::VariableShowStmt(n as *mut _),
            NodeEnum::DiscardStmt(n) => NodeMut::DiscardStmt(n as *mut _),
            NodeEnum::CreateTrigStmt(n) => NodeMut::CreateTrigStmt(&mut **n as *mut _),
            NodeEnum::CreatePlangStmt(n) => NodeMut::CreatePlangStmt(n as *mut _),
            NodeEnum::CreateRoleStmt(n) => NodeMut::CreateRoleStmt(n as *mut _),
            NodeEnum::AlterRoleStmt(n) => NodeMut::AlterRoleStmt(n as *mut _),
            NodeEnum::DropRoleStmt(n) => NodeMut::DropRoleStmt(n as *mut _),
            NodeEnum::LockStmt(n) => NodeMut::LockStmt(n as *mut _),
            NodeEnum::ConstraintsSetStmt(n) => NodeMut::ConstraintsSetStmt(n as *mut _),
            NodeEnum::ReindexStmt(n) => NodeMut::ReindexStmt(n as *mut _),
            NodeEnum::CheckPointStmt(n) => NodeMut::CheckPointStmt(n as *mut _),
            NodeEnum::CreateSchemaStmt(n) => NodeMut::CreateSchemaStmt(n as *mut _),
            NodeEnum::AlterDatabaseStmt(n) => NodeMut::AlterDatabaseStmt(n as *mut _),
            NodeEnum::AlterDatabaseSetStmt(n) => NodeMut::AlterDatabaseSetStmt(n as *mut _),
            NodeEnum::AlterRoleSetStmt(n) => NodeMut::AlterRoleSetStmt(n as *mut _),
            NodeEnum::CreateConversionStmt(n) => NodeMut::CreateConversionStmt(n as *mut _),
            NodeEnum::CreateCastStmt(n) => NodeMut::CreateCastStmt(n as *mut _),
            NodeEnum::CreateOpClassStmt(n) => NodeMut::CreateOpClassStmt(n as *mut _),
            NodeEnum::CreateOpFamilyStmt(n) => NodeMut::CreateOpFamilyStmt(n as *mut _),
            NodeEnum::AlterOpFamilyStmt(n) => NodeMut::AlterOpFamilyStmt(n as *mut _),
            NodeEnum::PrepareStmt(n) => NodeMut::PrepareStmt(&mut **n as *mut _),
            NodeEnum::ExecuteStmt(n) => NodeMut::ExecuteStmt(n as *mut _),
            NodeEnum::DeallocateStmt(n) => NodeMut::DeallocateStmt(n as *mut _),
            NodeEnum::DeclareCursorStmt(n) => NodeMut::DeclareCursorStmt(&mut **n as *mut _),
            NodeEnum::CreateTableSpaceStmt(n) => NodeMut::CreateTableSpaceStmt(n as *mut _),
            NodeEnum::DropTableSpaceStmt(n) => NodeMut::DropTableSpaceStmt(n as *mut _),
            NodeEnum::AlterObjectDependsStmt(n) => NodeMut::AlterObjectDependsStmt(&mut **n as *mut _),
            NodeEnum::AlterObjectSchemaStmt(n) => NodeMut::AlterObjectSchemaStmt(&mut **n as *mut _),
            NodeEnum::AlterOwnerStmt(n) => NodeMut::AlterOwnerStmt(&mut **n as *mut _),
            NodeEnum::AlterOperatorStmt(n) => NodeMut::AlterOperatorStmt(n as *mut _),
            NodeEnum::AlterTypeStmt(n) => NodeMut::AlterTypeStmt(n as *mut _),
            NodeEnum::DropOwnedStmt(n) => NodeMut::DropOwnedStmt(n as *mut _),
            NodeEnum::ReassignOwnedStmt(n) => NodeMut::ReassignOwnedStmt(n as *mut _),
            NodeEnum::CompositeTypeStmt(n) => NodeMut::CompositeTypeStmt(n as *mut _),
            NodeEnum::CreateEnumStmt(n) => NodeMut::CreateEnumStmt(n as *mut _),
            NodeEnum::CreateRangeStmt(n) => NodeMut::CreateRangeStmt(n as *mut _),
            NodeEnum::AlterEnumStmt(n) => NodeMut::AlterEnumStmt(n as *mut _),
            NodeEnum::AlterTsdictionaryStmt(n) => NodeMut::AlterTsdictionaryStmt(n as *mut _),
            NodeEnum::AlterTsconfigurationStmt(n) => NodeMut::AlterTsconfigurationStmt(n as *mut _),
            NodeEnum::CreateFdwStmt(n) => NodeMut::CreateFdwStmt(n as *mut _),
            NodeEnum::AlterFdwStmt(n) => NodeMut::AlterFdwStmt(n as *mut _),
            NodeEnum::CreateForeignServerStmt(n) => NodeMut::CreateForeignServerStmt(n as *mut _),
            NodeEnum::AlterForeignServerStmt(n) => NodeMut::AlterForeignServerStmt(n as *mut _),
            NodeEnum::CreateUserMappingStmt(n) => NodeMut::CreateUserMappingStmt(n as *mut _),
            NodeEnum::AlterUserMappingStmt(n) => NodeMut::AlterUserMappingStmt(n as *mut _),
            NodeEnum::DropUserMappingStmt(n) => NodeMut::DropUserMappingStmt(n as *mut _),
            NodeEnum::AlterTableSpaceOptionsStmt(n) => NodeMut::AlterTableSpaceOptionsStmt(n as *mut _),
            NodeEnum::AlterTableMoveAllStmt(n) => NodeMut::AlterTableMoveAllStmt(n as *mut _),
            NodeEnum::SecLabelStmt(n) => NodeMut::SecLabelStmt(&mut **n as *mut _),
            NodeEnum::CreateForeignTableStmt(n) => NodeMut::CreateForeignTableStmt(n as *mut _),
            NodeEnum::ImportForeignSchemaStmt(n) => NodeMut::ImportForeignSchemaStmt(n as *mut _),
            NodeEnum::CreateExtensionStmt(n) => NodeMut::CreateExtensionStmt(n as *mut _),
            NodeEnum::AlterExtensionStmt(n) => NodeMut::AlterExtensionStmt(n as *mut _),
            NodeEnum::AlterExtensionContentsStmt(n) => NodeMut::AlterExtensionContentsStmt(&mut **n as *mut _),
            NodeEnum::CreateEventTrigStmt(n) => NodeMut::CreateEventTrigStmt(n as *mut _),
            NodeEnum::AlterEventTrigStmt(n) => NodeMut::AlterEventTrigStmt(n as *mut _),
            NodeEnum::RefreshMatViewStmt(n) => NodeMut::RefreshMatViewStmt(n as *mut _),
            NodeEnum::ReplicaIdentityStmt(n) => NodeMut::ReplicaIdentityStmt(n as *mut _),
            NodeEnum::AlterSystemStmt(n) => NodeMut::AlterSystemStmt(n as *mut _),
            NodeEnum::CreatePolicyStmt(n) => NodeMut::CreatePolicyStmt(&mut **n as *mut _),
            NodeEnum::AlterPolicyStmt(n) => NodeMut::AlterPolicyStmt(&mut **n as *mut _),
            NodeEnum::CreateTransformStmt(n) => NodeMut::CreateTransformStmt(n as *mut _),
            NodeEnum::CreateAmStmt(n) => NodeMut::CreateAmStmt(n as *mut _),
            NodeEnum::CreatePublicationStmt(n) => NodeMut::CreatePublicationStmt(n as *mut _),
            NodeEnum::AlterPublicationStmt(n) => NodeMut::AlterPublicationStmt(n as *mut _),
            NodeEnum::CreateSubscriptionStmt(n) => NodeMut::CreateSubscriptionStmt(n as *mut _),
            NodeEnum::AlterSubscriptionStmt(n) => NodeMut::AlterSubscriptionStmt(n as *mut _),
            NodeEnum::DropSubscriptionStmt(n) => NodeMut::DropSubscriptionStmt(n as *mut _),
            NodeEnum::CreateStatsStmt(n) => NodeMut::CreateStatsStmt(n as *mut _),
            NodeEnum::AlterCollationStmt(n) => NodeMut::AlterCollationStmt(n as *mut _),
            NodeEnum::CallStmt(n) => NodeMut::CallStmt(&mut **n as *mut _),
            NodeEnum::AlterStatsStmt(n) => NodeMut::AlterStatsStmt(&mut **n as *mut _),
            NodeEnum::AExpr(n) => NodeMut::AExpr(&mut **n as *mut _),
            NodeEnum::ColumnRef(n) => NodeMut::ColumnRef(n as *mut _),
            NodeEnum::ParamRef(n) => NodeMut::ParamRef(n as *mut _),
            NodeEnum::AConst(n) => NodeMut::AConst(n as *mut _),
            NodeEnum::FuncCall(n) => NodeMut::FuncCall(&mut **n as *mut _),
            NodeEnum::AStar(n) => NodeMut::AStar(n as *mut _),
            NodeEnum::AIndices(n) => NodeMut::AIndices(&mut **n as *mut _),
            NodeEnum::AIndirection(n) => NodeMut::AIndirection(&mut **n as *mut _),
            NodeEnum::AArrayExpr(n) => NodeMut::AArrayExpr(n as *mut _),
            NodeEnum::ResTarget(n) => NodeMut::ResTarget(&mut **n as *mut _),
            NodeEnum::MultiAssignRef(n) => NodeMut::MultiAssignRef(&mut **n as *mut _),
            NodeEnum::TypeCast(n) => NodeMut::TypeCast(&mut **n as *mut _),
            NodeEnum::CollateClause(n) => NodeMut::CollateClause(&mut **n as *mut _),
            NodeEnum::SortBy(n) => NodeMut::SortBy(&mut **n as *mut _),
            NodeEnum::WindowDef(n) => NodeMut::WindowDef(&mut **n as *mut _),
            NodeEnum::RangeSubselect(n) => NodeMut::RangeSubselect(&mut **n as *mut _),
            NodeEnum::RangeFunction(n) => NodeMut::RangeFunction(n as *mut _),
            NodeEnum::RangeTableSample(n) => NodeMut::RangeTableSample(&mut **n as *mut _),
            NodeEnum::RangeTableFunc(n) => NodeMut::RangeTableFunc(&mut **n as *mut _),
            NodeEnum::RangeTableFuncCol(n) => NodeMut::RangeTableFuncCol(&mut **n as *mut _),
            NodeEnum::TypeName(n) => NodeMut::TypeName(n as *mut _),
            NodeEnum::ColumnDef(n) => NodeMut::ColumnDef(&mut **n as *mut _),
            NodeEnum::IndexElem(n) => NodeMut::IndexElem(&mut **n as *mut _),
            NodeEnum::Constraint(n) => NodeMut::Constraint(&mut **n as *mut _),
            NodeEnum::DefElem(n) => NodeMut::DefElem(&mut **n as *mut _),
            NodeEnum::RangeTblEntry(n) => NodeMut::RangeTblEntry(&mut **n as *mut _),
            NodeEnum::RangeTblFunction(n) => NodeMut::RangeTblFunction(&mut **n as *mut _),
            NodeEnum::TableSampleClause(n) => NodeMut::TableSampleClause(&mut **n as *mut _),
            NodeEnum::WithCheckOption(n) => NodeMut::WithCheckOption(&mut **n as *mut _),
            NodeEnum::SortGroupClause(n) => NodeMut::SortGroupClause(n as *mut _),
            NodeEnum::GroupingSet(n) => NodeMut::GroupingSet(n as *mut _),
            NodeEnum::WindowClause(n) => NodeMut::WindowClause(&mut **n as *mut _),
            NodeEnum::ObjectWithArgs(n) => NodeMut::ObjectWithArgs(n as *mut _),
            NodeEnum::AccessPriv(n) => NodeMut::AccessPriv(n as *mut _),
            NodeEnum::CreateOpClassItem(n) => NodeMut::CreateOpClassItem(n as *mut _),
            NodeEnum::TableLikeClause(n) => NodeMut::TableLikeClause(n as *mut _),
            NodeEnum::FunctionParameter(n) => NodeMut::FunctionParameter(&mut **n as *mut _),
            NodeEnum::LockingClause(n) => NodeMut::LockingClause(n as *mut _),
            NodeEnum::RowMarkClause(n) => NodeMut::RowMarkClause(n as *mut _),
            NodeEnum::XmlSerialize(n) => NodeMut::XmlSerialize(&mut **n as *mut _),
            NodeEnum::WithClause(n) => NodeMut::WithClause(n as *mut _),
            NodeEnum::InferClause(n) => NodeMut::InferClause(&mut **n as *mut _),
            NodeEnum::OnConflictClause(n) => NodeMut::OnConflictClause(&mut **n as *mut _),
            NodeEnum::CommonTableExpr(n) => NodeMut::CommonTableExpr(&mut **n as *mut _),
            NodeEnum::RoleSpec(n) => NodeMut::RoleSpec(n as *mut _),
            NodeEnum::TriggerTransition(n) => NodeMut::TriggerTransition(n as *mut _),
            NodeEnum::PartitionElem(n) => NodeMut::PartitionElem(&mut **n as *mut _),
            NodeEnum::PartitionSpec(n) => NodeMut::PartitionSpec(n as *mut _),
            NodeEnum::PartitionBoundSpec(n) => NodeMut::PartitionBoundSpec(n as *mut _),
            NodeEnum::PartitionRangeDatum(n) => NodeMut::PartitionRangeDatum(&mut **n as *mut _),
            NodeEnum::PartitionCmd(n) => NodeMut::PartitionCmd(n as *mut _),
            NodeEnum::VacuumRelation(n) => NodeMut::VacuumRelation(n as *mut _),
            NodeEnum::InlineCodeBlock(n) => NodeMut::InlineCodeBlock(n as *mut _),
            NodeEnum::CallContext(n) => NodeMut::CallContext(n as *mut _),
            NodeEnum::Integer(n) => NodeMut::Integer(n as *mut _),
            NodeEnum::Float(n) => NodeMut::Float(n as *mut _),
            NodeEnum::Boolean(n) => NodeMut::Boolean(n as *mut _),
            NodeEnum::String(n) => NodeMut::String(n as *mut _),
            NodeEnum::BitString(n) => NodeMut::BitString(n as *mut _),
            NodeEnum::List(n) => NodeMut::List(n as *mut _),
            NodeEnum::IntList(n) => NodeMut::IntList(n as *mut _),
            NodeEnum::OidList(n) => NodeMut::OidList(n as *mut _),
            NodeEnum::MergeStmt(n) => NodeMut::MergeStmt(&mut **n as *mut _),
            NodeEnum::MergeAction(n) => NodeMut::MergeAction(&mut **n as *mut _),
            NodeEnum::AlterDatabaseRefreshCollStmt(n) => NodeMut::AlterDatabaseRefreshCollStmt(n as *mut _),
            NodeEnum::ReturnStmt(n) => NodeMut::ReturnStmt(&mut **n as *mut _),
            NodeEnum::PlassignStmt(n) => NodeMut::PlassignStmt(&mut **n as *mut _),
            NodeEnum::StatsElem(n) => NodeMut::StatsElem(&mut **n as *mut _),
            NodeEnum::CtesearchClause(n) => NodeMut::CtesearchClause(n as *mut _),
            NodeEnum::CtecycleClause(n) => NodeMut::CtecycleClause(&mut **n as *mut _),
            NodeEnum::MergeWhenClause(n) => NodeMut::MergeWhenClause(&mut **n as *mut _),
            NodeEnum::PublicationObjSpec(n) => NodeMut::PublicationObjSpec(&mut **n as *mut _),
            NodeEnum::PublicationTable(n) => NodeMut::PublicationTable(&mut **n as *mut _),
            NodeEnum::JsonFormat(n) => NodeMut::JsonFormat(&mut *n as *mut _),
            NodeEnum::JsonReturning(n) => NodeMut::JsonReturning(&mut *n as *mut _),
            NodeEnum::JsonValueExpr(n) => NodeMut::JsonValueExpr(&mut **n as *mut _),
            NodeEnum::JsonConstructorExpr(n) => NodeMut::JsonConstructorExpr(&mut **n as *mut _),
            NodeEnum::JsonIsPredicate(n) => NodeMut::JsonIsPredicate(&mut **n as *mut _),
            NodeEnum::JsonOutput(n) => NodeMut::JsonOutput(&mut *n as *mut _),
            NodeEnum::JsonKeyValue(n) => NodeMut::JsonKeyValue(&mut **n as *mut _),
            NodeEnum::JsonObjectConstructor(n) => NodeMut::JsonObjectConstructor(&mut *n as *mut _),
            NodeEnum::JsonArrayConstructor(n) => NodeMut::JsonArrayConstructor(&mut *n as *mut _),
            NodeEnum::JsonArrayQueryConstructor(n) => NodeMut::JsonArrayQueryConstructor(&mut **n as *mut _),
            NodeEnum::JsonAggConstructor(n) => NodeMut::JsonAggConstructor(&mut **n as *mut _),
            NodeEnum::JsonObjectAgg(n) => NodeMut::JsonObjectAgg(&mut **n as *mut _),
            NodeEnum::JsonArrayAgg(n) => NodeMut::JsonArrayAgg(&mut **n as *mut _),
            NodeEnum::RtepermissionInfo(n) => NodeMut::RtepermissionInfo(&mut *n as *mut _),
            NodeEnum::WindowFuncRunCondition(n) => NodeMut::WindowFuncRunCondition(&mut **n as *mut _),
            NodeEnum::MergeSupportFunc(n) => NodeMut::MergeSupportFunc(&mut **n as *mut _),
            NodeEnum::JsonBehavior(n) => NodeMut::JsonBehavior(&mut **n as *mut _),
            NodeEnum::JsonExpr(n) => NodeMut::JsonExpr(&mut **n as *mut _),
            NodeEnum::JsonTablePath(n) => NodeMut::JsonTablePath(&mut *n as *mut _),
            NodeEnum::JsonTablePathScan(n) => NodeMut::JsonTablePathScan(&mut **n as *mut _),
            NodeEnum::JsonTableSiblingJoin(n) => NodeMut::JsonTableSiblingJoin(&mut **n as *mut _),
            NodeEnum::SinglePartitionSpec(n) => NodeMut::SinglePartitionSpec(&mut *n as *mut _),
            NodeEnum::JsonArgument(n) => NodeMut::JsonArgument(&mut **n as *mut _),
            NodeEnum::JsonFuncExpr(n) => NodeMut::JsonFuncExpr(&mut **n as *mut _),
            NodeEnum::JsonTablePathSpec(n) => NodeMut::JsonTablePathSpec(&mut **n as *mut _),
            NodeEnum::JsonTable(n) => NodeMut::JsonTable(&mut **n as *mut _),
            NodeEnum::JsonTableColumn(n) => NodeMut::JsonTableColumn(&mut **n as *mut _),
            NodeEnum::JsonParseExpr(n) => NodeMut::JsonParseExpr(&mut **n as *mut _),
            NodeEnum::JsonScalarExpr(n) => NodeMut::JsonScalarExpr(&mut **n as *mut _),
            NodeEnum::JsonSerializeExpr(n) => NodeMut::JsonSerializeExpr(&mut **n as *mut _),
        }
    }
}
