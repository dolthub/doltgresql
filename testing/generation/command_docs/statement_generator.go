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
	"errors"
	"fmt"
	"strings"
)

// StatementGenerator represents a statement, and is able to produce all valid variations of the statement.
type StatementGenerator interface {
	// AddChildren adds the given children to the generator. Not all generators accept all children, so this may error.
	AddChildren(child ...StatementGenerator) error
	// Consume returns true when the generator is able to produce a unique mutation, and false if it is not. Only one
	// generator should mutate per call, meaning a parent generator should only mutate when its children return false.
	// If the top-level generator returns false, then all permutations have been created.
	Consume() bool
	// String returns a string based on the current permutation.
	String() string
	// Copy returns a copy of the given generator (along with all of its children) in its original setting. This means
	// that the copy is in the same state that the target would be in if it had never called Consume.
	Copy() StatementGenerator
	// Reset sets the StatementGenerator back to its original state, which would be as though Consume was never called.
	Reset()
	// SourceString returns a string that may be used to recreate the StatementGenerator in a Go source file.
	SourceString() string
}

// TextGen is a generator that returns a simple string.
type TextGen string

var _ StatementGenerator = (*TextGen)(nil)

// Text creates a new StatementGenerator representing a simple string.
func Text(str string) *TextGen {
	gen := TextGen(str)
	return &gen
}

// AddChildren implements the interface StatementGenerator.
func (t *TextGen) AddChildren(children ...StatementGenerator) error {
	return fmt.Errorf("text cannot have children")
}

// Consume implements the interface StatementGenerator.
func (t *TextGen) Consume() bool {
	return false
}

// Copy implements the interface StatementGenerator.
func (t *TextGen) Copy() StatementGenerator {
	if t == nil {
		return nil
	}
	return Text(string(*t))
}

// String implements the interface StatementGenerator.
func (t *TextGen) String() string {
	return string(*t)
}

// Reset implements the interface StatementGenerator.
func (t *TextGen) Reset() {}

// SourceString implements the interface StatementGenerator.
func (t *TextGen) SourceString() string {
	return fmt.Sprintf(`Text("%s")`, string(*t))
}

// OrGen is a generator that contains multiple child generators, and will print only one at a time. Consuming will
// cycle to the next child.
type OrGen struct {
	children []StatementGenerator
	index    int
}

var _ StatementGenerator = (*OrGen)(nil)

// Or creates a new StatementGenerator representing an OrGen.
func Or(children ...StatementGenerator) *OrGen {
	return &OrGen{
		children: copyGenerators(children),
		index:    0,
	}
}

// AddChildren implements the interface StatementGenerator.
func (o *OrGen) AddChildren(children ...StatementGenerator) error {
	o.children = append(o.children, removeNilGenerators(children)...)
	return nil
}

// Consume implements the interface StatementGenerator.
func (o *OrGen) Consume() bool {
	if len(o.children) == 0 {
		return false
	}
	if o.children[o.index].Consume() {
		return true
	}
	o.index++
	if o.index >= len(o.children) {
		o.index = 0
		return false
	}
	return true
}

// Copy implements the interface StatementGenerator.
func (o *OrGen) Copy() StatementGenerator {
	if o == nil {
		return nil
	}
	return Or(o.children...)
}

// String implements the interface StatementGenerator.
func (o *OrGen) String() string {
	return o.children[o.index].String()
}

// Reset implements the interface StatementGenerator.
func (o *OrGen) Reset() {
	o.index = 0
	for _, child := range o.children {
		child.Reset()
	}
}

// SourceString implements the interface StatementGenerator.
func (o *OrGen) SourceString() string {
	return fmt.Sprintf(`Or(%s)`, sourceGenerators(o.children))
}

// VariableGen represents a variable in the synopsis. Its values are user-configurable if they cannot be deduced from
// the synopsis.
type VariableGen struct {
	name    string
	options *OrGen
}

var _ StatementGenerator = (*VariableGen)(nil)

// Variable creates a new StatementGenerator representing a VariableGen.
func Variable(name string, children ...StatementGenerator) *VariableGen {
	return &VariableGen{
		name:    name,
		options: Or(children...),
	}
}

