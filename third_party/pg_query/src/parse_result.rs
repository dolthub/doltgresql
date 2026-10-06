use std::collections::HashMap;
use std::collections::HashSet;
use std::iter::FromIterator;
use std::string::String;

use itertools::join;

use crate::*;

macro_rules! cast {
    ($target: expr, $pat: path) => {{
        if let $pat(a) = $target {
            // #1
            a
        } else {
            panic!("mismatch variant when cast to {}", stringify!($pat)); // #2
        }
    }};
}

impl protobuf::ParseResult {
    pub fn deparse(&self) -> Result<String> {
        crate::deparse(self)
    }

    // Note: this doesn't iterate over every possible node type, since we only care about a subset of nodes.
    pub fn nodes(&self) -> Vec<(NodeRef<'_>, i32, Context, bool)> {
        self.stmts
            .iter()
            .filter_map(|s|
            // RawStmt  ->  Node   ->    NodeEnum           ->              NodeRef
            s.stmt.as_ref().and_then(|s| s.node.as_ref()).map(|n| n.nodes()))
            .flatten()
            .collect()
    }

    /// Returns a mutable reference to nested nodes.
    ///
    /// # Safety
    ///
    /// The caller may have to deal with dangling pointers, and passing an
    /// invalid tree back to libpg_query may cause it to panic.
    pub unsafe fn nodes_mut(&mut self) -> Vec<(NodeMut, i32, Context)> {
        self.stmts
            .iter_mut()
            .filter_map(|s|
            // RawStmt  ->  Node   ->    NodeEnum           ->              NodeMut
            s.stmt.as_mut().and_then(|s| s.node.as_mut()).map(|n| n.nodes_mut()))
            .flatten()
            .collect()
    }
}

/// Result from calling [parse]
#[derive(Debug)]
pub struct ParseResult {
    pub protobuf: protobuf::ParseResult,
    pub warnings: Vec<String>,
    pub tables: Vec<(String, Context)>,
    pub aliases: HashMap<String, String>,
    pub cte_names: Vec<String>,
    pub functions: Vec<(String, Context)>,
    pub filter_columns: Vec<(Option<String>, String)>,
}

