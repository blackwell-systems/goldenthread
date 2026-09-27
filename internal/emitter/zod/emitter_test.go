// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package zod_test

import (
	"strings"
	"testing"

	"github.com/blackwell-systems/goldenthread/internal/emitter/zod"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

func TestEmit_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		schema   *schema.Schema
		expected []string // Strings that should appear in output
	}{
		{
			name: "string field",
			schema: &schema.Schema{
				Name:        "User",
				PackageName: "test",
				Fields: []schema.Field{
					{
						GoName:   "Name",
						JSONName: "name",
						Type:     schema.Type{Kind: schema.TypeString},
						Optional: false,
					},
				},
			},
			expected: []string{
				"export const UserSchema = z.object({",
				"name: z.string()",
				"})",
				"export type User = z.infer<typeof UserSchema>",
			},
		},
		{
			name: "number field",
			schema: &schema.Schema{
				Name:        "Product",
				PackageName: "test",
				Fields: []schema.Field{
					{
						GoName:   "Price",
						JSONName: "price",
						Type:     schema.Type{Kind: schema.TypeFloat},
						Optional: false,
					},
				},
			},
			expected: []string{
				"price: z.number()",
			},
		},
		{
			name: "boolean field",
			schema: &schema.Schema{
				Name:        "Config",
				PackageName: "test",
				Fields: []schema.Field{
					{
						GoName:   "Enabled",
						JSONName: "enabled",
						Type:     schema.Type{Kind: schema.TypeBool},
						Optional: false,
					},
				},
			},
			expected: []string{
				"enabled: z.boolean()",
			},
		},
		{
			name: "optional field",
			schema: &schema.Schema{
				Name:        "User",
				PackageName: "test",
				Fields: []schema.Field{
					{
						GoName:   "Bio",
						JSONName: "bio",
						Type:     schema.Type{Kind: schema.TypeString},
						Optional: true,
					},
				},
			},
			expected: []string{
				"bio: z.string().optional()",
			},
		},
	}

	emitter := zod.NewEmitter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := emitter.Emit(tt.schema)
			if err != nil {
				t.Fatalf("Emit() error = %v", err)
			}

			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Output missing expected string: %q\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

