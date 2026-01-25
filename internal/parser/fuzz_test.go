// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package parser_test

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/blackwell-systems/goldenthread/internal/load"
	"github.com/blackwell-systems/goldenthread/internal/parser"
)

// FuzzParsePackages fuzzes the main parser entry point with random Go struct definitions
func FuzzParsePackages(f *testing.F) {
	// Seed corpus with various struct tag formats
	f.Add("User", "Username", "username", `gt:"required"`)
	f.Add("Product", "Name", "name", `gt:"optional"`)
	f.Add("Task", "Status", "status", `gt:"enum:a,b,c"`)
	f.Add("Range", "Value", "value", `gt:"min:0,max:100"`)
	f.Add("Text", "Body", "body", `gt:"len:5..100"`)
	f.Add("Email", "Addr", "addr", `gt:"email"`)
	f.Add("S", "F", "f", ``)
	f.Add("Empty", "Field", "field", `gt:""`)
	f.Add("Bad", "Field", "field", `gt:"invalid,syntax,here"`)
	f.Add("Weird", "Field", "field", `gt:"min:,max:,len:,,"`)

	f.Fuzz(func(t *testing.T, structName, fieldName, jsonName, gtTag string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(structName) || !utf8.ValidString(fieldName) ||
			!utf8.ValidString(jsonName) || !utf8.ValidString(gtTag) {
			return
		}

		// Skip empty struct names (not valid Go)
		if structName == "" || fieldName == "" {
			return
		}

		// Skip names with spaces or special chars (not valid Go identifiers)
		for _, r := range structName {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return
			}
		}
		for _, r := range fieldName {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return
			}
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked on struct=%q field=%q json=%q tag=%q: %v",
					structName, fieldName, jsonName, gtTag, r)
			}
		}()

		// Create temporary test module
		tmpDir := t.TempDir()

		goMod := filepath.Join(tmpDir, "go.mod")
		if err := os.WriteFile(goMod, []byte("module fuzztest\n\ngo 1.23\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Generate Go code
		code := "package fuzztest\n\ntype " + structName + " struct {\n\t" +
			fieldName + " string `json:\"" + jsonName + "\" " + gtTag + "`\n}\n"

		testFile := filepath.Join(tmpDir, "test.go")
		if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
			t.Fatal(err)
		}

		// Try to parse - should never panic
		pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
		if err != nil {
			// Load errors are acceptable (invalid Go syntax)
			return
		}

		p := parser.NewParser()
		_, err = p.ParsePackages(pkgs)
		// Any result is acceptable - just shouldn't panic
		_ = err
	})
}

// FuzzNormalizeDoc fuzzes the documentation normalizer
func FuzzNormalizeDoc(f *testing.F) {
	// Seed corpus
	f.Add("Normal documentation")
	f.Add("")
	f.Add("Multiple\nlines\nof\ntext")
	f.Add("With\ttabs")
	f.Add("// Comment style")
	f.Add("   Leading spaces")
	f.Add("Trailing spaces   ")
	f.Add("\n\n\n")
	f.Add("Unicode: 日本語")

	f.Fuzz(func(t *testing.T, doc string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(doc) {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("normalizeDoc panicked on input %q: %v", doc, r)
			}
		}()

		// Call via integration since normalizeDoc is not exported
		// Just verify parser doesn't panic with various doc strings
		tmpDir := t.TempDir()

		goMod := filepath.Join(tmpDir, "go.mod")
		os.WriteFile(goMod, []byte("module test\n\ngo 1.23\n"), 0644)

		code := "package test\n\n// " + doc + "\ntype Test struct {\n\tField string `gt:\"required\"`\n}\n"
		testFile := filepath.Join(tmpDir, "test.go")
		os.WriteFile(testFile, []byte(code), 0644)

		pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
		if err != nil {
			return
		}

		p := parser.NewParser()
		_, _ = p.ParsePackages(pkgs)
	})
}
