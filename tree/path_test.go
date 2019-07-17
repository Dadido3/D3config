// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"reflect"
	"testing"
)

func TestPathSplit(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"A", args{"foo"}, []string{"foo"}},
		{"B", args{"foo.bar"}, []string{"foo", "bar"}},
		{"C", args{"foo."}, []string{"foo", ""}},
		{"D", args{".bar"}, []string{"", "bar"}},
		{"E", args{"."}, []string{"", ""}},
		{"F", args{""}, []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathSplit(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PathSplit() = %v, want %v", got, tt.want)
			}
		})
	}
}
