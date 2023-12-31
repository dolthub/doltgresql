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

package utils

// Stack is a generic stack.
type Stack[T any] struct {
	values []T
}

// NewStack creates a new, empty stack.
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

// Len returns the size of the stack.
func (s *Stack[T]) Len() int {
	return len(s.values)
}

// Peek returns the top value on the stack without removing it.
func (s *Stack[T]) Peek() (value T) {
	if len(s.values) == 0 {
		return
	}
	return s.values[len(s.values)-1]
}

// PeekDepth returns the n-th value from the top. PeekDepth(0) is equivalent to the standard Peek().
func (s *Stack[T]) PeekDepth(depth int) (value T) {
	if len(s.values) <= depth {
		return
	}
	return s.values[len(s.values)-(1+depth)]
}

// PeekReference returns a reference to the top value on the stack without removing it.
func (s *Stack[T]) PeekReference() *T {
	if len(s.values) == 0 {
		return nil
	}
	return &s.values[len(s.values)-1]
}

// Pop returns the top value on the stack while also removing it from the stack.
func (s *Stack[T]) Pop() (value T) {
	if len(s.values) == 0 {
		return
	}
	value = s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return
}

// Push adds the given value to the stack.
func (s *Stack[T]) Push(value T) {
	s.values = append(s.values, value)
}

// Empty returns whether the stack is empty.
func (s *Stack[T]) Empty() bool {
	return len(s.values) == 0
}
