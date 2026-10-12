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

//! The advisory lock functions.

use super::Function;
use crate::advisory::Key;
use crate::error::{PgError, Result};
use crate::oid::{BOOL, INT4, INT8};
use crate::query::Ctx;
use crate::types::Value;

/// VOID is the type of functions that return nothing.
const VOID: u32 = 2278;

/// f declares a strict advisory lock function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the advisory lock functions, each taking a bigint key or a pair of integer keys.
pub const FUNCTIONS: &[Function] = &[
    f("pg_advisory_lock", &[INT8], VOID, lock_exclusive_session),
    f("pg_advisory_lock", &[INT4, INT4], VOID, lock_exclusive_session),
    f("pg_advisory_lock_shared", &[INT8], VOID, lock_shared_session),
    f("pg_advisory_lock_shared", &[INT4, INT4], VOID, lock_shared_session),
    f("pg_advisory_xact_lock", &[INT8], VOID, lock_exclusive_transaction),
    f("pg_advisory_xact_lock", &[INT4, INT4], VOID, lock_exclusive_transaction),
    f("pg_advisory_xact_lock_shared", &[INT8], VOID, lock_shared_transaction),
    f("pg_advisory_xact_lock_shared", &[INT4, INT4], VOID, lock_shared_transaction),
    f("pg_try_advisory_lock", &[INT8], BOOL, try_exclusive_session),
    f("pg_try_advisory_lock", &[INT4, INT4], BOOL, try_exclusive_session),
    f("pg_try_advisory_lock_shared", &[INT8], BOOL, try_shared_session),
    f("pg_try_advisory_lock_shared", &[INT4, INT4], BOOL, try_shared_session),
    f("pg_try_advisory_xact_lock", &[INT8], BOOL, try_exclusive_transaction),
    f("pg_try_advisory_xact_lock", &[INT4, INT4], BOOL, try_exclusive_transaction),
    f("pg_try_advisory_xact_lock_shared", &[INT8], BOOL, try_shared_transaction),
    f("pg_try_advisory_xact_lock_shared", &[INT4, INT4], BOOL, try_shared_transaction),
    f("pg_advisory_unlock", &[INT8], BOOL, unlock_exclusive),
    f("pg_advisory_unlock", &[INT4, INT4], BOOL, unlock_exclusive),
    f("pg_advisory_unlock_shared", &[INT8], BOOL, unlock_shared),
    f("pg_advisory_unlock_shared", &[INT4, INT4], BOOL, unlock_shared),
    f("pg_advisory_unlock_all", &[], VOID, unlock_all),
];

/// key returns the lock key of the arguments in the session's database.
fn key(ctx: &Ctx<'_>, args: &[Value]) -> Key {
    let ints: Vec<i64> = args
        .iter()
        .map(|v| match v {
            Value::Int8(i) => *i,
            Value::Int4(i) => *i as i64,
            _ => 0,
        })
        .collect();
    Key::new(&ctx.session.database, &ints)
}

/// try_lock takes a lock, reporting whether it could.
fn try_lock(ctx: &mut Ctx<'_>, args: &[Value], exclusive: bool, transaction: bool) -> bool {
    let key = key(ctx, args);
    ctx.session.advisory.acquire(ctx.session.id, key, exclusive, transaction)
}

/// lock takes a lock, failing when another session holds it, which happens only when the statement could not wait
/// for the lock before it started because the key is not a constant.
fn lock(ctx: &mut Ctx<'_>, args: &[Value], exclusive: bool, transaction: bool) -> Result<Value> {
    if !try_lock(ctx, args, exclusive, transaction) {
        return Err(PgError::new(
            "55P03",
            "could not obtain an advisory lock that another session holds, since Doltgres cannot wait for one",
        ));
    }
    Ok(Value::Text(String::new()))
}

/// lock_exclusive_session takes an exclusive lock for the session.
fn lock_exclusive_session(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    lock(ctx, args, true, false)
}

/// lock_shared_session takes a shared lock for the session.
fn lock_shared_session(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    lock(ctx, args, false, false)
}

/// lock_exclusive_transaction takes an exclusive lock for the transaction.
fn lock_exclusive_transaction(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    lock(ctx, args, true, true)
}

/// lock_shared_transaction takes a shared lock for the transaction.
fn lock_shared_transaction(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    lock(ctx, args, false, true)
}

/// try_exclusive_session tries to take an exclusive lock for the session.
fn try_exclusive_session(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(try_lock(ctx, args, true, false)))
}

/// try_shared_session tries to take a shared lock for the session.
fn try_shared_session(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(try_lock(ctx, args, false, false)))
}

/// try_exclusive_transaction tries to take an exclusive lock for the transaction.
fn try_exclusive_transaction(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(try_lock(ctx, args, true, true)))
}

/// try_shared_transaction tries to take a shared lock for the transaction.
fn try_shared_transaction(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(try_lock(ctx, args, false, true)))
}

/// unlock releases a session-level lock, warning as Postgres does when the session does not hold it.
fn unlock(ctx: &mut Ctx<'_>, args: &[Value], exclusive: bool) -> Result<Value> {
    let key = key(ctx, args);
    let released = ctx.session.advisory.release(ctx.session.id, &key, exclusive);
    if !released {
        let mode = if exclusive { "ExclusiveLock" } else { "ShareLock" };
        ctx.session.notice(PgError {
            severity: "WARNING",
            ..PgError::new("01000", format!("you don't own a lock of type {mode}"))
        });
    }
    Ok(Value::Bool(released))
}

/// unlock_exclusive releases an exclusive session-level lock.
fn unlock_exclusive(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    unlock(ctx, args, true)
}

/// unlock_shared releases a shared session-level lock.
fn unlock_shared(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    unlock(ctx, args, false)
}

/// unlock_all releases every session-level lock of the session.
fn unlock_all(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    ctx.session.advisory.release_all(ctx.session.id, false, false);
    Ok(Value::Text(String::new()))
}
