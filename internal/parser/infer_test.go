// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package parser_test

import (
	"testing"

	"github.com/blackwell-systems/goldenthread/internal/load"
	"github.com/blackwell-systems/goldenthread/internal/parser"
	"github.com/blackwell-systems/goldenthread/internal/schema"
)

// topologyFixture mirrors bide's plan.Topology: a package of structs that carry
// only standard json: tags, with no gt: tags anywhere. It exercises nested
// struct references, slices of structs, slices of scalars, ints, bools, and
// ,omitempty optionality.
const topologyFixture = `package test

type Topology struct {
	Flow  string ` + "`json:\"flow\"`" + `
	Entry string ` + "`json:\"entry\"`" + `
	In    string ` + "`json:\"in\"`" + `
	Out   string ` + "`json:\"out\"`" + `

	Nodes []TopologyNode ` + "`json:\"nodes\"`" + `
	Edges []TopologyEdge ` + "`json:\"edges\"`" + `

	Branches []TopologyBranch ` + "`json:\"branches,omitempty\"`" + `
	Joins    []TopologyJoin   ` + "`json:\"joins,omitempty\"`" + `
	Loops    []TopologyLoop   ` + "`json:\"loops,omitempty\"`" + `
}

type TopologyNode struct {
	Name string ` + "`json:\"name\"`" + `
	Kind string ` + "`json:\"kind\"`" + `
	In   string ` + "`json:\"in\"`" + `
	Out  string ` + "`json:\"out\"`" + `
}

type TopologyEdge struct {
	From string ` + "`json:\"from\"`" + `
	To   string ` + "`json:\"to\"`" + `
	Type string ` + "`json:\"type,omitempty\"`" + `
}

type TopologyBranch struct {
	Over string             ` + "`json:\"over\"`" + `
	Arms []TopologyBranchArm ` + "`json:\"arms\"`" + `
	Else string             ` + "`json:\"else,omitempty\"`" + `
}

type TopologyBranchArm struct {
	Target   string ` + "`json:\"target\"`" + `
	Else     bool   ` + "`json:\"else,omitempty\"`" + `
	LoopBack bool   ` + "`json:\"loopBack,omitempty\"`" + `
	LoopMax  int    ` + "`json:\"loopMax,omitempty\"`" + `
}

type TopologyJoin struct {
	Name    string   ` + "`json:\"name\"`" + `
	Inputs  []string ` + "`json:\"inputs\"`" + `
	InTypes []string ` + "`json:\"inTypes\"`" + `
	Out     string   ` + "`json:\"out\"`" + `
}

type TopologyLoop struct {
	Head string   ` + "`json:\"head\"`" + `
	Over string   ` + "`json:\"over\"`" + `
	Body []string ` + "`json:\"body\"`" + `
	Max  int      ` + "`json:\"max\"`" + `
}
`

