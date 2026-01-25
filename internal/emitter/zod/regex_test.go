// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package zod

import (
	"strings"
	"testing"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)

func TestEmit_PatternEscaping(t *testing.T) {
	tests := []struct {
		name            string
		pattern         string
		expectedInRegex string
	}{
		{
			name:            "newline",
			pattern:         "\n",
			expectedInRegex: `regex(/\n/)`,
		},
		{
			name:            "tab",
			pattern:         "\t",
			expectedInRegex: `regex(/\t/)`,
		},
		{
			name:            "carriage return",
			pattern:         "\r",
			expectedInRegex: `regex(/\r/)`,
		},
		{
			name:            "forward slash",
			pattern:         "/test/",
			expectedInRegex: `regex(/\/test\//)`,
		},
		{
			name:            "backslash",
			pattern:         `\d+`,
			expectedInRegex: `regex(/\\d+/)`,
		},
		{
			name:            "mixed special chars",
			pattern:         "/path\n\t",
			expectedInRegex: `regex(/\/path\n\t/)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Name:        "Test",
				PackageName: "test",
				Fields: []schema.Field{
					{
						GoName:   "Value",
						JSONName: "value",
						Type:     schema.Type{Kind: schema.TypeString},
						Rules:    schema.FieldRules{Pattern: &tt.pattern},
					},
				},
			}

			emitter := NewEmitter()
			output, err := emitter.Emit(s)
			if err != nil {
				t.Fatalf("Emit() error = %v", err)
			}

			if !strings.Contains(output, tt.expectedInRegex) {
				t.Errorf("Expected regex %q not found in output:\n%s", tt.expectedInRegex, output)
			}

			// Verify no unescaped newlines break the output
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "regex(") {
					// Regex line should not be broken
					if !strings.Contains(line, "/)") {
						t.Error("Regex pattern appears to be broken across lines")
					}
				}
			}
		})
	}
}
