// Copyright 2026 Dolthub, Inc.
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

package tree

import (
	"math"
	"strconv"
)

// CursorOptions is the set of options given to DECLARE CURSOR.
type CursorOptions uint8

const (
	// CursorOptionBinary is set by BINARY.
	CursorOptionBinary CursorOptions = 1 << iota
	// CursorOptionAsensitive is set by ASENSITIVE.
	CursorOptionAsensitive
	// CursorOptionInsensitive is set by INSENSITIVE.
	CursorOptionInsensitive
	// CursorOptionNoScroll is set by NO SCROLL.
	CursorOptionNoScroll
	// CursorOptionScroll is set by SCROLL.
	CursorOptionScroll
	// CursorOptionHold is set by WITH HOLD.
	CursorOptionHold
)

// DeclareCursor represents a DECLARE CURSOR statement.
type DeclareCursor struct {
	Name    Name
	Options CursorOptions
	Select  *Select
}

var _ Statement = &DeclareCursor{}

// Format implements the NodeFormatter interface.
func (node *DeclareCursor) Format(ctx *FmtCtx) {
	ctx.WriteString("DECLARE ")
	ctx.FormatNode(&node.Name)
	if node.Options&CursorOptionBinary != 0 {
		ctx.WriteString(" BINARY")
	}
	if node.Options&CursorOptionAsensitive != 0 {
		ctx.WriteString(" ASENSITIVE")
	}
	if node.Options&CursorOptionInsensitive != 0 {
		ctx.WriteString(" INSENSITIVE")
	}
	if node.Options&CursorOptionNoScroll != 0 {
		ctx.WriteString(" NO SCROLL")
	}
	if node.Options&CursorOptionScroll != 0 {
		ctx.WriteString(" SCROLL")
	}
	ctx.WriteString(" CURSOR")
	if node.Options&CursorOptionHold != 0 {
		ctx.WriteString(" WITH HOLD")
	}
	ctx.WriteString(" FOR ")
	ctx.FormatNode(node.Select)
}

// FetchDirection is the direction that FETCH and MOVE travel through a cursor.
type FetchDirection uint8

const (
	// FetchDirectionForward moves forward by the count.
	FetchDirectionForward FetchDirection = iota
	// FetchDirectionBackward moves backward by the count.
	FetchDirectionBackward
	// FetchDirectionAbsolute moves to the row at the count, counting from the end when negative.
	FetchDirectionAbsolute
	// FetchDirectionRelative moves to the row that is the count away from the current row.
	FetchDirectionRelative
)

// FetchAll is the count used by ALL, which moves through every remaining row.
const FetchAll int64 = math.MaxInt64

// FetchCursor represents a FETCH or MOVE statement.
type FetchCursor struct {
	Name      Name
	Direction FetchDirection
	Count     int64
	IsMove    bool
}

var _ Statement = &FetchCursor{}

// Format implements the NodeFormatter interface.
func (node *FetchCursor) Format(ctx *FmtCtx) {
	if node.IsMove {
		ctx.WriteString("MOVE ")
	} else {
		ctx.WriteString("FETCH ")
	}
	switch node.Direction {
	case FetchDirectionForward:
		ctx.WriteString("FORWARD ")
	case FetchDirectionBackward:
		ctx.WriteString("BACKWARD ")
	case FetchDirectionAbsolute:
		ctx.WriteString("ABSOLUTE ")
	case FetchDirectionRelative:
		ctx.WriteString("RELATIVE ")
	}
	if node.Count == FetchAll {
		ctx.WriteString("ALL")
	} else {
		ctx.WriteString(strconv.FormatInt(node.Count, 10))
	}
	ctx.WriteString(" FROM ")
	ctx.FormatNode(&node.Name)
}

// CloseCursor represents a CLOSE statement. An empty name closes every cursor.
type CloseCursor struct {
	Name Name
}

var _ Statement = &CloseCursor{}

// Format implements the NodeFormatter interface.
func (node *CloseCursor) Format(ctx *FmtCtx) {
	ctx.WriteString("CLOSE ")
	if node.Name == "" {
		ctx.WriteString("ALL")
	} else {
		ctx.FormatNode(&node.Name)
	}
}
