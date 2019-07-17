// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"regexp"
	"strings"
)

// Separator delimits single path elements
const Separator = "."

var regexRemove = regexp.MustCompile(`^\.+|\.+$|\r|\n`)
var regexReplacePeriods = regexp.MustCompile(`\.{2,}`)

// Path is used to locate any element in a tree
//
// A period is used to delimit single elements.
type Path string

// NewPath creates a new path from strings or other Path objects
func NewPath(elem ...Path) Path {
	str := make([]string, len(elem))
	for i, v := range elem {
		str[i] = string(v) // TODO: Convert elem to interface and add typeswitch with casting
	}

	return Path(strings.Join(str, Separator)).Clean()
}

// Clean removes any empty or invalid elements from a path
func (p Path) Clean() Path {
	str := string(p)
	str = regexRemove.ReplaceAllString(str, "")
	str = regexReplacePeriods.ReplaceAllString(str, Separator)
	return Path(str)
}
