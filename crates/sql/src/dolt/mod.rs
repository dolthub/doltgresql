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

//! Dolt's version control: its procedures, system tables, and commit graph.

pub mod admin;
pub mod args;
pub mod artifacts;
pub mod branch_control;
pub mod conflicts;
pub mod diff;
pub mod docs;
pub mod history;
pub mod ignore;
pub mod merge;
pub mod objmerge;
pub mod patch;
pub mod procedures;
pub mod querydiff;
pub mod remotes;
pub mod revert;
pub mod stash;
pub mod tables;
