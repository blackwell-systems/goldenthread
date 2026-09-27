// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package parser extracts schema definitions from Go source code using AST analysis.
package parser

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"github.com/blackwell-systems/goldenthread/internal/load"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// Parser extracts schemas from Go source files.
type Parser struct {
	// TagName is the struct tag to look for (default: "gt")
	TagName string

	// FallbackTags are alternative tags to check (e.g., "validate")
	FallbackTags []string

	// TypeInfo provides go/types information for proper type resolution (optional)
	TypeInfo *load.TypeInfo
}

// NewParser creates a new parser with default settings.
func NewParser() *Parser {
	return &Parser{
		TagName:      "gt",
		FallbackTags: []string{"validate"},
	}
}

// ParsePackages parses schemas from loaded packages with full type information.
func (p *Parser) ParsePackages(pkgs []*load.Package) ([]*schema.Schema, error) {
	var allSchemas []*schema.Schema

	for _, pkg := range pkgs {
		// Set type info for this package
		p.TypeInfo = pkg.GetTypeInfo()

		// Parse each file in the package
		for i, file := range pkg.Pkg.Syntax {
			// Use GoFiles if available, otherwise use a generic path
			filePath := ""
			if i < len(pkg.Pkg.GoFiles) {
				filePath = pkg.Pkg.GoFiles[i]
			} else if len(pkg.Pkg.CompiledGoFiles) > 0 {
				filePath = pkg.Pkg.CompiledGoFiles[0] // Fallback
			} else {
				filePath = pkg.Pkg.PkgPath // Last resort
			}

			schemas, err := p.extractSchemasInternal(pkg.Fset, file, filePath, pkg.Pkg.PkgPath)
			if err != nil {
				return nil, err
			}
			allSchemas = append(allSchemas, schemas...)
		}
	}

	return allSchemas, nil
}

// extractSchemasInternal walks the AST and extracts schema definitions.
func (p *Parser) extractSchemasInternal(fset *token.FileSet, file *ast.File, path string, pkgPath string) ([]*schema.Schema, error) {
	var schemas []*schema.Schema
	var extractErr error

	ast.Inspect(file, func(n ast.Node) bool {
		// Look for type declarations
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		// Only process struct types
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Check if any field has gt: tags
		hasGTTags := false
		for _, field := range structType.Fields.List {
			if field.Tag != nil && p.hasRelevantTag(field.Tag.Value) {
				hasGTTags = true
				break
			}
		}

		if !hasGTTags {
			return true
		}

		// Extract schema
		s, err := p.extractSchema(fset, typeSpec, structType, pkgPath, path)
		if err != nil {
			extractErr = err
			return false
		}
		if s != nil {
			schemas = append(schemas, s)
		}

		return true
	})

	if extractErr != nil {
		return nil, extractErr
	}

	return schemas, nil
}

// extractSchema converts an AST struct type to a Schema.
// Returns (nil, error) if a field or the discriminated-union shape is invalid.
func (p *Parser) extractSchema(fset *token.FileSet, typeSpec *ast.TypeSpec, structType *ast.StructType, pkg, path string) (*schema.Schema, error) {
	s := &schema.Schema{
		Name:        typeSpec.Name.Name,
		PackageName: pkg,
		Pos: schema.SourcePos{
			File:   path,
			Line:   fset.Position(typeSpec.Pos()).Line,
			Column: fset.Position(typeSpec.Pos()).Column,
		},
	}

	// Extract documentation
	if typeSpec.Doc != nil {
		s.Documentation = normalizeDoc(typeSpec.Doc.Text())
	}

	// Extract fields
	for _, field := range structType.Fields.List {
		// Handle embedded fields (no names)
		if len(field.Names) == 0 {
			// Embedded field
			f, err := p.extractEmbeddedField(fset, field, path)
			if err != nil {
				return nil, err
			}
			if f != nil {
				s.Fields = append(s.Fields, *f)
			}
		} else {
			// Regular fields
			for _, name := range field.Names {
				f, err := p.extractField(fset, field, name.Name, path)
				if err != nil {
					return nil, err
				}
				if f != nil {
					s.Fields = append(s.Fields, *f)
				}
			}
		}
	}

	// Assemble discriminated-union metadata from discriminator/variant fields.
	if err := p.buildDiscriminatedUnion(s); err != nil {
		return nil, err
	}

	return s, nil
}

