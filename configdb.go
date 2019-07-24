// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// TODO: Cache trees of files, so it operations don't need to go through the file system

package configdb

import (
	"fmt"
	"log"
	"sync"

	"github.com/Dadido3/configdb/tree"
)

type eventReset struct {
	path       string
	resultChan chan<- error
}

type eventSet struct {
	path       string
	object     interface{}
	resultChan chan<- error
}

type eventRegister struct {
	paths      []string
	callback   func(c *Config, modified, added, removed []string)
	resultChan chan<- int
}

type eventUnregister struct {
	id         int
	resultChan chan<- struct{}
}

// Config stores any data structure as tree internally.
// Create this by using NewConfig().
//
// Use the Set(), Reset() and Get() methods to interact with that tree.
// Changes made to the config are immediately stored in the configuration file.
type Config struct {
	eventChan    chan interface{}
	listenerChan chan interface{}

	tree      tree.Node // Tree is only modified by the "Tree update handler" goroutine, to prevent deadlocks and out of sync data
	treeMutex sync.RWMutex

	waitGroup sync.WaitGroup
}

type listener struct {
	paths    []string
	callback func(c *Config, modified, added, removed []string)
}

// NewConfig returns a new Config object.
//
// It takes a list of File objects that can be created with UseJSONFile(path) and similar functions.
// All files in the list are merged into one big configuration tree.
// The priority of the files is descending, so elements/variables from the first files will overwrite the ones in files with lower priority.
//
// If a file is changed on disk, it is reloaded.
// Changes in the configuration tree will be broadcasted to any listener.
//
// If any of these files couldn't be read, this function will return an error.
func NewConfig(files []File) (*Config, error) {
	c := &Config{
		eventChan:    make(chan interface{}),
		listenerChan: make(chan interface{}),
	}

	readConfig := func(files []File) (tree.Node, error) {
		result := tree.Node{}

		for i := len(files) - 1; i >= 0; i-- {
			file := files[i]
			t, err := file.read()
			if err != nil {
				return nil, err
			}
			result.Merge(t)
		}

		return result, nil
	}

	setObject := func(files []File, path string, obj interface{}) error {
		if len(files) <= 0 {
			return fmt.Errorf("There are no files to write to")
		}
		file := files[0]

		t, err := file.read()
		if err != nil {
			return err
		}

		if err := t.Set(path, obj); err != nil {
			return err
		}

		if err := file.write(t); err != nil {
			return err
		}

		return nil
	}

	resetObject := func(files []File, path string) error {
		if len(files) <= 0 {
			return fmt.Errorf("There are no files to write to")
		}
		file := files[0]

		t, err := file.read()
		if err != nil {
			return err
		}

		if err := t.Remove(path); err != nil {
			return err
		}

		if err := file.write(t); err != nil {
			return err
		}

		return nil
	}

	// Try to read files and build config tree
	if tree, err := readConfig(files); err == nil {
		c.tree = tree // No need to lock mutex here, as nothing else can access the tree
	} else {
		return nil, err
	}

	treeChan := make(chan tree.Node, 1) // New (already merged) trees are put here to be compared and distributed to listeners

	// Event handler goroutine
	c.waitGroup.Add(1)
	go func() {
		defer c.waitGroup.Done()
		defer close(treeChan)

		changeChan := make(chan struct{}, 1) // Channel for file changes that trigger a reload of the config tree
		defer close(changeChan)

		for _, file := range files {
			file.registerWatcher(changeChan) // TODO: Handle error
			defer file.registerWatcher(nil)
		}

		for {
			select {
			case <-changeChan:
				tree, err := readConfig(files)
				if err != nil {
					// TODO: Handle error
					log.Printf("ConfigDB: %v", err)
					continue
				}
				// Write tree into tree channel, or replace the queued element if the goroutine is busy. This is non blocking
				select {
				case treeChan <- tree:
				default:
					select {
					case <-treeChan:
					default:
					}
					treeChan <- tree
				}

			case u, ok := <-c.eventChan:
				if !ok {
					return
				}
				switch u := u.(type) {
				case eventReset:
					err := resetObject(files, u.path)
					u.resultChan <- err
					// Write to changeChan in a non blocking way
					/*select {
					case changeChan <- struct{}{}:
					default:
					}*/

				case eventSet:
					err := setObject(files, u.path, u.object)
					u.resultChan <- err
					// Write to changeChan in a non blocking way
					/*select {
					case changeChan <- struct{}{}:
					default:
					}*/

				default:
					log.Panicf("Got invalid element %v of type %T in event channel", u, u)
				}
			}
		}
	}()

	sendChanges := func(l listener, modified, added, removed []string) {
		for _, lPath := range l.paths {
			tempModified, tempAdded, tempRemoved := []string{}, []string{}, []string{}
			for _, path := range modified {
				if tree.PathContains(path, lPath) {
					tempModified = append(tempModified, path)
				}
			}
			for _, path := range added {
				if tree.PathContains(path, lPath) {
					tempAdded = append(tempAdded, path)
				}
			}
			for _, path := range removed {
				if tree.PathContains(path, lPath) {
					tempRemoved = append(tempRemoved, path)
				}
			}
			if len(tempModified) > 0 || len(tempAdded) > 0 || len(tempRemoved) > 0 {
				l.callback(c, tempModified, tempAdded, tempRemoved)
			}
		}
	}

	// Tree update handler goroutine (Also distributes tree events to listeners)
	c.waitGroup.Add(1)
	go func() {
		defer c.waitGroup.Done()

		listeners := make(map[int]listener) // List of registered listeners
		listenersCounter := 0

		for {
			select {
			case t, ok := <-treeChan:
				if !ok {
					return
				}

				modified, added, removed := c.tree.Compare(t) // No mutex needed, as the tree is only modified in this goroutine
				c.treeMutex.Lock()
				c.tree = t
				c.treeMutex.Unlock()

				wg := sync.WaitGroup{}
				for _, l := range listeners {
					wg.Add(1)
					go func(l listener) {
						defer wg.Done()
						sendChanges(l, modified, added, removed)
					}(l)
				}
				wg.Wait()

			case e := <-c.listenerChan:
				switch e := e.(type) {
				case eventRegister:
					l := listener{e.paths, e.callback}
					if len(l.paths) == 0 {
						l.paths = []string{""} // Add at least one empty path that fits all, if there are not paths defined
					}
					listeners[listenersCounter] = l
					e.resultChan <- listenersCounter
					listenersCounter++
					modified, added, removed := tree.Node{}.Compare(c.tree) // Compare empty tree with current one
					sendChanges(l, modified, added, removed)                // No mutex needed, as the tree is only modified in this goroutine

				case eventUnregister:
					delete(listeners, e.id)
					e.resultChan <- struct{}{}

				default:
					log.Panicf("Got invalid element %v of type %T in listener channel", e, e)

				}
			}
		}
	}()

	return c, nil
}

