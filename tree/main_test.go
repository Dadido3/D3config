// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"reflect"
	"testing"
)

var testNode = &Node{
	Name: "foo.bar",
	Children: []*Node{
		{
			Name: "baz",
			Children: []*Node{
				{
					Name:     "bay.bax",
					Children: []*Node{},
				},
			},
		},
		{
			Name: "bam",
			Children: []*Node{
				{
					Name:     "bat",
					Children: []*Node{},
				},
			},
		},
	},
}

var testNodeSplittedBySeparators = &Node{
	Name: "foo",
	Children: []*Node{
		{
			Name: "bar",
			Children: []*Node{
				{
					Name: "baz",
					Children: []*Node{
						{
							Name: "bay",
							Children: []*Node{
								{
									Name:     "bax",
									Children: []*Node{},
								},
							},
						},
					},
				},
				{
					Name: "bam",
					Children: []*Node{
						{
							Name:     "bat",
							Children: []*Node{},
						},
					},
				},
			},
		},
	},
}

var testNodeJoinedBySeparators = &Node{
	Name: "foo.bar",
	Children: []*Node{
		{
			Name:     "baz.bay.bax",
			Children: []*Node{},
		},
		{
			Name:     "bam.bat",
			Children: []*Node{},
		},
	},
}

func TestNode_SplitBySeparators(t *testing.T) {
	tests := []struct {
		name    string
		n       *Node
		wantNew *Node
	}{
		{"1", testNode, testNodeSplittedBySeparators},
		{"2", testNodeSplittedBySeparators, testNodeSplittedBySeparators},
		{"3", testNodeJoinedBySeparators, testNodeSplittedBySeparators},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNew := tt.n.SplitBySeparators(); !reflect.DeepEqual(gotNew, tt.wantNew) {
				t.Errorf("Node.SplitBySeparators() = %v, want %v", gotNew, tt.wantNew)
			}
		})
	}
}

func TestNode_JoinBySeparators(t *testing.T) {
	tests := []struct {
		name    string
		n       *Node
		wantNew *Node
	}{
		{"1", testNode, testNodeJoinedBySeparators},
		{"2", testNodeSplittedBySeparators, testNodeJoinedBySeparators},
		{"3", testNodeJoinedBySeparators, testNodeJoinedBySeparators},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNew := tt.n.JoinBySeparators(); !reflect.DeepEqual(gotNew, tt.wantNew) {
				t.Errorf("Node.JoinBySeparators() = %v, want %v", gotNew, tt.wantNew)
			}
		})
	}
}
