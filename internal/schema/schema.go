// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package schema defines the intermediate representation (IR) for domain schemas.
// This IR is language-agnostic and serves as the bridge between Go AST parsing
// and code generation for target languages (Zod, TypeScript, OpenAPI).
package schema

// Schema represents a complete domain model extracted from a Go struct.
type Schema struct {
	// Name is the struct name (e.g., "User")
	Name string

	// PackageName is the Go package name (not import path)
	PackageName string

	// Fields are the struct fields with their validation rules
	Fields []Field

	// Rules are schema-level validation rules (cross-field constraints)
	Rules []Rule

	// Documentation is the comment above the struct
	Documentation string

	// Pos tracks where this schema was defined
	Pos SourcePos
}

// Field represents a single struct field with validation rules.
type Field struct {
	// GoName is the Go field name (e.g., "Username")
	GoName string

	// JSONName is the JSON tag name (e.g., "username")
	JSONName string

	// Type is the field's type information
	Type Type

	// Optional indicates if this field can be absent
	// Derived from: pointer types, omitempty tag, or !required
	Optional bool

	// Rules are field-specific validation rules
	Rules FieldRules

	// Documentation is the comment above the field
	Documentation string

	// Tags preserves original Go struct tags
	Tags map[string]string

	// Pos tracks where this field was defined
	Pos SourcePos
	
	// Embedded indicates if this field is an embedded struct
	// Embedded fields should be flattened during normalization
	Embedded bool
	
	// EmbeddedType is the type reference for embedded structs
	// Only set if Embedded is true
	EmbeddedType *TypeRef
}

// Type describes a field's type structure.
// Separates type references (named types) from shapes (scalars/composites).
type Type struct {
	// Kind is the base type category
	Kind TypeKind

	// Ref is the reference to a named type (structs, custom types)
	// Non-nil when Kind == TypeNamed
	Ref *TypeRef

	// Elem is the element type for arrays (when Kind == TypeArray)
	Elem *Type

	// Key and Value types for maps (when Kind == TypeMap)
	Key   *Type
	Value *Type

	// Fields for inline object types (when Kind == TypeObject)
	Fields []Field
}

// TypeRef references a named Go type.
type TypeRef struct {
	// PackageQualifier is the package qualifier or import path
	// For AST-only parsing: may be package name ("time", "uuid")
	// With go/packages: full import path ("github.com/google/uuid")
	// Empty string means local package
	PackageQualifier string

	// Name is the type name (e.g., "Time", "User")
	Name string
}

// TypeKind categorizes field types.
type TypeKind int

const (
	// Scalar types
	TypeString TypeKind = iota
	TypeInt
	TypeUint
	TypeFloat
	TypeBool
	TypeTime
	TypeUUID

	// Composite types
	TypeArray
	TypeMap
	TypeObject

	// Named type reference
	TypeNamed

	// TypeAny represents interface{} or any
	TypeAny
)

// String returns the string representation of TypeKind.
func (k TypeKind) String() string {
	switch k {
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeUint:
		return "uint"
	case TypeFloat:
		return "float"
	case TypeBool:
		return "bool"
	case TypeTime:
		return "time"
	case TypeUUID:
		return "uuid"
	case TypeArray:
		return "array"
	case TypeMap:
		return "map"
	case TypeObject:
		return "object"
	case TypeNamed:
		return "named"
	case TypeAny:
		return "any"
	default:
		return "unknown"
	}
}

// FieldRules contains all validation rules for a field.
type FieldRules struct {
	// String rules
	MinLength *int
	MaxLength *int
	Pattern   *string
	Format    *Format

	// Numeric rules
	Min *float64
	Max *float64

	// Array rules
	MinItems    *int
	MaxItems    *int
	UniqueItems bool

	// Enum values (for oneof)
	Enum []string

	// Custom validators (function names)
	CustomValidators []string
}

// Format represents standardized string formats.
type Format string

const (
	FormatEmail    Format = "email"
	FormatUUID     Format = "uuid"
	FormatURL      Format = "url"
	FormatDate     Format = "date"
	FormatDateTime Format = "datetime"
	FormatIPv4     Format = "ipv4"
	FormatIPv6     Format = "ipv6"
)

// Rule represents a schema-level validation rule (cross-field constraints).
type Rule struct {
	// Type is the rule category
	Type RuleType

	// Expression is the validation logic (implementation-specific)
	Expression string

	// Message is the error message when validation fails
	Message string
}

// RuleType categorizes schema-level rules.
type RuleType int

const (
	// RuleRefine is a custom validation function
	RuleRefine RuleType = iota

	// RuleDepends indicates conditional field requirements
	RuleDepends
)

// SourcePos tracks where a schema element was defined in source code.
// Used for readable generated comments and better error messages.
type SourcePos struct {
	// File is the absolute path to the Go file
	File string

	// Line is the line number
	Line int

	// Column is the column number
	Column int
}

// String returns a file:line string for error messages.
func (p SourcePos) String() string {
	if p.File == "" {
		return ""
	}
	return p.File + ":" + intToString(p.Line)
}

// Validate checks if a Schema is semantically valid.
// Enforces:
// - Non-empty names
// - No duplicate field names
// - No conflicting optional/required flags
// - Embedded struct collision detection
func (s *Schema) Validate() error {
	if s.Name == "" {
		return &ValidationError{
			Field:   "Name",
			Message: "schema name cannot be empty",
			Pos:     s.Pos,
		}
	}

	seen := make(map[string]SourcePos)
	for i, field := range s.Fields {
		if field.GoName == "" {
			return &ValidationError{
				Field:   "Fields",
				Message: "field name cannot be empty",
				Index:   &i,
				Pos:     field.Pos,
			}
		}

		if prevPos, exists := seen[field.GoName]; exists {
			return &ValidationError{
				Field:   field.GoName,
				Message: "duplicate field name (previously defined at " + prevPos.String() + ")",
				Index:   &i,
				Pos:     field.Pos,
			}
		}
		seen[field.GoName] = field.Pos

		// Validate optional/required consistency
		if !field.Optional {
			hasRequired := false
			if gtTag, ok := field.Tags["gt"]; ok {
				if contains(gtTag, "required") {
					hasRequired = true
				}
			}

			hasOmitEmpty := false
			if jsonTag, ok := field.Tags["json"]; ok {
				if contains(jsonTag, "omitempty") {
					hasOmitEmpty = true
				}
			}

			if !hasRequired && hasOmitEmpty {
				return &ValidationError{
					Field:   field.GoName,
					Message: "field has omitempty but is not optional (add required tag to override, or remove omitempty)",
					Index:   &i,
					Pos:     field.Pos,
				}
			}
		}
	}

	return nil
}

// ValidationError represents a schema validation error.
type ValidationError struct {
	Field   string
	Message string
	Index   *int
	Pos     SourcePos
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	prefix := "schema validation error"
	if e.Pos.File != "" {
		prefix += " at " + e.Pos.String()
	}

	if e.Index != nil {
		return prefix + ": " + e.Field + "[" + intToString(*e.Index) + "]: " + e.Message
	}
	return prefix + ": " + e.Field + ": " + e.Message
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf) - 1
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	if neg {
		buf[i] = '-'
		i--
	}
	return string(buf[i+1:])
}