// AddChildren implements the interface StatementGenerator.
func (v *VariableGen) AddChildren(children ...StatementGenerator) error {
	children = removeNilGenerators(children)
	if len(children) == 0 {
		return nil
	}
	if len(children) > 1 {
		return fmt.Errorf("attempting to give variable `%s` too many children", v.name)
	}
	if v.options != nil {
		return fmt.Errorf("variable `%s` has already been assigned", v.name)
	}
	orChild, ok := children[0].(*OrGen)
	if !ok {
		return fmt.Errorf("variable `%s` was given an invalid child type `%T`", v.name, children[0])
	}
	v.options = orChild
	return nil
}

// Consume implements the interface StatementGenerator.
func (v *VariableGen) Consume() bool {
	if v.options != nil {
		return v.options.Consume()
	}
	return false
}

// Copy implements the interface StatementGenerator.
func (v *VariableGen) Copy() StatementGenerator {
	if v == nil {
		return nil
	}
	return Variable(v.name, v.options.children...)
}

// String implements the interface StatementGenerator.
func (v *VariableGen) String() string {
	if len(v.options.children) > 0 {
		return v.options.String()
	} else {
		return v.name
	}
}

// Reset implements the interface StatementGenerator.
func (v *VariableGen) Reset() {
	v.options.Reset()
}

// SourceString implements the interface StatementGenerator.
func (v *VariableGen) SourceString() string {
	if len(v.options.children) > 0 {
		return fmt.Sprintf(`Variable("%s", %s)`, v.name, sourceGenerators(v.options.children))
	} else {
		return fmt.Sprintf(`Variable("%s")`, v.name)
	}
}

// CollectionGen is a generator that contains multiple child generators, and will print all of its children.
type CollectionGen struct {
	children []StatementGenerator
}

var _ StatementGenerator = (*CollectionGen)(nil)

// Collection creates a new StatementGenerator representing a CollectionGen.
func Collection(children ...StatementGenerator) *CollectionGen {
	return &CollectionGen{
		children: copyGenerators(children),
	}
}

// AddChildren implements the interface StatementGenerator.
func (c *CollectionGen) AddChildren(children ...StatementGenerator) error {
	c.children = append(c.children, removeNilGenerators(children)...)
	return nil
}

// Consume implements the interface StatementGenerator.
func (c *CollectionGen) Consume() bool {
	for i := range c.children {
		if c.children[i].Consume() {
			return true
		}
	}
	return false
}

// Copy implements the interface StatementGenerator.
func (c *CollectionGen) Copy() StatementGenerator {
	if c == nil {
		return nil
	}
	return Collection(c.children...)
}

// String implements the interface StatementGenerator.
func (c *CollectionGen) String() string {
	var childrenStrings []string
	for i := range c.children {
		childString := c.children[i].String()
		if len(childString) > 0 {
			childrenStrings = append(childrenStrings, childString)
		}
	}
	return strings.Join(childrenStrings, " ")
}

// Reset implements the interface StatementGenerator.
func (c *CollectionGen) Reset() {
	for _, child := range c.children {
		child.Reset()
	}
}

// SourceString implements the interface StatementGenerator.
func (c *CollectionGen) SourceString() string {
	return fmt.Sprintf(`Collection(%s)`, sourceGenerators(c.children))
}

// RepeatGen is a generator that will repeat its children up to the limit, starting with no repetition.
type RepeatGen struct {
	template *CollectionGen
	children *CollectionGen
	start    int
	current  int
	limit    int
}

var _ StatementGenerator = (*RepeatGen)(nil)

// Repeat creates a new StatementGenerator representing a RepeatGen. The start count must be either zero or one, and is
// bounded to whichever is closest. The limit cannot be less than the start count, and is set to the start count in such
// cases.
func Repeat(startCount int, limit int, children ...StatementGenerator) *RepeatGen {
	if startCount < 0 {
		startCount = 0
	} else if startCount > 1 {
		startCount = 1
	}
	if limit < startCount {
		limit = startCount
	}
	repeatGen := &RepeatGen{
		template: Collection(children...),
		children: Collection(),
		start:    startCount,
		current:  startCount,
		limit:    limit,
	}
	if startCount == 1 {
		_ = repeatGen.children.AddChildren(children...)
	}
	return repeatGen
}

