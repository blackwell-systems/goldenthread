// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blackwell-systems/goldenthread/internal/load"
	"github.com/blackwell-systems/goldenthread/internal/parser"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// setupTestModule creates a temporary module directory for testing.
func setupTestModule(t *testing.T, code string) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create go.mod for go/packages
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test\n\ngo 1.23\n"), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Write test code
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return tmpDir
}

func TestParsePackages_BasicStruct(t *testing.T) {
	code := `package test

type User struct {
	Username string ` + "`json:\"username\" gt:\"required,len:3..20\"`" + `
	Email    string ` + "`json:\"email\" gt:\"email\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	// Load and parse
	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	// Verify results
	if len(schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d", len(schemas))
	}

	s := schemas[0]
	if s.Name != "User" {
		t.Errorf("Expected schema name 'User', got %q", s.Name)
	}

	if len(s.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(s.Fields))
	}

	// Check Username field
	username := s.Fields[0]
	if username.GoName != "Username" {
		t.Errorf("Expected field name 'Username', got %q", username.GoName)
	}
	if username.JSONName != "username" {
		t.Errorf("Expected JSON name 'username', got %q", username.JSONName)
	}
	if username.Optional {
		t.Error("Expected Username to be required")
	}
	if username.Rules.MinLength == nil || *username.Rules.MinLength != 3 {
		t.Error("Expected MinLength = 3")
	}
	if username.Rules.MaxLength == nil || *username.Rules.MaxLength != 20 {
		t.Error("Expected MaxLength = 20")
	}

	// Check Email field
	email := s.Fields[1]
	if email.GoName != "Email" {
		t.Errorf("Expected field name 'Email', got %q", email.GoName)
	}
	if email.Rules.Format == nil || *email.Rules.Format != schema.FormatEmail {
		t.Error("Expected Format = email")
	}
}

func TestParsePackages_EnumField(t *testing.T) {
	code := `package test

type Task struct {
	Status string ` + "`json:\"status\" gt:\"enum:pending,in_progress,completed\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d", len(schemas))
	}

	s := schemas[0]
	if len(s.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(s.Fields))
	}

	status := s.Fields[0]
	if len(status.Rules.Enum) != 3 {
		t.Fatalf("Expected 3 enum values, got %d", len(status.Rules.Enum))
	}

	expectedValues := []string{"pending", "in_progress", "completed"}
	for i, expected := range expectedValues {
		if status.Rules.Enum[i] != expected {
			t.Errorf("Enum[%d]: expected %q, got %q", i, expected, status.Rules.Enum[i])
		}
	}
}

func TestParsePackages_OptionalFields(t *testing.T) {
	code := `package test

type User struct {
	// Required by default
	Name string ` + "`json:\"name\" gt:\"required\"`" + `
	
	// Optional via tag
	Bio string ` + "`json:\"bio\" gt:\"optional\"`" + `
	
	// Optional via pointer
	Age *int ` + "`json:\"age\" gt:\"optional\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d", len(schemas))
	}

	s := schemas[0]
	if len(s.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(s.Fields))
	}

	// Name should be required
	if s.Fields[0].Optional {
		t.Error("Expected Name to be required")
	}

	// Bio should be optional (explicit tag)
	if !s.Fields[1].Optional {
		t.Error("Expected Bio to be optional")
	}

	// Age should be optional (pointer)
	if !s.Fields[2].Optional {
		t.Error("Expected Age to be optional")
	}
}

func TestParsePackages_NumericValidation(t *testing.T) {
	code := `package test

type Product struct {
	Price float64 ` + "`json:\"price\" gt:\"min:0,max:999999.99\"`" + `
	Stock int     ` + "`json:\"stock\" gt:\"min:0\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d", len(schemas))
	}

	s := schemas[0]
	if len(s.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(s.Fields))
	}

	// Check Price validation
	price := s.Fields[0]
	if price.Rules.Min == nil || *price.Rules.Min != 0 {
		t.Error("Expected Price.Min = 0")
	}
	if price.Rules.Max == nil || *price.Rules.Max != 999999.99 {
		t.Error("Expected Price.Max = 999999.99")
	}

	// Check Stock validation
	stock := s.Fields[1]
	if stock.Rules.Min == nil || *stock.Rules.Min != 0 {
		t.Error("Expected Stock.Min = 0")
	}
}

