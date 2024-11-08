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

package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/auth/rfc5802"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
)

// SCRAM authentication is defined in RFC-5802:
// https://datatracker.ietf.org/doc/html/rfc5802

// These are mechanisms that are used for SASL authentication.
const (
	SASLMechanism_SCRAM_SHA_256      = "SCRAM-SHA-256"
	SASLMechanism_SCRAM_SHA_256_PLUS = "SCRAM-SHA-256-PLUS"
)

// EnableAuthentication handles whether authentication is enabled. If enabled, it verifies that the given user exists,
// and checks that the encrypted password is derivable from the stored encrypted password.
var EnableAuthentication = true

// SASLBindingFlag are the flags for gs2-cbind-flag, used in SASL authentication.
type SASLBindingFlag string

const (
	SASLBindingFlag_NoClientSupport        SASLBindingFlag = "n"
	SASLBindingFlag_AssumedNoServerSupport SASLBindingFlag = "y"
	SASLBindingFlag_Used                   SASLBindingFlag = "p"
)

// SASLInitial is the structured form of the input given by *pgproto3.SASLInitialResponse.
type SASLInitial struct {
	Flag     SASLBindingFlag
	BindName string // Only set when Flag is SASLBindingFlag_Used
	Binding  string // Base64 encoding of cbind-input
	Authzid  string // Authorization ID, currently ignored in favor of the startup message's username
	Username string // Prepared using SASLprep, currently ignored in favor of the startup message's username
	Nonce    string
	RawData  []byte // The bytes that were received in the message
}

// SASLContinue is the structured form of the output for *pgproto3.SASLInitialResponse.
type SASLContinue struct {
	Nonce      string
	Salt       string // Base64 encoded salt
	Iterations uint32
}

// SASLResponse is the structured form of the input given by *pgproto3.SASLResponse.
type SASLResponse struct {
	GS2Header   string
	Nonce       string
	ClientProof string // Base64 encoded
	RawData     []byte // The bytes that were received in the message
}

// handleAuthentication handles authentication for the given user
func (h *ConnectionHandler) handleAuthentication(startupMessage *pgproto3.StartupMessage) error {
	var username string
	var host string
	var ok bool
	if username, ok = startupMessage.Parameters["user"]; ok && len(username) > 0 {
		if h.Conn().RemoteAddr().Network() == "unix" {
			host = "localhost"
		} else {
			host, _, _ = net.SplitHostPort(h.Conn().RemoteAddr().String())
			if len(host) == 0 {
				host = "localhost"
			}
		}
	} else {
		username = "doltgres" // TODO: should we use this, or the default "postgres" since programs may default to it?
		host = "localhost"
	}
	h.mysqlConn.User = username
	h.mysqlConn.UserData = sql.MysqlConnectionUser{
		User: username,
		Host: host,
	}
	// Currently, regression tests disable authentication, since we can't just replay the messages due to nonces.
	if !EnableAuthentication {
		return h.send(&pgproto3.AuthenticationOk{})
	}
	// We only support one mechanism for now.
	if err := h.send(&pgproto3.AuthenticationSASL{
		AuthMechanisms: []string{
			SASLMechanism_SCRAM_SHA_256,
		},
	}); err != nil {
		return err
	}
	if err := h.backend.SetAuthType(pgproto3.AuthTypeSASL); err != nil {
		return err
	}
	// Even though we can determine whether the role exists at this point, we delay the actual error for additional security.
	role := auth.GetRole(username)
	var saslInitial SASLInitial
	var saslContinue SASLContinue
	var saslResponse SASLResponse
	for {
		initialResponse, err := h.backend.Receive()
		if err != nil {
			return err
		}
		switch response := initialResponse.(type) {
		case *pgproto3.SASLInitialResponse:
			saslInitial, err = readSASLInitial(response)
			if err != nil {
				_ = h.send(&pgproto3.ErrorResponse{
					Severity: "FATAL",
					Code:     "XX000",
					Message:  err.Error(),
				})
				return err
			}
			var salt string
			if role.Password != nil {
				salt = role.Password.Salt.ToBase64()
			} else {
				// We do this to get a stable salt. An unstable salt could be used to determine whether a username exists.
				salt = rfc5802.H(rfc5802.OctetString(username))[:16].ToBase64()
			}
			saslContinue = SASLContinue{
				Nonce:      saslInitial.Nonce + auth.GenerateRandomOctetString(16).ToBase64(),
				Salt:       salt,
				Iterations: 4096,
			}
			if err = h.send(saslContinue.Encode()); err != nil {
				return err
			}
			if err = h.backend.SetAuthType(pgproto3.AuthTypeSASLContinue); err != nil {
				return err
			}
		case *pgproto3.SASLResponse:
			saslResponse, err = readSASLResponse(saslInitial.Base64Header(), saslContinue.Nonce, response)
			if err != nil {
				_ = h.send(&pgproto3.ErrorResponse{
					Severity: "FATAL",
					Code:     "XX000",
					Message:  err.Error(),
				})
				return err
			}
			serverSignature, err := verifySASLClientProof(role, saslInitial, saslContinue, saslResponse)
			if err != nil {
				_ = h.send(&pgproto3.ErrorResponse{
					Severity: "FATAL",
					Code:     "28P01",
					Message:  err.Error(),
				})
				return err
			}
			if err = h.send(&pgproto3.AuthenticationSASLFinal{
				Data: []byte("v=" + serverSignature),
			}); err != nil {
				return err
			}
			return h.send(&pgproto3.AuthenticationOk{})
		default:
			return fmt.Errorf("unknown message type encountered during SASL authentication: %T", response)
		}
	}
}

