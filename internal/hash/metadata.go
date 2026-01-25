// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package hash

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// Metadata stores hashes of generated schemas for drift detection.
type Metadata struct {
	// Version of goldenthread used to generate
	Version string `json:"version"`
	
	// Schemas maps schema name to its hash
	Schemas map[string]SchemaMetadata `json:"schemas"`
}

// SchemaMetadata stores metadata for a single schema.
type SchemaMetadata struct {
	// Hash of the schema content
	Hash string `json:"hash"`
	
	// SourceFile is the Go source file
	SourceFile string `json:"source_file"`
	
	// OutputFile is the generated output file
	OutputFile string `json:"output_file"`
}

const metadataFilename = ".goldenthread.json"

// WriteMetadata writes metadata to the output directory.
func WriteMetadata(outDir string, schemas []*schema.Schema, version string) error {
	md := Metadata{
		Version: version,
		Schemas: make(map[string]SchemaMetadata),
	}
	
	for _, s := range schemas {
		hash := ComputeSchemaHash(s)
		md.Schemas[s.Name] = SchemaMetadata{
			Hash:       hash,
			SourceFile: s.Pos.File,
			OutputFile: toKebabCase(s.Name) + ".ts",
		}
	}
	
	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}
	
	metadataPath := filepath.Join(outDir, metadataFilename)
	return os.WriteFile(metadataPath, data, 0644)
}

// ReadMetadata reads metadata from the output directory.
func ReadMetadata(outDir string) (*Metadata, error) {
	metadataPath := filepath.Join(outDir, metadataFilename)
	
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}
	
	var md Metadata
	if err := json.Unmarshal(data, &md); err != nil {
		return nil, err
	}
	
	return &md, nil
}

// toKebabCase converts PascalCase to kebab-case.
func toKebabCase(s string) string {
	if s == "" {
		return ""
	}
	
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	
	return string(result)
}
