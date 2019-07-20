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

func (e *ErrElementNotFound) Error() string {
	return fmt.Sprintf("Element at %v not found", e.path)
}

// ErrPathInsideValue is returned if a path is pointing inside a value.
type ErrPathInsideValue struct {
	path string
}

func (e *ErrPathInsideValue) Error() string {
	return fmt.Sprintf("Element at %v is pointing inside value", e.path)
}

// ErrUnexpectedType is returned if a type differs from the expected type.
type ErrUnexpectedType struct {
	path          string
	got, expected string
}

func (e *ErrUnexpectedType) Error() string {
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

// ErrKeyIsNotString is returned if a key of a map is not of type string.
type ErrKeyIsNotString struct {
	key interface{}
}

func (e *ErrKeyIsNotString) Error() string {
	return fmt.Sprintf("Key %v is of type %T. Only strings are supported", e.key, e.key)
}

// ErrNonPointerOrNil is returned when trying to write into a nil or non pointer value.
type ErrNonPointerOrNil struct {
	v interface{}
}

func (e *ErrNonPointerOrNil) Error() string {
	return fmt.Sprintf("Trying to write into non pointer or nil value %v of type %T", e.v, e.v)
}
