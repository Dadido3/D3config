// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package configdb

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		"f": []tree.Node{
			tree.Node{"sub": tree.Node{}},
			tree.Node{"val": true},
		},
		"g": []tree.Node{
			tree.Node{"sub": tree.Node{"val": false}},
			tree.Node{"val": true},
		},
	},
}

func TestMarshallingAndUnmarshalling(t *testing.T) {
	if err := treeA.Check(); err != nil {
		t.Errorf("Illegal element in tree: %v", err)
	}

	bytes, err := json.MarshalIndent(treeA, "", "  ")
	if err != nil {
		t.Errorf("json.Marshal() failed: %v", err)
	}
	f, err := os.Create(filepath.Join(".", "testfiles", "json", "a.json"))
	if err != nil {
		t.Errorf("os.Create() failed: %v", err)
	}
	defer f.Close()

	f.Write(bytes)

	result := tree.Node{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		t.Errorf("json.Unmarshal() failed: %v", err)
	}
	if err := result.Check(); err != nil {
		t.Errorf("Illegal element in tree: %v", err)
	}

	bytes, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Errorf("json.Marshal() failed: %v", err)
	}
	f, err = os.Create(filepath.Join(".", "testfiles", "json", "a_readback.json"))
	if err != nil {
		t.Errorf("os.Create() failed: %v", err)
	}
	defer f.Close()

	f.Write(bytes)
}
