// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package normalize_test

import (
	"testing"

	"github.com/blackwell-systems/goldenthread/internal/normalize"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

func TestValidateJSONNames_NoCollision(t *testing.T) {
	schemas := []*schema.Schema{
		{
			Name:        "User",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: "ID", JSONName: "id"},
				{GoName: "Name", JSONName: "name"},
				{GoName: "Email", JSONName: "email"},
			},
		},
	}

	if err := normalize.ValidateJSONNames(schemas); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateJSONNames_Collision(t *testing.T) {
	schemas := []*schema.Schema{
		{
			Name:        "User",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: "UserID", JSONName: "userId", Pos: schema.SourcePos{File: "test.go", Line: 10}},
				{GoName: "UserId", JSONName: "userId", Pos: schema.SourcePos{File: "test.go", Line: 13}},
			},
		},
	}

	err := normalize.ValidateJSONNames(schemas)
	if err == nil {
		t.Fatal("expected collision error, got nil")
	}

	if _, ok := err.(*schema.ValidationError); !ok {
		t.Errorf("expected ValidationError, got: %T", err)
	}
}

func TestValidateJSONNames_FallbackToGoName(t *testing.T) {
	schemas := []*schema.Schema{
		{
			Name:        "User",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: "ID", JSONName: ""},   // No JSON tag, falls back to "ID"
				{GoName: "Name", JSONName: ""}, // No JSON tag, falls back to "Name"
				{GoName: "Email", JSONName: "email"},
			},
		},
	}

	if err := normalize.ValidateJSONNames(schemas); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateJSONNames_FallbackCollision(t *testing.T) {
	schemas := []*schema.Schema{
		{
			Name:        "User",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: "Name", JSONName: "", Pos: schema.SourcePos{File: "test.go", Line: 10}},
				{GoName: "OtherField", JSONName: "Name", Pos: schema.SourcePos{File: "test.go", Line: 13}},
			},
		},
	}

	// First field falls back to "Name", second explicitly uses "Name" - collision!
	err := normalize.ValidateJSONNames(schemas)
	if err == nil {
		t.Fatal("expected collision error, got nil")
	}
}

func TestValidateGoNames_NoCollision(t *testing.T) {
	schemas := []*schema.Schema{
		{
			Name:        "User",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: "ID", JSONName: "id"},
				{GoName: "Name", JSONName: "name"},
				{GoName: "Email", JSONName: "email"},
			},
		},
	}

	if err := normalize.ValidateGoNames(schemas); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateGoNames_Collision(t *testing.T) {
	schemas := []*schema.Schema{
		{
			Name:        "User",
			PackageName: "test",
			Fields: []schema.Field{
				{GoName: "Name", JSONName: "name", Pos: schema.SourcePos{File: "test.go", Line: 10}},
				{GoName: "Name", JSONName: "full_name", Pos: schema.SourcePos{File: "test.go", Line: 13}},
			},
		},
	}

	err := normalize.ValidateGoNames(schemas)
	if err == nil {
		t.Fatal("expected collision error, got nil")
	}

	if _, ok := err.(*schema.ValidationError); !ok {
		t.Errorf("expected ValidationError, got: %T", err)
	}
}
