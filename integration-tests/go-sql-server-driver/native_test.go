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

package main

import "testing"

// These run only when DOLTGRES_NATIVE_BINDIR names the Postgres fork's bin
// directory; see nativedef.go.

func TestNativeCluster(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/cluster.yaml")
}

func TestNativeLargeBlobReplication(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/large-blob-replication.yaml")
}

func TestNativeLargeJSONReplication(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/large-json-replication.yaml")
}

func TestNativeLargeTextReplication(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/large-text-replication.yaml")
}

func TestNativeLargeMultiColumnReplication(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/large-multi-column-replication.yaml")
}

func TestNativeLargeValuesFailover(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/large-values-failover.yaml")
}

func TestNativeWideTableFailover(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/wide-table-failover.yaml")
}

func TestNativeTypeDiversity(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/type-diversity.yaml")
}

func TestNativeClusterRootObjects(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/cluster-root-objects.yaml")
}

func TestNativeClusterAuthReplication(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/cluster-auth-replication.yaml")
}

func TestNativeClusterReadOnly(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/cluster-read-only.yaml")
}

func TestNativeClusterFailover(t *testing.T) {
	t.Parallel()
	RunNativeTestsFile(t, "tests/native/cluster-failover.yaml")
}
