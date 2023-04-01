// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var treeEmpty = Node{}

var treeA = Node{
	"someString": "someString",
	"someNumber": Number("123"),
	"subnode": Node{
		"a": Node{
			"foo": "bar",
		},
		"b": Node{
			"someFloat": Number("123.456"),
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
		"f": []interface{}{
			Node{"sub": Node{}},
			Node{"val": true},
		},
		"g": []interface{}{
			Node{"sub": Node{"val": false}},
			Node{"val": true},
		},
	},
}

var treeB = Node{
	"someString": "someStringEdited",
	"someNumber": Number("123"),
	"subnode": Node{
		"a": Node{
			"foo": Node{
				"sub": Node{},
			},
		},
		"b": Node{
			"someFloat": Number("123.4567"),
		},
		"c": "NothingToSeeHere",
		"d": Node{
			"sub": Node{
				"sub": Node{},
			},
		},
		"f": []interface{}{
			Node{"sub": Node{"val": false}},
			Node{"val": true},
		},
		"g": []interface{}{
			Node{"sub": Node{"val": false}},
			Node{"val": true},
		},
	},
}

var treeAB = Node{
	"someString": "someStringEdited",
	"someNumber": Number("123"),
	"subnode": Node{
		"a": Node{
			"foo": Node{
				"sub": Node{},
			},
		},
		"b": Node{
			"someFloat": Number("123.4567"),
		},
		"c": "NothingToSeeHere",
		"d": Node{
			"sub": Node{
				"sub": Node{},
			},
		},
		"e": Node{
			"sub": Node{
				"val": "string",
			},
		},
		"f": []interface{}{
			Node{"sub": Node{"val": false}},
			Node{"val": true},
		},
		"g": []interface{}{
			Node{"sub": Node{"val": false}},
			Node{"val": true},
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
			[]string{".someString", ".subnode.a.foo", ".subnode.b.someFloat", ".subnode.c", ".subnode.f"},
			[]string{".subnode.a.foo.sub", ".subnode.d", ".subnode.d.sub", ".subnode.d.sub.sub"},
			[]string{".subnode.c.sub", ".subnode.c.sub.sub", ".subnode.e", ".subnode.e.sub", ".subnode.e.sub.val"},
		},
		{"A -> 0", treeA, treeEmpty,
			[]string{},
			[]string{},
			[]string{".someString", ".someNumber", ".subnode", ".subnode.a", ".subnode.b", ".subnode.c", ".subnode.e", ".subnode.a.foo", ".subnode.b.someFloat", ".subnode.c.sub", ".subnode.c.sub.sub", ".subnode.e.sub", ".subnode.e.sub.val", ".subnode.f", ".subnode.g"},
		},
		{"0 -> A", treeEmpty, treeA,
			[]string{},
			[]string{".someString", ".someNumber", ".subnode", ".subnode.a", ".subnode.b", ".subnode.c", ".subnode.e", ".subnode.a.foo", ".subnode.b.someFloat", ".subnode.c.sub", ".subnode.c.sub.sub", ".subnode.e.sub", ".subnode.e.sub.val", ".subnode.f", ".subnode.g"},
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
		{"A", Node{}, ".test.123.foo.bar", Node{}, Node{"test": Node{"123": Node{"foo": Node{"bar": Node{}}}}}},
		{"B", Node{"foo": Node{"value": "string", "bar": Number("1234")}}, ".foo.bar.baz", Node{}, Node{"foo": Node{"value": "string", "bar": Node{"baz": Node{}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.CreatePath(tt.path)
			if err != nil {
				t.Errorf("CreatePath() failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.CreatePath() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(tt.n, tt.wantThis) {
				t.Errorf("this = %v, want %v", tt.n, tt.wantThis)
			}
			if err := got.Check(); err != nil {
				t.Errorf("Illegal element in tree: %v", err)
			}
			if err := tt.n.Check(); err != nil {
				t.Errorf("Illegal element in tree: %v", err)
			}
		})
	}
}

type customType string

func TestNode_Set(t *testing.T) {
	type args struct {
		path    string
		element interface{}
	}

	tempNode, err := Node{}.CreatePath(".foo.bar.baz")
	if err != nil {
		t.Fatalf("CreatePath() failed: %v", err)
	}

	tests := []struct {
		name     string
		n        Node
		args     args
		wantErr  bool
		wantThis Node
	}{
		{"A", Node{}, args{".foo.bar", "test"}, false, Node{"foo": Node{"bar": "test"}}},
		{"B", tempNode, args{".foo.bar", 123}, false, Node{"foo": Node{"bar": Number("123")}}},
		{"C", Node{}, args{".foo.bar", customType("test")}, false, Node{"foo": Node{"bar": "test"}}},
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
			if err := tt.n.Check(); err != nil {
				t.Errorf("Illegal element in tree: %v", err)
			}
		})
	}
}

func TestSetGet_Struct(t *testing.T) {
	type subStruct struct {
		SomeString string
	}
	type testStruct struct {
		SomeString        string
		SomeInt           int8
		SomeBool          bool
		SomeFloat         float64 `conf:"somethingRenamed"`
		SomeSubstruct     subStruct
		SomePointerStruct *subStruct
		SomeMap           map[string]int
		SomeSlice         []bool
		SomeArray         [5]bool
		SomeNilPointer    *subStruct
	}
	s := testStruct{
		"test",
		-5,
		true,
		123.456,
		subStruct{"bla"},
		&subStruct{"foo"},
		map[string]int{"a": -1, "b": 0, "c": 1},
		[]bool{true, false, true},
		[5]bool{true, false, true},
		nil,
	}

	tree := Node{}

	err := tree.Set(".somePath", s)
	if err != nil {
		t.Errorf("tree.Set() failed: %v", err)
	}
	if err := tree.Check(); err != nil {
		t.Errorf("Illegal element in tree: %v", err)
	}

	want := Node{
		"somePath": Node{
			"SomeString":       "test",
			"SomeInt":          Number("-5"),
			"SomeBool":         true,
			"somethingRenamed": Number("123.456"),
			"SomeSubstruct": Node{
				"SomeString": "bla",
			},
			"SomePointerStruct": Node{
				"SomeString": "foo",
			},
			"SomeMap": Node{
				"a": Number("-1"),
				"b": Number("0"),
				"c": Number("1"),
			},
			"SomeSlice":      []interface{}{true, false, true},
			"SomeArray":      []interface{}{true, false, true, false, false},
			"SomeNilPointer": nil,
		},
	}

	if !reflect.DeepEqual(tree, want) {
		t.Errorf("Got %v, want %v", tree, want)
	}

	var sResult testStruct

	err = tree.Get(".somePath", &sResult)
	if err != nil {
		t.Errorf("tree.Set() failed: %v", err)
	}
	if !reflect.DeepEqual(s, sResult) {
		t.Errorf("Original value is %v, but got back %v", s, sResult)
	}
}

func TestNode_Merge(t *testing.T) {
	tests := []struct {
		name     string
		n        Node
		new      Node
		wantThis Node
	}{
		{"A", treeA, treeB, treeAB},
		{"B", treeEmpty, treeA, treeA},
		{"C", treeA, treeEmpty, treeA},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.Merge(tt.new)
			if !reflect.DeepEqual(tt.n, tt.wantThis) {
				t.Errorf("this = %v, want %v", tt.n, tt.wantThis)
			}
			if err := tt.n.Check(); err != nil {
				t.Errorf("Illegal element in tree: %v", err)
			}
		})
	}
}

func TestMarshallingAndUnmarshalling(t *testing.T) {
	if err := treeA.Check(); err != nil {
		t.Errorf("Illegal element in tree: %v", err)
	}

	bytes, err := json.MarshalIndent(treeA, "", "    ")
	if err != nil {
		t.Errorf("json.Marshal() failed: %v", err)
	}
	f, err := os.Create(filepath.Join(".", "..", "testfiles", "json", "a.json"))
	if err != nil {
		t.Errorf("os.Create() failed: %v", err)
	}
	defer f.Close()

	f.Write(bytes)

	result := Node{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		t.Errorf("json.Unmarshal() failed: %v", err)
	}
	if err := result.Check(); err != nil {
		t.Errorf("Illegal element in tree: %v", err)
	}

	bytes, err = json.MarshalIndent(result, "", "    ")
	if err != nil {
		t.Errorf("json.Marshal() failed: %v", err)
	}
	f, err = os.Create(filepath.Join(".", "..", "testfiles", "json", "a_readback.json"))
	if err != nil {
		t.Errorf("os.Create() failed: %v", err)
	}
	defer f.Close()

	f.Write(bytes)
}

func TestNode_Remove(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name     string
		n        Node
		args     args
		wantErr  bool
		wantThis Node
	}{
		{"A", treeA.Copy(), args{".subnode.e"}, false, treeB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Remove(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Node.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		if err := tt.n.Check(); err != nil {
			t.Errorf("Illegal element in tree: %v", err)
		}
		if !reflect.DeepEqual(tt.n, tt.wantThis) {
			t.Errorf("this = %v, want %v", tt.n, tt.wantThis)
		}
	}
}