// AddChildren implements the interface StatementGenerator.
func (r *RepeatGen) AddChildren(children ...StatementGenerator) error {
	err1 := r.template.AddChildren(children...)
	var err2 error
	if r.start == 1 {
		err2 = r.children.AddChildren(children...)
	}
	return errors.Join(err1, err2)
}

// Consume implements the interface StatementGenerator.
func (r *RepeatGen) Consume() bool {
	if r.children.Consume() {
		return true
	}
	if r.current < r.limit {
		_ = r.children.AddChildren(r.template.Copy())
		r.current++
		return true
	}
	return false
}

// Copy implements the interface StatementGenerator.
func (r *RepeatGen) Copy() StatementGenerator {
	if r == nil {
		return nil
	}
	return Repeat(r.start, r.limit, r.template.children...)
}

// String implements the interface StatementGenerator.
func (r *RepeatGen) String() string {
	return r.children.String()
}

// Reset implements the interface StatementGenerator.
func (r *RepeatGen) Reset() {
	r.current = r.start
	r.children = Collection()
	if r.start == 1 {
		_ = r.children.AddChildren(r.template.children...)
	}
}

// SourceString implements the interface StatementGenerator.
func (r *RepeatGen) SourceString() string {
	if len(r.template.children) > 0 {
		return fmt.Sprintf(`Repeat(%d, %s)`, r.limit, sourceGenerators(r.template.children))
	} else {
		return fmt.Sprintf(`Repeat(%d)`, r.limit)
	}
}

// OptionalGen is a generator that will toggle between displaying its children and not displaying its children.
type OptionalGen struct {
	children *CollectionGen
	display  bool
}

var _ StatementGenerator = (*OptionalGen)(nil)

// Optional creates a new StatementGenerator representing an OptionalGen.
func Optional(children ...StatementGenerator) *OptionalGen {
	return &OptionalGen{
		children: Collection(children...),
		display:  false,
	}
}

// AddChildren implements the interface StatementGenerator.
func (o *OptionalGen) AddChildren(children ...StatementGenerator) error {
	return o.children.AddChildren(children...)
}

// Consume implements the interface StatementGenerator.
func (o *OptionalGen) Consume() bool {
	if !o.display {
		o.display = true
		return true
	} else if o.children.Consume() {
		return true
	} else {
		o.display = false
		return false
	}
}

// Copy implements the interface StatementGenerator.
func (o *OptionalGen) Copy() StatementGenerator {
	if o == nil {
		return nil
	}
	return Optional(o.children.children...)
}

// String implements the interface StatementGenerator.
func (o *OptionalGen) String() string {
	if o.display {
		return o.children.String()
	} else {
		return ""
	}
}

// Reset implements the interface StatementGenerator.
func (o *OptionalGen) Reset() {
	o.display = false
	o.children.Reset()
}

// SourceString implements the interface StatementGenerator.
func (o *OptionalGen) SourceString() string {
	return fmt.Sprintf(`Optional(%s)`, sourceGenerators(o.children.children))
}

// copyGenerators returns a full copy of the given slice of generators. Each generator will be in its original state.
func copyGenerators(gens []StatementGenerator) []StatementGenerator {
	gens = removeNilGenerators(gens)
	if len(gens) == 0 {
		return nil
	}
	newGens := make([]StatementGenerator, len(gens))
	for i, gen := range gens {
		newGens[i] = gen.Copy()
	}
	return newGens
}

// sourceGenerators returns a comma-separated SourceString from the given generator slice.
func sourceGenerators(gens []StatementGenerator) string {
	gens = removeNilGenerators(gens)
	if len(gens) == 0 {
		return ""
	}
	sourceStrs := make([]string, len(gens))
	for i, gen := range gens {
		sourceStrs[i] = gen.SourceString()
	}
	return strings.Join(sourceStrs, ", ")
}

// removeNilGenerators returns a new slice of generators with all nils removed.
func removeNilGenerators(gens []StatementGenerator) []StatementGenerator {
	newGens := make([]StatementGenerator, 0, len(gens))
	for i := range gens {
		if gens[i] != nil {
			newGens = append(newGens, gens[i])
		}
	}
	if len(newGens) == 0 {
		return nil
	}
	return newGens
}