func TestEmit_ValidationRules(t *testing.T) {
	tests := []struct {
		name     string
		field    schema.Field
		expected string
	}{
		{
			name: "string min length",
			field: schema.Field{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{MinLength: intPtr(3)},
			},
			expected: "username: z.string().min(3)",
		},
		{
			name: "string max length",
			field: schema.Field{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{MaxLength: intPtr(20)},
			},
			expected: "username: z.string().max(20)",
		},
		{
			name: "string length range",
			field: schema.Field{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{MinLength: intPtr(3), MaxLength: intPtr(20)},
			},
			expected: "username: z.string().min(3).max(20)",
		},
		{
			name: "string pattern",
			field: schema.Field{
				GoName:   "SKU",
				JSONName: "sku",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{Pattern: strPtr("^[A-Z0-9]+$")},
			},
			expected: "sku: z.string().regex(/^[A-Z0-9]+$/)",
		},
		{
			name: "numeric min",
			field: schema.Field{
				GoName:   "Age",
				JSONName: "age",
				Type:     schema.Type{Kind: schema.TypeInt},
				Rules:    schema.FieldRules{Min: floatPtr(13)},
			},
			expected: "age: z.number().min(13)",
		},
		{
			name: "numeric max",
			field: schema.Field{
				GoName:   "Age",
				JSONName: "age",
				Type:     schema.Type{Kind: schema.TypeInt},
				Rules:    schema.FieldRules{Max: floatPtr(130)},
			},
			expected: "age: z.number().max(130)",
		},
		{
			name: "numeric range",
			field: schema.Field{
				GoName:   "Age",
				JSONName: "age",
				Type:     schema.Type{Kind: schema.TypeInt},
				Rules:    schema.FieldRules{Min: floatPtr(13), Max: floatPtr(130)},
			},
			expected: "age: z.number().min(13).max(130)",
		},
	}

	emitter := zod.NewEmitter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Name:        "Test",
				PackageName: "test",
				Fields:      []schema.Field{tt.field},
			}

			output, err := emitter.Emit(s)
			if err != nil {
				t.Fatalf("Emit() error = %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Output missing expected string: %q\nGot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestEmit_Formats(t *testing.T) {
	tests := []struct {
		name     string
		format   schema.Format
		expected string
	}{
		{
			name:     "email format",
			format:   schema.FormatEmail,
			expected: "z.string().email()",
		},
		{
			name:     "uuid format",
			format:   schema.FormatUUID,
			expected: "z.string().uuid()",
		},
		{
			name:     "url format",
			format:   schema.FormatURL,
			expected: "z.string().url()",
		},
		{
			name:     "datetime format",
			format:   schema.FormatDateTime,
			expected: "z.string().datetime()",
		},
		{
			name:     "date format",
			format:   schema.FormatDate,
			expected: "z.string().regex(/^\\\\d{4}-\\\\d{2}-\\\\d{2}$/)",
		},
		{
			name:     "ipv4 format",
			format:   schema.FormatIPv4,
			expected: "z.string().ip({ version: 'v4' })",
		},
		{
			name:     "ipv6 format",
			format:   schema.FormatIPv6,
			expected: "z.string().ip({ version: 'v6' })",
		},
	}

	emitter := zod.NewEmitter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := tt.format
			s := &schema.Schema{
				Name:        "Test",
				PackageName: "test",
				Fields: []schema.Field{
					{
						GoName:   "Field",
						JSONName: "field",
						Type:     schema.Type{Kind: schema.TypeString},
						Rules:    schema.FieldRules{Format: &format},
					},
				},
			}

			output, err := emitter.Emit(s)
			if err != nil {
				t.Fatalf("Emit() error = %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Output missing expected string: %q\nGot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestEmit_Enum(t *testing.T) {
	s := &schema.Schema{
		Name:        "Task",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Status",
				JSONName: "status",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules:    schema.FieldRules{Enum: []string{"pending", "in_progress", "completed"}},
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	expected := "status: z.enum(['pending', 'in_progress', 'completed'])"
	if !strings.Contains(output, expected) {
		t.Errorf("Output missing expected enum: %q\nGot:\n%s", expected, output)
	}
}

func TestEmit_Array(t *testing.T) {
	elemType := schema.Type{Kind: schema.TypeString}
	s := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Tags",
				JSONName: "tags",
				Type:     schema.Type{Kind: schema.TypeArray, Elem: &elemType},
				Optional: false,
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	expected := "tags: z.array(z.string())"
	if !strings.Contains(output, expected) {
		t.Errorf("Output missing expected array: %q\nGot:\n%s", expected, output)
	}
}

func TestEmit_ArrayWithValidation(t *testing.T) {
	elemType := schema.Type{Kind: schema.TypeString}
	s := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Tags",
				JSONName: "tags",
				Type:     schema.Type{Kind: schema.TypeArray, Elem: &elemType},
				Rules:    schema.FieldRules{MinItems: intPtr(1), MaxItems: intPtr(10)},
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	if !strings.Contains(output, "z.array(z.string()).min(1).max(10)") {
		t.Errorf("Output missing array validation\nGot:\n%s", output)
	}
}

func TestEmit_Map(t *testing.T) {
	keyType := schema.Type{Kind: schema.TypeString}
	valueType := schema.Type{Kind: schema.TypeString}
	s := &schema.Schema{
		Name:        "Config",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Settings",
				JSONName: "settings",
				Type:     schema.Type{Kind: schema.TypeMap, Key: &keyType, Value: &valueType},
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	expected := "settings: z.record(z.string(), z.string())"
	if !strings.Contains(output, expected) {
		t.Errorf("Output missing expected map: %q\nGot:\n%s", expected, output)
	}
}

func TestEmit_Documentation(t *testing.T) {
	s := &schema.Schema{
		Name:          "User",
		PackageName:   "test",
		Documentation: "User represents a system user.",
		Fields: []schema.Field{
			{
				GoName:        "Username",
				JSONName:      "username",
				Type:          schema.Type{Kind: schema.TypeString},
				Documentation: "Username is the unique identifier.",
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	if !strings.Contains(output, "/**") {
		t.Error("Output missing JSDoc comment opening")
	}
	if !strings.Contains(output, "Username is the unique identifier.") {
		t.Error("Output missing field documentation")
	}
	if !strings.Contains(output, "*/") {
		t.Error("Output missing JSDoc comment closing")
	}
}

func TestEmit_SpecialTypes(t *testing.T) {
	tests := []struct {
		name     string
		typeKind schema.TypeKind
		expected string
	}{
		{
			name:     "time.Time",
			typeKind: schema.TypeTime,
			expected: "z.string().datetime()",
		},
		{
			name:     "UUID type",
			typeKind: schema.TypeUUID,
			expected: "z.string().uuid()",
		},
	}

	emitter := zod.NewEmitter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Name:        "Test",
				PackageName: "test",
				Fields: []schema.Field{
					{
						GoName:   "Field",
						JSONName: "field",
						Type:     schema.Type{Kind: tt.typeKind},
					},
				},
			}

			output, err := emitter.Emit(s)
			if err != nil {
				t.Fatalf("Emit() error = %v", err)
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Output missing expected type: %q\nGot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestEmit_NamedTypeReference(t *testing.T) {
	s := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "Profile",
				JSONName: "profile",
				Type: schema.Type{
					Kind: schema.TypeNamed,
					Ref:  &schema.TypeRef{PackageQualifier: "", Name: "Profile"},
				},
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	expected := "profile: ProfileSchema"
	if !strings.Contains(output, expected) {
		t.Errorf("Output missing type reference: %q\nGot:\n%s", expected, output)
	}
}

func TestEmit_MultipleFields(t *testing.T) {
	s := &schema.Schema{
		Name:        "User",
		PackageName: "test",
		Fields: []schema.Field{
			{
				GoName:   "ID",
				JSONName: "id",
				Type:     schema.Type{Kind: schema.TypeString},
			},
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
			},
			{
				GoName:   "Age",
				JSONName: "age",
				Type:     schema.Type{Kind: schema.TypeInt},
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	// Check all fields are present
	expected := []string{
		"id: z.string()",
		"username: z.string()",
		"age: z.number()",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Output missing field: %q\nGot:\n%s", exp, output)
		}
	}
}

func TestEmit_DiscriminatedUnion(t *testing.T) {
	edgePayload := schema.Field{
		GoName:   "Edge",
		JSONName: "edge",
		Type: schema.Type{
			Kind: schema.TypeNamed,
			Ref:  &schema.TypeRef{Name: "EdgeSpec"},
		},
	}
	switchPayload := schema.Field{
		GoName:   "Switch",
		JSONName: "switch",
		Type: schema.Type{
			Kind: schema.TypeNamed,
			Ref:  &schema.TypeRef{Name: "SwitchSpec"},
		},
	}
	joinPayload := schema.Field{
		GoName:   "Join",
		JSONName: "join",
		Type: schema.Type{
			Kind: schema.TypeNamed,
			Ref:  &schema.TypeRef{Name: "JoinSpec"},
		},
	}

	s := &schema.Schema{
		Name:        "WiringElement",
		PackageName: "test",
		Discriminator: &schema.DiscriminatedUnion{
			DiscriminatorName: "kind",
			Variants: []schema.Variant{
				{Value: "edge", PayloadField: &edgePayload},
				{Value: "switch", PayloadField: &switchPayload},
				{Value: "join", PayloadField: &joinPayload},
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	expected := []string{
		"export const WiringElementSchema = z.discriminatedUnion('kind', [",
		"kind: z.literal('edge'),",
		"edge: EdgeSpecSchema",
		"kind: z.literal('switch'),",
		"switch: SwitchSpecSchema",
		"kind: z.literal('join'),",
		"join: JoinSpecSchema",
		"export type WiringElement = z.infer<typeof WiringElementSchema>",
	}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Output missing expected string: %q\nGot:\n%s", exp, output)
		}
	}

	// Must not fall back to a plain z.object schema constant.
	if strings.Contains(output, "z.object({\n  kind:") {
		t.Errorf("Discriminated union should not emit a top-level z.object\nGot:\n%s", output)
	}
}

func TestEmit_DiscriminatedUnion_VariantWithoutPayload(t *testing.T) {
	activePayload := schema.Field{
		GoName:   "Active",
		JSONName: "active",
		Type: schema.Type{
			Kind: schema.TypeNamed,
			Ref:  &schema.TypeRef{Name: "ActiveState"},
		},
	}

	s := &schema.Schema{
		Name:        "State",
		PackageName: "test",
		Discriminator: &schema.DiscriminatedUnion{
			DiscriminatorName: "kind",
			Variants: []schema.Variant{
				{Value: "active", PayloadField: &activePayload},
				// idle variant carries only the discriminator literal.
				{Value: "idle", PayloadField: nil},
			},
		},
	}

	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	if !strings.Contains(output, "kind: z.literal('idle')") {
		t.Errorf("Output missing payload-free variant literal\nGot:\n%s", output)
	}
	// A payload-free variant object must close cleanly with no dangling field.
	if !strings.Contains(output, "z.literal('idle')\n  })") {
		t.Errorf("Payload-free variant object malformed\nGot:\n%s", output)
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
