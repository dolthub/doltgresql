// Copyright 2024 Dolthub, Inc.
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

package auth

import (
	crptorand "crypto/rand"
	mathrand "math/rand"

	"github.com/dolthub/doltgresql/server/auth/rfc5802"
)

// GenerateRandomOctetString generates an OctetString filled with random bytes.
func GenerateRandomOctetString(length int) rfc5802.OctetString {
	if length <= 0 {
		return nil
	}
	str := make(rfc5802.OctetString, length)
	_, err := crptorand.Read(str)
	if err != nil {
		// If crypto/rand errors for some reason, we'll use the non-cryptographic version since it will never error.
		// Considering this is used for salts, etc. it is fine to use the non-cryptographic version. We still prefer
		// the other though, this is just backup.
		_, _ = mathrand.Read(str) //lint:ignore SA1019 Intentional usage
	}
	return str
}
