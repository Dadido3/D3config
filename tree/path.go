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