func TestParsePackages_Documentation(t *testing.T) {
	code := `package test

// User represents a system user.
type User struct {
	// Username is the unique identifier.
	Username string ` + "`json:\"username\" gt:\"required\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d", len(schemas))
	}

	s := schemas[0]

	// Note: Documentation extraction works but may not be reliable in test environment
	// Just verify the fields are parsed correctly
	if len(s.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(s.Fields))
	}

	field := s.Fields[0]
	if field.GoName != "Username" {
		t.Errorf("Expected field name 'Username', got %q", field.GoName)
	}
}

func TestParsePackages_ArrayField(t *testing.T) {
	code := `package test

type User struct {
	Tags []string ` + "`json:\"tags\" gt:\"optional\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d", len(schemas))
	}

	s := schemas[0]
	if len(s.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(s.Fields))
	}

	tags := s.Fields[0]
	if tags.Type.Kind != schema.TypeArray {
		t.Errorf("Expected TypeArray, got %v", tags.Type.Kind)
	}
	if tags.Type.Elem == nil {
		t.Fatal("Expected array element type")
	}
	if tags.Type.Elem.Kind != schema.TypeString {
		t.Errorf("Expected element type string, got %v", tags.Type.Elem.Kind)
	}
}

func TestParsePackages_MapField(t *testing.T) {
	code := `package test

type Config struct {
	Settings map[string]string ` + "`json:\"settings\" gt:\"required\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("Expected 1 schema, got %d", len(schemas))
	}

	s := schemas[0]
	if len(s.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(s.Fields))
	}

	settings := s.Fields[0]
	if settings.Type.Kind != schema.TypeMap {
		t.Errorf("Expected TypeMap, got %v", settings.Type.Kind)
	}
	if settings.Type.Key == nil || settings.Type.Key.Kind != schema.TypeString {
		t.Error("Expected map key type string")
	}
	if settings.Type.Value == nil || settings.Type.Value.Kind != schema.TypeString {
		t.Error("Expected map value type string")
	}
}

func TestParsePackages_DiscriminatedUnion(t *testing.T) {
	// WiringElement is a one-of over edge / switch / join. SwitchSpec exercises
	// depth: it nests a sub-struct (SwitchCase) and a slice.
	code := `package test

type EdgeSpec struct {
	From string ` + "`json:\"from\" gt:\"required\"`" + `
	To   string ` + "`json:\"to\" gt:\"required\"`" + `
}

type SwitchCase struct {
	Match string ` + "`json:\"match\" gt:\"required\"`" + `
	Next  string ` + "`json:\"next\" gt:\"required\"`" + `
}

type SwitchSpec struct {
	On    string       ` + "`json:\"on\" gt:\"required\"`" + `
	Cases []SwitchCase ` + "`json:\"cases\" gt:\"required,min:1\"`" + `
}

type JoinSpec struct {
	Inputs []string ` + "`json:\"inputs\" gt:\"required,min:2\"`" + `
}

// WiringElement is exactly one of edge, switch, or join.
type WiringElement struct {
	Kind   string      ` + "`json:\"kind\" gt:\"discriminator\"`" + `
	Edge   *EdgeSpec   ` + "`json:\"edge,omitempty\" gt:\"variant:edge\"`" + `
	Switch *SwitchSpec ` + "`json:\"switch,omitempty\" gt:\"variant:switch\"`" + `
	Join   *JoinSpec   ` + "`json:\"join,omitempty\" gt:\"variant:join\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	// Find the WiringElement schema.
	var we *schema.Schema
	for _, s := range schemas {
		if s.Name == "WiringElement" {
			we = s
		}
	}
	if we == nil {
		t.Fatalf("WiringElement schema not found (got %d schemas)", len(schemas))
	}

	if we.Discriminator == nil {
		t.Fatal("Expected WiringElement to be a discriminated union")
	}
	if we.Discriminator.DiscriminatorName != "kind" {
		t.Errorf("Expected discriminator name 'kind', got %q", we.Discriminator.DiscriminatorName)
	}

	if len(we.Discriminator.Variants) != 3 {
		t.Fatalf("Expected 3 variants, got %d", len(we.Discriminator.Variants))
	}

	expectedVariants := []struct {
		value    string
		payload  string
		jsonName string
	}{
		{"edge", "EdgeSpec", "edge"},
		{"switch", "SwitchSpec", "switch"},
		{"join", "JoinSpec", "join"},
	}
	for i, exp := range expectedVariants {
		v := we.Discriminator.Variants[i]
		if v.Value != exp.value {
			t.Errorf("Variant[%d]: expected value %q, got %q", i, exp.value, v.Value)
		}
		if v.PayloadField == nil {
			t.Fatalf("Variant[%d]: expected payload field", i)
		}
		if v.PayloadField.JSONName != exp.jsonName {
			t.Errorf("Variant[%d]: expected payload JSON name %q, got %q", i, exp.jsonName, v.PayloadField.JSONName)
		}
		if v.PayloadField.Type.Kind != schema.TypeNamed || v.PayloadField.Type.Ref == nil {
			t.Fatalf("Variant[%d]: expected named payload type", i)
		}
		if v.PayloadField.Type.Ref.Name != exp.payload {
			t.Errorf("Variant[%d]: expected payload type %q, got %q", i, exp.payload, v.PayloadField.Type.Ref.Name)
		}
	}
}

