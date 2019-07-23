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

type JSONFile struct {
	path string

	watcher *fsnotify.Watcher
}

func UseJSONFile(path string) *JSONFile {
	f := &JSONFile{
		path:    path,
		watcher: nil,
	}

	return f
}

func (f *JSONFile) read() (tree.Node, error) {
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

func (f *JSONFile) write(t tree.Node) error {
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

func (f *JSONFile) registerWatcher(change chan<- struct{}) error {
	// Close previous element, if there is one
	if f.watcher != nil {
		err := f.watcher.Close()
		if err != nil {
			return err
		}
		f.watcher = nil
	}

	// If there is no channel, just do nothing
	if change == nil {
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
				change <- struct{}{}
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
