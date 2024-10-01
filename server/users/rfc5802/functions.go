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

package rfc5802

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"

	"github.com/xdg-go/stringprep"
	"golang.org/x/crypto/pbkdf2"
)

var (
	clientKeyConstant = OctetString("Client Key")
	serverKeyConstant = OctetString("Server Key")
)

// OctetString is equivalent to a byte slice. An octet, as defined in the RFC, is an 8-bit byte. Go only supports 8-bit
// bytes. Additionally, an octet string is defined as a sequence of octets, which we can represent as a slice.
type OctetString []byte

// Base64ToOctetString returns the original octet string from its base64 encoded form.
func Base64ToOctetString(base64String string) OctetString {
	decoded, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		// If we've encountered an error, then we'll return the equivalent of an empty hash. This will fail in a later step.
		return make(OctetString, 32)
	}
	return decoded
}

// ClientKey returns the client key created using the salted password and a specific constant.
func ClientKey(saltedPassword OctetString) OctetString {
	return HMAC(saltedPassword, clientKeyConstant)
}

// ClientProof returns the client proof by xor'ing the client key and client signature.
func ClientProof(clientKey OctetString, clientSignature OctetString) OctetString {
	if len(clientKey) != len(clientSignature) {
		return make(OctetString, 32)
	}
	return clientKey.Xor(clientSignature)
}

// ClientSignature returns the client signature using the given stored key and auth message.
func ClientSignature(storedKey OctetString, authMessage string) OctetString {
	return HMAC(storedKey, OctetString(authMessage))
}

// H performs the SHA256 hash function, which is the hash function used by Postgres.
func H(str OctetString) OctetString {
	ret := sha256.Sum256(str)
	return ret[:]
}

// Hi is, essentially, PBKDF2 with HMAC as the pseudorandom function.
func Hi(str OctetString, salt OctetString, i uint32) OctetString {
	return pbkdf2.Key(str, salt, int(i), 32, sha256.New)
}

// HMAC applies the HMAC keyed hash algorithm on the given octet strings.
func HMAC(key OctetString, str OctetString) OctetString {
	mac := hmac.New(sha256.New, key)
	mac.Write(str)
	return mac.Sum(nil)
}

// Normalize runs the SASLprep profile (https://datatracker.ietf.org/doc/html/rfc4013) of the stringprep algorithm
// (https://datatracker.ietf.org/doc/html/rfc3454). This accepts a standard UTF8 encoded string, unlike other functions
// which may take an OctetString.
func Normalize(str string) (string, error) {
	return stringprep.SASLprep.Prepare(str)
}

// SaltedPassword returns the salted password. The password should not have been normalized, as it is normalized within
// the function.
func SaltedPassword(password string, salt OctetString, i uint32) OctetString {
	normalizedPassword, err := Normalize(password)
	if err != nil {
		// If there is an error, then it should be fine to return the zero hash
		return make(OctetString, 32)
	}
	return Hi(OctetString(normalizedPassword), salt, i)
}

// ServerKey returns the server key created using the salted password and a specific constant.
func ServerKey(saltedPassword OctetString) OctetString {
	return HMAC(saltedPassword, serverKeyConstant)
}

// ServerSignature returns the server signature using the given server key and auth message.
func ServerSignature(serverKey OctetString, authMessage string) OctetString {
	return HMAC(serverKey, OctetString(authMessage))
}

// StoredKey returns the stored key created using the client key.
func StoredKey(clientKey OctetString) OctetString {
	return H(clientKey)
}

// AppendInteger appends the given integer.
func (os OctetString) AppendInteger(val uint32) OctetString {
	result := make(OctetString, len(os)+4)
	binary.BigEndian.PutUint32(result[len(os):], val)
	return result
}

// Equals returns whether the calling octet string is equal to the given octet string.
func (os OctetString) Equals(other OctetString) bool {
	return bytes.Equal(os, other)
}

// ToBase64 returns the OctetString as a base64 encoded UTF8 string.
func (os OctetString) ToBase64() string {
	return base64.StdEncoding.EncodeToString(os)
}

// ToHex returns the OctetString as a hex encoded UTF8 string (lowercase).
func (os OctetString) ToHex() string {
	return hex.EncodeToString(os)
}

// Xor applies "exclusive or" for every octet between both strings. Assumes that both strings have the same length, so
// perform any checks before calling this function.
func (os OctetString) Xor(other OctetString) OctetString {
	result := make(OctetString, len(os))
	for i := range os {
		result[i] = os[i] ^ other[i]
	}
	return result
}
