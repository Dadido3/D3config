// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package configdb

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Dadido3/configdb/tree"
)

var treeA = tree.Node{
	"someString": "someString",
	"someNumber": tree.Number("123"),
	"subnode": tree.Node{
		"a": tree.Node{
			"foo": "bar",
		},
		"b": tree.Node{
			"someFloat": tree.Number("123.456"),
		},
		"c": tree.Node{
			"sub": tree.Node{
				"sub": tree.Node{},
			},
		},
		"e": tree.Node{
			"sub": tree.Node{
				"val": "string",
			},
		},
		"f": []interface{}{
			tree.Node{"sub": tree.Node{}},
			tree.Node{"val": true},
		},
		"g": []interface{}{
			tree.Node{"sub": tree.Node{"val": false}},
			tree.Node{"val": true},
		},
	},
}

func TestYAML(t *testing.T) {
	c, err := New([]Storage{UseYAMLFile(filepath.Join(".", "testfiles", "yaml", "a.yml"))})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer c.Close()

	if err := c.Set("", treeA); err != nil {
		t.Errorf("Set() failed: %v", err)
	}

	var readBack tree.Node

	if err := c.Get("", &readBack); err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if err := readBack.Check(); err != nil {
		t.Errorf("Illegal element in tree: %v", err)
	}
	if !reflect.DeepEqual(treeA, readBack) {
		t.Errorf("got %#v, want %#v", treeA, readBack)
	}

}
