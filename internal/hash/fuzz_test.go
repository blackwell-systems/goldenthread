// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package hash

import (
	"testing"
	"unicode/utf8"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// FuzzComputeSchemaHash fuzzes the schema hashing function
func FuzzComputeSchemaHash(f *testing.F) {
	// Seed corpus
	f.Add("User", "test", "ID", "id")
	f.Add("", "", "", "")
	f.Add("A", "a", "B", "b")
	f.Add("VeryLongName", "pkg", "VeryLongFieldName", "very_long_json_name")
	f.Add("Unicode日本語", "パッケージ", "フィールド", "json")
	f.Add("Emoji🎉", "pkg", "field🚀", "json")

	f.Fuzz(func(t *testing.T, schemaName, pkgName, fieldGoName, fieldJSONName string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(schemaName) || !utf8.ValidString(pkgName) ||
			!utf8.ValidString(fieldGoName) || !utf8.ValidString(fieldJSONName) {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ComputeSchemaHash panicked on schema=%q pkg=%q field=%q json=%q: %v",
					schemaName, pkgName, fieldGoName, fieldJSONName, r)
			}
		}()

		s := &schema.Schema{
			Name:        schemaName,
			PackageName: pkgName,
			Fields: []schema.Field{
				{
					GoName:   fieldGoName,
					JSONName: fieldJSONName,
					Type:     schema.Type{Kind: schema.TypeString},
					Optional: false,
				},
			},
		}

		hash1 := ComputeSchemaHash(s)

		// Verify hash properties
		if len(hash1) != 64 {
			t.Errorf("Expected 64-char hash, got %d", len(hash1))
		}

		// Verify hash is deterministic
		hash2 := ComputeSchemaHash(s)
		if hash1 != hash2 {
			t.Error("Hash not deterministic")
		}

		// Verify hash is hex
		for _, ch := range hash1 {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
				t.Errorf("Hash contains non-hex character: %c", ch)
			}
		}

		// Verify hash is not empty/zero
		if hash1 == "0000000000000000000000000000000000000000000000000000000000000000" {
			t.Error("Hash is all zeros")
		}
	})
}

// FuzzComputeSchemaHash_Stability fuzzes hash stability across modifications
func FuzzComputeSchemaHash_Stability(f *testing.F) {
	// Seed corpus for testing what should/shouldn't change hash
	f.Add("User", "original doc", "file1.go", "new doc", "file2.go")
	f.Add("Schema", "doc1", "a.go", "doc2", "b.go")
	f.Add("", "", "", "", "")

	f.Fuzz(func(t *testing.T, name, doc1, file1, doc2, file2 string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(name) || !utf8.ValidString(doc1) ||
			!utf8.ValidString(file1) || !utf8.ValidString(doc2) || !utf8.ValidString(file2) {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Hash stability test panicked: %v", r)
			}
		}()

		// Create two schemas that differ only in documentation and position
		s1 := &schema.Schema{
			Name:          name,
			PackageName:   "test",
			Documentation: doc1,
			Pos:           schema.SourcePos{File: file1, Line: 10},
			Fields: []schema.Field{
				{
					GoName:        "Field",
					JSONName:      "field",
					Type:          schema.Type{Kind: schema.TypeString},
					Documentation: doc1,
					Pos:           schema.SourcePos{File: file1, Line: 11},
				},
			},
		}

		s2 := &schema.Schema{
			Name:          name,
			PackageName:   "test",
			Documentation: doc2,
			Pos:           schema.SourcePos{File: file2, Line: 20},
			Fields: []schema.Field{
				{
					GoName:        "Field",
					JSONName:      "field",
					Type:          schema.Type{Kind: schema.TypeString},
					Documentation: doc2,
					Pos:           schema.SourcePos{File: file2, Line: 21},
				},
			},
		}

		hash1 := ComputeSchemaHash(s1)
		hash2 := ComputeSchemaHash(s2)

		// Documentation and position should NOT affect hash
		if hash1 != hash2 {
			t.Errorf("Hash changed due to doc/position changes: %s != %s", hash1, hash2)
		}
	})
}

