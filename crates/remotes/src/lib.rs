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

//! Dolt's remotes API: the gRPC chunk store service and the HTTP table file transfers that Dolt remotes use, as a
//! server over local databases and as a chunk store for `http`, `https`, and `ssh` remotes.

pub mod client;
pub mod cluster;
mod sealer;
pub mod server;
mod smux;
mod ssh;

/// remotesapi is the code that tonic-build generates from Dolt's chunk store and credentials protos.
#[allow(clippy::all)]
pub mod remotesapi {
    include!("proto/dolt.services.remotesapi.v1alpha1.rs");
}

/// replicationapi is the code that tonic-build generates from Dolt's replication proto.
#[allow(clippy::all)]
pub mod replicationapi {
    include!("proto/dolt.services.replicationapi.v1alpha1.rs");
}
