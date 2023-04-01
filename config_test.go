// Copyright (c) 2019-2023 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package config

import (
	"path/filepath"
	"testing"
)

func TestSimple1(t *testing.T) {
	testStruct := struct {
		F float64 `conf:"someFloat"`
	}{}

	c, err := New([]Storage{UseJSONFile(filepath.Join(".", "testfiles", "json", "b.json"))})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer c.Close()

	if err := c.Get(".subnode.b", &testStruct); err != nil {
		t.Errorf("Get() failed: %v", err)
	}

	if err := c.Set(".subnode.e.sub", []string{"foo, bar"}); err != nil {
		t.Errorf("Set() failed: %v", err)
	}
}