// buildDiscriminatedUnion inspects a schema's fields for discriminator and
// variant markers and, when present, populates s.Discriminator. It enforces
// the shape: exactly one gt:"discriminator" field, one or more gt:"variant:<name>"
// fields, and unique variant names.
func (p *Parser) buildDiscriminatedUnion(s *schema.Schema) error {
	var discriminator *schema.Field
	var discriminatorIdx int
	var variants []schema.Field

	for i := range s.Fields {
		f := &s.Fields[i]
		if f.Rules.IsDiscriminator {
			if discriminator != nil {
				return &schema.ValidationError{
					Field:   f.GoName,
					Message: "multiple discriminator fields (previously defined at " + discriminator.Pos.String() + ")",
					Pos:     f.Pos,
				}
			}
			discriminator = f
			discriminatorIdx = i
		}
		if f.Rules.Variant != "" {
			variants = append(variants, *f)
		}
	}

	// No discriminated-union markers: plain struct.
	if discriminator == nil && len(variants) == 0 {
		return nil
	}

	if discriminator == nil {
		return &schema.ValidationError{
			Field:   s.Name,
			Message: "variant fields present but no discriminator field (add a gt:\"discriminator\" field)",
			Pos:     s.Pos,
		}
	}

	if len(variants) == 0 {
		return &schema.ValidationError{
			Field:   discriminator.GoName,
			Message: "discriminator field present but no variant fields (add gt:\"variant:<name>\" fields)",
			Pos:     discriminator.Pos,
		}
	}

	discName := s.Fields[discriminatorIdx].JSONName
	if discName == "" {
		discName = camelCaseForJSON(discriminator.GoName)
	}

	du := &schema.DiscriminatedUnion{
		DiscriminatorName: discName,
	}

	seen := make(map[string]schema.SourcePos)
	for i := range variants {
		v := variants[i]
		if prevPos, exists := seen[v.Rules.Variant]; exists {
			return &schema.ValidationError{
				Field:   v.GoName,
				Message: "duplicate variant value " + v.Rules.Variant + " (previously defined at " + prevPos.String() + ")",
				Pos:     v.Pos,
			}
		}
		seen[v.Rules.Variant] = v.Pos

		payload := v
		du.Variants = append(du.Variants, schema.Variant{
			Value:        v.Rules.Variant,
			PayloadField: &payload,
		})
	}

	s.Discriminator = du
	return nil
}

// camelCaseForJSON derives a default JSON name from a Go field name when no
// json tag is present, matching the emitter's fallback behavior.
func camelCaseForJSON(goName string) string {
	if goName == "" {
		return ""
	}
	runes := []rune(goName)
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	return string(runes)
}