// FuzzComputeSchemaHash_TypeChanges fuzzes that type changes DO affect hash
func FuzzComputeSchemaHash_TypeChanges(f *testing.F) {
	// Seed corpus with type kind values
	f.Add(int8(0), int8(1))  // TypeUnknown vs TypeString
	f.Add(int8(1), int8(2))  // TypeString vs TypeInt
	f.Add(int8(5), int8(6))  // TypeBool vs TypeFloat
	f.Add(int8(0), int8(13)) // TypeUnknown vs TypeMap

	f.Fuzz(func(t *testing.T, kind1, kind2 int8) {
		// Limit to valid TypeKind range (0-13)
		// Valid TypeKind range: 0 (TypeString) to 11 (TypeAny)
		if kind1 < 0 || kind1 > 11 || kind2 < 0 || kind2 > 11 {
			return
		}

		// Skip if same type
		if kind1 == kind2 {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Type change hash test panicked on kinds %d,%d: %v", kind1, kind2, r)
			}
		}()

		s1 := &schema.Schema{
			Name:        "Test",
			PackageName: "test",
			Fields: []schema.Field{
				{
					GoName:   "Value",
					JSONName: "value",
					Type:     schema.Type{Kind: schema.TypeKind(kind1)},
				},
			},
		}

		s2 := &schema.Schema{
			Name:        "Test",
			PackageName: "test",
			Fields: []schema.Field{
				{
					GoName:   "Value",
					JSONName: "value",
					Type:     schema.Type{Kind: schema.TypeKind(kind2)},
				},
			},
		}

		hash1 := ComputeSchemaHash(s1)
		hash2 := ComputeSchemaHash(s2)

		// Different types MUST produce different hashes
		if hash1 == hash2 {
			t.Errorf("Type change did not affect hash: kind %d vs %d", kind1, kind2)
		}
	})
}

// FuzzComputeSchemaHash_FieldOrder fuzzes that field order doesn't affect hash
func FuzzComputeSchemaHash_FieldOrder(f *testing.F) {
	// Seed corpus
	f.Add("Field1", "Field2", "json1", "json2")
	f.Add("A", "B", "a", "b")
	f.Add("", "X", "", "x")
	f.Add("Unicode", "日本語", "uni", "kanji")

	f.Fuzz(func(t *testing.T, go1, go2, json1, json2 string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(go1) || !utf8.ValidString(go2) ||
			!utf8.ValidString(json1) || !utf8.ValidString(json2) {
			return
		}

		// Skip if fields are identical
		if go1 == go2 {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Field order hash test panicked: %v", r)
			}
		}()

		// Create schema with fields in one order
		s1 := &schema.Schema{
			Name:        "Test",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: go1, JSONName: json1, Type: schema.Type{Kind: schema.TypeString}},
				{GoName: go2, JSONName: json2, Type: schema.Type{Kind: schema.TypeString}},
			},
		}

		// Create schema with fields in reverse order
		s2 := &schema.Schema{
			Name:        "Test",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: go2, JSONName: json2, Type: schema.Type{Kind: schema.TypeString}},
				{GoName: go1, JSONName: json1, Type: schema.Type{Kind: schema.TypeString}},
			},
		}

		hash1 := ComputeSchemaHash(s1)
		hash2 := ComputeSchemaHash(s2)

		// Field order should NOT affect hash (fields are sorted internally)
		if hash1 != hash2 {
			t.Errorf("Field order affected hash: %s != %s", hash1, hash2)
		}
	})
}

// FuzzComputeSchemaHash_OptionalChange fuzzes that optional flag changes hash
func FuzzComputeSchemaHash_OptionalChange(f *testing.F) {
	// Seed corpus
	f.Add("Field", "json")
	f.Add("", "")
	f.Add("VeryLongFieldName", "json_name")

	f.Fuzz(func(t *testing.T, goName, jsonName string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(goName) || !utf8.ValidString(jsonName) {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Optional flag test panicked: %v", r)
			}
		}()

		s1 := &schema.Schema{
			Name:        "Test",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: goName, JSONName: jsonName, Type: schema.Type{Kind: schema.TypeString}, Optional: false},
			},
		}

		s2 := &schema.Schema{
			Name:        "Test",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: goName, JSONName: jsonName, Type: schema.Type{Kind: schema.TypeString}, Optional: true},
			},
		}

		hash1 := ComputeSchemaHash(s1)
		hash2 := ComputeSchemaHash(s2)

		// Optional flag MUST affect hash
		if hash1 == hash2 {
			t.Error("Optional flag change did not affect hash")
		}
	})
}