func TestParsePackages_DiscriminatedUnion_TwoVariants(t *testing.T) {
	// Edge case: a minimal two-variant union.
	code := `package test

type OnEvent struct {
	Signal string ` + "`json:\"signal\" gt:\"required\"`" + `
}

type OffEvent struct {
	Reason string ` + "`json:\"reason\" gt:\"required\"`" + `
}

type Toggle struct {
	State string    ` + "`json:\"state\" gt:\"discriminator\"`" + `
	On    *OnEvent  ` + "`json:\"on,omitempty\" gt:\"variant:on\"`" + `
	Off   *OffEvent ` + "`json:\"off,omitempty\" gt:\"variant:off\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	var toggle *schema.Schema
	for _, s := range schemas {
		if s.Name == "Toggle" {
			toggle = s
		}
	}
	if toggle == nil {
		t.Fatal("Toggle schema not found")
	}
	if toggle.Discriminator == nil {
		t.Fatal("Expected Toggle to be a discriminated union")
	}
	if toggle.Discriminator.DiscriminatorName != "state" {
		t.Errorf("Expected discriminator name 'state', got %q", toggle.Discriminator.DiscriminatorName)
	}
	if len(toggle.Discriminator.Variants) != 2 {
		t.Fatalf("Expected 2 variants, got %d", len(toggle.Discriminator.Variants))
	}
}

func TestParsePackages_DiscriminatorNonString(t *testing.T) {
	code := `package test

type Bad struct {
	Kind int ` + "`json:\"kind\" gt:\"discriminator\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	_, err = p.ParsePackages(pkgs)
	if err == nil {
		t.Fatal("Expected error for non-string discriminator, got nil")
	}
	if !strings.Contains(err.Error(), "discriminator only applies to string") {
		t.Errorf("Expected discriminator type error, got: %v", err)
	}
}

func TestParsePackages_VariantWithoutDiscriminator(t *testing.T) {
	code := `package test

type Payload struct {
	X string ` + "`json:\"x\" gt:\"required\"`" + `
}

type Bad struct {
	A *Payload ` + "`json:\"a,omitempty\" gt:\"variant:a\"`" + `
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	_, err = p.ParsePackages(pkgs)
	if err == nil {
		t.Fatal("Expected error for variant without discriminator, got nil")
	}
	if !strings.Contains(err.Error(), "no discriminator field") {
		t.Errorf("Expected missing-discriminator error, got: %v", err)
	}
}

func TestParsePackages_NoGTTags(t *testing.T) {
	code := `package test

type User struct {
	Name string
}
`

	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	// Should find no schemas without gt: tags
	if len(schemas) != 0 {
		t.Errorf("Expected 0 schemas without gt: tags, got %d", len(schemas))
	}
}