// extractField converts an AST field to a schema Field.
// Returns (nil, nil) if field should be skipped (no tags).
// Returns (nil, error) if field has invalid configuration.
func (p *Parser) extractField(fset *token.FileSet, field *ast.Field, name string, path string) (*schema.Field, error) {
	if field.Tag == nil {
		return nil, nil
	}

	tagValue := field.Tag.Value
	tags := p.parseTags(tagValue)

	// Check for gt: tag
	gtTag, hasGT := tags[p.TagName]
	if !hasGT {
		// Check fallback tags
		for _, fallback := range p.FallbackTags {
			if _, hasFallback := tags[fallback]; hasFallback {
				gtTag = tags[fallback]
				hasGT = true
				break
			}
		}
	}

	if !hasGT {
		return nil, nil
	}

	pos := fset.Position(field.Pos())
	f := &schema.Field{
		GoName:   name,
		JSONName: p.extractJSONName(tags),
		Tags:     tags,
		Pos: schema.SourcePos{
			File:   path,
			Line:   pos.Line,
			Column: pos.Column,
		},
	}

	// Parse validation rules from tag with conflict detection
	var parseErr error
	f.Rules, parseErr = p.parseRulesWithValidation(gtTag, field.Type)
	if parseErr != nil {
		if valErr, ok := parseErr.(*schema.ValidationError); ok {
			valErr.Field = name
			valErr.Pos = f.Pos
		}
		return nil, parseErr
	}

	// Determine optional semantics with conflict detection
	isPtr := p.isPointer(field.Type)
	omitEmpty := p.hasOmitEmpty(tags)
	hasOptionalTag := p.containsToken(gtTag, "optional")
	hasRequiredTag := p.containsToken(gtTag, "required")

	// Conflict: both required and optional
	if hasRequiredTag && hasOptionalTag {
		return nil, &schema.ValidationError{
			Field:   name,
			Message: "field has both required and optional tags",
			Pos:     f.Pos,
		}
	}

	// Precedence:
	// 1. required → Optional = false
	// 2. optional → Optional = true
	// 3. pointer → Optional = true
	// 4. omitempty → Optional = true
	// 5. default → Optional = false
	if hasRequiredTag {
		f.Optional = false
	} else if hasOptionalTag {
		f.Optional = true
	} else if isPtr {
		f.Optional = true
	} else if omitEmpty {
		f.Optional = true
	} else {
		f.Optional = false
	}

	// Extract type information using go/types
	f.Type = p.extractTypeWithInfo(field.Type, p.TypeInfo)

	// Extract documentation (prefer Doc, fallback to Comment)
	if field.Doc != nil {
		f.Documentation = normalizeDoc(field.Doc.Text())
	} else if field.Comment != nil {
		f.Documentation = normalizeDoc(field.Comment.Text())
	}

	return f, nil
}

// extractEmbeddedField handles embedded struct fields.
func (p *Parser) extractEmbeddedField(fset *token.FileSet, field *ast.Field, path string) (*schema.Field, error) {
	// Extract type to determine the embedded type name
	pos := fset.Position(field.Pos())

	// Get the type using go/types
	fieldType := p.extractTypeWithInfo(field.Type, p.TypeInfo)

	// For embedded fields, we mark them specially
	// The field name will be the type name
	typeName := ""
	var embeddedRef *schema.TypeRef

	if fieldType.Kind == schema.TypeNamed && fieldType.Ref != nil {
		typeName = fieldType.Ref.Name
		embeddedRef = fieldType.Ref
	} else {
		// Can't determine embedded type name, skip
		return nil, nil
	}

	f := &schema.Field{
		GoName:   typeName,
		JSONName: "", // Embedded fields don't have JSON names (fields are promoted)
		Type:     fieldType,
		Optional: false, // Embedded structs themselves are not optional
		Rules:    schema.FieldRules{},
		Tags:     make(map[string]string),
		Pos: schema.SourcePos{
			File:   path,
			Line:   pos.Line,
			Column: pos.Column,
		},
		Embedded:     true,
		EmbeddedType: embeddedRef,
	}

	// Extract documentation
	if field.Doc != nil {
		f.Documentation = normalizeDoc(field.Doc.Text())
	} else if field.Comment != nil {
		f.Documentation = normalizeDoc(field.Comment.Text())
	}

	return f, nil
}

// extractType converts an AST type expression to Type.
func (p *Parser) extractType(expr ast.Expr) schema.Type {
	switch t := expr.(type) {
	case *ast.Ident:
		return p.identToFieldType(t.Name)
	case *ast.StarExpr:
		// Pointer type - unwrap and extract the underlying type
		return p.extractType(t.X)
	case *ast.SelectorExpr:
		// Handle package.Type (e.g., time.Time)
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				return schema.Type{Kind: schema.TypeTime}
			}
			// Other package-qualified types are named types
			return schema.Type{
				Kind: schema.TypeNamed,
				Ref: &schema.TypeRef{
					PackageQualifier: ident.Name,
					Name:             t.Sel.Name,
				},
			}
		}
		return schema.Type{Kind: schema.TypeAny}
	case *ast.ArrayType:
		return schema.Type{
			Kind: schema.TypeArray,
			Elem: &[]schema.Type{p.extractType(t.Elt)}[0],
		}
	case *ast.MapType:
		return schema.Type{
			Kind:  schema.TypeMap,
			Key:   &[]schema.Type{p.extractType(t.Key)}[0],
			Value: &[]schema.Type{p.extractType(t.Value)}[0],
		}
	default:
		return schema.Type{Kind: schema.TypeAny}
	}
}

