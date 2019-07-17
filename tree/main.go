// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"strings"
)

// Node contains a list of other nodes OR some value of arbitrary type
//
// If a node has no children, it represents its value, even if it is nil
type Node struct {
	Name string

	Value    interface{}
	Children []*Node
}

// Compare compares the current tree with the one in new and returns a list of paths for elements that were modified, added or removed
func (n *Node) Compare(new *Node) (modified, added, removed []Path, err error) {
	return
}

// SplitBySeparators will split all nodes that contain separators in their name
func (n *Node) SplitBySeparators() (new *Node) {
	names := strings.Split(n.Name, Separator)

	name, names := names[len(names)-1], names[:len(names)-1]
	new = &Node{
		Name:     name,
		Value:    n.Value, // TODO: Copy value
		Children: []*Node{},
	}

	for _, child := range n.Children {
		new.Children = append(new.Children, child.SplitBySeparators())
	}

	for len(names) > 0 {
		name, names = names[len(names)-1], names[:len(names)-1]
		new = &Node{
			Name:     name,
			Value:    nil,
			Children: []*Node{new},
		}
	}

	return
}

// JoinBySeparators will join all nodes that contain only one parent node or a value
func (n *Node) JoinBySeparators() (new *Node) {
	if len(n.Children) == 1 {
		new = n.Children[0].JoinBySeparators()
		new.Name = strings.Join([]string{n.Name, new.Name}, Separator)
		return
	}

	new = &Node{
		Name:     n.Name,
		Value:    n.Value, // TODO: Copy value
		Children: []*Node{},
	}

	for _, child := range n.Children {
		new.Children = append(new.Children, child.JoinBySeparators())
	}

	return
}