impl ParseResult {
    pub fn new(protobuf: protobuf::ParseResult, stderr: String) -> Self {
        let warnings = stderr.lines().filter_map(|l| if l.starts_with("WARNING") { Some(l.trim().into()) } else { None }).collect();
        let mut tables: HashSet<(String, Context)> = HashSet::new();
        let mut aliases: HashMap<String, String> = HashMap::new();
        let mut cte_names: HashSet<String> = HashSet::new();
        let mut functions: HashSet<(String, Context)> = HashSet::new();
        let mut filter_columns: HashSet<(Option<String>, String)> = HashSet::new();

        for (node, _depth, context, has_filter_columns) in protobuf.nodes().into_iter() {
            match node {
                NodeRef::CommonTableExpr(s) => {
                    cte_names.insert(s.ctename.to_owned());
                }
                NodeRef::RangeVar(v) => {
                    // TODO: this incorrectly returns no tables: parse('with f as (select * from f limit 1) select * from f')
                    let table = if !v.schemaname.is_empty() { format!("{}.{}", v.schemaname, v.relname) } else { v.relname.to_owned() };
                    if cte_names.contains(&table) {
                        continue;
                    }
                    tables.insert((table.to_owned(), context));
                    v.alias.as_ref().and_then(|alias| aliases.insert(alias.aliasname.to_owned(), table));
                }
                NodeRef::FuncCall(c) => {
                    let funcname = join(c.funcname.iter().filter_map(|n| n.node.as_ref().map(|n| &cast!(n, NodeEnum::String).sval)), ".");
                    functions.insert((funcname, Context::Call));
                }
                NodeRef::DropStmt(s) => {
                    match protobuf::ObjectType::try_from(s.remove_type) {
                        Ok(protobuf::ObjectType::ObjectTable) => {
                            for o in &s.objects {
                                if let Some(NodeEnum::List(list)) = &o.node {
                                    let table =
                                        join(list.items.iter().filter_map(|i| i.node.as_ref().map(|n| &cast!(n, NodeEnum::String).sval)), ".");
                                    tables.insert((table, Context::DDL));
                                };
                            }
                        }
                        Ok(protobuf::ObjectType::ObjectRule) | Ok(protobuf::ObjectType::ObjectTrigger) => {
                            for o in &s.objects {
                                if let Some(NodeEnum::List(list)) = &o.node {
                                    // Unlike ObjectTable, this ignores the last string (the rule/trigger name)
                                    let table = join(
                                        list.items[0..list.items.len() - 1]
                                            .iter()
                                            .filter_map(|i| i.node.as_ref().map(|n| &cast!(n, NodeEnum::String).sval)),
                                        ".",
                                    );
                                    tables.insert((table, Context::DDL));
                                };
                            }
                        }
                        Ok(protobuf::ObjectType::ObjectFunction) => {
                            // Only one function can be dropped in a statement
                            if let Some(NodeEnum::ObjectWithArgs(object)) = &s.objects[0].node {
                                if let Some(NodeEnum::String(string)) = &object.objname[0].node {
                                    functions.insert((string.sval.to_string(), Context::DDL));
                                }
                            }
                        }
                        _ => (),
                    }
                }
                NodeRef::CreateFunctionStmt(s) => {
                    if let Some(NodeEnum::String(string)) = &s.funcname[0].node {
                        functions.insert((string.sval.to_string(), Context::DDL));
                    }
                }
                NodeRef::RenameStmt(s) => {
                    if let Ok(protobuf::ObjectType::ObjectFunction) = protobuf::ObjectType::try_from(s.rename_type) {
                        if let Some(object) = &s.object {
                            if let Some(NodeEnum::ObjectWithArgs(object)) = &object.node {
                                if let Some(NodeEnum::String(string)) = &object.objname[0].node {
                                    functions.insert((string.sval.to_string(), Context::DDL));
                                    functions.insert((s.newname.to_string(), Context::DDL));
                                }
                            }
                        }
                    }
                }
                NodeRef::ColumnRef(c) => {
                    if !has_filter_columns {
                        continue;
                    }
                    let f: Vec<String> = c
                        .fields
                        .iter()
                        .filter_map(|n| match n.node.as_ref() {
                            Some(NodeEnum::String(s)) => Some(s.sval.to_string()),
                            _ => None,
                        })
                        .rev()
                        .collect();
                    if f.len() > 0 {
                        filter_columns.insert((f.get(1).cloned(), f[0].to_string()));
                    }
                }
                _ => (),
            }
        }

        Self {
            protobuf,
            warnings,
            tables: Vec::from_iter(tables),
            aliases,
            cte_names: Vec::from_iter(cte_names),
            functions: Vec::from_iter(functions),
            filter_columns: Vec::from_iter(filter_columns),
        }
    }

    /// Returns all referenced tables in the query
    pub fn tables(&self) -> Vec<String> {
        let mut tables = HashSet::new();
        self.tables.iter().for_each(|(t, _c)| {
            tables.insert(t.to_string());
        });
        Vec::from_iter(tables)
    }

    /// Returns only tables that were selected from
    pub fn select_tables(&self) -> Vec<String> {
        self.tables
            .iter()
            .filter_map(|(table, context)| match context {
                Context::Select => Some(table.to_string()),
                _ => None,
            })
            .collect()
    }

    /// Returns only tables that were modified by the query
    pub fn dml_tables(&self) -> Vec<String> {
        self.tables
            .iter()
            .filter_map(|(table, context)| match context {
                Context::DML => Some(table.to_string()),
                _ => None,
            })
            .collect()
    }

    /// Returns only tables that were modified by DDL statements
    pub fn ddl_tables(&self) -> Vec<String> {
        self.tables
            .iter()
            .filter_map(|(table, context)| match context {
                Context::DDL => Some(table.to_string()),
                _ => None,
            })
            .collect()
    }

    /// Returns all function references
    pub fn functions(&self) -> Vec<String> {
        let mut functions = HashSet::new();
        self.functions.iter().for_each(|(f, _c)| {
            functions.insert(f.to_string());
        });
        Vec::from_iter(functions)
    }

    /// Returns DDL functions
    pub fn ddl_functions(&self) -> Vec<String> {
        self.functions
            .iter()
            .filter_map(|(function, context)| match context {
                Context::DDL => Some(function.to_string()),
                _ => None,
            })
            .collect()
    }

