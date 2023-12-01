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

package main

import (
	"fmt"

	"github.com/dolthub/doltgresql/utils"
)

// StatementGeneratorStackElement represents an element within the StatementGeneratorStack.
type StatementGeneratorStackElement struct {
	depth      int
	generators []StatementGenerator
}

// StatementGeneratorStack handles the creation of a StatementGenerator by contextually applying operations based on the
// current state of the internal stack whenever calls are made.
type StatementGeneratorStack struct {
	stack *utils.Stack[*StatementGeneratorStackElement]
	depth int
}

// NewStatementGeneratorStack returns a new *StatementGeneratorStack.
func NewStatementGeneratorStack() *StatementGeneratorStack {
	sgs := &StatementGeneratorStack{
		stack: utils.NewStack[*StatementGeneratorStackElement](),
		depth: 0,
	}
	sgs.stack.Push(NewStatementGeneratorStackElement(0))
	return sgs
}

// NewStatementGeneratorStackElement returns a new *StatementGeneratorStackElement.
func NewStatementGeneratorStackElement(depth int, gens ...StatementGenerator) *StatementGeneratorStackElement {
	return &StatementGeneratorStackElement{
		depth:      depth,
		generators: gens,
	}
}

// AddText creates a new TextGen at the current depth.
func (sgs *StatementGeneratorStack) AddText(text string) {
	sgs.stack.Peek().Append(Text(text))
}

// AddVariable creates a new VariableGen at the current depth.
func (sgs *StatementGeneratorStack) AddVariable(name string) {
	sgs.stack.Peek().Append(Variable(name, nil))
}

// Or will take all items from the current depth and add them to a parent OrGen. Either the previous depth is an OrGen,
// or the stack will increment the sub depth to insert an OrGen.
func (sgs *StatementGeneratorStack) Or() error {
	if sgs.stack.Empty() {
		return fmt.Errorf("cannot apply Or to an empty stack")
	}
	parentElement := sgs.stack.PeekDepth(1)
	current := sgs.aggregate(sgs.stack.Pop().generators)
	if parentElement == nil {
		// We're the root, so we put an OrGen as the root, and make the current a child of the new OrGen
		sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth, Or(current)))
		sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth))
	} else if len(parentElement.generators) == 0 {
		if parentElement.depth == sgs.depth {
			// We're still at the same depth, so it's safe to append an OrGen
			parentElement.Append(Or(current))
		} else {
			// We should retain the depth, so we need to create a new element since there's a depth boundary
			sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth, Or(current)))
		}
		sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth))
	} else {
		if parentElement.depth == sgs.depth {
			// We're still at the same depth, so we need to check if the parent is an OrGen or not
			if orGen, ok := parentElement.LastGenerator().(*OrGen); ok {
				// The parent is an OrGen, so we can simply add this as another option
				if err := orGen.AddChildren(current); err != nil {
					return err
				}
			} else {
				// The parent is not an OrGen, so we need to add to the depth
				sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth, Or(current)))
			}
		} else {
			// We should retain the depth, so we need to create a new element since there's a depth boundary
			sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth, Or(current)))
		}
		sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth))
	}
	return nil
}

// NewScope increases the depth.
func (sgs *StatementGeneratorStack) NewScope() {
	sgs.depth++
	sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth))
}

// NewOptionalScope increases the depth, creating an OptionalGen at the depth root, and adding elements to the sub depth.
func (sgs *StatementGeneratorStack) NewOptionalScope() {
	sgs.depth++
	sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth, Optional()))
	sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth))
}

// NewParenScope increases the depth, creating a CollectionGen at the depth root, adding Text("(") to the CollectionGen,
// and adding any new elements to the sub depth.
func (sgs *StatementGeneratorStack) NewParenScope() {
	sgs.depth++
	sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth, Collection(Text("("))))
	sgs.stack.Push(NewStatementGeneratorStackElement(sgs.depth))
}

// ExitScope decrements the depth, writing all elements from the current depth (and all sub depths) to the preceding
// depth.
func (sgs *StatementGeneratorStack) ExitScope() error {
	defer func() {
		sgs.depth--
	}()
	current := sgs.aggregate(sgs.stack.Pop().generators)
	for parentElement := sgs.stack.Peek(); parentElement.depth == sgs.depth; parentElement = sgs.stack.Peek() {
		if orGen, ok := parentElement.LastGenerator().(*OrGen); ok {
			if err := orGen.AddChildren(current); err != nil {
				return err
			}
		} else {
			parentElement.Append(current)
		}
		current = sgs.aggregate(sgs.stack.Pop().generators)
	}
	sgs.stack.Peek().Append(current)
	return nil
}

// ExitOptionalScope decrements the depth, writing all elements from the current depth (and all sub depths) to the
// preceding depth. This will fail if the root of the current depth is not an OptionalGen.
func (sgs *StatementGeneratorStack) ExitOptionalScope() error {
	defer func() {
		sgs.depth--
	}()
	current := sgs.aggregate(sgs.stack.Pop().generators)
	for parentElement := sgs.stack.Peek(); sgs.stack.PeekDepth(1).depth == sgs.depth; parentElement = sgs.stack.Peek() {
		if orGen, ok := parentElement.LastGenerator().(*OrGen); ok {
			if err := orGen.AddChildren(current); err != nil {
				return err
			}
		} else {
			parentElement.Append(current)
		}
		current = sgs.aggregate(sgs.stack.Pop().generators)
	}
	optionalElement := sgs.stack.Pop()
	if optionalElement.depth != sgs.depth {
		return fmt.Errorf("internal bookkeeping error, attempted to exit from optional scope but the depth is incorrect")
	}
	if optionalGen, ok := optionalElement.LastGenerator().(*OptionalGen); ok {
		if err := optionalGen.AddChildren(current); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("internal bookkeeping error, attempted to exit from optional scope but missing OptionalGen")
	}
	sgs.stack.Peek().Append(sgs.aggregate(optionalElement.generators))
	return nil
}

