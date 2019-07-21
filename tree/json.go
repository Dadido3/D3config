// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
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

	new, err := anyToTree(reflect.ValueOf(root))
	if err != nil {
		return err
	}
	newRoot, ok := new.(Node)
	if !ok {
		return &ErrUnexpectedType{"", fmt.Sprintf("%T", new), "Node"}
	}

	for k := range n {
		delete(n, k)
	}
	for k, v := range newRoot {
		n[k] = v
	}

	return nil
}
