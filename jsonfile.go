// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package configdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Dadido3/configdb/tree"
	"github.com/fsnotify/fsnotify"
)

// JSONFile represents a json file on disk.
type JSONFile struct {
	path string

	watcher *fsnotify.Watcher
}

// UseJSONFile returns a JSONFile object.
func UseJSONFile(path string) Storage {
	f := &JSONFile{
		path:    path,
		watcher: nil,
	}

	return f
}

// Read returns the tree representation of its content.
func (f *JSONFile) Read() (tree.Node, error) {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		return tree.Node{}, nil // Not existent file behaves like an empty tree
	}

	buf, err := ioutil.ReadFile(f.path)
	if err != nil {
		return nil, fmt.Errorf("Reading file %v failed: %v", f.path, err)
	}

	node := tree.Node{}
	if err := json.Unmarshal(buf, &node); err != nil {
		return nil, err
	}

	return node, nil
}

// Write takes a tree and stores it in some shape and form.
func (f *JSONFile) Write(t tree.Node) error {
	buf, err := json.MarshalIndent(t, "", "    ")
	if err != nil {
		return err
	}

	tempPath := f.path + ".tmp"
	if err := ioutil.WriteFile(tempPath, buf, 0644); err != nil {
		return fmt.Errorf("Writing file %v failed: %v", tempPath, err)
	}

	if err := os.Rename(tempPath, f.path); err != nil {
		return fmt.Errorf("Renaming file %v to %v failed: %v", tempPath, f.path, err)
	}

	return nil
}

// RegisterWatcher takes a channel that is used to signal changes/modifications on the data.
// Only one channel can be registered at a time.
//
// A nil value can be passed to unregister the listener.
func (f *JSONFile) RegisterWatcher(changeChan chan<- struct{}) error {
	// Close previous element, if there is one
	if f.watcher != nil {
		err := f.watcher.Close()
		if err != nil {
			return err
		}
		f.watcher = nil
	}

	// If there is no channel, just do nothing
	if changeChan == nil {
		return nil
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func(w *fsnotify.Watcher) {
		for {
			select {
			case _, ok := <-w.Events:
				if !ok {
					return
				}
				// Write to changeChan in a non blocking way
				select {
				case changeChan <- struct{}{}:
				default:
				}
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
				// TODO: Handle error
			}
		}
	}(w)

	err = w.Add(f.path)
	if err != nil {
		w.Close()
		return err
	}

	f.watcher = w

	return nil
}
