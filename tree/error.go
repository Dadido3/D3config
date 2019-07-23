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
	Path string
}

func (e *ErrElementNotFound) Error() string {
	return fmt.Sprintf("Element at %v not found", e.Path)
}

// ErrPathInsideValue is returned if a path is pointing inside a value.
type ErrPathInsideValue struct {
	Path string
}

func (e *ErrPathInsideValue) Error() string {
	return fmt.Sprintf("Element at %v is pointing inside value", e.Path)
}

// ErrPathInvalid is returned if a path is invalid.
type ErrPathInvalid struct {
	Path   string
	Reason string
}

func (e *ErrPathInvalid) Error() string {
	return fmt.Sprintf("Path %v is invalid: %v", e.Path, e.Reason)
}

// ErrUnexpectedType is returned if a type differs from the expected type.
type ErrUnexpectedType struct {
	Path          string
	Got, Expected string
}

func (e *ErrUnexpectedType) Error() string {
	if e.Path != "" {
		if e.Expected != "" {
			return fmt.Sprintf("Element at %v is of type %v instead of %v", e.Path, e.Got, e.Expected)
		}
		return fmt.Sprintf("Element at %v is of unexpected type %v", e.Path, e.Got)
	}
	if e.Expected != "" {
		return fmt.Sprintf("Element is of type %v instead of %v", e.Got, e.Expected)
	}
	return fmt.Sprintf("Element is of unexpected type %v", e.Got)
}

// ErrKeyIsNotString is returned if a key of a map is not of type string.
type ErrKeyIsNotString struct {
	Key, Type string
}

func (e *ErrKeyIsNotString) Error() string {
	return fmt.Sprintf("Key %v is of type %v. Only strings are supported", e.Key, e.Type)
}

// ErrCannotModify is returned when trying to write into a nil or non pointer value.
type ErrCannotModify struct {
	Value, Type string
}

func (e *ErrCannotModify) Error() string {
	return fmt.Sprintf("Trying to write into non pointer or nil value %v of type %v", e.Value, e.Type)
}
