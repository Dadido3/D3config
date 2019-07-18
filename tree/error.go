// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"fmt"
)

// ErrElementNotFound is returned if the element couldn't be found at the given path.
type ErrElementNotFound struct {
	path string
}

func (e ErrElementNotFound) Error() string {
	return fmt.Sprintf("Element at %v not found", e.path)
}

// ErrPathInsideValue is returned if a path is pointing inside a value.
type ErrPathInsideValue struct {
	path string
}

func (e ErrPathInsideValue) Error() string {
	return fmt.Sprintf("Element at %v is pointing inside value", e.path)
}

// ErrUnexpectedType is returned if a type differs from the expected type.
type ErrUnexpectedType struct {
	path          string
	got, expected string
}

func (e ErrUnexpectedType) Error() string {
	if e.path != "" {
		if e.expected != "" {
			return fmt.Sprintf("Element at %v is of type %v instead of %v", e.path, e.got, e.expected)
		}
		return fmt.Sprintf("Element at %v is of unexpected type %v", e.path, e.got)
	}
	if e.expected != "" {
		return fmt.Sprintf("Element is of type %v instead of %v", e.got, e.expected)
	}
	return fmt.Sprintf("Element is of unexpected type %v", e.got)
}
