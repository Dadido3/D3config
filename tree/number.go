// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Number can represent any integer or float type.
type Number string

// NumberCreate takes any number type (integers, floats), and returns a number object that represents the exact value.
func NumberCreate(num interface{}) (Number, error) {
	switch v := num.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return Number(fmt.Sprintf("%d", v)), nil
	case float32, float64:
		return Number(fmt.Sprintf("%f", v)), nil
	case json.Number:
		return Number(v), nil
	}

	return "", ErrUnexpectedType{"", fmt.Sprintf("%T", num), ""}
}

// Int returns the number interpreted as int, or an error if the type doesn't match.
func (n Number) Int() (int, error) {
	result, err := strconv.ParseInt(string(n), 0, 0)
	return int(result), err
}

// Int64 returns the number interpreted as int64, or an error if the type doesn't match.
func (n Number) Int64() (int64, error) {
	result, err := strconv.ParseInt(string(n), 0, 64)
	return result, err
}

// Uint returns the number interpreted as uint, or an error if the type doesn't match.
func (n Number) Uint() (uint, error) {
	result, err := strconv.ParseUint(string(n), 0, 0)
	return uint(result), err
}

// Uint64 returns the number interpreted as uint64, or an error if the type doesn't match.
func (n Number) Uint64() (uint64, error) {
	result, err := strconv.ParseUint(string(n), 0, 64)
	return result, err
}

// Float64 returns the number interpreted as float64, or an error if the type doesn't match.
func (n Number) Float64() (float64, error) {
	result, err := strconv.ParseFloat(string(n), 64)
	return result, err
}

func (n Number) String() (string, error) {
	return string(n), nil
}

// MarshalJSON writes the raw string of the Number type
func (n Number) MarshalJSON() ([]byte, error) {
	return []byte(n), nil
}
