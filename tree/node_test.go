// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"reflect"
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

func TestNode_CreatePath(t *testing.T) {
	tests := []struct {
		name     string
		n        Node
		path     string
		want     Node
		wantThis Node
	}{
		{"A", Node{}, "test.123.foo.bar", Node{}, Node{"test": Node{"123": Node{"foo": Node{"bar": Node{}}}}}},
		{"B", Node{"foo": Node{"value": "string", "bar": 1234}}, "foo.bar.baz", Node{}, Node{"foo": Node{"value": "string", "bar": Node{"baz": Node{}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.CreatePath(tt.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.CreatePath() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.n, tt.wantThis) {
				t.Errorf("this = %v, want %v", tt.n, tt.wantThis)
			}
		})
	}
}

type somethingInvalid string

func TestNode_Set(t *testing.T) {
	type args struct {
		path    string
		element interface{}
	}
	tests := []struct {
		name     string
		n        Node
		args     args
		wantErr  bool
		wantThis Node
	}{
		{"A", Node{}, args{"foo.bar", "test"}, false, Node{"foo": Node{"bar": "test"}}},
		{"B", Node{}.CreatePath("foo.bar.baz"), args{"foo.bar", 123}, false, Node{"foo": Node{"bar": int64(123)}}},
		{"C", Node{}, args{"foo.bar", somethingInvalid("test")}, true, Node{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.n.Set(tt.args.path, tt.args.element)
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.n, tt.wantThis) {
				t.Errorf("this = %v, want %v", tt.n, tt.wantThis)
			}
		})
	}
}