// TestInferJSON_Topology verifies that with InferJSON on, a package of
// json-only structs (bide's plan.Topology shape) produces a full set of
// schemas: every struct is generated, every field is present, ,omitempty maps
// to optional, []string becomes an array of strings, []Struct becomes an array
// of named references, ints map to number, and bools map to boolean.
func TestInferJSON_Topology(t *testing.T) {
	tmpDir := setupTestModule(t, topologyFixture)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	p.InferJSON = true
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	byName := make(map[string]*schema.Schema)
	for _, s := range schemas {
		byName[s.Name] = s
	}

	// All 7 structs must be generated.
	wantStructs := []string{
		"Topology", "TopologyNode", "TopologyEdge", "TopologyBranch",
		"TopologyBranchArm", "TopologyJoin", "TopologyLoop",
	}
	if len(schemas) != len(wantStructs) {
		t.Fatalf("expected %d schemas, got %d (%v)", len(wantStructs), len(schemas), schemaNames(schemas))
	}
	for _, name := range wantStructs {
		if byName[name] == nil {
			t.Errorf("missing schema %q (got %v)", name, schemaNames(schemas))
		}
	}

	// Topology: field count and per-field shape.
	topo := byName["Topology"]
	if topo == nil {
		t.Fatal("Topology schema not found")
	}
	if len(topo.Fields) != 9 {
		t.Fatalf("Topology: expected 9 fields, got %d", len(topo.Fields))
	}

	fields := fieldsByJSONName(topo)

	// flow: required string.
	assertField(t, fields, "flow", schema.TypeString, false)

	// nodes: required array of a named reference (TopologyNode).
	nodes := requireField(t, fields, "nodes")
	if nodes.Optional {
		t.Error("Topology.nodes: expected required")
	}
	if nodes.Type.Kind != schema.TypeArray || nodes.Type.Elem == nil {
		t.Fatalf("Topology.nodes: expected array, got %v", nodes.Type.Kind)
	}
	if nodes.Type.Elem.Kind != schema.TypeNamed || nodes.Type.Elem.Ref == nil ||
		nodes.Type.Elem.Ref.Name != "TopologyNode" {
		t.Errorf("Topology.nodes: expected []TopologyNode reference, got %+v", nodes.Type.Elem)
	}

	// branches: optional (omitempty) array of TopologyBranch references.
	branches := requireField(t, fields, "branches")
	if !branches.Optional {
		t.Error("Topology.branches: expected optional (omitempty)")
	}
	if branches.Type.Kind != schema.TypeArray || branches.Type.Elem == nil ||
		branches.Type.Elem.Kind != schema.TypeNamed ||
		branches.Type.Elem.Ref == nil || branches.Type.Elem.Ref.Name != "TopologyBranch" {
		t.Errorf("Topology.branches: expected []TopologyBranch reference, got %+v", branches.Type)
	}

	// TopologyJoin: []string fields map to arrays of strings.
	join := byName["TopologyJoin"]
	if join == nil {
		t.Fatal("TopologyJoin schema not found")
	}
	joinFields := fieldsByJSONName(join)
	inputs := requireField(t, joinFields, "inputs")
	if inputs.Type.Kind != schema.TypeArray || inputs.Type.Elem == nil ||
		inputs.Type.Elem.Kind != schema.TypeString {
		t.Errorf("TopologyJoin.inputs: expected []string, got %+v", inputs.Type)
	}
	if inputs.Optional {
		t.Error("TopologyJoin.inputs: expected required")
	}

	// TopologyBranchArm: ints map to number, bools map to boolean, omitempty is optional.
	arm := byName["TopologyBranchArm"]
	if arm == nil {
		t.Fatal("TopologyBranchArm schema not found")
	}
	armFields := fieldsByJSONName(arm)
	// target: required string.
	assertField(t, armFields, "target", schema.TypeString, false)
	// else: bool, optional via omitempty.
	elseField := requireField(t, armFields, "else")
	if elseField.Type.Kind != schema.TypeBool {
		t.Errorf("TopologyBranchArm.else: expected bool, got %v", elseField.Type.Kind)
	}
	if !elseField.Optional {
		t.Error("TopologyBranchArm.else: expected optional (omitempty)")
	}
	// loopBack: bool, optional.
	loopBack := requireField(t, armFields, "loopBack")
	if loopBack.Type.Kind != schema.TypeBool {
		t.Errorf("TopologyBranchArm.loopBack: expected bool, got %v", loopBack.Type.Kind)
	}
	// loopMax: int, optional.
	loopMax := requireField(t, armFields, "loopMax")
	if loopMax.Type.Kind != schema.TypeInt {
		t.Errorf("TopologyBranchArm.loopMax: expected int, got %v", loopMax.Type.Kind)
	}
	if !loopMax.Optional {
		t.Error("TopologyBranchArm.loopMax: expected optional (omitempty)")
	}

	// TopologyLoop.max is a required int.
	loop := byName["TopologyLoop"]
	if loop == nil {
		t.Fatal("TopologyLoop schema not found")
	}
	loopFields := fieldsByJSONName(loop)
	assertField(t, loopFields, "max", schema.TypeInt, false)
}

