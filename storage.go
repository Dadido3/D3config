// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package configdb

import "github.com/Dadido3/configdb/tree"

// Storage interface provides arbitrary ways to store/read hierarchical data.
type Storage interface {
	Read() (tree.Node, error)
	Write(t tree.Node) error
	RegisterWatcher(changeChan chan<- struct{}) error
}