// ExitParenScope decrements the depth, writing all elements from the current depth (and all sub depths) to the
// preceding depth, while appending an ending Text(")").
func (sgs *StatementGeneratorStack) ExitParenScope() error {
	defer func() {
		sgs.depth--
	}()
	current := sgs.aggregate(sgs.stack.Pop().generators)
	for parentElement := sgs.stack.Peek(); sgs.stack.PeekDepth(1).depth == sgs.depth; parentElement = sgs.stack.Peek() {
		if orGen, ok := parentElement.LastGenerator().(*OrGen); ok {
			if err := orGen.AddChildren(current); err != nil {
				return err
			}
		} else {
			parentElement.Append(current)
		}
		current = sgs.aggregate(sgs.stack.Pop().generators)
	}
	collectionElement := sgs.stack.Pop()
	if collectionElement.depth != sgs.depth {
		return fmt.Errorf("internal bookkeeping error, attempted to exit from paren scope but the depth is incorrect")
	}
	if collectionGen, ok := collectionElement.LastGenerator().(*CollectionGen); ok {
		if err := collectionGen.AddChildren(current, Text(")")); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("internal bookkeeping error, attempted to exit from paren scope but missing CollectionGen")
	}
	sgs.stack.Peek().Append(sgs.aggregate(collectionElement.generators))
	return nil
}

// Repeat will add the last StatementGenerator at the current depth to an OptionalGen.
func (sgs *StatementGeneratorStack) Repeat() error {
	current := sgs.stack.Peek()
	lastGen := current.LastGenerator()
	if lastGen == nil {
		return fmt.Errorf("unable to repeat as no generators exist at the current depth")
	}
	current.Append(Optional(lastGen.Copy()))
	return nil
}

// OptionalRepeat will add the last StatementGenerator at the current depth to an OptionalGen. If a prefix is given,
// then it will be added as a TextGen before the repeated generator.
func (sgs *StatementGeneratorStack) OptionalRepeat(prefix string) error {
	current := sgs.stack.Peek()
	lastGen := current.LastGenerator()
	if lastGen == nil {
		return fmt.Errorf("unable to optionally repeat as no generators exist at the current depth")
	}
	if len(prefix) > 0 {
		current.Append(Optional(Collection(Text(prefix), lastGen.Copy())))
	} else {
		current.Append(Optional(lastGen.Copy()))
	}
	return nil
}

// Finish returns a StatementGenerator containing all of the generators that have been added. This returns an error if
// the stack is not at the root depth, as that indicates a missing scope exit, or too many scope entries.
func (sgs *StatementGeneratorStack) Finish() (StatementGenerator, error) {
	if sgs.depth != 0 {
		if sgs.depth < 0 {
			return nil, fmt.Errorf("depth is invalid, too many scope exits")
		} else {
			return nil, fmt.Errorf("depth is invalid, too many scope entries")
		}
	} else if !sgs.stack.Empty() && sgs.stack.Peek().depth != 0 {
		return nil, fmt.Errorf("internal bookkeeping error, stack depth does not match handle depth")
	}
	if sgs.stack.Len() == 1 && len(sgs.stack.Peek().generators) == 0 {
		return nil, nil
	}
	var lastDepth []StatementGenerator
	for !sgs.stack.Empty() {
		currentDepth := sgs.stack.Pop()
		if len(currentDepth.generators) == 0 {
			return nil, fmt.Errorf("internal bookkeeping error, stack has a depth with no generators")
		}
		if lastGen := currentDepth.LastGenerator(); lastGen != nil {
			if orGen, ok := lastGen.(*OrGen); ok {
				if err := orGen.AddChildren(sgs.aggregate(lastDepth)); err != nil {
					return nil, err
				}
			} else {
				currentDepth.Append(sgs.aggregate(lastDepth))
			}
		} else {
			currentDepth.Append(sgs.aggregate(lastDepth))
		}
		lastDepth = currentDepth.generators
	}
	return sgs.aggregate(lastDepth), nil
}

// aggregate returns an aggregate of the given generators. Returns nil if the slice is empty.
func (sgs *StatementGeneratorStack) aggregate(gens []StatementGenerator) StatementGenerator {
	gens = removeNilGenerators(gens)
	if len(gens) == 0 {
		return nil
	} else if len(gens) == 1 {
		return gens[0]
	} else {
		return Collection(gens...)
	}
}

// LastGenerator returns the last StatementGenerator contained within its slice.
func (sgse *StatementGeneratorStackElement) LastGenerator() StatementGenerator {
	if sgse == nil || len(sgse.generators) == 0 {
		return nil
	}
	return sgse.generators[len(sgse.generators)-1]
}

// Append adds the given child to this element's collection.
func (sgse *StatementGeneratorStackElement) Append(child StatementGenerator) {
	sgse.generators = append(sgse.generators, child)
}
