// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package normalize handles schema normalization including embedded struct flattening.
package normalize

import (
	"fmt"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)


// SchemaRegistry holds all parsed schemas for cross-reference resolution.
type SchemaRegistry struct {
	// schemas maps (PackageQualifier, Name) to Schema
	schemas map[string]*schema.Schema
}

// NewRegistry creates a new schema registry.
func NewRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]*schema.Schema),
	}
}

// Add adds a schema to the registry.
func (r *SchemaRegistry) Add(s *schema.Schema) {
	key := makeKey(s.PackageName, s.Name)
	r.schemas[key] = s
}

// Lookup finds a schema by type reference.
func (r *SchemaRegistry) Lookup(ref *schema.TypeRef) *schema.Schema {
	if ref == nil {
		return nil
	}
	// Try with package qualifier first, then without (local package)
	key := makeKey(ref.PackageQualifier, ref.Name)
	if s, ok := r.schemas[key]; ok {
		return s
	}
	// Try without package (same package reference)
	key = makeKey("", ref.Name)
	return r.schemas[key]
}

// makeKey creates a lookup key from package and name.
func makeKey(pkg, name string) string {
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

// FlattenEmbedded flattens embedded struct fields into their parent schemas.
// This must be called after all schemas are parsed and added to the registry.
func FlattenEmbedded(schemas []*schema.Schema) error {
	registry := NewRegistry()
	
	// First pass: add all schemas to registry
	for _, s := range schemas {
		registry.Add(s)
	}
	
	// Second pass: flatten embedded fields
	for _, s := range schemas {
		if err := flattenSchemaWithContext(s, s.PackageName, registry, make(map[string]bool)); err != nil {
			return err
		}
	}
	
	return nil
}

// flattenSchemaWithContext flattens embedded fields with package context.
func flattenSchemaWithContext(s *schema.Schema, currentPkg string, registry *SchemaRegistry, visiting map[string]bool) error {
	key := makeKey(s.PackageName, s.Name)
	
	// Detect cycles
	if visiting[key] {
		return &schema.ValidationError{
			Message: "cycle detected in embedded structs: " + key,
			Pos:     s.Pos,
		}
	}
	
	visiting[key] = true
	defer delete(visiting, key)
	
	// Track field names for collision detection
	fieldNames := make(map[string]schema.SourcePos)
	var newFields []schema.Field
	
	for _, field := range s.Fields {
		if field.Embedded && field.EmbeddedType != nil {
			// Look up the embedded schema
			// If PackageQualifier is empty, try current package
			ref := field.EmbeddedType
			if ref.PackageQualifier == "" {
				// Try with current package name
				ref = &schema.TypeRef{
					PackageQualifier: currentPkg,
					Name:             ref.Name,
				}
			}
			
			embeddedSchema := registry.Lookup(ref)
			if embeddedSchema == nil {
				// Can't resolve embedded type - skip flattening
				// This happens for embedded types from other packages not in the registry
				newFields = append(newFields, field)
				continue
			}
			
			// Recursively flatten the embedded schema first
			if err := flattenSchemaWithContext(embeddedSchema, currentPkg, registry, visiting); err != nil {
				return err
			}
			
			// Promote embedded schema's fields
			for _, embeddedField := range embeddedSchema.Fields {
				// Skip embedded fields that are themselves embedded (already flattened)
				if embeddedField.Embedded {
					continue
				}
				
				// Check for collisions
				if prevPos, exists := fieldNames[embeddedField.GoName]; exists {
					return &schema.ValidationError{
						Field:   embeddedField.GoName,
						Message: fmt.Sprintf("field name collision (previously defined at %s)", prevPos.String()),
						Pos:     field.Pos,
					}
				}
				
				// Also check JSON name collisions
				if embeddedField.JSONName != "" {
					for _, existing := range newFields {
						if existing.JSONName == embeddedField.JSONName {
							return &schema.ValidationError{
								Field:   embeddedField.JSONName,
								Message: fmt.Sprintf("JSON field name collision with %s", existing.GoName),
								Pos:     field.Pos,
							}
						}
					}
				}
				
				// Promote the field
				promotedField := embeddedField
				// Note: we keep the field's original documentation and tags
				
				newFields = append(newFields, promotedField)
				fieldNames[promotedField.GoName] = promotedField.Pos
			}
		} else {
			// Regular field - check for collisions and add
			if prevPos, exists := fieldNames[field.GoName]; exists {
				return &schema.ValidationError{
					Field:   field.GoName,
					Message: fmt.Sprintf("duplicate field name (previously defined at %s)", prevPos.String()),
					Pos:     field.Pos,
				}
			}
			
			newFields = append(newFields, field)
			fieldNames[field.GoName] = field.Pos
		}
	}
	
	// Replace fields with flattened version
	s.Fields = newFields
	
	return nil
}
