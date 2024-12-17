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

package id

import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"
)

// Internal uses one of two formats. Which format is being used is marked by the upper Section bit being either 0 or 1.
// Often, an ID contains information that will commonly be accessed by the item, so the first format is tailored for
// efficient retrieval of specific segments. If an item is larger than the size limit (255, size is stored as an uint8),
// then we use the second format, which inserts a separator between items. This allows Internal to hold any data in case
// the need arises in the future, but in practice we'll only see the first format (since data will usually be
// identifiers or smaller embedded IDs). Internal IDs will be accessed far more often than they'll be created, hence the
// focus on efficient retrieval rather than simplicity of storage.
//
// First format (upper bit is 0):
//     The first byte is the section
//     The second byte contains the number of segments N (up to 255 segments)
//     The next N bytes contain the length of each respective segment (up to 255 bytes)
//     The remaining bytes are the original string data, stored contiguously
// Second format (upper bit is 1):
//     The first byte is the section
//     The remaining bytes are the original string data, stored with the separator between each segment

const (
	// idSeparator marks the different data sections in an Internal ID. This is the null byte since that byte is invalid in
	// all identifiers, so we can guarantee that it's safe to use as a separator. This is used when an individual data
	// segment is larger than 254 bytes.
	idSeparator = "\x00"
	// formatMask is the upper bit that determines whether we're using the first or second format.
	formatMask = uint8(0x80)
	// Null is an empty, invalid ID.
	Null Internal = ""
)

// Internal is an ID that is used within Doltgres. This ID is never exposed to clients through any normal means, and
// exists solely for internal operations to be able to identify specific items. This functions as an internal
// replacement for Postgres' OIDs.
type Internal string

// NewInternal constructs an Internal ID using the given section and data.
func NewInternal(section Section, data ...string) Internal {
	if section == Section_Null {
		// It's easier if there's only one canonical way to represent a null ID, so we'll return our constant instead of
		// creating a new string
		return Null
	}
	if len(data) > 255 {
		return newInternalSecondFormat(section, data)
	}
	buf := bytes.Buffer{}
	buf.WriteByte(uint8(section))
	buf.WriteByte(uint8(len(data)))
	for _, segment := range data {
		segmentLength := len(segment)
		if segmentLength > 255 {
			return newInternalSecondFormat(section, data)
		}
		buf.WriteByte(uint8(segmentLength))
	}
	for _, segment := range data {
		buf.WriteString(segment)
	}
	return Internal(buf.Bytes())
}

// newInternalSecondFormat constructs an Internal ID using the given section and data. This always returns the second
// format (using the separator).
func newInternalSecondFormat(section Section, data []string) Internal {
	buf := bytes.Buffer{}
	buf.WriteByte(uint8(section) | formatMask)
	for i, segment := range data {
		if i > 0 {
			buf.WriteString(idSeparator)
		}
		buf.WriteString(segment)
	}
	return Internal(buf.Bytes())
}

// IsValid returns whether the Internal ID is valid.
func (id Internal) IsValid() bool {
	// We don't allow setting the section to Section_Null, so we can do a simple length check
	return len(id) > 0
}

// Section returns the Section for this Internal ID.
func (id Internal) Section() Section {
	if len(id) == 0 {
		return Section_Null
	}
	return Section(id[0] & (^formatMask))
}

// Data returns the original data used to create this Internal ID.
func (id Internal) Data() []string {
	if len(id) <= 1 {
		return nil
	}
	if id[0]&formatMask == formatMask {
		// Second format
		return strings.Split(string(id[1:]), idSeparator)
	} else {
		// First format
		segmentCount := int(id[1])
		data := id[2+segmentCount:] // We skip 2 for the section and count bytes, then the number of segment counts
		segments := make([]string, segmentCount)
		start := 0
		for i := 0; i < segmentCount; i++ {
			length := int(id[2+i])
			segments[i] = string(data[start : start+length])
			start += length
		}
		return segments
	}
}

// SegmentCount returns the number of segments that were in the original data.
func (id Internal) SegmentCount() int {
	if len(id) <= 1 {
		return 0
	}
	if id[0]&formatMask == formatMask {
		// Second format
		return len(id.Data())
	} else {
		// First format
		return int(id[1])
	}
}

// Segment returns the segment from the given index. An empty string is returned for an index not contained by the ID.
func (id Internal) Segment(index int) string {
	if index < 0 || len(id) <= 1 {
		return ""
	}
	if id[0]&formatMask == formatMask {
		// Second format
		data := id.Data()
		if index >= len(data) {
			return ""
		}
		return data[index]
	} else {
		// First format
		segmentCount := int(id[1])
		data := id[2+segmentCount:] // We skip 2 for the section and count bytes, then the number of segment counts
		if index >= segmentCount {
			return ""
		}
		start := 0
		currentLength := 0
		for i := 0; i <= index; i++ {
			start += currentLength
			currentLength = int(id[2+i])
		}
		return string(data[start : start+currentLength])
	}
}

// String returns a display-suitable version of the ID. Although the ID is implemented as a string, it should not be
// treated as a string except for the purposes of storage and retrieval.
func (id Internal) String() string {
	data := id.Data()
	if len(data) == 0 {
		return fmt.Sprintf(`{%s:[]}`, id.Section().String())
	}
	return fmt.Sprintf(`{%s:["%s"]}`, id.Section().String(), strings.Join(data, `","`))
}

// CaseString returns a quoted string that may be used to represent this ID in a switch-case.
func (id Internal) CaseString() string {
	if len(id) == 0 {
		return `""`
	}
	if id[0]&formatMask == formatMask {
		// Second format
		data := strings.ReplaceAll(string(id[1:]), "\x00", `\x00`)
		data = strings.ReplaceAll(data, `"`, `\x22`)
		return fmt.Sprintf(`"\x%02x%s"`, id[0], data)
	} else {
		// First format
		sb := strings.Builder{}
		sb.Grow(len(id) + 32)
		sb.WriteRune('"')
		count := int(id[1])
		sb.WriteString(fmt.Sprintf(`\x%02x\x%02x`, id[0], count))
		for i := 0; i < count; i++ {
			sb.WriteString(fmt.Sprintf(`\x%02x`, id[2+i]))
		}
		sb.WriteString(strings.ReplaceAll(string(id[2+count:]), `"`, `\x22`))
		sb.WriteRune('"')
		return sb.String()
	}
}

// UnderlyingBytes returns the underlying bytes for the ID. These must not be modified, as this is intended solely for
// efficient usage of operations that require byte slices.
func (id Internal) UnderlyingBytes() []byte {
	return unsafe.Slice(unsafe.StringData(string(id)), len(id))
}

// usesSecondFormat returns whether the separator is used, which is the second format.
func (id Internal) usesSecondFormat() bool {
	return len(id) > 0 && id[0]&formatMask == formatMask
}
