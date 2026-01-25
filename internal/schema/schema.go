// Package schema defines the intermediate representation (IR) for domain schemas.
// This IR is language-agnostic and serves as the bridge between Go AST parsing
// and code generation for target languages (Zod, TypeScript, OpenAPI).
package schema

// Schema represents a complete domain model extracted from a Go struct.
type Schema struct {
	// Name is the struct name (e.g., "User")
	Name string

	// Package is the Go package name
	Package string

	// Fields are the struct fields with their validation rules
	Fields []Field

	// Rules are schema-level validation rules (cross-field constraints)
	Rules []Rule

	// Documentation is the comment above the struct
	Documentation string

	// Location tracks where this schema was defined
	Location SourceLocation
}

// Field represents a single struct field with validation rules.
type Field struct {
	// Name is the Go field name (e.g., "Username")
	Name string

	// JSONName is the JSON tag name (e.g., "username")
	JSONName string

	// Type is the field's type information
	Type FieldType

	// Required indicates if this field must be present
	Required bool

	// Rules are field-specific validation rules
	Rules FieldRules

	// Documentation is the comment above the field
	Documentation string

	// Tags preserves original Go struct tags
	Tags map[string]string
}

// FieldType describes a field's type structure.
type FieldType struct {
	// Kind is the base type category
	Kind TypeKind

	// Element is the type for array elements (when Kind == TypeArray)
	Element *FieldType

	// Properties are nested fields (when Kind == TypeObject)
	Properties []Field

	// KeyType and ValueType for maps (when Kind == TypeMap)
	KeyType   *FieldType
	ValueType *FieldType
}

// TypeKind categorizes field types.
type TypeKind int

const (
	// TypeString represents string types
	TypeString TypeKind = iota

	// TypeInt represents integer types (int, int8, int16, int32, int64)
	TypeInt

	// TypeUint represents unsigned integer types
	TypeUint

	// TypeFloat represents floating-point types (float32, float64)
	TypeFloat

	// TypeBool represents boolean type
	TypeBool

	// TypeArray represents slice/array types
	TypeArray

	// TypeObject represents struct/object types
	TypeObject

	// TypeMap represents map types
	TypeMap

	// TypeTime represents time.Time
	TypeTime

	// TypeUUID represents UUID types
	TypeUUID

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
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	case TypeMap:
		return "map"
	case TypeTime:
		return "time"
	case TypeUUID:
		return "uuid"
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

	// Custom validators (function names)
	CustomValidators []string
}

// Format represents standardized string formats.
type Format string

const (
	// FormatEmail represents email addresses
	FormatEmail Format = "email"

	// FormatUUID represents UUID strings
	FormatUUID Format = "uuid"

	// FormatURL represents URLs
	FormatURL Format = "url"

	// FormatDate represents ISO 8601 dates
	FormatDate Format = "date"

	// FormatDateTime represents ISO 8601 date-times
	FormatDateTime Format = "datetime"

	// FormatIPv4 represents IPv4 addresses
	FormatIPv4 Format = "ipv4"

	// FormatIPv6 represents IPv6 addresses
	FormatIPv6 Format = "ipv6"
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

// SourceLocation tracks where a schema was defined in source code.
type SourceLocation struct {
	// File is the absolute path to the Go file
	File string

	// Line is the line number where the struct starts
	Line int

	// Column is the column number
	Column int
}

// Validate checks if a Schema is semantically valid.
func (s *Schema) Validate() error {
	if s.Name == "" {
		return &ValidationError{Field: "Name", Message: "schema name cannot be empty"}
	}

	for i, field := range s.Fields {
		if field.Name == "" {
			return &ValidationError{
				Field:   "Fields",
				Message: "field name cannot be empty",
				Index:   &i,
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
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Index != nil {
		return "schema validation error: " + e.Field + "[" + string(rune(*e.Index)) + "]: " + e.Message
	}
	return "schema validation error: " + e.Field + ": " + e.Message
}
