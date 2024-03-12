// Copyright 2023-2024 Dolthub, Inc.
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

package utils

import (
	"fmt"
	"math"
	"math/big"
	"sort"
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
	// SetConsumeIterations is equivalent to calling Copy then Consume the given number of times, without allocating a
	// new StatementGenerator. This allows you to generate a specific statement efficiently, rather than calling Consume
	// the given number of times. If the count is <= 0, then the statement will be in its original state (the same state
	// as a StatementGenerator copy).
	SetConsumeIterations(count *big.Int)
	// SetConsumeIterationsFast is the same as SetConsumeIterations, except far more efficient due to using uint64,
	// however it only works for iteration counts <= MAX_SIZE(uint64).
	SetConsumeIterationsFast(count uint64)
	// String returns a string based on the current permutation.
	String() string
	// Copy returns a copy of the given generator (along with all of its children) in its original setting. This means
	// that the copy is in the same state that the target would be in if it had never called Consume.
	Copy() StatementGenerator
	// Reset sets the StatementGenerator back to its original state, which would be as though Consume was never called.
	// This is equivalent to calling SetConsumeIterations(0), albeit slightly more efficient.
	Reset()
	// SourceString returns a string that may be used to recreate the StatementGenerator in a Go source file.
	SourceString() string
	// Permutations returns the number of unique permutations that the generator can return.
	Permutations() *big.Int
	// PermutationsUint64 returns the number of unique permutations that the generator can return. Returns true if the
	// number fits within an uint64, false if it's larger than an uint64.
	PermutationsUint64() (uint64, bool)
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

// SetConsumeIterations implements the interface StatementGenerator.
func (t *TextGen) SetConsumeIterations(count *big.Int) {}

// SetConsumeIterationsFast implements the interface StatementGenerator.
func (t *TextGen) SetConsumeIterationsFast(count uint64) {}

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

// Permutations implements the interface StatementGenerator.
func (t *TextGen) Permutations() *big.Int {
	return BigIntOne
}

// PermutationsUint64 implements the interface StatementGenerator.
func (t *TextGen) PermutationsUint64() (uint64, bool) {
	return 1, true
}

// OrGen is a generator that contains multiple child generators, and will print only one at a time. Consuming will
// cycle to the next child.
type OrGen struct {
	children []StatementGenerator
	index    int
	localInt *big.Int
}

var _ StatementGenerator = (*OrGen)(nil)

// Or creates a new StatementGenerator representing an OrGen.
func Or(children ...StatementGenerator) *OrGen {
	return &OrGen{
		children: copyGenerators(children),
		index:    0,
		localInt: new(big.Int),
	}
}

// AddChildren implements the interface StatementGenerator.
func (o *OrGen) AddChildren(children ...StatementGenerator) error {
	o.children = append(o.children, copyGenerators(children)...)
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

// SetConsumeIterations implements the interface StatementGenerator.
func (o *OrGen) SetConsumeIterations(count *big.Int) {
	// If we're given zero, then we'll just call Reset
	if count.Cmp(BigIntZero) <= 0 {
		o.Reset()
		return
	}
	count = o.localInt.Mod(count, o.Permutations())
	for i, child := range o.children {
		// The index is equal to whichever child we stop on
		o.index = i
		childPermutations := child.Permutations()
		if childPermutations.Cmp(count) > 0 {
			// The child has more permutations than the count, so we'll stop here
			if count.Cmp(BigIntMaxUint64) <= 0 {
				child.SetConsumeIterationsFast(count.Uint64())
			} else {
				child.SetConsumeIterations(count)
			}
			break
		} else {
			// The child's permutations are <= the count, so we'll reset it and subtract it from the total.
			// Subtraction here is the opposite of the addition we do to determine the permutation count.
			// Important to note that the permutations equaling the count means that the index increments to the next
			// item, but since the count will be zero, it matches the original state of that item.
			child.Reset()
			count.Sub(count, childPermutations)
		}
	}
	// We still need to reset any children that we never looped over
	for i := o.index + 1; i < len(o.children); i++ {
		o.children[i].Reset()
	}
}

// SetConsumeIterationsFast implements the interface StatementGenerator.
func (o *OrGen) SetConsumeIterationsFast(count uint64) {
	// This is a copy of SetConsumeIterations, except rewritten to use uint64
	if count <= 0 {
		o.Reset()
		return
	}
	permutations, _ := o.PermutationsUint64()
	count = count % permutations
	for i, child := range o.children {
		o.index = i
		childPermutations, _ := child.PermutationsUint64()
		if childPermutations > count {
			child.SetConsumeIterationsFast(count)
			break
		} else {
			child.Reset()
			count -= childPermutations
		}
	}
	for i := o.index + 1; i < len(o.children); i++ {
		o.children[i].Reset()
	}
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

// Permutations implements the interface StatementGenerator.
func (o *OrGen) Permutations() *big.Int {
	sum := big.NewInt(0)
	for _, child := range o.children {
		sum.Add(sum, child.Permutations())
	}
	return sum
}

// PermutationsUint64 implements the interface StatementGenerator.
func (o *OrGen) PermutationsUint64() (uint64, bool) {
	sum := uint64(0)
	for _, child := range o.children {
		childCount, ok := child.PermutationsUint64()
		if !ok || sum > (math.MaxUint64-childCount) {
			return math.MaxUint64, false
		}
		sum += childCount
	}
	return sum, true
}

// VariableGen represents a variable in the synopsis. Its values are user-configurable if they cannot be deduced from
// the synopsis.
type VariableGen struct {
	name    string
	options StatementGenerator
}

var _ StatementGenerator = (*VariableGen)(nil)

// Variable creates a new StatementGenerator representing a VariableGen.
func Variable(name string, child StatementGenerator) *VariableGen {
	if child != nil {
		return &VariableGen{
			name:    name,
			options: child.Copy(),
		}
	} else {
		return &VariableGen{
			name:    name,
			options: nil,
		}
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
	v.options = children[0].Copy()
	return nil
}

// Consume implements the interface StatementGenerator.
func (v *VariableGen) Consume() bool {
	if v.options != nil {
		return v.options.Consume()
	}
	return false
}

// SetConsumeIterations implements the interface StatementGenerator.
func (v *VariableGen) SetConsumeIterations(count *big.Int) {
	if v.options != nil {
		v.options.SetConsumeIterations(count)
	}
}

// SetConsumeIterationsFast implements the interface StatementGenerator.
func (v *VariableGen) SetConsumeIterationsFast(count uint64) {
	if v.options != nil {
		v.options.SetConsumeIterationsFast(count)
	}
}

// Copy implements the interface StatementGenerator.
func (v *VariableGen) Copy() StatementGenerator {
	if v == nil {
		return nil
	}
	return Variable(v.name, v.options)
}

// String implements the interface StatementGenerator.
func (v *VariableGen) String() string {
	if v.options != nil {
		return v.options.String()
	} else {
		return v.name
	}
}

// Reset implements the interface StatementGenerator.
func (v *VariableGen) Reset() {
	if v.options != nil {
		v.options.Reset()
	}
}

// SourceString implements the interface StatementGenerator.
func (v *VariableGen) SourceString() string {
	if v.options != nil {
		return fmt.Sprintf(`Variable("%s", %s)`, v.name, v.options.SourceString())
	} else {
		return fmt.Sprintf(`Variable("%s", nil)`, v.name)
	}
}

// Permutations implements the interface StatementGenerator.
func (v *VariableGen) Permutations() *big.Int {
	if v.options != nil {
		return v.options.Permutations()
	} else {
		return BigIntOne
	}
}

// PermutationsUint64 implements the interface StatementGenerator.
func (v *VariableGen) PermutationsUint64() (uint64, bool) {
	if v.options != nil {
		return v.options.PermutationsUint64()
	} else {
		return 1, true
	}
}

// CollectionGen is a generator that contains multiple child generators, and will print all of its children.
type CollectionGen struct {
	children []StatementGenerator
	localInt *big.Int
}

var _ StatementGenerator = (*CollectionGen)(nil)

// Collection creates a new StatementGenerator representing a CollectionGen.
func Collection(children ...StatementGenerator) *CollectionGen {
	return &CollectionGen{
		children: copyGenerators(children),
		localInt: new(big.Int),
	}
}

// AddChildren implements the interface StatementGenerator.
func (c *CollectionGen) AddChildren(children ...StatementGenerator) error {
	c.children = append(c.children, copyGenerators(children)...)
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

// SetConsumeIterations implements the interface StatementGenerator.
func (c *CollectionGen) SetConsumeIterations(count *big.Int) {
	// We handle this one as though it's a non-uniform numbering system (binary and decimal are uniform systems).
	// In a traditional number system like binary, you can find each bit's value using the following:
	//
	// bit = number % 2; number = number / 2;
	//
	// Collections behave similarly to that system, where we increment the second generator after fully incrementing the
	// first generator. Then we have to iterate over the first generator again before we can increment the second
	// generator again. Do this until the second generator has exhausted its permutations, and then the third generator
	// can increment.
	//
	// Going back to our binary example, we can achieve that same counting effect by replacing 2 with the permutation
	// count. This lets us have our non-uniform numbering system, and allows us to efficiently find the exact number for
	// each generator.
	count = c.localInt.Mod(count, c.Permutations())
	index := 0
	for i, child := range c.children {
		// The index is equal to whichever child we stop on
		index = i
		childPermutations := child.Permutations()
		// We give the child the modulo of the count versus its permutation count, which will determine how many
		// iterations it's supposed to simulate from the total.
		childIterations := new(big.Int).Mod(count, childPermutations)
		if childIterations.Cmp(BigIntMaxUint64) <= 0 {
			child.SetConsumeIterationsFast(childIterations.Uint64())
		} else {
			child.SetConsumeIterations(childIterations)
		}
		// We divide the count by this child's permutation count to move to the next "base".
		count.Div(count, childPermutations)
		// If we're at zero now, then this child used up the remaining count, so we'll stop here
		if count.Cmp(BigIntZero) <= 0 {
			break
		}
	}
	// We still need to reset any children that we never looped over
	for index += 1; index < len(c.children); index++ {
		c.children[index].Reset()
	}
}

// SetConsumeIterationsFast implements the interface StatementGenerator.
func (c *CollectionGen) SetConsumeIterationsFast(count uint64) {
	// This is a copy of SetConsumeIterations, except rewritten to use uint64
	permutations, _ := c.PermutationsUint64()
	count = count % permutations
	index := 0
	for i, child := range c.children {
		index = i
		childPermutations, _ := child.PermutationsUint64()
		child.SetConsumeIterationsFast(count % childPermutations)
		count /= childPermutations
		if count <= 0 {
			break
		}
	}
	for index += 1; index < len(c.children); index++ {
		c.children[index].Reset()
	}
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

// Permutations implements the interface StatementGenerator.
func (c *CollectionGen) Permutations() *big.Int {
	total := big.NewInt(1)
	for _, child := range c.children {
		childPermutations := child.Permutations()
		if childPermutations.Cmp(BigIntZero) != 0 {
			total.Mul(total, childPermutations)
		}
	}
	return total
}

// PermutationsUint64 implements the interface StatementGenerator.
func (c *CollectionGen) PermutationsUint64() (uint64, bool) {
	total := uint64(1)
	for _, child := range c.children {
		childPermutations, ok := child.PermutationsUint64()
		if !ok {
			return math.MaxUint64, false
		}
		if childPermutations == 0 {
			continue
		}
		if total > math.MaxUint64/childPermutations {
			return math.MaxUint64, false
		}
		total *= childPermutations
	}
	return total, true
}

// OptionalGen is a generator that will toggle between displaying its children and not displaying its children.
type OptionalGen struct {
	children *CollectionGen
	display  bool
	localInt *big.Int
}

var _ StatementGenerator = (*OptionalGen)(nil)

// Optional creates a new StatementGenerator representing an OptionalGen.
func Optional(children ...StatementGenerator) *OptionalGen {
	return &OptionalGen{
		children: Collection(children...),
		display:  false,
		localInt: new(big.Int),
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

// SetConsumeIterations implements the interface StatementGenerator.
func (o *OptionalGen) SetConsumeIterations(count *big.Int) {
	// If we're given zero, then we'll just call Reset
	if count.Cmp(BigIntZero) <= 0 {
		o.Reset()
		return
	}
	// The count is >= 1, so display will be true
	o.display = true
	count = o.localInt.Mod(count, o.Permutations())
	// Setting display to true uses a single Consume, so we subtract it before passing the count to the child
	count.Sub(count, BigIntOne)
	// We'll pass the rest of the remaining count to the child, which will be >= 0
	o.children.SetConsumeIterations(count)
}

// SetConsumeIterationsFast implements the interface StatementGenerator.
func (o *OptionalGen) SetConsumeIterationsFast(count uint64) {
	// This is a copy of SetConsumeIterations, except rewritten to use uint64
	if count <= 0 {
		o.Reset()
		return
	}
	o.display = true
	permutations, _ := o.PermutationsUint64()
	count = count % permutations
	count -= 1
	o.children.SetConsumeIterationsFast(count)
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

// Permutations implements the interface StatementGenerator.
func (o *OptionalGen) Permutations() *big.Int {
	return new(big.Int).Add(BigIntOne, o.children.Permutations())
}

// PermutationsUint64 implements the interface StatementGenerator.
func (o *OptionalGen) PermutationsUint64() (uint64, bool) {
	childCount, ok := o.children.PermutationsUint64()
	if !ok || childCount == math.MaxUint64 {
		return math.MaxUint64, false
	}
	return 1 + childCount, true
}

// ApplyVariableDefinition applies the given map of variable definitions to the statement generator. This modifies the
// statement generator, rather than returning a copy.
func ApplyVariableDefinition(gen StatementGenerator, definitions map[string]StatementGenerator) error {
	if len(definitions) == 0 {
		return nil
	}
	switch gen := gen.(type) {
	case *CollectionGen:
		for _, child := range gen.children {
			if err := ApplyVariableDefinition(child, definitions); err != nil {
				return err
			}
		}
	case *OptionalGen:
		if err := ApplyVariableDefinition(gen.children, definitions); err != nil {
			return err
		}
	case *OrGen:
		for _, child := range gen.children {
			if err := ApplyVariableDefinition(child, definitions); err != nil {
				return err
			}
		}
	case *TextGen:
		// Nothing to do here
	case *VariableGen:
		if gen.options == nil {
			if definition, ok := definitions[gen.name]; ok {
				if err := gen.AddChildren(definition); err != nil {
					return err
				}
				if err := ApplyVariableDefinition(gen.options, definitions); err != nil {
					return err
				}
			}
		} else {
			if err := ApplyVariableDefinition(gen.options, definitions); err != nil {
				return err
			}
		}
	case nil:
		return nil
	default:
		return fmt.Errorf("unknown generator encountered: %T", gen)
	}
	return nil
}

// UnsetVariables returns the name of all variables that do not have a definition. Sorted in ascending order.
func UnsetVariables(gen StatementGenerator) ([]string, error) {
	varNames := make(map[string]struct{})
	switch gen := gen.(type) {
	case *CollectionGen:
		for _, child := range gen.children {
			children, err := UnsetVariables(child)
			if err != nil {
				return nil, err
			}
			for _, childName := range children {
				varNames[childName] = struct{}{}
			}
		}
	case *OptionalGen:
		return UnsetVariables(gen.children)
	case *OrGen:
		return UnsetVariables(Collection(gen.children...))
	case *TextGen:
		// Nothing to do here
	case *VariableGen:
		if gen.options == nil {
			return []string{gen.name}, nil
		} else {
			return UnsetVariables(gen.options)
		}
	default:
		return nil, fmt.Errorf("unknown generator encountered: %T", gen)
	}
	var varNamesSlice []string
	for varName := range varNames {
		varNamesSlice = append(varNamesSlice, varName)
	}
	sort.Strings(varNamesSlice)
	return varNamesSlice, nil
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
