// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package normalize handles schema validation and normalization.
package normalize

import (
	"fmt"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// ValidateJSONNames checks for JSON name collisions within each schema.
// This must be called after embedded field flattening.
func ValidateJSONNames(schemas []*schema.Schema) error {
	for _, s := range schemas {
		if err := validateSchemaJSONNames(s); err != nil {
			return err
		}
	}
	return nil
}

// validateSchemaJSONNames checks a single schema for JSON name collisions.
func validateSchemaJSONNames(s *schema.Schema) error {
	jsonNames := make(map[string]*schema.Field)
	
	for i := range s.Fields {
		field := &s.Fields[i]
		
		// Skip embedded fields (should already be flattened)
		if field.Embedded {
			continue
		}
		
		// Get the JSON name for this field
		jsonName := field.JSONName
		if jsonName == "" {
			// Fallback to Go field name if no JSON tag
			jsonName = field.GoName
		}
		
		// Check for collision
		if existingField, exists := jsonNames[jsonName]; exists {
			return &schema.ValidationError{
				Field: jsonName,
				Message: fmt.Sprintf(
					"JSON name collision: field '%s' (from %s:%d) conflicts with field '%s' (from %s:%d)",
					field.GoName,
					field.Pos.File, field.Pos.Line,
					existingField.GoName,
					existingField.Pos.File, existingField.Pos.Line,
				),
				Pos: field.Pos,
			}
		}
		
		jsonNames[jsonName] = field
	}
	
	return nil
}

// ValidateGoNames checks for Go field name collisions within each schema.
// This catches issues before embedded field flattening.
func ValidateGoNames(schemas []*schema.Schema) error {
	for _, s := range schemas {
		if err := validateSchemaGoNames(s); err != nil {
			return err
		}
	}
	return nil
}

// validateSchemaGoNames checks a single schema for Go field name duplicates.
func validateSchemaGoNames(s *schema.Schema) error {
	goNames := make(map[string]*schema.Field)
	
	for i := range s.Fields {
		field := &s.Fields[i]
		
		// Check for collision
		if existingField, exists := goNames[field.GoName]; exists {
			return &schema.ValidationError{
				Field: field.GoName,
				Message: fmt.Sprintf(
					"duplicate field name (previously defined at %s:%d)",
					existingField.Pos.File, existingField.Pos.Line,
				),
				Pos: field.Pos,
			}
		}
		
		goNames[field.GoName] = field
	}
	
	return nil
}
