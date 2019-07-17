// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var treeEmpty = Node{}

var treeA = Node{
	"someString": "someString",
	"someNumber": 123,
	"subnode": Node{
		"a": Node{
			"foo": "bar",
		},
		"b": Node{
			"someFloat": 123.456,
		},
		"c": Node{
			"sub": Node{
				"sub": Node{},
			},
		},
		"e": Node{
			"sub": Node{
				"val": "string",
			},
		},
	},
}

var treeB = Node{
	"someString": "someStringEdited",
	"someNumber": 123,
	"subnode": Node{
		"a": Node{
			"foo": Node{
				"sub": Node{},
			},
		},
		"b": Node{
			"someFloat": 123.4567,
		},
		"c": "NothingToSeeHere",
		"d": Node{
			"sub": Node{
				"sub": Node{},
			},
		},
	},
}

func TestNode_Compare(t *testing.T) {
	tests := []struct {
		name         string
		n            Node
		new          Node
		wantModified []string
		wantAdded    []string
		wantRemoved  []string
	}{
		{"A -> B", treeA, treeB,
			[]string{"someString", "subnode.a.foo", "subnode.b.someFloat", "subnode.c"},
			[]string{"subnode.a.foo.sub", "subnode.d", "subnode.d.sub", "subnode.d.sub.sub"},
			[]string{"subnode.c.sub", "subnode.c.sub.sub", "subnode.e", "subnode.e.sub", "subnode.e.sub.val"},
		},
		{"A -> 0", treeA, treeEmpty,
			[]string{},
			[]string{},
			[]string{"someString", "someNumber", "subnode", "subnode.a", "subnode.b", "subnode.c", "subnode.e", "subnode.a.foo", "subnode.b.someFloat", "subnode.c.sub", "subnode.c.sub.sub", "subnode.e.sub", "subnode.e.sub.val"},
		},
		{"0 -> A", treeEmpty, treeA,
			[]string{},
			[]string{"someString", "someNumber", "subnode", "subnode.a", "subnode.b", "subnode.c", "subnode.e", "subnode.a.foo", "subnode.b.someFloat", "subnode.c.sub", "subnode.c.sub.sub", "subnode.e.sub", "subnode.e.sub.val"},
			[]string{},
		},
	}
	sortOption := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModified, gotAdded, gotRemoved := tt.n.Compare(tt.new)
			if !cmp.Equal(gotModified, tt.wantModified, sortOption, cmpopts.EquateEmpty()) {
				t.Errorf("Node.Compare() gotModified = %v, want %v", gotModified, tt.wantModified)
			}
			if !cmp.Equal(gotAdded, tt.wantAdded, sortOption, cmpopts.EquateEmpty()) {
				t.Errorf("Node.Compare() gotAdded = %v, want %v", gotAdded, tt.wantAdded)
			}
			if !cmp.Equal(gotRemoved, tt.wantRemoved, sortOption, cmpopts.EquateEmpty()) {
				t.Errorf("Node.Compare() gotRemoved = %v, want %v", gotRemoved, tt.wantRemoved)
			}
		})
	}
}
