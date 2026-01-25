// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package goldenthread_test

import (
	"fmt"

	"github.com/blackwell-systems/goldenthread/internal/emitter/zod"
	"github.com/blackwell-systems/goldenthread/internal/normalize"
	"github.com/blackwell-systems/goldenthread/internal/parser"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// Example demonstrates programmatic usage of goldenthread as a library.
// Most users should use the CLI tool instead, but this shows the underlying API.
func Example() {
	// Create a simple schema programmatically
	s := &schema.Schema{
		Name:        "User",
		PackageName: "models",
		Fields: []schema.Field{
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules: schema.FieldRules{
					Required:  ptr(true),
					MinLength: ptr(3),
					MaxLength: ptr(20),
				},
			},
			{
				GoName:   "Email",
				JSONName: "email",
				Type:     schema.Type{Kind: schema.TypeString},
				Rules: schema.FieldRules{
					Email: ptr(true),
				},
			},
		},
	}

	// Generate Zod schema
	emitter := zod.NewEmitter()
	output, err := emitter.Emit(s)
	if err != nil {
		panic(err)
	}

	fmt.Println(output)
}

// ExampleParser demonstrates parsing Go source code to extract schemas.
func ExampleParser() {
	// In practice, use load.LoadPackages to get real go/packages
	// This is just a demonstration of the API
	p := parser.NewParser()
	_ = p

	fmt.Println("Parser extracts schemas from Go source using go/packages")
	// Output: Parser extracts schemas from Go source using go/packages
}

// ExampleNormalize demonstrates embedded struct flattening.
func Example_normalize() {
	// Create schema with embedded field
	base := &schema.Schema{
		Name:        "Base",
		PackageName: "models",
		Fields: []schema.Field{
			{
				GoName:   "ID",
				JSONName: "id",
				Type:     schema.Type{Kind: schema.TypeString},
			},
		},
	}

	user := &schema.Schema{
		Name:        "User",
		PackageName: "models",
		Fields: []schema.Field{
			{
				GoName:     "Base",
				JSONName:   "",
				Type:       schema.Type{Kind: schema.TypeNamed, GoType: "Base"},
				Embedded:   true,
			},
			{
				GoName:   "Username",
				JSONName: "username",
				Type:     schema.Type{Kind: schema.TypeString},
			},
		},
	}

	// Register and normalize
	registry := normalize.NewSchemaRegistry()
	registry.Register(base)
	registry.Register(user)

	normalized, err := normalize.FlattenEmbedded(user, registry)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Original fields: %d\n", len(user.Fields))
	fmt.Printf("Normalized fields: %d\n", len(normalized.Fields))
	// Output:
	// Original fields: 2
	// Normalized fields: 2
}

func ptr[T any](v T) *T {
	return &v
}
