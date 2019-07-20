// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

// Node contains children that are either nodes, slices, or values of the following types:
// - bool
// - string
// - number
//
// Valid child names must not contain periods (PathSeparator).
type Node map[string]interface{}

// CreatePath makes sure that a given path exists by creating nodes and overwriting existing values.
//
// The function will return the node the path points to.
func (n Node) CreatePath(path string) Node {
	elements := PathSplit(path)

	node := n
	for _, e := range elements {
		child, ok := node[e]
		if !ok {
			// Create child if it doesn't exist
			tempNode := Node{}
			node[e] = tempNode
			node = tempNode
		} else {
			tempNode, ok := child.(Node)
			if !ok {
				// Child is not a node, so overwrite it
				tempNode = Node{}
				node[e] = tempNode
			}
			node = tempNode
		}
	}

	return node
}

// Set creates all needed nodes and sets the element at the given path.
func (n Node) Set(path string, obj interface{}) error {
	var newElement interface{}

	newElement, err := anyToTree(reflect.ValueOf(obj))
	if err != nil {
		return err
	}

	pathElements := PathSplit(path)
	lastElement := pathElements[len(pathElements)-1]
	node := n
	if len(pathElements) > 1 {
		node = n.CreatePath(PathJoin(pathElements[:len(pathElements)-1]...))
	}
	node[lastElement] = newElement

	return nil
}

// Get reads the element at the path, and writes it into the given object obj.
func (n Node) Get(path string, obj interface{}) error {
	elements := PathSplit(path)

	inter := interface{}(n)
	for _, e := range elements {
		var ok bool
		node, ok := inter.(Node)
		if !ok {
			return &ErrPathInsideValue{path} // Path points inside a value
		}
		inter, ok = node[e]
		if !ok {
			return &ErrElementNotFound{path} // Element at path doesn't exist
		}
	}

	return treeToAny(inter, reflect.ValueOf(obj))
}

// GetBool returns the bool at the given path.
// In case of an error, the fallback is returned.
func (n Node) GetBool(path string, fallback bool) (result bool) {
	if err := n.Get(path, &result); err != nil {
		result = fallback
	}
	return
}

// GetString returns the string at the given path.
// In case of an error, the fallback is returned.
func (n Node) GetString(path string, fallback string) (result string) {
	if err := n.Get(path, &result); err != nil {
		result = fallback
	}
	return
}

// GetInt64 returns the integer at the given path.
// In case of an error, the fallback is returned.
func (n Node) GetInt64(path string, fallback int64) (result int64) {
	if err := n.Get(path, &result); err != nil {
		result = fallback
	}
	return
}

// GetFloat64 returns the float at the given path.
// In case of an error, the fallback is returned.
func (n Node) GetFloat64(path string, fallback float64) (result float64) {
	if err := n.Get(path, &result); err != nil {
		result = fallback
	}
	return
}

// Compare compares the current tree with the one in new and returns a list of paths for elements that were modified, added or removed.
//
// A change of the content/sub-content of a slice is returned as change of the slice itself.
func (n Node) Compare(new Node) (modified, added, removed []string) {
	return n.compare(new, "")
}

func (n Node) compare(new Node, prefix string) (modified, added, removed []string) {
	// Look for modified or removed elements
	for k, v := range n {
		vNew, foundNew := new[k]

		if foundNew {
			nodeA, aIsNode := v.(Node)
			nodeB, bIsNode := vNew.(Node)
			if aIsNode && bIsNode {
				// If both elements are nodes, check recursively.
				mod, add, rem := nodeA.compare(nodeB, prefix+k+PathSeparator) // Prefix is not really a path, as it can have a path separator at the end
				modified, added, removed = append(modified, mod...), append(added, add...), append(removed, rem...)
			} else if aIsNode {
				// If only a is a node, it got overwritten by a value
				modified = append(modified, prefix+k)
				_, _, rem := nodeA.compare(Node{}, prefix+k+PathSeparator)
				removed = append(removed, rem...)
			} else if bIsNode {
				// If only b is a node, it replaced a value
				modified = append(modified, prefix+k)
				_, add, _ := Node{}.compare(nodeB, prefix+k+PathSeparator)
				added = append(added, add...)
			} else if !cmp.Equal(v, vNew) {
				// If the two values are not equal
				modified = append(modified, prefix+k)
			}
			continue
		}

		// Not found, add to removed list
		removed = append(removed, prefix+k)
		if nodeA, ok := v.(Node); ok {
			_, _, rem := nodeA.compare(Node{}, prefix+k+PathSeparator)
			removed = append(removed, rem...)
		}
	}

	// Look for added elements
	for k, vNew := range new {
		_, found := n[k]

		if !found {
			added = append(added, prefix+k)
			if nodeB, ok := vNew.(Node); ok {
				_, add, _ := Node{}.compare(nodeB, prefix+k+PathSeparator)
				added = append(added, add...)
			}
		}
	}

	return
}

// Merge merges this tree with the new one.
//
// The following rules apply:
// - If both elements are nodes, their children are merged
// - Otherwise, the element of the new tree is written
// - If there is some element in the old, but not in the new tree, the old one is kept
// - If there is some element in the new, but not in the old tree, the new one is written
//
// Slices will not be merged, but new ones will overwrite old ones.
func (n Node) Merge(new Node) {
	for k, vNew := range new {
		v, found := n[k]

		if found {
			nodeA, aIsNode := v.(Node)
			nodeB, bIsNode := vNew.(Node)
			if aIsNode && bIsNode {
				// If both elements are nodes, merge recursively.
				nodeA.Merge(nodeB)
			} else {
				// If only one or none of the elements is a node, replace the old with the new one
				n[k] = vNew
			}
			continue
		}

		// Element not found in old tree
		n[k] = vNew
	}
}

// Check returns an error when a tree contains any malformed or illegal elements.
//
// Paths returned in errors are not valid paths, as they start with root and can contain numbers for slice elements.
func (n Node) Check() error {
	var recursive func(v interface{}, path string) error
	recursive = func(v interface{}, path string) error {
		switch v := v.(type) {
		case Node:
			for k, child := range v {
				err := recursive(child, PathJoin(path, k))
				if err != nil {
					return err
				}
			}
		case bool, string, Number:
		default:
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				s := reflect.ValueOf(v)
				for i := 0; i < s.Len(); i++ {
					child := s.Index(i).Interface()
					err := recursive(child, PathJoin(path, fmt.Sprint(i))) // Pseudo path for slice elements, not really a valid path
					if err != nil {
						return err
					}
				}
			} else {
				return &ErrUnexpectedType{path, fmt.Sprintf("%T", v), ""}
			}
		}

		return nil
	}

	return recursive(n, "root")
}
