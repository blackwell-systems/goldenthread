// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package parser

import (
	"testing"
)

func TestParseTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tagToken
	}{
		{
			name:  "simple flags",
			input: "required,optional",
			expected: []tagToken{
				{value: "required", isKeyValue: false},
				{value: "optional", isKeyValue: false},
			},
		},
		{
			name:  "key-value without commas",
			input: "required,min:5,max:10",
			expected: []tagToken{
				{value: "required", isKeyValue: false},
				{value: "min:5", isKeyValue: true, key: "min", valueAfterColon: "5"},
				{value: "max:10", isKeyValue: true, key: "max", valueAfterColon: "10"},
			},
		},
		{
			name:  "enum with commas in value",
			input: "required,enum:active,inactive",
			expected: []tagToken{
				{value: "required", isKeyValue: false},
				{value: "enum:active,inactive", isKeyValue: true, key: "enum", valueAfterColon: "active,inactive"},
			},
		},
		{
			name:  "enum with multiple values",
			input: "required,enum:pending,in_progress,completed,cancelled",
			expected: []tagToken{
				{value: "required", isKeyValue: false},
				{value: "enum:pending,in_progress,completed,cancelled", isKeyValue: true, key: "enum", valueAfterColon: "pending,in_progress,completed,cancelled"},
			},
		},
		{
			name:  "enum followed by flag",
			input: "enum:a,b,c,optional",
			expected: []tagToken{
				{value: "enum:a,b,c", isKeyValue: true, key: "enum", valueAfterColon: "a,b,c"},
				{value: "optional", isKeyValue: false},
			},
		},
		{
			name:  "enum followed by key-value",
			input: "enum:x,y,min:5",
			expected: []tagToken{
				{value: "enum:x,y", isKeyValue: true, key: "enum", valueAfterColon: "x,y"},
				{value: "min:5", isKeyValue: true, key: "min", valueAfterColon: "5"},
			},
		},
	}

	p := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := p.parseTokens(tt.input)

			if len(tokens) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d", len(tt.expected), len(tokens))
				for i, tok := range tokens {
					t.Logf("  tokens[%d]: %q (isKV=%v, key=%q, val=%q)", i, tok.value, tok.isKeyValue, tok.key, tok.valueAfterColon)
				}
				return
			}

			for i, expected := range tt.expected {
				got := tokens[i]
				if got.value != expected.value || got.isKeyValue != expected.isKeyValue {
					t.Errorf("token %d: expected %+v, got %+v", i, expected, got)
				}
				if got.isKeyValue {
					if got.key != expected.key || got.valueAfterColon != expected.valueAfterColon {
						t.Errorf("token %d: expected key=%q val=%q, got key=%q val=%q",
							i, expected.key, expected.valueAfterColon, got.key, got.valueAfterColon)
					}
				}
			}
		})
	}
}
