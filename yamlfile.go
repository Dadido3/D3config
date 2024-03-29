// Copyright (c) 2019-2023 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Dadido3/D3config/tree"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// YAMLFile represents a json file on disk.
type YAMLFile struct {
	path string

	watcher *fsnotify.Watcher
}

// UseYAMLFile returns a YAMLFile object.
func UseYAMLFile(path string) Storage {
	f := &YAMLFile{
		path:    path,
		watcher: nil,
	}

	return f
}

// Read returns the tree representation of its content.
func (f *YAMLFile) Read() (tree.Node, error) {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		return tree.Node{}, nil // Not existent file behaves like an empty tree.
	}

	buf, err := ioutil.ReadFile(f.path)
	if err != nil {
		return nil, fmt.Errorf("reading YAML file %v failed: %w", f.path, err)
	}

	node := tree.Node{}
	if err := yaml.Unmarshal(buf, &node); err != nil {
		return nil, fmt.Errorf("unmarshalling %v failed: %w", f.path, err)
	}

	return node, nil
}

// Write takes a tree and stores it in some shape and form.
func (f *YAMLFile) Write(t tree.Node) error {
	buf, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshalling into %v failed: %w", f.path, err)
	}

	tempPath := f.path + ".tmp"
	if err := ioutil.WriteFile(tempPath, buf, 0644); err != nil {
		return fmt.Errorf("writing file %v failed: %w", tempPath, err)
	}

	if err := os.Rename(tempPath, f.path); err != nil {
		return fmt.Errorf("renaming file %v to %v failed: %w", tempPath, f.path, err)
	}

	return nil
}

// RegisterWatcher takes a channel that is used to signal changes/modifications of the data.
// Only one channel can be registered at a time.
//
// A nil value can be passed to unregister the listener.
func (f *YAMLFile) RegisterWatcher(changeChan chan<- struct{}) error {
	// Close previous element, if there is one.
	if f.watcher != nil {
		err := f.watcher.Close()
		if err != nil {
			return err
		}
		f.watcher = nil
	}

	// If there is no channel, just do nothing.
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
				// Write to changeChan in a non blocking way.
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
