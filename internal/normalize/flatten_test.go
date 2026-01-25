// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package normalize_test

import (
	"testing"

	"github.com/blackwell-systems/goldenthread/internal/normalize"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

func TestFlattenEmbedded_NoEmbedded(t *testing.T) {
	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "ID", JSONName: "id", Type: schema.Type{Kind: schema.TypeString}},
			{GoName: "Name", JSONName: "name", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	schemas := []*schema.Schema{user}
	err := normalize.FlattenEmbedded(schemas)
	if err != nil {
		t.Fatalf("FlattenEmbedded() error = %v", err)
	}

	// Should have same fields (no changes)
	if len(user.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(user.Fields))
	}
}

func TestFlattenEmbedded_SimpleEmbed(t *testing.T) {
	base := &schema.Schema{
		Name:        "Base",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "ID", JSONName: "id", Type: schema.Type{Kind: schema.TypeString}},
			{GoName: "CreatedAt", JSONName: "created_at", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "Base",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Base"},
			},
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	schemas := []*schema.Schema{base, user}
	err := normalize.FlattenEmbedded(schemas)
	if err != nil {
		t.Fatalf("FlattenEmbedded() error = %v", err)
	}

	// User should now have 3 fields: ID, CreatedAt (from Base), Username
	if len(user.Fields) != 3 {
		t.Fatalf("Expected 3 fields after flattening, got %d", len(user.Fields))
	}

	// Check field order and names
	expectedFields := []string{"ID", "CreatedAt", "Username"}
	for i, expected := range expectedFields {
		if user.Fields[i].GoName != expected {
			t.Errorf("Field[%d]: expected %q, got %q", i, expected, user.Fields[i].GoName)
		}
	}
}

func TestFlattenEmbedded_NestedEmbed(t *testing.T) {
	timestamps := &schema.Schema{
		Name:        "Timestamps",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "CreatedAt", JSONName: "created_at", Type: schema.Type{Kind: schema.TypeString}},
			{GoName: "UpdatedAt", JSONName: "updated_at", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	base := &schema.Schema{
		Name:        "Base",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "Timestamps",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Timestamps"},
			},
			{GoName: "ID", JSONName: "id", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "Base",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Base"},
			},
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	schemas := []*schema.Schema{timestamps, base, user}
	err := normalize.FlattenEmbedded(schemas)
	if err != nil {
		t.Fatalf("FlattenEmbedded() error = %v", err)
	}

	// User should have: CreatedAt, UpdatedAt, ID (from Base which includes Timestamps), Username
	if len(user.Fields) != 4 {
		t.Fatalf("Expected 4 fields after nested flattening, got %d", len(user.Fields))
	}

	expectedFields := []string{"CreatedAt", "UpdatedAt", "ID", "Username"}
	for i, expected := range expectedFields {
		if user.Fields[i].GoName != expected {
			t.Errorf("Field[%d]: expected %q, got %q", i, expected, user.Fields[i].GoName)
		}
	}
}

func TestFlattenEmbedded_MultipleEmbeds(t *testing.T) {
	timestamps := &schema.Schema{
		Name:        "Timestamps",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "CreatedAt", JSONName: "created_at", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	metadata := &schema.Schema{
		Name:        "Metadata",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "Version", JSONName: "version", Type: schema.Type{Kind: schema.TypeInt}},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "Timestamps",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Timestamps"},
			},
			{
				GoName:       "Metadata",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Metadata"},
			},
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	schemas := []*schema.Schema{timestamps, metadata, user}
	err := normalize.FlattenEmbedded(schemas)
	if err != nil {
		t.Fatalf("FlattenEmbedded() error = %v", err)
	}

	// User should have: CreatedAt, Version, Username
	if len(user.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(user.Fields))
	}

	expectedFields := []string{"CreatedAt", "Version", "Username"}
	for i, expected := range expectedFields {
		if user.Fields[i].GoName != expected {
			t.Errorf("Field[%d]: expected %q, got %q", i, expected, user.Fields[i].GoName)
		}
	}
}

