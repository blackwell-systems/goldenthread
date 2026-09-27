// Copyright 2025 Dayna Blackwell / Blackwell Systems
// SPDX-License-Identifier: Apache-2.0 OR MIT

package models

// EdgeSpec connects one node to another.
type EdgeSpec struct {
	From string `json:"from" gt:"required"`
	To   string `json:"to" gt:"required"`
}

// SwitchCase is one branch of a switch.
type SwitchCase struct {
	Match string `json:"match" gt:"required"`
	Next  string `json:"next" gt:"required"`
}

// SwitchSpec routes to one of several cases based on a value.
type SwitchSpec struct {
	On    string       `json:"on" gt:"required"`
	Cases []SwitchCase `json:"cases" gt:"required,min:1"`
}

// JoinSpec merges several inputs into one.
type JoinSpec struct {
	Inputs []string `json:"inputs" gt:"required,min:2"`
}

// WiringElement is exactly one of edge, switch, or join, selected by kind.
type WiringElement struct {
	// Kind selects the active variant.
	Kind string `json:"kind" gt:"discriminator"`

	// Edge is present when kind == "edge".
	Edge *EdgeSpec `json:"edge,omitempty" gt:"variant:edge"`

	// Switch is present when kind == "switch".
	Switch *SwitchSpec `json:"switch,omitempty" gt:"variant:switch"`

	// Join is present when kind == "join".
	Join *JoinSpec `json:"join,omitempty" gt:"variant:join"`
}