// identToFieldType maps Go type names to TypeKind.
func (p *Parser) identToFieldType(name string) schema.Type {
	switch name {
	case "string":
		return schema.Type{Kind: schema.TypeString}
	case "int", "int8", "int16", "int32", "int64":
		return schema.Type{Kind: schema.TypeInt}
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return schema.Type{Kind: schema.TypeUint}
	case "float32", "float64":
		return schema.Type{Kind: schema.TypeFloat}
	case "bool":
		return schema.Type{Kind: schema.TypeBool}
	default:
		// Unknown identifiers are named types (e.g., UserID, Email, custom types)
		return schema.Type{
			Kind: schema.TypeNamed,
			Ref: &schema.TypeRef{
				PackageQualifier: "", // Local package, will be resolved with go/packages
				Name:             name,
			},
		}
	}
}

// parseTags extracts struct tags using reflect.StructTag for correct parsing.
func (p *Parser) parseTags(tagLit string) map[string]string {
	tagLit = strings.Trim(tagLit, "`")
	tag := reflect.StructTag(tagLit)

	out := make(map[string]string)

	// Extract only the tags we care about
	if v, ok := tag.Lookup("json"); ok {
		out["json"] = v
	}
	if v, ok := tag.Lookup(p.TagName); ok {
		out[p.TagName] = v
	}
	for _, fallback := range p.FallbackTags {
		if v, ok := tag.Lookup(fallback); ok {
			out[fallback] = v
		}
	}

	return out
}

