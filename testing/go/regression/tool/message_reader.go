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

package main

import "github.com/jackc/pgx/v5/pgproto3"

// MessageReader acts like an iterator over a collection of messages.
type MessageReader struct {
	messages []pgproto3.Message
	queue    []pgproto3.Message
	idx      int
}

// NewMessageReader returns a new *MessageReader.
func NewMessageReader(messages []pgproto3.Message) *MessageReader {
	return &MessageReader{
		messages: messages,
		idx:      0,
	}
}

// Decrement sets the reader back a position, so that the next call to Next will return the previous message.
func (mr *MessageReader) Decrement() {
	mr.idx--
	if mr.idx < 0 {
		mr.idx = 0
	}
}

// IsEmpty returns true when all messages have been returned.
func (mr *MessageReader) IsEmpty() bool {
	return len(mr.queue) == 0 && mr.idx >= len(mr.messages)
}

// Next returns the next message, or nil if the reader has been exhausted.
func (mr *MessageReader) Next() pgproto3.Message {
	if len(mr.queue) > 0 {
		if len(mr.queue) == 1 {
			m := mr.queue[0]
			mr.queue = nil
			return m
		} else {
			m := mr.queue[0]
			mr.queue = mr.queue[1:]
			return m
		}
	}
	if mr.idx >= len(mr.messages) {
		return nil
	}
	mr.idx++
	return mr.messages[mr.idx-1]
}

// Peek returns the next message without moving the reader forward, or nil if the reader has been exhausted.
func (mr *MessageReader) Peek() pgproto3.Message {
	if len(mr.queue) > 0 {
		return mr.queue[0]
	}
	if mr.idx >= len(mr.messages) {
		return nil
	}
	return mr.messages[mr.idx]
}

// Previous returns the message that was last returned by Next. Returns nil if the reader is at the beginning message.
func (mr *MessageReader) Previous() pgproto3.Message {
	if mr.idx <= 0 {
		return nil
	}
	return mr.messages[mr.idx-1]
}

// PushQueue adds the given messages to the queue, which will be returned when calling Next. Both Previous and Decrement
// ignore the queue.
func (mr *MessageReader) PushQueue(messages ...pgproto3.Message) {
	mr.queue = append(mr.queue, messages...)
}

// SyncToNextQuery advances the reader forward to the next query.
func (mr *MessageReader) SyncToNextQuery() {
	for {
		switch mr.Next().(type) {
		case *pgproto3.ReadyForQuery:
			return
		case *pgproto3.Terminate:
			mr.Decrement()
			return
		case nil:
			return
		}
	}
}