// readSASLInitial reads the initial SASL response from the client.
func readSASLInitial(r *pgproto3.SASLInitialResponse) (SASLInitial, error) {
	if r.AuthMechanism != SASLMechanism_SCRAM_SHA_256 {
		return SASLInitial{}, fmt.Errorf("SASL mechanism not supported: %s", r.AuthMechanism)
	}
	saslInitial := SASLInitial{}
	sections := strings.Split(string(r.Data), ",")
	if len(sections) < 3 {
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: too few sections")
	}

	// gs2-cbind-flag is the first section
	gs2CbindFlag := sections[0]
	if len(gs2CbindFlag) == 0 {
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-cbind-flag")
	}
	switch gs2CbindFlag[0] {
	case 'n':
		saslInitial.Flag = SASLBindingFlag_NoClientSupport
	case 'p':
		if len(gs2CbindFlag) < 3 {
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-cbind-flag channel binding")
		}
		saslInitial.Flag = SASLBindingFlag_Used
		saslInitial.BindName = gs2CbindFlag[2:]
	case 'y':
		saslInitial.Flag = SASLBindingFlag_AssumedNoServerSupport
	default:
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-cbind-flag options (%c)", gs2CbindFlag[0])
	}

	// authzid is the second section
	authzid := sections[1]
	if len(authzid) > 0 {
		if len(authzid) < 3 {
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed authzid")
		}
		saslInitial.Authzid = authzid[2:]
	}

	// Read the gs2-header
	for i := 2; i < len(sections); i++ {
		if len(sections[i]) < 2 {
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: malformed gs2-header")
		}
		switch sections[i][0] {
		case 'c':
			saslInitial.Binding = sections[i][2:]
		case 'n':
			saslInitial.Username = sections[i][2:]
		case 'r':
			saslInitial.Nonce = sections[i][2:]
		default:
			return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: unknown gs2-header option (%c)", sections[i][0])
		}
	}

	// Validate that all required options have been read
	if len(saslInitial.Nonce) == 0 {
		return SASLInitial{}, fmt.Errorf("invalid SASLInitialResponse: missing nonce")
	}
	// Copy the message bytes, since the backend may re-use the slice for future responses
	saslInitial.RawData = make([]byte, len(r.Data))
	copy(saslInitial.RawData, r.Data)
	return saslInitial, nil
}

