// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// goldenthread CLI tool
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackwell-systems/goldenthread/internal/emitter/zod"
	"github.com/blackwell-systems/goldenthread/internal/parser"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		if err := generate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := check(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "init":
		if err := initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("goldenthread v0.1.0")
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usage := `goldenthread - The golden thread of truth through your system

USAGE:
    goldenthread <command> [options]

COMMANDS:
    generate <dir>    Generate schemas from Go source files
    check <dir>       Verify generated schemas are up-to-date
    init              Create goldenthread.yaml configuration
    version           Print version information
    help              Show this help message

EXAMPLES:
    goldenthread generate ./models
    goldenthread generate --target=zod ./api
    goldenthread check ./models

For more information, visit: https://github.com/blackwell-systems/goldenthread
`
	fmt.Print(usage)
}

func generate(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	outDir := fs.String("out", "./gen", "output directory for generated files")
	target := fs.String("target", "zod", "generation target (zod, typescript, openapi)")
	recursive := fs.Bool("recursive", false, "recursively process subdirectories")
	
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: goldenthread generate [options] <directory>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	
	if err := fs.Parse(args); err != nil {
		return err
	}
	
	if fs.NArg() == 0 {
		return fmt.Errorf("generate requires a directory argument")
	}
	
	inputDir := fs.Arg(0)
	
	fmt.Printf("goldenthread - generating schemas\n")
	fmt.Printf("  Source: %s\n", inputDir)
	fmt.Printf("  Output: %s\n", *outDir)
	fmt.Printf("  Target: %s\n", *target)
	fmt.Println()
	
	// Only support zod for v0.1
	if *target != "zod" {
		return fmt.Errorf("only 'zod' target is supported in v0.1")
	}
	
	// Create parser
	p := parser.NewParser()
	
	// Parse schemas
	var schemas []*schema.Schema
	var err error
	
	if *recursive {
		schemas, err = p.ParseDirRecursive(inputDir)
	} else {
		schemas, err = p.ParseDir(inputDir)
	}
	
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %w", err)
	}
	
	if len(schemas) == 0 {
		fmt.Println("No schemas found with gt: tags")
		return nil
	}
	
	fmt.Printf("Found %d schema(s)\n", len(schemas))
	
	// Create output directory
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate schemas
	emitter := zod.NewEmitter()
	successCount := 0
	
	for _, schema := range schemas {
		fmt.Printf("  Generating %s...\n", schema.Name)
		
		// Validate schema
		if err := schema.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: skipping %s: %v\n", schema.Name, err)
			continue
		}
		
		// Generate Zod schema
		output, err := emitter.Emit(schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to emit %s: %v\n", schema.Name, err)
			continue
		}
		
		// Write to file
		outFile := filepath.Join(*outDir, toKebabCase(schema.Name)+".ts")
		if err := os.WriteFile(outFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to write %s: %v\n", outFile, err)
			continue
		}
		
		successCount++
	}
	
	fmt.Printf("\nGenerated %d/%d schemas successfully\n", successCount, len(schemas))
	
	if successCount < len(schemas) {
		return fmt.Errorf("some schemas failed to generate")
	}
	
	return nil
}

func check(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("check requires a directory argument")
	}

	fmt.Println("🧵 goldenthread - checking schemas...")
	fmt.Printf("   Source: %s\n", args[0])
	fmt.Println("   This is a placeholder - implementation coming soon")

	return nil
}

func initialize() error {
	fmt.Println("🧵 goldenthread - initializing configuration...")
	fmt.Println("   This is a placeholder - implementation coming soon")
	return nil
}

// toKebabCase converts PascalCase to kebab-case for filenames.
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
			result = append(result, r+32) // Convert to lowercase
		} else {
			result = append(result, r)
		}
	}
	
	return string(result)
}
