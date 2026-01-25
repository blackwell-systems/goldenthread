// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package load uses go/packages to load Go packages with proper type information.
package load

import (
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Package represents a loaded Go package with syntax and type information.
type Package struct {
	// Pkg is the loaded package
	Pkg *packages.Package

	// Fset is the file set for position information
	Fset *token.FileSet
}

// LoadPackages loads Go packages from the given directory patterns.
// Patterns can be:
// - "./models" - single directory
// - "./..." - recursive
// - "github.com/foo/bar" - import path
func LoadPackages(patterns ...string) ([]*Package, error) {
	return LoadPackagesWithDir("", patterns...)
}

// LoadPackagesWithDir loads packages with a specific working directory.
// If dir is empty, uses current directory.
func LoadPackagesWithDir(dir string, patterns ...string) ([]*Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports,
		Tests: false,
		Dir:   dir,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Check for errors in loaded packages
	var errs []error
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				errs = append(errs, e)
			}
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("package loading errors: %v", errs)
	}

	// Convert to our Package type
	result := make([]*Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		result = append(result, &Package{
			Pkg:  pkg,
			Fset: pkg.Fset,
		})
	}

	return result, nil
}

// TypeInfo provides type information for resolving types correctly.
type TypeInfo struct {
	// TypesInfo from go/packages
	Info *types.Info

	// Package for scope resolution
	Pkg *types.Package
}

// GetTypeInfo returns type information for a package.
func (p *Package) GetTypeInfo() *TypeInfo {
	return &TypeInfo{
		Info: p.Pkg.TypesInfo,
		Pkg:  p.Pkg.Types,
	}
}
