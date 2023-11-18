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

package connection

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/dolthub/doltgresql/postgres/connection/iobufpool"
	"github.com/dolthub/doltgresql/utils"
)

var BufferSize = 2048

// connBuffers maintains a pool of buffers, reusable between connections. These are only used for processing 
// fixed-length messages where we know we won't exceed the buffer size on a single read.
var connBuffers = sync.Pool{
	New: func() any {
		return make([]byte, BufferSize)
	},
}

const headerSize = 5

// headerBuffers maintains a pool of buffers, reusable between connections, that are used for reading message headers
// (the first 5 bytes of a client message)
var headerBuffers = sync.Pool{
	New: func() any {
		return make([]byte, headerSize)
	},
}


var sliceOfZeroes = make([]byte, BufferSize)

// Receive returns all messages that were sent from the given connection. This checks with all messages that have a 
// Header, and have called AddMessageHeader within their init() function. Returns a nil slice if no messages were 
// matched. This is the recommended way to check for messages when a specific message is not expected.
// Use ReceiveInto or ReceiveIntoAny when expecting specific messages, where it would be an error to receive messages
// different from the expectation.
func Receive(conn net.Conn) (Message, error) {
	header := headerBuffers.Get().([]byte)
	defer headerBuffers.Put(header)
	
	n, err := conn.Read(header)
	if err != nil {
		return nil, err
	}

	if n < headerSize {
		return nil, errors.New("received message header is too short")
	}

	message, ok := allMessageHeaders[header[0]]
	if !ok {
		return nil, fmt.Errorf("received message header is not recognized: %v", header[0])
	}

	// TODO: possibly not every message has a length in this position, need an easy interface to tell us if so
	messageLen := int(binary.BigEndian.Uint32(header[1:])) - 4
	
	buffer := iobufpool.Get(messageLen)
	defer iobufpool.Put(buffer)
	
	msgBuffer := (*buffer)[:messageLen]
	n, err = conn.Read(msgBuffer)
	if err != nil {
		return nil, err
	}

	if n < messageLen {
		return nil, fmt.Errorf("received message body is too short: expected %d bytes but read %d", messageLen, n)
	}

	db := newDecodeBuffer(msgBuffer)
	db.skipHeader = true

	return receiveFromBuffer(db, message)
}

// ReceiveInto reads the given Message from the connection. This should only be used when a specific message is expected,
// and that message did not call AddMessageHeader in its init() function. In addition, if the client sends multiple
// messages at once, then this will only read the first message. If multiple messages are expected, use ReceiveIntoAny.
func ReceiveInto[T Message](conn net.Conn, message T) (out T, err error) {
	buffer := connBuffers.Get().([]byte)
	defer connBuffers.Put(zeroBuffer(buffer))

	if _, err := conn.Read(buffer); err != nil {
		return out, err
	}
	db := newDecodeBuffer(buffer)
	outMessage, err := receiveFromBuffer(db, message)
	return outMessage, err
}

// ReceiveIntoAny returns all messages that were sent from the given connection. This should only be used when one of
// several specific messages are expected, and those messages did not call AddMessageHeader in their init() functions.
// Messages given first have a higher matching priority. Returns a nil slice if no messages matched. Only returns an
// error on connection errors. This will not error if the connection sends extra data or unspecified messages.
func ReceiveIntoAny(conn net.Conn, messages ...Message) ([]Message, error) {
	buffer := connBuffers.Get().([]byte)
	defer connBuffers.Put(zeroBuffer(buffer))

	if _, err := conn.Read(buffer); err != nil {
		return nil, err
	}
	db := newDecodeBuffer(buffer)

	// This dual loop is used to process the given messages from the buffer.
	// For each step of the buffer, we try decoding each message given. If that message properly decodes, then we
	// finalize the changes on the buffer, and check the next message. As the first message has priority, we always
	// start the loop over the first message. If any message does not properly parse, then the buffer is reset to the
	// beginning of that specific loop. With this setup, the only way we will reach the break is once none of the
	// messages parse.
	var outMessages []Message
OuterLoop:
	for {
		for _, message := range messages {
			outMessage, err := receiveFromBuffer(db, message)
			if err == nil {
				outMessages = append(outMessages, outMessage)
				db.next()
				continue OuterLoop
			} else {
				db.reset()
			}
		}
		// If we've reached this point, then we've attempted to match all of the given messages.
		break
	}

	return outMessages, nil
}

