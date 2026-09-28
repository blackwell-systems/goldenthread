// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package main

import (
	"reflect"
	"testing"
)

func TestReorderFlagArgs(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"flags already first", []string{"--infer-json", "./models"}, []string{"--infer-json", "./models"}},
		{"bool flag after path", []string{"./models", "--infer-json"}, []string{"--infer-json", "./models"}},
		{"value flag after path (space form)", []string{"./models", "--out", "./gen"}, []string{"--out", "./gen", "./models"}},
		{"value flag after path (equals form)", []string{"./models", "--out=./gen"}, []string{"--out=./gen", "./models"}},
		{"mixed flags around path", []string{"--recursive", "./models", "--out", "./gen"}, []string{"--recursive", "--out", "./gen", "./models"}},
		{"double dash terminator", []string{"./models", "--", "--not-a-flag"}, []string{"./models", "--", "--not-a-flag"}},
		{"no flags", []string{"./models"}, []string{"./models"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := reorderFlagArgs(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("reorderFlagArgs(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
