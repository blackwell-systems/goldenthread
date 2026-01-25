// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package parser extracts schema definitions from Go source code using AST analysis.
package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// Parser extracts schemas from Go source files.
type Parser struct {
	// TagName is the struct tag to look for (default: "gt")
	TagName string

	// FallbackTags are alternative tags to check (e.g., "validate")
	FallbackTags []string
}

// NewParser creates a new parser with default settings.
func NewParser() *Parser {
	return &Parser{
		TagName:      "gt",
		FallbackTags: []string{"validate"},
	}
}

// ParseFile parses a single Go file and extracts all schemas.
func (p *Parser) ParseFile(path string) ([]*schema.Schema, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return p.extractSchemas(fset, file, path)
}

// ParseDir parses all Go files in a directory (non-recursive).
func (p *Parser) ParseDir(dir string) ([]*schema.Schema, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var schemas []*schema.Schema
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		fileSchemas, err := p.ParseFile(path)
		if err != nil {
			return nil, err
		}

		schemas = append(schemas, fileSchemas...)
	}

	return schemas, nil
}

// ParseDirRecursive parses all Go files in a directory tree.
func (p *Parser) ParseDirRecursive(root string) ([]*schema.Schema, error) {
	var schemas []*schema.Schema

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		fileSchemas, err := p.ParseFile(path)
		if err != nil {
			return err
		}

		schemas = append(schemas, fileSchemas...)
		return nil
	})

	return schemas, err
}

// extractSchemas walks the AST and extracts schema definitions.
func (p *Parser) extractSchemas(fset *token.FileSet, file *ast.File, path string) ([]*schema.Schema, error) {
	var schemas []*schema.Schema

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
		s := p.extractSchema(fset, typeSpec, structType, file.Name.Name, path)
		if s != nil {
			schemas = append(schemas, s)
		}

		return true
	})

	return schemas, nil
}

// extractSchema converts an AST struct type to a Schema.
func (p *Parser) extractSchema(fset *token.FileSet, typeSpec *ast.TypeSpec, structType *ast.StructType, pkg, path string) *schema.Schema {
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
		s.Documentation = typeSpec.Doc.Text()
	}

	// Extract fields
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			f := p.extractField(fset, field, name.Name, path)
			if f != nil {
				s.Fields = append(s.Fields, *f)
			}
		}
	}

	return s
}

// extractField converts an AST field to a schema Field.
func (p *Parser) extractField(fset *token.FileSet, field *ast.Field, name string, path string) *schema.Field {
	if field.Tag == nil {
		return nil
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
		return nil
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

	// Parse validation rules from tag
	f.Rules = p.parseRules(gtTag)

	// Determine optional semantics:
	// Default: required
	// Optional if: pointer OR omitempty OR gt:"optional"
	// gt:"required" overrides
	isPtr := p.isPointer(field.Type)
	omitEmpty := p.hasOmitEmpty(tags)
	hasOptionalTag := p.containsToken(gtTag, "optional")
	hasRequiredTag := p.containsToken(gtTag, "required")

	f.Optional = isPtr || omitEmpty || hasOptionalTag
	if hasRequiredTag {
		f.Optional = false
	}

	// Extract type information
	f.Type = p.extractType(field.Type)

	// Extract documentation
	if field.Doc != nil {
		f.Documentation = field.Doc.Text()
	}

	return f
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
					PackagePath: ident.Name,
					Name:        t.Sel.Name,
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
	case "Time":
		return schema.Type{Kind: schema.TypeTime}
	default:
		return schema.Type{Kind: schema.TypeAny}
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

// parseRules extracts validation rules from a tag value.
func (p *Parser) parseRules(tagValue string) schema.FieldRules {
	rules := schema.FieldRules{}

	parts := strings.Split(tagValue, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Handle key:value rules
		if idx := strings.Index(part, ":"); idx != -1 {
			key := part[:idx]
			value := part[idx+1:]
			p.applyRule(&rules, key, value)
		} else {
			// Handle boolean flags (required, email, etc.)
			p.applyFlag(&rules, part)
		}
	}

	return rules
}

// applyRule applies a key:value rule to FieldRules.
func (p *Parser) applyRule(rules *schema.FieldRules, key, value string) {
	switch key {
	case "min":
		if f := parseFloat(value); f != nil {
			rules.Min = f
		}
	case "max":
		if f := parseFloat(value); f != nil {
			rules.Max = f
		}
	case "len":
		// Parse range like "3..20"
		if idx := strings.Index(value, ".."); idx != -1 {
			minStr := value[:idx]
			maxStr := value[idx+2:]
			if min := parseInt(minStr); min != nil {
				rules.MinLength = min
			}
			if max := parseInt(maxStr); max != nil {
				rules.MaxLength = max
			}
		}
	case "pattern":
		rules.Pattern = &value
	}
}

// applyFlag applies a boolean flag to FieldRules.
func (p *Parser) applyFlag(rules *schema.FieldRules, flag string) {
	format := schema.Format(flag)
	switch format {
	case schema.FormatEmail, schema.FormatUUID, schema.FormatURL,
		schema.FormatDate, schema.FormatDateTime,
		schema.FormatIPv4, schema.FormatIPv6:
		rules.Format = &format
	}
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