// Register will add the given callback to the internal listener list.
// A list of paths can be defined to ignore all events that are not inside the given paths.
//
// An integer is returned, that can be used to Unregister() the callback.
func (c *Config) Register(paths []string, callback func(c *Config, modified, added, removed []string)) int {
	resultChan := make(chan int)
	c.listenerChan <- eventRegister{paths, callback, resultChan}
	return <-resultChan
}

// Unregister removes a callback from the internal listener list.
func (c *Config) Unregister(id int) {
	resultChan := make(chan struct{})
	c.listenerChan <- eventUnregister{id, resultChan}
	<-resultChan
}

// Set changes the element at the given path.
//
// It's possible to modify the root node, with the path "", if the passed object is a map or a structure.
//
// Changes are written immediately to the configuration files.
func (c *Config) Set(path string, object interface{}) error {
	resultChan := make(chan error)
	c.eventChan <- eventSet{path, object, resultChan}
	return <-resultChan
}

// Reset will remove the element at the given path.
// If there are lower priority configuration files
func (c *Config) Reset(path string) error {
	resultChan := make(chan error)
	c.eventChan <- eventReset{path, resultChan}
	return <-resultChan
}

// Get will marshal the elements at path into the given object.
func (c *Config) Get(path string, object interface{}) error {
	c.treeMutex.RLock()
	defer c.treeMutex.RUnlock()

	return c.tree.Get(path, object)
}

// Close will free all ressources/watchers.
func (c *Config) Close() {
	close(c.eventChan)
	c.waitGroup.Wait()
}
