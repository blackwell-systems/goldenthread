// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package hash_test

import (
	"testing"

	"github.com/blackwell-systems/goldenthread/internal/hash"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

func TestComputeSchemaHash_Deterministic(t *testing.T) {
	s := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Optional: false,
			},
			{
				GoName:   "Email",
				JSONName: "email",
				Type:     schema.Type{Kind: schema.TypeString},
				Optional: false,
			},
		},
	}

	// Compute hash multiple times
	hash1 := hash.ComputeSchemaHash(s)
	hash2 := hash.ComputeSchemaHash(s)
	hash3 := hash.ComputeSchemaHash(s)

	// Should be identical
	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("Hashes not deterministic: %q, %q, %q", hash1, hash2, hash3)
	}

	// Should be 64 hex characters (SHA-256)
	if len(hash1) != 64 {
		t.Errorf("Expected 64-char hash, got %d", len(hash1))
	}
}

func TestComputeSchemaHash_FieldOrderIndependent(t *testing.T) {
	s1 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
			{GoName: "Email", JSONName: "email", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	s2 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "Email", JSONName: "email", Type: schema.Type{Kind: schema.TypeString}},
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	hash1 := hash.ComputeSchemaHash(s1)
	hash2 := hash.ComputeSchemaHash(s2)

	// Should be identical (fields are sorted by name internally)
	if hash1 != hash2 {
		t.Errorf("Hashes differ with different field order: %q vs %q", hash1, hash2)
	}
}

func TestComputeSchemaHash_DocumentationIgnored(t *testing.T) {
	s1 := &schema.Schema{
		Name:          "User",
		PackageName:   "test",
		Documentation: "This is a user.",
		Fields: []schema.Field{
			{
				GoName:        "Username",
				JSONName:      "username",
				Type:          schema.Type{Kind: schema.TypeString},
				Documentation: "The username.",
			},
		},
	}

	s2 := &schema.Schema{
		Name:          "User",
		PackageName:   "test",
		Documentation: "A different description.",
		Fields: []schema.Field{
			{
				GoName:        "Username",
				JSONName:      "username",
				Type:          schema.Type{Kind: schema.TypeString},
				Documentation: "A different field description.",
			},
		},
	}

	hash1 := hash.ComputeSchemaHash(s1)
	hash2 := hash.ComputeSchemaHash(s2)

	// Documentation changes should not affect hash
	if hash1 != hash2 {
		t.Errorf("Hashes differ despite only doc changes: %q vs %q", hash1, hash2)
	}
}

func TestComputeSchemaHash_PositionIgnored(t *testing.T) {
	s1 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Pos:         schema.SourcePos{File: "file1.go", Line: 10},
		Fields: []schema.Field{
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Pos:      schema.SourcePos{File: "file1.go", Line: 11},
			},
		},
	}

	s2 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Pos:         schema.SourcePos{File: "file2.go", Line: 20},
		Fields: []schema.Field{
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Pos:      schema.SourcePos{File: "file2.go", Line: 21},
			},
		},
	}

	hash1 := hash.ComputeSchemaHash(s1)
	hash2 := hash.ComputeSchemaHash(s2)

	// Position changes should not affect hash
	if hash1 != hash2 {
		t.Errorf("Hashes differ despite only position changes: %q vs %q", hash1, hash2)
	}
}

func TestComputeSchemaHash_ValidationChanges(t *testing.T) {
	s1 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{MinLength: intPtr(3), MaxLength: intPtr(20)},
			},
		},
	}

	s2 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{MinLength: intPtr(5), MaxLength: intPtr(30)},
			},
		},
	}

	hash1 := hash.ComputeSchemaHash(s1)
	hash2 := hash.ComputeSchemaHash(s2)

	// Validation rule changes SHOULD affect hash
	if hash1 == hash2 {
		t.Error("Hashes should differ when validation rules change")
	}
}

func TestComputeSchemaHash_TypeChanges(t *testing.T) {
	s1 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "Age", JSONName: "age", Type: schema.Type{Kind: schema.TypeInt}},
		},
	}

	s2 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "Age", JSONName: "age", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	hash1 := hash.ComputeSchemaHash(s1)
	hash2 := hash.ComputeSchemaHash(s2)

	// Type changes SHOULD affect hash
	if hash1 == hash2 {
		t.Error("Hashes should differ when field types change")
	}
}

func TestComputeSchemaHash_OptionalChanges(t *testing.T) {
	s1 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "Bio", JSONName: "bio", Type: schema.Type{Kind: schema.TypeString}, Optional: false},
		},
	}

	s2 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "Bio", JSONName: "bio", Type: schema.Type{Kind: schema.TypeString}, Optional: true},
		},
	}

	hash1 := hash.ComputeSchemaHash(s1)
	hash2 := hash.ComputeSchemaHash(s2)

	// Optional flag changes SHOULD affect hash
	if hash1 == hash2 {
		t.Error("Hashes should differ when Optional flag changes")
	}
}

func TestComputeSchemaHash_EnumOrderIndependent(t *testing.T) {
	s1 := &schema.Schema{
		Name:        "Task",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Status",
				JSONName: "status",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{Enum: []string{"pending", "completed", "cancelled"}},
			},
		},
	}

	s2 := &schema.Schema{
		Name:        "Task",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Status",
				JSONName: "status",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{Enum: []string{"cancelled", "pending", "completed"}},
			},
		},
	}

	hash1 := hash.ComputeSchemaHash(s1)
	hash2 := hash.ComputeSchemaHash(s2)

	// Enum values are sorted, so order shouldn't matter
	if hash1 != hash2 {
		t.Errorf("Hashes should be identical for different enum order: %q vs %q", hash1, hash2)
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
