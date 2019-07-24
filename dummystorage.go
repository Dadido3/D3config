// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package configdb

import (
	"log"

	"github.com/Dadido3/configdb/tree"
)

// DummyStorage is a virtual storage container.
// The data is only kept in RAM, so anything stored in it will be lost once it is freed.
type DummyStorage struct {
	tree tree.Node
}

// UseDummyStorage returns a virtual storage container that only stores data in the RAM.
// Use it if you want the ability to overwrite settings, without writing those settings somewhere permanently.
//
// The storage can be initialized with a data structure that will be written at initPath.
func UseDummyStorage(initPath string, initData interface{}) Storage {
	f := &DummyStorage{
		tree: tree.Node{},
	}

	if initData != nil {
		if err := f.tree.Set(initPath, initData); err != nil {
			log.Panicf("Failed to initialize the dummy storage: %v", err)
		}
	}

	return f
}

// Read returns the tree representation of its content.
func (f *DummyStorage) Read() (tree.Node, error) {
	return f.tree, nil
}

// Write takes a tree and stores it in some shape and form.
func (f *DummyStorage) Write(t tree.Node) error {
	f.tree = t
	return nil
}

// RegisterWatcher takes a channel that is used to signal changes/modifications of the data.
// Only one channel can be registered at a time.
//
// A nil value can be passed to unregister the listener.
func (f *DummyStorage) RegisterWatcher(changeChan chan<- struct{}) error {
	return nil
}
