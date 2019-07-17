// Copyright (c) 2019 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tree

import (
	"reflect"
	"testing"
)

func TestNewPath(t *testing.T) {
	type args struct {
		elem []Path
	}
	tests := []struct {
		name string
		args args
		want Path
	}{
		{"1", args{elem: []Path{"foo", "bar"}}, "foo.bar"},
		{"2", args{elem: []Path{"foo", ".bar"}}, "foo.bar"},
		{"3", args{elem: []Path{"foo.", "bar"}}, "foo.bar"},
		{"4", args{elem: []Path{".foo.", ".bar."}}, "foo.bar"},
		{"5", args{elem: []Path{"foo.", "foo.baz", "bar"}}, "foo.foo.baz.bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPath(tt.args.elem...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Clean(t *testing.T) {
	tests := []struct {
		name string
		p    Path
		want Path
	}{
		{"1", "foo.bar", "foo.bar"},
		{"2", ".foo.bar.", "foo.bar"},
		{"3", ".foo.\n\r.bar.", "foo.bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Clean(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Path.Clean() = %v, want %v", got, tt.want)
			}
		})
	}
}
