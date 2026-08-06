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

package v1_1

import (
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// namespaceParams are the parameters that uuid_generate_v3 and uuid_generate_v5 share.
var namespaceParams = []extdef.Parameter{{Name: "namespace", Type: "uuid"}, {Name: "name", Type: "text"}}

// Extension returns the definition of the emulated extension.
func Extension() *extdef.Extension {
	return &extdef.Extension{
		Name: "uuid-ossp",
		Control: extdef.Control{
			DefaultVersion: "1.1",
			Comment:        "generate universally unique identifiers (UUIDs)",
			Superuser:      true,
			Trusted:        true,
			Relocatable:    true,
		},
		Routines: []extdef.Routine{
			{Name: "uuid_nil", Symbol: "uuid_nil", Returns: "uuid", Strict: true, Impl: uuidNil},
			{Name: "uuid_ns_dns", Symbol: "uuid_ns_dns", Returns: "uuid", Strict: true, Impl: uuidNsDns},
			{Name: "uuid_ns_url", Symbol: "uuid_ns_url", Returns: "uuid", Strict: true, Impl: uuidNsURL},
			{Name: "uuid_ns_oid", Symbol: "uuid_ns_oid", Returns: "uuid", Strict: true, Impl: uuidNsOID},
			{Name: "uuid_ns_x500", Symbol: "uuid_ns_x500", Returns: "uuid", Strict: true, Impl: uuidNsX500},
			{Name: "uuid_generate_v1", Symbol: "uuid_generate_v1", Returns: "uuid", Strict: true, Impl: uuidGenerateV1},
			{Name: "uuid_generate_v1mc", Symbol: "uuid_generate_v1mc", Returns: "uuid", Strict: true, Impl: uuidGenerateV1mc},
			{Name: "uuid_generate_v3", Symbol: "uuid_generate_v3", Parameters: namespaceParams, Returns: "uuid", Strict: true, Impl: uuidGenerateV3},
			{Name: "uuid_generate_v4", Symbol: "uuid_generate_v4", Returns: "uuid", Strict: true, Impl: uuidGenerateV4},
			{Name: "uuid_generate_v5", Symbol: "uuid_generate_v5", Parameters: namespaceParams, Returns: "uuid", Strict: true, Impl: uuidGenerateV5},
		},
	}
}

// uuidNil implements uuid_nil, which returns the "nil" UUID.
func uuidNil(ctx *sql.Context, args ...any) (any, error) {
	return uuid.Nil, nil
}

// uuidNsDns implements uuid_ns_dns, which returns the RFC 4122 namespace identifier for DNS names.
func uuidNsDns(ctx *sql.Context, args ...any) (any, error) {
	return uuid.NamespaceDNS, nil
}

// uuidNsURL implements uuid_ns_url, which returns the RFC 4122 namespace identifier for URLs.
func uuidNsURL(ctx *sql.Context, args ...any) (any, error) {
	return uuid.NamespaceURL, nil
}

// uuidNsOID implements uuid_ns_oid, which returns the RFC 4122 namespace identifier for ISO OIDs.
func uuidNsOID(ctx *sql.Context, args ...any) (any, error) {
	return uuid.NamespaceOID, nil
}

// uuidNsX500 implements uuid_ns_x500, which returns the RFC 4122 namespace identifier for X.500 names.
func uuidNsX500(ctx *sql.Context, args ...any) (any, error) {
	return uuid.NamespaceX500, nil
}

// uuidGenerateV1 implements uuid_generate_v1, which returns a version 1 UUID built from the timestamp,
// a clock sequence, and this computer's MAC address.
func uuidGenerateV1(ctx *sql.Context, args ...any) (any, error) {
	return uuid.NewV1()
}

// uuidGenerateV1mc implements uuid_generate_v1mc, which returns a version 1 UUID whose node is a random
// address with the IEEE 802 multicast and locally-administered bits set.
func uuidGenerateV1mc(ctx *sql.Context, args ...any) (any, error) {
	newUUID, err := uuid.NewV1()
	if err != nil {
		return nil, err
	}
	randomUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	copy(newUUID[10:], randomUUID[10:])
	newUUID[10] |= 0x03
	return newUUID, nil
}

// uuidGenerateV3 implements uuid_generate_v3, which returns a version 3 UUID formed from the MD5 hash
// of the given namespace and name.
func uuidGenerateV3(ctx *sql.Context, args ...any) (any, error) {
	namespace, name, err := namespaceAndName(args)
	if err != nil {
		return nil, err
	}
	return uuid.NewV3(namespace, name), nil
}

// uuidGenerateV4 implements uuid_generate_v4, which returns a version 4 UUID composed entirely of
// random data.
func uuidGenerateV4(ctx *sql.Context, args ...any) (any, error) {
	return uuid.NewV4()
}

// uuidGenerateV5 implements uuid_generate_v5, which returns a version 5 UUID formed from the SHA-1
// hash of the given namespace and name.
func uuidGenerateV5(ctx *sql.Context, args ...any) (any, error) {
	namespace, name, err := namespaceAndName(args)
	if err != nil {
		return nil, err
	}
	return uuid.NewV5(namespace, name), nil
}

// namespaceAndName reads the namespace and name arguments that uuid_generate_v3 and uuid_generate_v5
// share.
func namespaceAndName(args []any) (uuid.UUID, string, error) {
	namespace, ok := args[0].(uuid.UUID)
	if !ok {
		return uuid.Nil, "", errors.Errorf("expected a UUID namespace, received `%T`", args[0])
	}
	name, ok := args[1].(string)
	if !ok {
		return uuid.Nil, "", errors.Errorf("expected a TEXT name, received `%T`", args[1])
	}
	return namespace, name, nil
}
