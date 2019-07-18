// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"bytes"
	"encoding/json"
)

// UnmarshalJSON will unmarshal a json string into a Node object.
// It converts types to Node or Number when needed.
func (n Node) UnmarshalJSON(b []byte) error {

	var root map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&root)
	if err != nil {
		return err
	}

	for k := range n {
		delete(n, k)
	}

	// Replace objects in the tree with Node or Number if possible
	var recursive func(v interface{}) interface{}
	recursive = func(v interface{}) interface{} {
		switch v := v.(type) {
		case map[string]interface{}: // Replace map with node element
			node := Node(v)

			for k, n := range node {
				node[k] = recursive(n)
			}

			return node
		case []interface{}: // Replace elements inside of an array
			for i, n := range v {
				v[i] = recursive(n)
			}

			return v
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number: // Replace any type of number with Number
			n, err := NumberCreate(v)
			if err != nil {
				return nil
			}
			return n
		}

		return v
	}
	newRoot := recursive(root).(Node) // If root isn't of type Node, something went totally wrong

	for k, v := range newRoot {
		n[k] = v
	}

	return nil
}
