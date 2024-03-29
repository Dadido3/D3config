// Copyright (c) 2019-2023 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config_test

import (
	"fmt"
	"testing"
	"time"

	config "github.com/Dadido3/D3config"
	"github.com/Dadido3/D3config/tree"
)

func TestInit(t *testing.T) {
	// The upper storage objects have higher priority as the lower ones.
	// So the properties/values of the upper will overwrite the ones in the lower entries.
	// One special case is the storage object at index 0, this is the one that changes are written into.
	storages := []config.Storage{
		config.UseJSONFile("testfiles/json/userconfig.json"),
		config.UseYAMLFile("testfiles/yaml/custom.yml"),
		config.UseJSONFile("testfiles/json/default.json"),
	}

	c, err := config.New(storages)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}

func Create(t *testing.T) *config.Config {
	storages := []config.Storage{
		config.UseJSONFile("testfiles/json/userconfig.json"),
		config.UseYAMLFile("testfiles/yaml/custom.yml"),
		config.UseJSONFile("testfiles/json/default.json"),
	}

	c, err := config.New(storages)
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func TestReadValue(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	var f float32

	// Pass a pointer to any object you want to read from the internal tree at the given path ".box.width".
	err := c.Get(".box.width", &f)
	if err != nil {
		t.Error(err)
	}
}

func TestReadStructure(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	// You can use tags to change the names, or exclude fields with "omit".
	var str struct {
		Width     float64 `conf:"width"`
		Height    float64 `conf:"height"`
		PlsIgnore string  `conf:",omit"`
	}

	// Pass a pointer to any object you want to read from the internal tree at the given path ".box".
	err := c.Get(".box", &str)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v", str)
}

func TestReadAny(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	// It also works with slices/arrays.
	// They can be any type, even arrays of arrays.
	var s []string

	err := c.Get(".box.names", &s)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%#v\n", s)

	// Maps have the limitation that the key has to be a string.
	// But the value type can be anything.
	var m map[string]interface{}

	err = c.Get(".box", &m)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%#v\n", m)

	// The lib supports all objects that support text (un)marshaller interface.
	var ti time.Time

	err = c.Get(".back.toTheFuture", &ti)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", ti)
}

func TestWriteValue(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	b := true

	// Pass an boolean to be written at the path ".todo.WriteCode".
	err := c.Set(".todo.WriteCode", b)
	if err != nil {
		t.Error(err)
	}

	ti := time.Date(2019, 7, 24, 14, 46, 24, 124, time.UTC)

	// Pass time object to be written at the path ".time.WriteCodeAt".
	err = c.Set(".time.WriteCodeAt", ti)
	if err != nil {
		t.Error(err)
	}
}

func TestWriteStruct(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	str := struct {
		Eat, Sleep bool
	}{true, false}

	// Pass an object to be written at the path ".todo".
	err := c.Set(".todo", str)
	if err != nil {
		t.Error(err)
	}
}

func TestWriteNil(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	// You can also overwrite anything with nil.
	err := c.Set(".todo", nil)
	if err != nil {
		t.Error(err)
	}
}

func TestReset(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	// Resets the element at the path ".todo".
	// This will restore any defaults, if there are any present in lower priority storage objects.
	err := c.Reset(".todo")
	if err != nil {
		t.Error(err)
	}

	// This will reset everything to default.
	err = c.Reset("")
	if err != nil {
		t.Error(err)
	}
}

func TestRegister(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	// Register callback to listen for events.
	// Once registered, the callback is called once to update the listener with the current state of the tree.
	id := c.RegisterCallback(nil, func(c *config.Config, modified, added, removed []string) {
		fmt.Printf("All m: %v, a: %v, r:%v\n", modified, added, removed)
	})
	// Use the result id to unregister later.
	defer c.UnregisterCallback(id)

	// Register callback to listen for events, but only inside the path ".something.to.watch".
	// This includes modifications to ".something.to.watch" itself.
	id = c.RegisterCallback([]string{".something.to.watch"}, func(c *config.Config, modified, added, removed []string) {
		fmt.Printf("Filtered m: %v, a: %v, r:%v\n", modified, added, removed)
	})
	// Use the result id to unregister later.
	defer c.UnregisterCallback(id)

	// Test the callback.
	err := c.Set(".something.to.watch.for", 125)
	if err != nil {
		t.Error(err)
	}

	// The event may not be sent immediately, wait a bit before terminating the program.
	time.Sleep(100 * time.Millisecond)
}

func TestTreeNode(t *testing.T) {
	c := Create(t)
	defer c.Close()

	// ---------

	var nodes []tree.Node

	// Get a list of tree.Node objects.
	// That will copy a subtree into the variable nodes.
	err := c.Get(".slicedNodes", &nodes)
	if err != nil {
		t.Fatal(err)
	}

	// Read value of that subtree.
	result := nodes[0].GetInt64(".something", 0)
	fmt.Println(result)
}

// Implement Storage interface.
type CustomStorage struct {
}

func (f *CustomStorage) Read() (tree.Node, error) {
	return tree.Node{
		"SomethingPermanent": tree.Node{
			"foo": tree.Number("123"),
			"bar": tree.Number("-123.456"),
		},
	}, nil
}

func (f *CustomStorage) Write(t tree.Node) error {
	return fmt.Errorf("Can't write into this storage object")
}

func (f *CustomStorage) RegisterWatcher(changeChan chan<- struct{}) error {
	return nil
}

func TestCustomStorage(t *testing.T) {
	// Use the custom made storage object along with others.
	// Be aware, that if you have a non writable storage at the top, the tree can't be modified anymore.
	storages := []config.Storage{
		config.UseJSONFile("testfiles/json/userconfig.json"),
		&CustomStorage{},
		config.UseJSONFile("testfiles/json/default.json"),
	}

	c, err := config.New(storages)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
}
