// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"strings"
)

// PathSeparator delimits single path elements.
const PathSeparator = "."

// PathJoin creates a new path from several path strings.
func PathJoin(elem ...string) string {
	return strings.Join(elem, PathSeparator)
}

// PathSplit splits a path into its elements.
func PathSplit(path string) []string {
	return strings.Split(path, PathSeparator)
}

// PathContains returns whether the path contains the subPath.
func PathContains(path, subPath string) bool {
	p, sp := PathSplit(path), PathSplit(subPath)

	if len(sp) > len(p) {
		return false
	}

	for i, e := range sp {
		if e != p[i] {
			return false
		}
	}

	return true
}