func TestFlattenEmbedded_GoNameCollision(t *testing.T) {
	base := &schema.Schema{
		Name:        "Base",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "ID", JSONName: "id", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "Base",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Base"},
				Pos:          schema.SourcePos{File: "test.go", Line: 10},
			},
			{
				GoName:   "ID",
				JSONName: "user_id",
				Type:     schema.Type{Kind: schema.TypeString},
				Pos:      schema.SourcePos{File: "test.go", Line: 11},
			},
		},
	}

	schemas := []*schema.Schema{base, user}
	err := normalize.FlattenEmbedded(schemas)
	
	// Should error due to field name collision
	if err == nil {
		t.Fatal("Expected error for field name collision, got nil")
	}

	// Check error type
	if valErr, ok := err.(*schema.ValidationError); ok {
		if valErr.Field != "ID" {
			t.Errorf("Expected error for field 'ID', got %q", valErr.Field)
		}
	} else {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestFlattenEmbedded_JSONNameCollision(t *testing.T) {
	base := &schema.Schema{
		Name:        "Base",
		PackageName: "test",
		Fields: []schema.Field{
			{GoName: "CreatedAt", JSONName: "created_at", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Timestamp",
				JSONName: "created_at",
				Type:     schema.Type{Kind: schema.TypeString},
				Pos:      schema.SourcePos{File: "test.go", Line: 10},
			},
			{
				GoName:       "Base",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Base"},
				Pos:          schema.SourcePos{File: "test.go", Line: 11},
			},
		},
	}

	schemas := []*schema.Schema{base, user}
	err := normalize.FlattenEmbedded(schemas)
	
	// Should error due to JSON name collision
	if err == nil {
		t.Fatal("Expected error for JSON name collision, got nil")
	}

	// Check error mentions collision
	if valErr, ok := err.(*schema.ValidationError); ok {
		if valErr.Field != "created_at" {
			t.Errorf("Expected error for field 'created_at', got %q", valErr.Field)
		}
	} else {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestFlattenEmbedded_CycleDetection(t *testing.T) {
	// Create a cycle: A embeds B, B embeds A
	schemaA := &schema.Schema{
		Name:        "A",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "B",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "B"},
			},
		},
	}

	schemaB := &schema.Schema{
		Name:        "B",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "A",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "A"},
			},
		},
	}

	schemas := []*schema.Schema{schemaA, schemaB}
	err := normalize.FlattenEmbedded(schemas)
	
	// Should error due to cycle
	if err == nil {
		t.Fatal("Expected error for cycle detection, got nil")
	}
}

func TestFlattenEmbedded_UnresolvableType(t *testing.T) {
	// User embeds Base, but Base is not in registry
	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "Base",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "other", Name: "Base"},
			},
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	schemas := []*schema.Schema{user}
	err := normalize.FlattenEmbedded(schemas)
	if err != nil {
		t.Fatalf("FlattenEmbedded() error = %v (should skip unresolvable types)", err)
	}

	// Should keep the embedded field as-is (not flatten unresolvable)
	if len(user.Fields) != 2 {
		t.Errorf("Expected 2 fields (unresolvable embed preserved), got %d", len(user.Fields))
	}
}

func TestFlattenEmbedded_PackageQualifier(t *testing.T) {
	// Schema from different package
	base := &schema.Schema{
		Name:        "Base",
		PackageName: "models",
		Fields: []schema.Field{
			{GoName: "ID", JSONName: "id", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "users",
		Fields: []schema.Field{
			{
				GoName:       "Base",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "models", Name: "Base"},
			},
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	schemas := []*schema.Schema{base, user}
	err := normalize.FlattenEmbedded(schemas)
	if err != nil {
		t.Fatalf("FlattenEmbedded() error = %v", err)
	}

	// Should flatten cross-package embed
	if len(user.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(user.Fields))
	}

	if user.Fields[0].GoName != "ID" {
		t.Errorf("Expected first field to be 'ID', got %q", user.Fields[0].GoName)
	}
}

func TestSchemaRegistry_Lookup(t *testing.T) {
	registry := normalize.NewRegistry()

	schema1 := &schema.Schema{
		Name:        "User",
		PackageName: "test",
	}
	schema2 := &schema.Schema{
		Name:        "Base",
		PackageName: "models",
	}
	schema3 := &schema.Schema{
		Name:        "Local",
		PackageName: "",
	}

	registry.Add(schema1)
	registry.Add(schema2)
	registry.Add(schema3)

	tests := []struct {
		name     string
		ref      *schema.TypeRef
		expected *schema.Schema
	}{
		{
			name:     "lookup with package qualifier",
			ref:      &schema.TypeRef{PackageQualifier: "models", Name: "Base"},
			expected: schema2,
		},
		{
			name:     "lookup with package test",
			ref:      &schema.TypeRef{PackageQualifier: "test", Name: "User"},
			expected: schema1,
		},
		{
			name:     "lookup without package (empty pkg schema)",
			ref:      &schema.TypeRef{PackageQualifier: "", Name: "Local"},
			expected: schema3,
		},
		{
			name:     "lookup nil ref",
			ref:      nil,
			expected: nil,
		},
		{
			name:     "lookup non-existent",
			ref:      &schema.TypeRef{PackageQualifier: "", Name: "NotFound"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.Lookup(tt.ref)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFlattenEmbedded_PreservesValidation(t *testing.T) {
	base := &schema.Schema{
		Name:        "Base",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Email",
				JSONName: "email",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{Format: formatPtr(schema.FormatEmail)},
			},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:       "Base",
				Embedded:     true,
				EmbeddedType: &schema.TypeRef{PackageQualifier: "", Name: "Base"},
			},
			{GoName: "Username", JSONName: "username", Type: schema.Type{Kind: schema.TypeString}},
		},
	}

	schemas := []*schema.Schema{base, user}
	err := normalize.FlattenEmbedded(schemas)
	if err != nil {
		t.Fatalf("FlattenEmbedded() error = %v", err)
	}

	// Check that Email field's validation is preserved
	emailField := user.Fields[0]
	if emailField.GoName != "Email" {
		t.Fatalf("Expected first field to be 'Email', got %q", emailField.GoName)
	}
	if emailField.Rules.Format == nil || *emailField.Rules.Format != schema.FormatEmail {
		t.Error("Expected Email field to preserve format validation")
	}
}

// Helper function
func formatPtr(f schema.Format) *schema.Format {
	return &f
}
