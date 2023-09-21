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

import "fmt"

func init() {
	initializeDefaultMessage(Describe{})
	addMessageHeader(Describe{})
}

// Describe represents a PostgreSQL message.
type Describe struct {
	IsPrepared bool // IsPrepared states whether we're describing a prepared statement or a portal.
	Target     string
}

var describeDefault = Message{
	Name: "Describe",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('D'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "DescribingTarget",
			Type: Byte1,
			Data: int32(0),
		},
		{
			Name: "TargetName",
			Type: String,
			Data: "",
		},
	},
}

var _ MessageType = Describe{}

// encode implements the interface MessageType.
func (m Describe) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	if m.IsPrepared {
		outputMessage.Field("DescribingTarget").MustWrite('S')
	} else {
		outputMessage.Field("DescribingTarget").MustWrite('P')
	}
	outputMessage.Field("TargetName").MustWrite(m.Target)
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m Describe) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	describingTarget := s.Field("DescribingTarget").MustGet().(int32)
	var isPrepared bool
	if describingTarget == 'S' {
		isPrepared = true
	} else if describingTarget == 'P' {
		isPrepared = false
	} else {
		return nil, fmt.Errorf("Unknown describing target in Describe message: %d", describingTarget)
	}
	return Describe{
		IsPrepared: isPrepared,
		Target:     s.Field("TargetName").MustGet().(string),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m Describe) defaultMessage() *Message {
	return &describeDefault
}
