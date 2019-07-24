<!--
 Copyright (c) 2019 David Vogel
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# ConfigDB [![Build Status](https://travis-ci.com/Dadido3/configdb.svg?branch=master)](https://travis-ci.com/Dadido3/configdb)

This is a small library to handle configuration files.
Any configuration data is stored in a tree, that serves as intermediate layer between the config files on disk, and data structures in your software.

This should make handling configuration data as simple as possible.
Ideally it is as simple to use as a database.
You can read/write any structure or data type, without caring when and how the data is stored somewhere.

## Features

- Marshal & unmarshal any structures or types.
- Can handle multiple configuration files. They are merged into one tree priorized by order. (e.g. user settings, default, ...)
- Has several storage types (JSON files, YAML files), and you can add your own storages.
- Changes are saved to disk automatically, and changes on disk are loaded automatically.
- Event system that signals changes to the tree.
- Safe against power loss while writing files to disk.
- Thread-safe and deadlock proof by design.

## Current state

The library is feature complete, but as it is really new and not much tested (Beside the unit tests) i can't guarantee that everything will work correctly.
If you encounter a bug, or some undocumented behavior, open an issue.

ToDo:

- [ ] `time.Time` and other objects that (un)marshal from/into text are not handled yet.
- [ ] Fix Dropbox to cause files to not save, as it shortly prevents write access.

## Usage

### Initialize

```go
// The upper storage objects have higher priority as the lower ones.
// So the properties/values of the upper will overwrite the ones in the lower entries.
// One special case is the storage object at index 0, this is the one that changes are written into.
storages := []configdb.Storage{
    configdb.UseJSONFile("testfiles/json/userconfig.json"),
    configdb.UseYAMLFile("testfiles/yaml/custom.yml"),
    configdb.UseJSONFile("testfiles/json/default.json"),
}

c, err := configdb.New(storages)
if err != nil {
    fmt.Fatal(err)
}
defer c.Close()
```

### Read value

```go
var f float32

// Pass a pointer to any object you want to read from the internal tree at the given path ".box.width"
err := c.Get(".box.width", &f)
if err != nil {
    t.Error(err)
}
```

This will write `123.456` into `f`, with json data like:

```json
{
    "box": {
        "width": 123.456,
        "height": 654.321
    }
}
```

### Read structure

```go
// You can use tags to change the names, or exclude fields with "omit"
var str struct {
    Width     float64 `cdb:"width"`
    Height    float64 `cdb:"height"`
    PlsIgnore string  `cdb:",omit"`
}

// Pass a pointer to any object you want to read from the internal tree at the given path ".box"
err := c.Get(".box", &str)
if err != nil {
    t.Error(err)
}
fmt.Printf("%+v", str)
```

With the same json data as above, this will result in:

```go
{Width:123.456 Height:654.321 PlsIgnore:}
```

### Read slices, maps and more

```go
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
```

Will result in:

```go
[]string{"Sam Sung", "Saad Maan", "Chris P. Bacon"}
map[string]interface {}{"height":"654.321", "names":[]interface {}{"Sam Sung", "Saad Maan", "Chris P. Bacon"}, "width":"123.456"}
```

### Write value

```go
var b bool

// Pass an object to be written at the path ".todo.WriteCode"
err := c.Set(".todo.WriteCode", b)
if err != nil {
    t.Error(err)
}
```

This will write the changes to disk immediately, but the internal tree may be updated later.
Therefore a `Get()` directly following a `Set()` may still result in old data.
In these cases it's better to rely on the event mechanism, which is explained a few steps below.

If config was created with `testfiles/json/userconfig.json` being the first file, the following content will be added to it:

```json
{
    "todo": {
        "WriteCode": false
    }
}
```

### Write structure

```go
str := struct {
    Eat, Sleep bool
}{true, false}

// Pass an object to be written at the path ".todo"
err := c.Set(".todo", str)
if err != nil {
    t.Error(err)
}
```

Which will result in the file `testfiles/json/userconfig.json` to look like: (Assuming `"WriteCode": true` was written before)

