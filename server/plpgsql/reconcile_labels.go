// Copyright 2025 Dolthub, Inc.
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

package plpgsql

import (
	"strconv"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/utils"
)

// continueTargetOption is the key, within a loop's ScopeBegin Options, holding the offset from that ScopeBegin
// to the loop's next-iteration operation, which is where a CONTINUE for that loop jumps. Block.AppendOperations
// sets it and reconcileLabels removes it, so it never reaches the persisted operations.
const continueTargetOption = "continue_target"

// labelStackItem is the stack item used while reconciling labels.
type labelStackItem struct {
	label  string
	start  int
	isLoop bool
}

// reconcileLabels handles all GOTO operations that may point to a label, since labels may be nested/shadowed. It's
// easier to perform this is a final step, rather than trying to reconcile them during the operation conversion step.
func reconcileLabels(ops []InterpreterOperation) error {
	labels := utils.NewStack[labelStackItem]()
	gotos := make(map[int]*InterpreterOperation)
	for opIndex, operation := range ops {
		switch operation.OpCode {
		case OpCode_Goto:
			// When this is true, we have a label
			if len(operation.PrimaryData) > 0 {
				if operation.Index < 0 {
					// This is a CONTINUE, so we already know the index that we need to go to
					found := false
					for i := 0; i < labels.Len(); i++ {
						stackItem := labels.PeekDepth(i)
						if stackItem.label == operation.PrimaryData {
							if !stackItem.isLoop {
								return errors.New("CONTINUE cannot be used outside a loop")
							}
							found = true
							ops[opIndex].Index = stackItem.start
							ops[opIndex].PrimaryData = ""
							break
						}
					}
					if !found {
						return errors.Errorf(`there is no label "%s" attached to any block or loop enclosing this statement`, operation.PrimaryData)
					}
				} else {
					// This is an EXIT, so we'll save it for later
					gotos[opIndex] = &ops[opIndex]
				}
			}
		case OpCode_ScopeBegin:
			// A CONTINUE defaults to the operation after this one, else we'll continually increase the scope.
			// Loops that begin with setup operations instead of their next-iteration step say where it is.
			start := opIndex + 1
			if offset, ok := operation.Options[continueTargetOption]; ok {
				parsedOffset, err := strconv.Atoi(offset)
				if err != nil {
					return errors.Wrapf(err, "invalid CONTINUE target for scope at operation %d", opIndex)
				}
				start = opIndex + parsedOffset
			}
			// We'll push the label and loop status to the stack
			labels.Push(labelStackItem{
				label:  operation.PrimaryData,
				start:  start,
				isLoop: len(operation.Target) > 0,
			})
			// We clear the label, loop status, and CONTINUE target since we only set them for reconciliation
			ops[opIndex].PrimaryData = ""
			ops[opIndex].Target = ""
			delete(ops[opIndex].Options, continueTargetOption)
			if len(ops[opIndex].Options) == 0 {
				ops[opIndex].Options = nil
			}
		case OpCode_ScopeEnd:
			stackItem := labels.Pop()
			for gotoIdx, gotoOp := range gotos {
				if gotoOp.PrimaryData == stackItem.label {
					gotoOp.Index = opIndex // We want to go to this operation, as we want to exit the scope
					gotoOp.PrimaryData = ""
					delete(gotos, gotoIdx)
				}
			}
		}
	}
	if len(gotos) > 0 {
		for _, op := range gotos {
			return errors.Errorf(`there is no label "%s" attached to any block or loop enclosing this statement`, op.PrimaryData)
		}
	}
	return nil
}