// ReceiveBruteForceMatches checks with every message to find matches. This is highly inefficient, and is intended to
// assist with debugging, due to code paths where we may not know which message to expect. The order of returned
// messages are not deterministic. Each []Message represents all matches that follow the first match. Only returns an
// error on connection errors.
func ReceiveBruteForceMatches(conn net.Conn) ([][]Message, error) {
	buffer := connBuffers.Get().([]byte)
	defer connBuffers.Put(zeroBuffer(buffer))

	if _, err := conn.Read(buffer); err != nil {
		return nil, err
	}

	var allPossibleMessages [][]Message
TopLevelLoop:
	for _, firstMessage := range allMessages {
		initialDB := newDecodeBuffer(buffer)
		var messages []Message
		var messageChain []Message
		if outMessage, err := receiveFromBuffer(initialDB, firstMessage); err == nil {
			messages = append(messages, outMessage)
		} else {
			continue TopLevelLoop
		}

		type stackInfo struct {
			index  int
			buffer *decodeBuffer
		}

		stack := utils.NewStack[stackInfo]()
		stack.Push(stackInfo{0, initialDB.copy()})
	StackLoop:
		for !stack.Empty() {
			if stack.Peek().index >= len(allMessages) {
				stack.Pop()
				if !isSubset(messages, messageChain) {
					allPossibleMessages = append(allPossibleMessages, messages)
					messageChain = messages
					newMessages := make([]Message, len(messages))
					copy(newMessages, messages)
					messages = newMessages
				}
				messages = messages[:len(messages)-1]
				continue StackLoop
			}

			message := allMessages[stack.Peek().index]
			db := stack.Peek().buffer.copy()
			stack.PeekReference().index++
			outMessage, err := receiveFromBuffer(db, message)
			if err == nil {
				messages = append(messages, outMessage)
				db.next()
				stack.Push(stackInfo{0, db.copy()})
			}
		}
	}
	return allPossibleMessages, nil
}

// Send sends the given message over the connection.
func Send(conn net.Conn, message Message) error {
	encodedMessage, err := message.Encode()
	if err != nil {
		return err
	}
	data, err := encode(encodedMessage)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

// receiveFromBuffer writes the contents of the buffer into the given Message.
func receiveFromBuffer[T Message](buffer *decodeBuffer, message T) (out T, err error) {
	defaultMessage := message.DefaultMessage()
	fields := defaultMessage.Copy().Fields
	if err = decode(buffer, []FieldGroup{fields}, 1); err != nil {
		return out, err
	}
	if len(buffer.data) > 0 {
		return out, errors.New("received extra data from buffer that was not handled by message")
	}
	decodedMessage, err := message.Decode(MessageFormat{defaultMessage.Name, fields, defaultMessage.info, false})
	if err != nil {
		return out, err
	}
	return decodedMessage.(T), nil
}

// zeroBuffer fills the given buffer with zeroes, returning the same buffer that was given.
func zeroBuffer(buffer []byte) []byte {
	copy(buffer, sliceOfZeroes)
	return buffer
}

// isSubset returns whether the first message slice is a subset of the second message slice. Equal slices are viewed as
// subsets.
func isSubset(subset []Message, full []Message) bool {
	if len(subset) > len(full) {
		return false
	}
	for i := range subset {
		if subset[i] != full[i] {
			return false
		}
	}
	return true
}