```json
{
    "todo": {
        "Eat": true,
        "Sleep": false,
        "WriteCode": true
    }
}
```

### Write nil

```go
// You can also overwrite anything with nil
err := c.Set(".todo", nil)
if err != nil {
    t.Error(err)
}
```

Which will result in:

```json
{
    "todo": null
}
```

This can be used to overwrite and disable any defaults from other storage objects.

### Reset element

```go
// Resets the element at the path ".todo".
// This will restore any defaults, if there are any present in lower priority storage objects.
err := c.Reset(".todo")
if err != nil {
    t.Error(err)
}

// This will reset everything to default.
// It has the same effect as deleting the highest priority file.
err = c.Reset("")
if err != nil {
    t.Error(err)
}
```

### Register and unregister event callback

```go
// Register callback to listen for events.
// Once registered, the callback is called once to update the listener with the current state of the tree.
id := c.RegisterCallback(nil, func(c *configdb.Config, modified, added, removed []string) {
    fmt.Printf("All m: %v, a: %v, r:%v\n", modified, added, removed)
})
// Use the result id to unregister later
defer c.UnregisterCallback(id)

// Register callback to listen for events, but only for path ".something.to.watch"
id = c.RegisterCallback([]string{".something.to.watch"}, func(c *configdb.Config, modified, added, removed []string) {
    fmt.Printf("Filtered m: %v, a: %v, r:%v\n", modified, added, removed)
})
// Use the result id to unregister later
defer c.UnregisterCallback(id)

// Test the callback
err := c.Set(".something.to.watch.for", 123)
if err != nil {
    t.Error(err)
}

// The event may not be sent immediately, wait a bit before terminating the program
time.Sleep(100 * time.Millisecond)
```

The output could look like this:

```
All m: [], a: [.box .box.width .box.height .box.names], r:[]
Filtered m: [], a: [.something.to.watch .something.to.watch.for], r:[]
All m: [], a: [.something .something.to .something.to.watch .something.to.watch.for], r:[]
```

The parameters are lists of paths (strings) that were either modified, added or deleted from the tree.
But in most cases these lists can be ignored.
It's enough to set the path (filter) to the object you want to watch for.

The first callback is just there to update the newly created listener with the current state of the tree.
If there is data at the set path, there will always be one call right after registering the callback.

With a filter, the callback is only called when the watched data changes.
Therefore the callback is suitable for triggering a reconnect or similar actions, as it is only called when necessary.

Additionally it is made sure that the tree is in sync with the changes. It's safe to use `c.Get()` or even `c.Set()`/`c.Reset()` inside the callback.

### Custom storage objects

```go
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
    storages := []configdb.Storage{
        configdb.UseJSONFile("testfiles/json/userconfig.json"),
        &CustomStorage{},
        configdb.UseJSONFile("testfiles/json/default.json"),
    }

    c, err := configdb.New(storages)
    if err != nil {
        t.Fatal(err)
    }
    defer c.Close()
}
```

## FAQ

### What are valid element names?

Any character except period `.` is allowed. Also empty names are valid too.

### How to address elements of an array or slice with a path?

You can't. Paths can only address map elements or structure fields.
But you can use `Get()` to read any slice or array.

If you need to register a callback on something inside an array or slice, you have to point on the array/slice itself.
E.g. `.someField.slice` will also trigger an event when some element or value several levels deep inside of that slice is modified.

### Is it really not possible to address elements inside arrays or slices?

With a trick it is:

- Import `"github.com/Dadido3/configdb/tree"`
- Use the following snippet:

```go
var nodes []tree.Node

// Get a list of tree.Node objects
// That will copy a subtree into the variable nodes
err := c.Get(".slicedNodes", &nodes)
if err != nil {
    t.Fatal(err)
}

// Read value of that subtree
result := nodes[0].GetInt64(".something", 0)
fmt.Println(result)
```

Any edits you do on `nodes` have no effect on the main tree.
You need to use `c.Set(path, nodes)` to write it back.

This way you can also create copies to work with while the configuration is being modified.
