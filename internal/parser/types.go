// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package parser

import (
	"go/ast"
	"go/types"

	"github.com/blackwell-systems/goldenthread/internal/load"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// extractTypeWithInfo extracts type information using go/types for proper resolution.
// Requires TypeInfo from go/packages - parser always uses ParsePackages which provides this.
func (p *Parser) extractTypeWithInfo(expr ast.Expr, typeInfo *load.TypeInfo) schema.Type {
	if typeInfo == nil || typeInfo.Info == nil {
		// TypeInfo should always be present when using ParsePackages
		// Fall back to best-effort AST parsing for robustness
		return p.extractType(expr)
	}

	// Get the type from go/types
	tv, ok := typeInfo.Info.Types[expr]
	if !ok {
		// Type info not found for this expression - use AST parsing
		return p.extractType(expr)
	}

	return p.convertGoType(tv.Type, typeInfo)
}

// convertGoType converts a go/types.Type to our schema.Type.
func (p *Parser) convertGoType(t types.Type, typeInfo *load.TypeInfo) schema.Type {
	// Unwrap pointer types
	if ptr, ok := t.(*types.Pointer); ok {
		return p.convertGoType(ptr.Elem(), typeInfo)
	}

	switch t := t.(type) {
	case *types.Basic:
		return p.convertBasicType(t)

	case *types.Named:
		return p.convertNamedType(t)

	case *types.Slice:
		elemType := p.convertGoType(t.Elem(), typeInfo)
		return schema.Type{
			Kind: schema.TypeArray,
			Elem: &elemType,
		}

	case *types.Array:
		elemType := p.convertGoType(t.Elem(), typeInfo)
		return schema.Type{
			Kind: schema.TypeArray,
			Elem: &elemType,
		}

	case *types.Map:
		keyType := p.convertGoType(t.Key(), typeInfo)
		valType := p.convertGoType(t.Elem(), typeInfo)
		return schema.Type{
			Kind:  schema.TypeMap,
			Key:   &keyType,
			Value: &valType,
		}

	case *types.Struct:
		// Inline struct - would need to extract fields
		return schema.Type{Kind: schema.TypeObject}

	default:
		return schema.Type{Kind: schema.TypeAny}
	}
}

// convertBasicType converts a basic Go type to our TypeKind.
func (p *Parser) convertBasicType(t *types.Basic) schema.Type {
	switch t.Kind() {
	case types.String:
		return schema.Type{Kind: schema.TypeString}
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		return schema.Type{Kind: schema.TypeInt}
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return schema.Type{Kind: schema.TypeUint}
	case types.Float32, types.Float64:
		return schema.Type{Kind: schema.TypeFloat}
	case types.Bool:
		return schema.Type{Kind: schema.TypeBool}
	default:
		return schema.Type{Kind: schema.TypeAny}
	}
}

// convertNamedType converts a named type (like time.Time, uuid.UUID, custom types).
func (p *Parser) convertNamedType(t *types.Named) schema.Type {
	obj := t.Obj()
	pkg := obj.Pkg()

	// Handle standard library types
	if pkg != nil {
		pkgPath := pkg.Path()
		typeName := obj.Name()

		// Special handling for time.Time
		if pkgPath == "time" && typeName == "Time" {
			return schema.Type{Kind: schema.TypeTime}
		}

		// Return as named type with real package path
		return schema.Type{
			Kind: schema.TypeNamed,
			Ref: &schema.TypeRef{
				PackageQualifier: pkgPath,
				Name:             typeName,
			},
		}
	}

	// Local package type (no import needed)
	return schema.Type{
		Kind: schema.TypeNamed,
		Ref: &schema.TypeRef{
			PackageQualifier: "",
			Name:             obj.Name(),
		},
	}
}
