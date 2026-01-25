// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

/*
Package goldenthread is a schema compiler that generates TypeScript/Zod validation from Go structs.

# Overview

goldenthread generates type-safe validation schemas from Go struct definitions.
Define your domain models once in Go with validation tags, and compile them into
Zod schemas for TypeScript—automatically.

# Quick Start

Define a Go struct with validation tags:

	type User struct {
	    Username string `json:"username" gt:"required,len:3..20"`
	    Email    string `json:"email" gt:"email"`
	    Age      int    `json:"age" gt:"min:13,max:130"`
	}

Generate Zod schemas:

	$ goldenthread generate ./models

Use the generated TypeScript:

	import { UserSchema, User } from './gen/user'

	const result = UserSchema.safeParse(data)
	if (!result.success) {
	    console.error('Validation failed:', result.error)
	}

# Installation

	go install github.com/blackwell-systems/goldenthread/cmd/goldenthread@latest

# Commands

  - generate: Generate Zod schemas from Go source
  - check: Verify generated schemas are in sync (CI-ready)
  - version: Display version information

# Validation Tags

goldenthread supports comprehensive validation rules in the gt: struct tag:

String validation:

	Name string `gt:"required,len:1..100"`
	SKU  string `gt:"pattern:^[A-Z0-9]+$"`

Numeric validation:

	Age   int     `gt:"min:0,max:150"`
	Price float64 `gt:"min:0,max:999999.99"`

Format validation:

	Email string `gt:"email"`
	ID    string `gt:"uuid,required"`
	URL   string `gt:"url"`

Enums:

	Status string `gt:"enum:draft,published,archived"`

Arrays and maps:

	Tags     []string          `gt:"min:1,max:10"`
	Metadata map[string]string `gt:"optional"`

# Type System

Supported Go types:

  - Primitives: string, int, int64, float64, bool
  - Time: time.Time (mapped to z.string().datetime())
  - Collections: []T, map[string]T
  - Nested structs with references
  - Embedded structs (automatically flattened)
  - Pointer types (become optional: *string → z.string().optional())

# CI Integration

Verify schemas stay in sync with source:

	goldenthread check ./models

Returns exit code 1 if schemas are out of sync, perfect for CI pipelines.

# Architecture

goldenthread uses a five-stage compilation pipeline:

	Go source → Parse → Normalize → Hash → Emit → TypeScript
	            (AST)    (IR)      (drift) (Zod)   (generated)

The intermediate representation (IR) is language-agnostic, enabling future
emitters for OpenAPI, JSON Schema, or other validation frameworks.

# Quality Assurance

goldenthread includes comprehensive testing:

  - 53.4% test coverage across all packages
  - 12 fuzz targets running continuously (every 30 minutes)
  - 2 bugs found and fixed by fuzzing before v0.1.0 release
  - CI/CD with automated testing and linting

# License

Dual-licensed under Apache-2.0 OR MIT - your choice.
*/
package goldenthread