// readSASLResponse reads the second SASL response from the client.
func readSASLResponse(gs2EncodedHeader string, nonce string, r *pgproto3.SASLResponse) (SASLResponse, error) {
	saslResponse := SASLResponse{}
	for _, section := range strings.Split(string(r.Data), ",") {
		if len(section) < 3 {
			return SASLResponse{}, fmt.Errorf("invalid SASLResponse: attribute too small")
		}
		switch section[0] {
		case 'c':
			saslResponse.GS2Header = section[2:]
			if saslResponse.GS2Header != gs2EncodedHeader {
				return SASLResponse{}, fmt.Errorf("invalid SASLResponse: inconsistent GS2 header")
			}
		case 'p':
			saslResponse.ClientProof = section[2:]
		case 'r':
			saslResponse.Nonce = section[2:]
			if saslResponse.Nonce != nonce {
				return SASLResponse{}, fmt.Errorf("invalid SASLResponse: nonce does not match authentication session")
			}
		default:
			return SASLResponse{}, fmt.Errorf("invalid SASLResponse: unknown attribute (%c)", section[0])
		}
	}

	// Validate that all required options have been read
	if len(saslResponse.Nonce) == 0 {
		return SASLResponse{}, fmt.Errorf("invalid SASLResponse: missing nonce")
	}
	if len(saslResponse.ClientProof) == 0 {
		return SASLResponse{}, fmt.Errorf("invalid SASLResponse: missing nonce")
	}
	// Copy the message bytes, since the backend may re-use the slice for future responses
	saslResponse.RawData = make([]byte, len(r.Data))
	copy(saslResponse.RawData, r.Data)
	return saslResponse, nil
}

// verifySASLClientProof verifies that the proof given by the client in valid. Returns the base64-encoded
// ServerSignature, which verifies (to the client) that the server has proper access to the client's authentication
// information.
func verifySASLClientProof(user auth.Role, saslInitial SASLInitial, saslContinue SASLContinue, saslResponse SASLResponse) (string, error) {
	if !user.CanLogin || user.Password == nil {
		return "", fmt.Errorf(`password authentication failed for user "%s"`, user.Name)
	}
	// TODO: check the "valid until" time
	clientProof := rfc5802.Base64ToOctetString(saslResponse.ClientProof)
	authMessage := fmt.Sprintf("%s,%s,%s", saslInitial.MessageBare(), saslContinue.Encode().Data, saslResponse.MessageWithoutProof())
	clientSignature := rfc5802.ClientSignature(user.Password.StoredKey, authMessage)
	if len(clientProof) != len(clientSignature) {
		return "", fmt.Errorf(`password authentication failed for user "%s"`, user.Name)
	}
	clientKey := clientSignature.Xor(clientProof)
	storedKey := rfc5802.StoredKey(clientKey)
	if !storedKey.Equals(user.Password.StoredKey) {
		return "", fmt.Errorf(`password authentication failed for user "%s"`, user.Name)
	}
	serverSignature := rfc5802.ServerSignature(user.Password.ServerKey, authMessage)
	return serverSignature.ToBase64(), nil
}

// Base64Header returns the base64-encoded GS2 header and channel binding data.
func (si SASLInitial) Base64Header() string {
	return base64.StdEncoding.EncodeToString(si.base64HeaderBytes())
}

// MessageBare returns the message without the GS2 header.
func (si SASLInitial) MessageBare() []byte {
	return bytes.TrimPrefix(si.RawData, si.base64HeaderBytes())
}

// base64HeaderBytes returns the GS2 header encoded as bytes.
func (si SASLInitial) base64HeaderBytes() []byte {
	bb := bytes.Buffer{}
	switch si.Flag {
	case SASLBindingFlag_NoClientSupport:
		bb.WriteString("n,")
	case SASLBindingFlag_AssumedNoServerSupport:
		bb.WriteString("y,")
	case SASLBindingFlag_Used:
		bb.WriteString(fmt.Sprintf("p=%s,", si.BindName))
	}
	bb.WriteString(si.Authzid)
	bb.WriteRune(',')
	return bb.Bytes()
}

// Encode returns the struct as an AuthenticationSASLContinue message.
func (sc SASLContinue) Encode() *pgproto3.AuthenticationSASLContinue {
	return &pgproto3.AuthenticationSASLContinue{
		Data: []byte(fmt.Sprintf("r=%s,s=%s,i=%d", sc.Nonce, sc.Salt, sc.Iterations)),
	}
}

// MessageWithoutProof returns the client-final-message-without-proof.
func (sr SASLResponse) MessageWithoutProof() []byte {
	// client-final-message is defined as:
	// client-final-message-without-proof "," proof
	// So we can simply search for ",p=" and exclude everything after that for well-conforming messages.
	// If the message does not conform, then an error will happen later in the pipeline.
	index := strings.LastIndex(string(sr.RawData), ",p=")
	if index == -1 {
		return sr.RawData
	}
	return sr.RawData[:index]
}
