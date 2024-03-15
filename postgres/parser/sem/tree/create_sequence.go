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

// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in licenses/BSD-vitess.txt.

// Portions of this file are additionally subject to the following
// license and copyright.
//
// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// This code was derived from https://github.com/youtube/vitess.

package tree

import "github.com/cockroachdb/errors"

var _ Statement = &CreateSequence{}

// CreateSequence represents a CREATE SEQUENCE statement.
type CreateSequence struct {
	IfNotExists bool
	Name        TableName
	Persistence Persistence
	Options     SequenceOptions
}

// Format implements the NodeFormatter interface.
func (node *CreateSequence) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")

	if node.Persistence == PersistenceTemporary {
		ctx.WriteString("TEMPORARY ")
	} else if node.Persistence == PersistenceUnlogged {
		ctx.WriteString("UNCLOGGED ")
	}

	ctx.WriteString("SEQUENCE ")

	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	ctx.FormatNode(&node.Name)
	ctx.FormatNode(&node.Options)
}

// SequenceOptions represents a list of sequence options.
type SequenceOptions []SequenceOption

// Format implements the NodeFormatter interface.
func (node *SequenceOptions) Format(ctx *FmtCtx) {
	for i := range *node {
		option := &(*node)[i]
		ctx.WriteByte(' ')
		switch option.Name {
		case SeqOptAs:
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')
			ctx.WriteString(option.AsType.SQLString())
		case SeqOptIncrementBy, SeqOptStartWith, SeqOptCache:
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')
			ctx.Printf("%d", *option.IntVal)
		case SeqOptMaxValue, SeqOptMinValue:
			if option.IntVal == nil {
				ctx.WriteString("NO ")
				ctx.WriteString(option.Name)
			} else {
				ctx.WriteString(option.Name)
				ctx.WriteByte(' ')
				ctx.Printf("%d", *option.IntVal)
			}
		case SeqOptCycle, SeqOptNoCycle:
			ctx.WriteString(option.Name)
		case SeqOptOwnedBy:
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')
			if option.ColumnItemVal == nil {
				ctx.WriteString("NONE")
			} else {
				ctx.FormatNode(option.ColumnItemVal)
			}
		default:
			panic(errors.AssertionFailedf("unexpected SequenceOption: %v", option))
		}
	}
}

// SequenceOption represents an option on a CREATE SEQUENCE statement.
type SequenceOption struct {
	Name          string
	IntVal        *int64
	OptionalWord  bool
	ColumnItemVal *ColumnItem
	AsType        ResolvableTypeReference
}

// Names of options on CREATE SEQUENCE.
const (
	SeqOptAs          = "AS"
	SeqOptCycle       = "CYCLE"
	SeqOptNoCycle     = "NO CYCLE"
	SeqOptOwnedBy     = "OWNED BY"
	SeqOptCache       = "CACHE"
	SeqOptIncrementBy = "INCREMENT BY"
	SeqOptMinValue    = "MINVALUE"
	SeqOptMaxValue    = "MAXVALUE"
	SeqOptStartWith   = "START WITH"
)
