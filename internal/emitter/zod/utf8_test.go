// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package zod

import (
	"testing"
	"unicode/utf8"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)

func TestEmit_UTF8_EmptyJSONName(t *testing.T) {
	s := &schema.Schema{
		Name:        "日本語",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "フィールド",
				JSONName: "",
				Type:     schema.Type{Kind: schema.TypeString},
				Optional: false,
			},
		},
	}

	emitter := NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	if !utf8.ValidString(output) {
		t.Errorf("Output contains invalid UTF-8")
		t.Logf("Output: %q", output)
		// Find the invalid byte
		for i, r := range output {
			if r == utf8.RuneError {
				t.Logf("Invalid UTF-8 at position %d", i)
			}
		}
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple ASCII",
			input:    "UserName",
			expected: "userName",
		},
		{
			name:     "single character",
			input:    "X",
			expected: "x",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already camelCase",
			input:    "userName",
			expected: "userName",
		},
		{
			name:     "Japanese characters",
			input:    "フィールド",
			expected: "フィールド",
		},
		{
			name:     "Mixed ASCII and Japanese",
			input:    "User日本語",
			expected: "user日本語",
		},
		{
			name:     "Emoji",
			input:    "🎉Party",
			expected: "🎉Party",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := camelCase(tt.input)
			if result != tt.expected {
				t.Errorf("camelCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
