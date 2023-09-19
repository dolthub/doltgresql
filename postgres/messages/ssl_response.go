// Copyright 2023 Dolthub, Inc.
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

package messages

// SSLResponse tells the client whether SSL is supported.
type SSLResponse struct {
	SupportsSSL bool
}

// Bytes returns SSLResponse as a byte slice, ready to be returned to the client.
func (sslr SSLResponse) Bytes() []byte {
	if sslr.SupportsSSL {
		return []byte{'Y'}
	} else {
		return []byte{'N'}
	}
}
