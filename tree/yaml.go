// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"fmt"
	"reflect"
)

// UnmarshalYAML will unmarshal yaml data into a Node object.
// It converts anything to valid tree objects.
func (n Node) UnmarshalYAML(unmarshal func(v interface{}) error) error {

	var root map[string]interface{}
	err := unmarshal(&root)
	if err != nil {
		return err
	}

	new, err := marshal(reflect.ValueOf(root))
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
