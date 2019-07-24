// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package configdb

import "github.com/Dadido3/configdb/tree"

// File interface.
type File interface {
	read() (tree.Node, error)
	write(t tree.Node) error
	registerWatcher(changeChan chan<- struct{}) error
}