// TestInferJSON_DisabledProducesNothing verifies the default is preserved:
// without inference, the same json-only package produces zero schemas.
func TestInferJSON_DisabledProducesNothing(t *testing.T) {
	tmpDir := setupTestModule(t, topologyFixture)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser() // InferJSON defaults to false
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 0 {
		t.Errorf("without inference, expected 0 schemas, got %d (%v)", len(schemas), schemaNames(schemas))
	}
}

// TestInferJSON_ExcludesJSONDash verifies that a field tagged json:"-" is
// dropped and unexported gt-less fields are ignored under inference.
func TestInferJSON_ExcludesJSONDash(t *testing.T) {
	code := `package test

type Row struct {
	ID     string ` + "`json:\"id\"`" + `
	Secret string ` + "`json:\"-\"`" + `
	hidden string
}
`
	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	p.InferJSON = true
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(schemas))
	}
	s := schemas[0]
	if len(s.Fields) != 1 {
		t.Fatalf("expected 1 field (id), got %d: %+v", len(s.Fields), s.Fields)
	}
	if s.Fields[0].JSONName != "id" {
		t.Errorf("expected only field 'id', got %q", s.Fields[0].JSONName)
	}
}

// TestInferJSON_GTTakesPrecedence verifies that a struct mixing gt-tagged and
// json-only fields keeps gt behavior for the gt field and infers the rest.
func TestInferJSON_GTTakesPrecedence(t *testing.T) {
	code := `package test

type Mixed struct {
	Name  string ` + "`json:\"name\" gt:\"required,len:3..20\"`" + `
	Notes string ` + "`json:\"notes,omitempty\"`" + `
}
`
	tmpDir := setupTestModule(t, code)

	pkgs, err := load.LoadPackagesWithDir(tmpDir, ".")
	if err != nil {
		t.Fatalf("LoadPackages() error = %v", err)
	}

	p := parser.NewParser()
	p.InferJSON = true
	schemas, err := p.ParsePackages(pkgs)
	if err != nil {
		t.Fatalf("ParsePackages() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(schemas))
	}
	fields := fieldsByJSONName(schemas[0])

	// gt field keeps its rules.
	name := requireField(t, fields, "name")
	if name.Rules.MinLength == nil || *name.Rules.MinLength != 3 {
		t.Error("Mixed.name: expected gt len rule (min 3) preserved")
	}
	if name.Optional {
		t.Error("Mixed.name: expected required from gt tag")
	}

	// json-only field is inferred as optional (omitempty).
	notes := requireField(t, fields, "notes")
	if !notes.Optional {
		t.Error("Mixed.notes: expected optional (omitempty)")
	}
	if notes.Type.Kind != schema.TypeString {
		t.Errorf("Mixed.notes: expected string, got %v", notes.Type.Kind)
	}
}

// --- helpers ---

func schemaNames(schemas []*schema.Schema) []string {
	names := make([]string, 0, len(schemas))
	for _, s := range schemas {
		names = append(names, s.Name)
	}
	return names
}

func fieldsByJSONName(s *schema.Schema) map[string]schema.Field {
	out := make(map[string]schema.Field, len(s.Fields))
	for _, f := range s.Fields {
		out[f.JSONName] = f
	}
	return out
}

func requireField(t *testing.T, fields map[string]schema.Field, jsonName string) schema.Field {
	t.Helper()
	f, ok := fields[jsonName]
	if !ok {
		t.Fatalf("missing field %q", jsonName)
	}
	return f
}

func assertField(t *testing.T, fields map[string]schema.Field, jsonName string, kind schema.TypeKind, optional bool) {
	t.Helper()
	f := requireField(t, fields, jsonName)
	if f.Type.Kind != kind {
		t.Errorf("field %q: expected kind %v, got %v", jsonName, kind, f.Type.Kind)
	}
	if f.Optional != optional {
		t.Errorf("field %q: expected optional=%v, got %v", jsonName, optional, f.Optional)
	}
}