// extractJSONName gets the JSON field name from tags.
func (p *Parser) extractJSONName(tags map[string]string) string {
	jsonTag, ok := tags["json"]
	if !ok {
		return ""
	}

	// Handle "json:\"name,omitempty\""
	parts := strings.Split(jsonTag, ",")
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

// tagToken represents a parsed token from a tag value.
type tagToken struct {
	value           string // The full token ("min:3" or "required")
	isKeyValue      bool   // True if this is a key:value token
	key             string // The key part ("min")
	valueAfterColon string // The value part after colon ("3" or "a,b,c")
}

// parseTokens splits a tag value into tokens, respecting key:value boundaries.
// This allows values to contain commas (e.g., "enum:a,b,c").
func (p *Parser) parseTokens(tagValue string) []tagToken {
	// Known token names that could appear as flags
	knownFlags := map[string]bool{
		"required": true, "optional": true,
		"email": true, "uuid": true, "url": true,
		"date": true, "datetime": true, "ipv4": true, "ipv6": true,
	}

	var tokens []tagToken
	var current strings.Builder
	var currentKey string
	inKeyValue := false

	for i := 0; i < len(tagValue); i++ {
		ch := tagValue[i]

		switch ch {
		case ':':
			// Start of value in key:value pair
			if !inKeyValue {
				// Extract the key before the colon
				currentKey = current.String()
				inKeyValue = true
			}
			current.WriteByte(ch)
		case ',':
			if !inKeyValue {
				// Regular token separator
				if current.Len() > 0 {
					tokens = append(tokens, p.makeToken(current.String()))
					current.Reset()
				}
			} else {
				// Check if this key allows comma-containing values
				// Both "enum" and "pattern" need special handling
				if currentKey == "pattern" {
					// For pattern, commas are ALWAYS part of the regex (e.g., {3,10})
					// Only end pattern if what follows is a new token (has colon)
					nextIsKey := false
					if i+1 < len(tagValue) {
						remaining := strings.TrimSpace(tagValue[i+1:])
						if remaining != "" && strings.Contains(remaining, ":") {
							colonIdx := strings.Index(remaining, ":")
							commaIdx := strings.Index(remaining, ",")
							// If colon comes before next comma (or no comma), it's a new key:value
							if commaIdx == -1 || colonIdx < commaIdx {
								nextIsKey = true
							}
						}
					}

					if nextIsKey {
						// End the pattern token
						if current.Len() > 0 {
							tokens = append(tokens, p.makeToken(current.String()))
							current.Reset()
							currentKey = ""
						}
						inKeyValue = false
					} else {
						// Comma is part of pattern value
						current.WriteByte(ch)
					}
				} else if currentKey == "enum" {
					// For enum, a comma ends the value ONLY if what follows is clearly a new token
					// New token = has a colon (key:value) OR has comma after it (flag,...)
					nextIsKey := false
					if i+1 < len(tagValue) {
						remaining := strings.TrimSpace(tagValue[i+1:])
						if remaining != "" {
							colonIdx := strings.Index(remaining, ":")
							commaIdx := strings.Index(remaining, ",")

							if colonIdx != -1 && (commaIdx == -1 || colonIdx < commaIdx) {
								// Colon before any comma = next is key:value
								nextIsKey = true
							} else if commaIdx != -1 {
								// Has a comma = could be "enumval,enumval" or "enumval,flag" or "enumval,key:val"
								if colonIdx != -1 && colonIdx > commaIdx {
									// Colon comes AFTER the first comma = the part before comma could be a token
									// Check what's before the comma
									wordBeforeComma := strings.TrimSpace(remaining[:commaIdx])
									if knownFlags[wordBeforeComma] {
										// Known flag before comma
										nextIsKey = true
									} else {
										// Not a known flag, assume it's an enum value
										nextIsKey = false
									}
								} else if colonIdx != -1 && colonIdx < commaIdx {
									// Colon comes BEFORE the first comma = would have been caught above
									// This shouldn't happen but handle it
									nextIsKey = true
								} else {
									// No colon after comma = check if what's before comma is a known flag
									wordBeforeComma := strings.TrimSpace(remaining[:commaIdx])
									if knownFlags[wordBeforeComma] {
										// It's a known flag token, end enum here
										nextIsKey = true
									} else {
										// Not a known flag = treat as enum value
										nextIsKey = false
									}
								}
							} else {
								// No comma, no colon = check if it's a known flag
								word := strings.TrimSpace(remaining)
								if knownFlags[word] {
									// Known flag token after enum
									nextIsKey = true
								}
								// else: unknown word with no comma/colon = last enum value
							}
						}
					}

					if nextIsKey {
						// End the enum token
						if current.Len() > 0 {
							tokens = append(tokens, p.makeToken(current.String()))
							current.Reset()
							currentKey = ""
						}
						inKeyValue = false
					} else {
						// Comma is part of enum values
						current.WriteByte(ch)
					}
				} else {
					// For other keys, comma ends the value
					if current.Len() > 0 {
						tokens = append(tokens, p.makeToken(current.String()))
						current.Reset()
						currentKey = ""
					}
					inKeyValue = false
				}
			}
		default:
			current.WriteByte(ch)
		}
	}

	// Add final token
	if current.Len() > 0 {
		tokens = append(tokens, p.makeToken(current.String()))
	}

	return tokens
}

// makeToken creates a tagToken from a raw token string.
func (p *Parser) makeToken(raw string) tagToken {
	idx := strings.Index(raw, ":")
	if idx == -1 {
		return tagToken{
			value:      raw,
			isKeyValue: false,
		}
	}

	return tagToken{
		value:           raw,
		isKeyValue:      true,
		key:             raw[:idx],
		valueAfterColon: raw[idx+1:],
	}
}

// parseRulesWithValidation extracts validation rules with conflict detection.
func (p *Parser) parseRulesWithValidation(tagValue string, fieldType ast.Expr) (schema.FieldRules, error) {
	rules := schema.FieldRules{}

	// Determine field type kind for validation
	extractedType := p.extractType(fieldType)
	isString := extractedType.Kind == schema.TypeString
	isNumeric := extractedType.Kind == schema.TypeInt ||
		extractedType.Kind == schema.TypeUint ||
		extractedType.Kind == schema.TypeFloat
	isArray := extractedType.Kind == schema.TypeArray

	var formatCount int
	knownTokens := map[string]bool{
		"required": true, "optional": true,
		"min": true, "max": true, "len": true, "pattern": true,
		"email": true, "uuid": true, "url": true,
		"date": true, "datetime": true, "ipv4": true, "ipv6": true,
		"enum": true,
		// Discriminated union markers
		"discriminator": true, "variant": true,
	}

	// Parse tokens - need to handle key:value where value contains commas
	tokens := p.parseTokens(tagValue)

	for _, token := range tokens {
		token.value = strings.TrimSpace(token.value)
		if token.value == "" {
			continue
		}

		// Handle key:value rules
		if token.isKeyValue {
			// Check if key is known
			if !knownTokens[token.key] {
				return rules, &schema.ValidationError{
					Message: "unknown tag token: " + token.key,
				}
			}

			if err := p.applyRuleWithValidation(&rules, token.key, token.valueAfterColon, isString, isNumeric, isArray); err != nil {
				return rules, err
			}
		} else {
			// Handle boolean flags
			tokenValue := token.value

			// Skip presence flags (handled elsewhere)
			if tokenValue == "required" || tokenValue == "optional" {
				continue
			}

			// Check if token is known
			if !knownTokens[tokenValue] {
				return rules, &schema.ValidationError{
					Message: "unknown tag token: " + tokenValue,
				}
			}

			if err := p.applyFlagWithValidation(&rules, tokenValue, isString, &formatCount); err != nil {
				return rules, err
			}
		}
	}

	// Validate min/max relationship
	if rules.Min != nil && rules.Max != nil && *rules.Min > *rules.Max {
		return rules, &schema.ValidationError{
			Message: "min value greater than max value",
		}
	}

	return rules, nil
}

// applyRuleWithValidation applies a key:value rule with type checking.
func (p *Parser) applyRuleWithValidation(rules *schema.FieldRules, key, value string, isString, isNumeric, isArray bool) error {
	switch key {
	case "min":
		if !isNumeric && !isArray {
			return &schema.ValidationError{
				Message: "min rule only applies to numeric or array types",
			}
		}
		if isArray {
			i := parseInt(value)
			if i == nil {
				return &schema.ValidationError{
					Message: "invalid min value: " + value,
				}
			}
			rules.MinItems = i
		} else {
			f := parseFloat(value)
			if f == nil {
				return &schema.ValidationError{
					Message: "invalid min value: " + value,
				}
			}
			rules.Min = f
		}
	case "max":
		if !isNumeric && !isArray {
			return &schema.ValidationError{
				Message: "max rule only applies to numeric or array types",
			}
		}
		if isArray {
			i := parseInt(value)
			if i == nil {
				return &schema.ValidationError{
					Message: "invalid max value: " + value,
				}
			}
			rules.MaxItems = i
		} else {
			f := parseFloat(value)
			if f == nil {
				return &schema.ValidationError{
					Message: "invalid max value: " + value,
				}
			}
			rules.Max = f
		}
	case "len":
		if !isString {
			return &schema.ValidationError{
				Message: "len rule only applies to string types",
			}
		}
		// Parse range like "3..20"
		idx := strings.Index(value, "..")
		if idx == -1 {
			return &schema.ValidationError{
				Message: "len rule must be in format M..N: " + value,
			}
		}
		minStr := value[:idx]
		maxStr := value[idx+2:]

		min := parseInt(minStr)
		if min == nil {
			return &schema.ValidationError{
				Message: "invalid len min value: " + minStr,
			}
		}

		max := parseInt(maxStr)
		if max == nil {
			return &schema.ValidationError{
				Message: "invalid len max value: " + maxStr,
			}
		}

		if *min > *max {
			return &schema.ValidationError{
				Message: "len min greater than max",
			}
		}

		rules.MinLength = min
		rules.MaxLength = max
	case "pattern":
		if !isString {
			return &schema.ValidationError{
				Message: "pattern rule only applies to string types",
			}
		}
		rules.Pattern = &value
	case "enum":
		if !isString {
			return &schema.ValidationError{
				Message: "enum rule only applies to string types",
			}
		}
		// Parse comma-separated enum values
		// Note: values with commas not currently supported
		enumValues := strings.Split(value, ",")
		for i, v := range enumValues {
			enumValues[i] = strings.TrimSpace(v)
		}
		// Filter out empty values
		var filtered []string
		for _, v := range enumValues {
			if v != "" {
				filtered = append(filtered, v)
			}
		}
		if len(filtered) == 0 {
			return &schema.ValidationError{
				Message: "enum must have at least one value",
			}
		}
		rules.Enum = filtered
	case "variant":
		name := strings.TrimSpace(value)
		if name == "" {
			return &schema.ValidationError{
				Message: "variant must name a discriminator value",
			}
		}
		rules.Variant = name
	}
	return nil
}

// applyFlagWithValidation applies a boolean flag with conflict checking.
func (p *Parser) applyFlagWithValidation(rules *schema.FieldRules, flag string, isString bool, formatCount *int) error {
	// Discriminator marker (string-only): identifies the tag field of a
	// discriminated union.
	if flag == "discriminator" {
		if !isString {
			return &schema.ValidationError{
				Message: "discriminator only applies to string types",
			}
		}
		rules.IsDiscriminator = true
		return nil
	}

	format := schema.Format(flag)

	// Check if it's a format flag
	switch format {
	case schema.FormatEmail, schema.FormatUUID, schema.FormatURL,
		schema.FormatDate, schema.FormatDateTime,
		schema.FormatIPv4, schema.FormatIPv6:

		if !isString {
			return &schema.ValidationError{
				Message: "format " + flag + " only applies to string types",
			}
		}

		*formatCount = *formatCount + 1
		if *formatCount > 1 {
			return &schema.ValidationError{
				Message: "multiple format constraints on single field",
			}
		}

		rules.Format = &format
	}

	return nil
}

// containsToken checks if a tag value contains a specific token.
func (p *Parser) containsToken(tagValue, token string) bool {
	parts := strings.Split(tagValue, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == token {
			return true
		}
	}
	return false
}

// isPointer checks if a field type is a pointer.
func (p *Parser) isPointer(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

// hasOmitEmpty checks if the json tag contains omitempty.
func (p *Parser) hasOmitEmpty(tags map[string]string) bool {
	jsonTag, ok := tags["json"]
	if !ok {
		return false
	}
	parts := strings.Split(jsonTag, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == "omitempty" {
			return true
		}
	}
	return false
}

// hasRelevantTag checks if a tag string contains relevant tags.
func (p *Parser) hasRelevantTag(tagString string) bool {
	tags := p.parseTags(tagString)

	if _, ok := tags[p.TagName]; ok {
		return true
	}

	for _, fallback := range p.FallbackTags {
		if _, ok := tags[fallback]; ok {
			return true
		}
	}

	return false
}

// Helper functions

// normalizeDoc normalizes documentation strings for consistent output.
func normalizeDoc(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Text() from ast already removes // and /* markers
	// Just normalize whitespace
	return s
}

func parseInt(s string) *int {
	var i int
	_, err := parseIntValue(s, &i)
	if err != nil {
		return nil
	}
	return &i
}

func parseFloat(s string) *float64 {
	var f float64
	_, err := parseFloatValue(s, &f)
	if err != nil {
		return nil
	}
	return &f
}

func parseIntValue(s string, i *int) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &schema.ValidationError{Message: "invalid integer"}
		}
		n = n*10 + int(c-'0')
	}
	*i = n
	return n, nil
}

func parseFloatValue(s string, f *float64) (float64, error) {
	// Simplified float parsing
	var n float64
	var decimal bool
	var divisor float64 = 1

	for _, c := range s {
		if c == '.' {
			decimal = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, &schema.ValidationError{Message: "invalid float"}
		}

		digit := float64(c - '0')
		if decimal {
			divisor *= 10
			n += digit / divisor
		} else {
			n = n*10 + digit
		}
	}

	*f = n
	return n, nil
}
