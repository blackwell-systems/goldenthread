// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package hash computes deterministic hashes of schemas for drift detection.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"sort"
	"strings"

	"github.com/blackwell-systems/goldenthread/internal/schema"
)

type hash256 interface {
	hash.Hash
	Sum(b []byte) []byte
}

// ComputeSchemaHash computes a deterministic hash of a schema's content.
// The hash includes:
// - Schema name and package
// - All fields (names, types, rules, optional flags) in sorted order
// - Does NOT include documentation or source positions
func ComputeSchemaHash(s *schema.Schema) string {
	h := sha256.New()

	// Schema name and package
	h.Write([]byte(s.Name))
	h.Write([]byte(s.PackageName))

	// Discriminated-union shape (discriminator name + ordered variant values)
	if s.Discriminator != nil {
		h.Write([]byte("discriminatedUnion"))
		h.Write([]byte(s.Discriminator.DiscriminatorName))
		for _, v := range s.Discriminator.Variants {
			h.Write([]byte(v.Value))
			if v.PayloadField != nil {
				h.Write([]byte(v.PayloadField.GoName))
				writeType(h, v.PayloadField.Type)
			}
		}
	}

	// Sort fields by name for deterministic output
	fields := make([]schema.Field, len(s.Fields))
	copy(fields, s.Fields)
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].GoName < fields[j].GoName
	})

	// Hash each field
	for _, field := range fields {
		h.Write([]byte(field.GoName))
		h.Write([]byte(field.JSONName))

		// Type information
		writeType(h, field.Type)

		// Optional flag
		if field.Optional {
			h.Write([]byte("optional"))
		} else {
			h.Write([]byte("required"))
		}

		// Rules
		writeRules(h, field.Rules)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// writeType writes type information to the hash.
func writeType(h hash256, t schema.Type) {
	h.Write([]byte(t.Kind.String()))

	if t.Ref != nil {
		h.Write([]byte(t.Ref.PackageQualifier))
		h.Write([]byte(t.Ref.Name))
	}

	if t.Elem != nil {
		writeType(h, *t.Elem)
	}

	if t.Key != nil {
		writeType(h, *t.Key)
	}

	if t.Value != nil {
		writeType(h, *t.Value)
	}

	// Hash inline object fields
	for _, field := range t.Fields {
		h.Write([]byte(field.GoName))
		writeType(h, field.Type)
	}
}

// writeRules writes field rules to the hash.
func writeRules(h hash256, rules schema.FieldRules) {
	if rules.MinLength != nil {
		h.Write([]byte(intToString(*rules.MinLength)))
	}
	if rules.MaxLength != nil {
		h.Write([]byte(intToString(*rules.MaxLength)))
	}
	if rules.Pattern != nil {
		h.Write([]byte(*rules.Pattern))
	}
	if rules.Format != nil {
		h.Write([]byte(*rules.Format))
	}
	if rules.Min != nil {
		h.Write([]byte(floatToString(*rules.Min)))
	}
	if rules.Max != nil {
		h.Write([]byte(floatToString(*rules.Max)))
	}
	if rules.MinItems != nil {
		h.Write([]byte(intToString(*rules.MinItems)))
	}
	if rules.MaxItems != nil {
		h.Write([]byte(intToString(*rules.MaxItems)))
	}
	if rules.UniqueItems {
		h.Write([]byte("unique"))
	}
	if len(rules.Enum) > 0 {
		// Sort enum values for deterministic hash
		enumCopy := make([]string, len(rules.Enum))
		copy(enumCopy, rules.Enum)
		sort.Strings(enumCopy)
		h.Write([]byte(strings.Join(enumCopy, ",")))
	}
	if rules.IsDiscriminator {
		h.Write([]byte("discriminator"))
	}
	if rules.Variant != "" {
		h.Write([]byte("variant:" + rules.Variant))
	}
}

// Helper functions for number conversion without stdlib

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

func floatToString(f float64) string {
	// Simple float64 to string for hashing (not for display)
	if f == 0 {
		return "0"
	}

	// Manual bit conversion
	sign := uint64(0)
	if f < 0 {
		sign = 1
		f = -f
	}

	// For hashing purposes, simple representation is fine
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * 1000000)

	return intToString(int(sign)) + ":" + intToString(intPart) + "." + intToString(fracPart)
}
