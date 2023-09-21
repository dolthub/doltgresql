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

// FieldType is the type of the field as defined by PostgreSQL.
type FieldType byte

const (
	Byte1    FieldType = iota // Byte1 is a single unsigned byte.
	ByteN                     // ByteN is a variable number of bytes. Allowed on the last field, or when a ByteCount-tagged field precedes it.
	Int8                      // Int8 is a single signed byte.
	Int16                     // Int16 are two bytes.
	Int32                     // Int32 are four bytes.
	String                    // String is a variable-length type, generally punctuated by a NULL terminator.
	Repeated                  // Repeated is a parent type that states its children will be repeated until the end of the message.

	Byte4 = Int32 //TODO: verify that this is correct, only used on one type
)

// FieldTags are special attributes that may be assigned to fields.
type FieldTags int32

const (
	Header                 FieldTags = 1 << iota // Header is the header of the message.
	MessageLengthInclusive                       // MessageLengthInclusive is the length of the message, including the count's size.
	MessageLengthExclusive                       // MessageLengthExclusive is the length of the message, excluding the count's size.
	ExcludeTerminator                            // ExcludeTerminator excludes the terminator for String types.
	ByteCount                                    // ByteCount signals that the following ByteN non-child field uses this field for its count.
	RepeatedTerminator                           // RepeatedTerminator states that the Repeated type always ends with a NULL terminator.
)

// Field is a field within the PostgreSQL message.
type Field struct {
	Name     string
	Type     FieldType
	Tags     FieldTags
	Data     any // Data may ONLY be one of the following: int32, string, []byte. Nil is not allowed.
	Children [][]*Field
}

// Copy returns a copy of this field, which is free to modify.
func (f *Field) Copy() *Field {
	fieldCopy := *f
	if len(f.Children) > 0 {
		newChildren := make([][]*Field, len(f.Children))
		for groupIndex, fieldGroup := range f.Children {
			newFields := make([]*Field, len(fieldGroup))
			for fieldIndex, field := range fieldGroup {
				newFields[fieldIndex] = field.Copy()
			}
			newChildren[groupIndex] = newFields
		}
		fieldCopy.Children = newChildren
	}
	return &fieldCopy
}

// extend lengthens the children to the new length, using the given default children to fill each newly-created entry.
// All new entries will be copied from the default, therefore they're free to modify. This modifies the calling Field
// in-place.
func (f *Field) extend(newLength int, defaultChildren []*Field) {
	for currentIndex := len(f.Children); currentIndex < newLength; currentIndex++ {
		newFields := make([]*Field, len(defaultChildren))
		for i, field := range defaultChildren {
			newFields[i] = field.Copy()
		}
		f.Children = append(f.Children, newFields)
	}
}