    /// Returns functions that were called
    pub fn call_functions(&self) -> Vec<String> {
        self.functions
            .iter()
            .filter_map(|(function, context)| match context {
                Context::Call => Some(function.to_string()),
                _ => None,
            })
            .collect()
    }

    /// Converts the parsed query back into a SQL string
    pub fn deparse(&self) -> Result<String> {
        crate::deparse(&self.protobuf)
    }

    /// Intelligently truncates queries to a max length.
    ///
    /// # Example
    ///
    /// ```rust
    /// let query = "INSERT INTO \"x\" (a, b, c, d, e, f) VALUES ($1)";
    /// let result = pg_query::parse(query).unwrap();
    /// assert_eq!(result.truncate(32).unwrap(), "INSERT INTO x (...) VALUES (...)")
    /// ```
    pub fn truncate(&self, max_length: usize) -> Result<String> {
        crate::truncate(&self.protobuf, max_length)
    }

    /// Returns all statement types in the query
    pub fn statement_types(&self) -> Vec<&str> {
        self.protobuf
            .stmts
            .iter()
            .filter_map(|s| match s.stmt.as_ref().and_then(|s| s.node.as_ref()) {
                Some(NodeEnum::InsertStmt(..)) => Some("InsertStmt"),
                Some(NodeEnum::DeleteStmt(..)) => Some("DeleteStmt"),
                Some(NodeEnum::UpdateStmt(..)) => Some("UpdateStmt"),
                Some(NodeEnum::SelectStmt(..)) => Some("SelectStmt"),
                Some(NodeEnum::MergeStmt(..)) => Some("MergeStmt"),
                Some(NodeEnum::AlterTableStmt(..)) => Some("AlterTableStmt"),
                Some(NodeEnum::AlterTableCmd(..)) => Some("AlterTableCmd"),
                Some(NodeEnum::AlterDomainStmt(..)) => Some("AlterDomainStmt"),
                Some(NodeEnum::SetOperationStmt(..)) => Some("SetOperationStmt"),
                Some(NodeEnum::GrantStmt(..)) => Some("GrantStmt"),
                Some(NodeEnum::GrantRoleStmt(..)) => Some("GrantRoleStmt"),
                Some(NodeEnum::AlterDefaultPrivilegesStmt(..)) => Some("AlterDefaultPrivilegesStmt"),
                Some(NodeEnum::ClosePortalStmt(..)) => Some("ClosePortalStmt"),
                Some(NodeEnum::ClusterStmt(..)) => Some("ClusterStmt"),
                Some(NodeEnum::CopyStmt(..)) => Some("CopyStmt"),
                Some(NodeEnum::CreateStmt(..)) => Some("CreateStmt"),
                Some(NodeEnum::DefineStmt(..)) => Some("DefineStmt"),
                Some(NodeEnum::DropStmt(..)) => Some("DropStmt"),
                Some(NodeEnum::TruncateStmt(..)) => Some("TruncateStmt"),
                Some(NodeEnum::CommentStmt(..)) => Some("CommentStmt"),
                Some(NodeEnum::FetchStmt(..)) => Some("FetchStmt"),
                Some(NodeEnum::IndexStmt(..)) => Some("IndexStmt"),
                Some(NodeEnum::CreateFunctionStmt(..)) => Some("CreateFunctionStmt"),
                Some(NodeEnum::AlterFunctionStmt(..)) => Some("AlterFunctionStmt"),
                Some(NodeEnum::DoStmt(..)) => Some("DoStmt"),
                Some(NodeEnum::RenameStmt(..)) => Some("RenameStmt"),
                Some(NodeEnum::RuleStmt(..)) => Some("RuleStmt"),
                Some(NodeEnum::NotifyStmt(..)) => Some("NotifyStmt"),
                Some(NodeEnum::ListenStmt(..)) => Some("ListenStmt"),
                Some(NodeEnum::UnlistenStmt(..)) => Some("UnlistenStmt"),
                Some(NodeEnum::TransactionStmt(..)) => Some("TransactionStmt"),
                Some(NodeEnum::ViewStmt(..)) => Some("ViewStmt"),
                Some(NodeEnum::LoadStmt(..)) => Some("LoadStmt"),
                Some(NodeEnum::CreateDomainStmt(..)) => Some("CreateDomainStmt"),
                Some(NodeEnum::CreatedbStmt(..)) => Some("CreatedbStmt"),
                Some(NodeEnum::DropdbStmt(..)) => Some("DropdbStmt"),
                Some(NodeEnum::VacuumStmt(..)) => Some("VacuumStmt"),
                Some(NodeEnum::ExplainStmt(..)) => Some("ExplainStmt"),
                Some(NodeEnum::CreateTableAsStmt(..)) => Some("CreateTableAsStmt"),
                Some(NodeEnum::CreateSeqStmt(..)) => Some("CreateSeqStmt"),
                Some(NodeEnum::AlterSeqStmt(..)) => Some("AlterSeqStmt"),
                Some(NodeEnum::VariableSetStmt(..)) => Some("VariableSetStmt"),
                Some(NodeEnum::VariableShowStmt(..)) => Some("VariableShowStmt"),
                Some(NodeEnum::DiscardStmt(..)) => Some("DiscardStmt"),
                Some(NodeEnum::CreateTrigStmt(..)) => Some("CreateTrigStmt"),
                // CreatePLangStmt is capitalized differently to match C implementation.
                Some(NodeEnum::CreatePlangStmt(..)) => Some("CreatePLangStmt"),
                Some(NodeEnum::CreateRoleStmt(..)) => Some("CreateRoleStmt"),
                Some(NodeEnum::AlterRoleStmt(..)) => Some("AlterRoleStmt"),
                Some(NodeEnum::DropRoleStmt(..)) => Some("DropRoleStmt"),
                Some(NodeEnum::LockStmt(..)) => Some("LockStmt"),
                Some(NodeEnum::ConstraintsSetStmt(..)) => Some("ConstraintsSetStmt"),
                Some(NodeEnum::ReindexStmt(..)) => Some("ReindexStmt"),
                Some(NodeEnum::CheckPointStmt(..)) => Some("CheckPointStmt"),
                Some(NodeEnum::CreateSchemaStmt(..)) => Some("CreateSchemaStmt"),
                Some(NodeEnum::AlterDatabaseStmt(..)) => Some("AlterDatabaseStmt"),
                Some(NodeEnum::AlterDatabaseSetStmt(..)) => Some("AlterDatabaseSetStmt"),
                Some(NodeEnum::AlterRoleSetStmt(..)) => Some("AlterRoleSetStmt"),
                Some(NodeEnum::CreateConversionStmt(..)) => Some("CreateConversionStmt"),
                Some(NodeEnum::CreateCastStmt(..)) => Some("CreateCastStmt"),
                Some(NodeEnum::CreateOpClassStmt(..)) => Some("CreateOpClassStmt"),
                Some(NodeEnum::CreateOpFamilyStmt(..)) => Some("CreateOpFamilyStmt"),
                Some(NodeEnum::AlterOpFamilyStmt(..)) => Some("AlterOpFamilyStmt"),
                Some(NodeEnum::PrepareStmt(..)) => Some("PrepareStmt"),
                Some(NodeEnum::ExecuteStmt(..)) => Some("ExecuteStmt"),
                Some(NodeEnum::DeallocateStmt(..)) => Some("DeallocateStmt"),
                Some(NodeEnum::DeclareCursorStmt(..)) => Some("DeclareCursorStmt"),
                Some(NodeEnum::CreateTableSpaceStmt(..)) => Some("CreateTableSpaceStmt"),
                Some(NodeEnum::DropTableSpaceStmt(..)) => Some("DropTableSpaceStmt"),
                Some(NodeEnum::AlterObjectDependsStmt(..)) => Some("AlterObjectDependsStmt"),
                Some(NodeEnum::AlterObjectSchemaStmt(..)) => Some("AlterObjectSchemaStmt"),
                Some(NodeEnum::AlterOwnerStmt(..)) => Some("AlterOwnerStmt"),
                Some(NodeEnum::AlterOperatorStmt(..)) => Some("AlterOperatorStmt"),
                Some(NodeEnum::AlterTypeStmt(..)) => Some("AlterTypeStmt"),
                Some(NodeEnum::DropOwnedStmt(..)) => Some("DropOwnedStmt"),
                Some(NodeEnum::ReassignOwnedStmt(..)) => Some("ReassignOwnedStmt"),
                Some(NodeEnum::CompositeTypeStmt(..)) => Some("CompositeTypeStmt"),
                Some(NodeEnum::CreateEnumStmt(..)) => Some("CreateEnumStmt"),
                Some(NodeEnum::CreateRangeStmt(..)) => Some("CreateRangeStmt"),
                Some(NodeEnum::AlterEnumStmt(..)) => Some("AlterEnumStmt"),
                // AlterTSDictionaryStmt is capitalized differently to match C implementation.
                Some(NodeEnum::AlterTsdictionaryStmt(..)) => Some("AlterTSDictionaryStmt"),
                // AlterTSConfigurationStmt is capitalized differently to match C implementation.
                Some(NodeEnum::AlterTsconfigurationStmt(..)) => Some("AlterTSConfigurationStmt"),
                Some(NodeEnum::CreateFdwStmt(..)) => Some("CreateFdwStmt"),
                Some(NodeEnum::AlterFdwStmt(..)) => Some("AlterFdwStmt"),
                Some(NodeEnum::CreateForeignServerStmt(..)) => Some("CreateForeignServerStmt"),
                Some(NodeEnum::AlterForeignServerStmt(..)) => Some("AlterForeignServerStmt"),
                Some(NodeEnum::CreateUserMappingStmt(..)) => Some("CreateUserMappingStmt"),
                Some(NodeEnum::AlterUserMappingStmt(..)) => Some("AlterUserMappingStmt"),
                Some(NodeEnum::DropUserMappingStmt(..)) => Some("DropUserMappingStmt"),
                Some(NodeEnum::AlterTableSpaceOptionsStmt(..)) => Some("AlterTableSpaceOptionsStmt"),
                Some(NodeEnum::AlterTableMoveAllStmt(..)) => Some("AlterTableMoveAllStmt"),
                Some(NodeEnum::SecLabelStmt(..)) => Some("SecLabelStmt"),
                Some(NodeEnum::CreateForeignTableStmt(..)) => Some("CreateForeignTableStmt"),
                Some(NodeEnum::ImportForeignSchemaStmt(..)) => Some("ImportForeignSchemaStmt"),
                Some(NodeEnum::CreateExtensionStmt(..)) => Some("CreateExtensionStmt"),
                Some(NodeEnum::AlterExtensionStmt(..)) => Some("AlterExtensionStmt"),
                Some(NodeEnum::AlterExtensionContentsStmt(..)) => Some("AlterExtensionContentsStmt"),
                Some(NodeEnum::CreateEventTrigStmt(..)) => Some("CreateEventTrigStmt"),
                Some(NodeEnum::AlterEventTrigStmt(..)) => Some("AlterEventTrigStmt"),
                Some(NodeEnum::RefreshMatViewStmt(..)) => Some("RefreshMatViewStmt"),
                Some(NodeEnum::ReplicaIdentityStmt(..)) => Some("ReplicaIdentityStmt"),
                Some(NodeEnum::AlterSystemStmt(..)) => Some("AlterSystemStmt"),
                Some(NodeEnum::CreatePolicyStmt(..)) => Some("CreatePolicyStmt"),
                Some(NodeEnum::AlterPolicyStmt(..)) => Some("AlterPolicyStmt"),
                Some(NodeEnum::CreateTransformStmt(..)) => Some("CreateTransformStmt"),
                Some(NodeEnum::CreateAmStmt(..)) => Some("CreateAmStmt"),
                Some(NodeEnum::CreatePublicationStmt(..)) => Some("CreatePublicationStmt"),
                Some(NodeEnum::AlterPublicationStmt(..)) => Some("AlterPublicationStmt"),
                Some(NodeEnum::CreateSubscriptionStmt(..)) => Some("CreateSubscriptionStmt"),
                Some(NodeEnum::AlterSubscriptionStmt(..)) => Some("AlterSubscriptionStmt"),
                Some(NodeEnum::DropSubscriptionStmt(..)) => Some("DropSubscriptionStmt"),
                Some(NodeEnum::CreateStatsStmt(..)) => Some("CreateStatsStmt"),
                Some(NodeEnum::AlterCollationStmt(..)) => Some("AlterCollationStmt"),
                Some(NodeEnum::CallStmt(..)) => Some("CallStmt"),
                Some(NodeEnum::AlterStatsStmt(..)) => Some("AlterStatsStmt"),
                _ => None,
            })
            .collect()
    }
}
