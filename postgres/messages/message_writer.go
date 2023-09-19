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

// MessageWriter is used to easily (and safely) interact with the contents of a Message.
type MessageWriter struct {
	message    Message
	fieldQueue []messageWriterChildPosition
}

// messageWriterChildPosition contains the name and position of a queued entry in a MessageWriter.
type messageWriterChildPosition struct {
	name     string
	position int
}

// Child returns a new MessageWriter pointing to the child of the field chain provided so far. As there may be multiple
// children, the position determines which child will be referenced. If a child position is given that does not exist,
// then a child at that position will be created. If the position is more than an increment, then this will also create
// all children up to the position, by giving them their default values.
func (mw MessageWriter) Child(name string, position int) MessageWriter {
	fieldQueue := make([]messageWriterChildPosition, len(mw.fieldQueue)+1)
	copy(fieldQueue, mw.fieldQueue)
	fieldQueue[len(fieldQueue)-1] = messageWriterChildPosition{name, position}
	return MessageWriter{
		message:    mw.message,
		fieldQueue: fieldQueue,
	}
}

// Write writes the given value to the field pointed to by the field chain provided. Only accepts values with the
// following types: int/8/16/32/64, uint/8/16/32/64, string, []byte (use an empty slice instead of nil).
func (mw MessageWriter) Write(value any) error {
	if mw.message.isDefault {
		return fmt.Errorf("Cannot write to the default message: %s", mw.message.Name)
	}

	var field *Field
	var defaultField *Field
	if fieldInfo, ok := mw.message.info.fieldInfo[mw.fieldQueue[0].name]; ok {
		field = mw.message.Fields[fieldInfo.RelativeIndex]
		defaultField = mw.message.info.defaultMessage.Fields[fieldInfo.RelativeIndex]
	} else {
		return fmt.Errorf(`The message "%s" does not contain a field named "%s"`, mw.message.Name, mw.fieldQueue[0].name)
	}
	fq := mw.fieldQueue[1:]
	for len(fq) > 0 {
		fieldInfo, ok := mw.message.info.fieldInfo[fq[0].name]
		if !ok {
			return fmt.Errorf(`The message "%s" does not contain a field named "%s"`, mw.message.Name, fq[0].name)
		}
		if fieldInfo.Parent != field.Name {
			return fmt.Errorf(`In the message "%s", the field "%s"" is not a child of the field "%s"`,
				mw.message.Name, fq[0].name, field.Name)
		}
		field.extend(fq[0].position+1, defaultField.Children[0]) // extend takes the length, so add 1 to the position
		field = field.Children[fq[0].position][fieldInfo.RelativeIndex]
		defaultField = defaultField.Children[0][fieldInfo.RelativeIndex]
		// Remove the child from the queue
		fq = fq[1:]
	}

	switch field.Type {
	case Byte1, Int8, Int16, Int32:
		switch value := value.(type) {
		case int:
			field.Data = int32(value)
		case int8:
			field.Data = int32(value)
		case int16:
			field.Data = int32(value)
		case int32:
			field.Data = value
		case int64:
			field.Data = int32(value)
		case uint:
			field.Data = int32(value)
		case uint8:
			field.Data = int32(value)
		case uint16:
			field.Data = int32(value)
		case uint32:
			field.Data = int32(value)
		case uint64:
			field.Data = int32(value)
		default:
			return fmt.Errorf("Attempted to write an invalid value of type `%T` into the following integer field: %s", value, field.Name)
		}
	case ByteN:
		switch value := value.(type) {
		case []byte:
			field.Data = value
		default:
			return fmt.Errorf("Attempted to write an invalid value of type `%T` into the following ByteN field: %s", value, field.Name)
		}
	case String:
		switch value := value.(type) {
		case string:
			field.Data = value
		default:
			return fmt.Errorf("Attempted to write an invalid value of type `%T` into the following String field: %s", value, field.Name)
		}
	default:
		panic("message type has not been defined")
	}
	return nil
}

// Get returns the value of the field pointed to by the field chain provided.
func (mw MessageWriter) Get() (any, error) {
	var field *Field
	if fieldInfo, ok := mw.message.info.fieldInfo[mw.fieldQueue[0].name]; ok {
		field = mw.message.Fields[fieldInfo.RelativeIndex]
	} else {
		return nil, fmt.Errorf(`The message "%s" does not contain a field named "%s"`, mw.message.Name, mw.fieldQueue[0].name)
	}
	fq := mw.fieldQueue[1:]
	for len(fq) > 0 {
		fieldInfo, ok := mw.message.info.fieldInfo[fq[0].name]
		if !ok {
			return nil, fmt.Errorf(`The message "%s" does not contain a field named "%s"`, mw.message.Name, fq[0].name)
		}
		if fieldInfo.Parent != field.Name {
			return nil, fmt.Errorf(`In the message "%s", the field "%s" is not a child of the field "%s"`,
				mw.message.Name, fq[0].name, field.Name)
		}
		if fq[0].position >= len(field.Children) {
			return nil, fmt.Errorf("Index out of bounds.\nMessage: %s\nField: %s\nIndex: %d\nLength: %d",
				mw.message.Name, field.Name, fq[0].position, len(field.Children))
		}
		field = field.Children[fq[0].position][fieldInfo.RelativeIndex]
		// Remove the child from the queue
		fq = fq[1:]
	}
	return field.Data, nil
}

// MustWrite is the same as Write, except that this panics on errors rather than returning them.
func (mw MessageWriter) MustWrite(value any) {
	if err := mw.Write(value); err != nil {
		panic(err)
	}
}

// MustGet is the same as Get, except that this panics on errors rather than returning them.
func (mw MessageWriter) MustGet() any {
	value, err := mw.Get()
	if err != nil {
		panic(err)
	}
	return value
}
