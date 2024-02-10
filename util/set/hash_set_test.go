// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package set

import (
	"slices"
	"testing"
)

type Int int

func (i Int) Equal(j Int) bool { return i == j }
func (i Int) Hash() uint64     { return uint64(i) }

func TestHashSet(t *testing.T) {
	s := NewHashSet[Int]()
	s.Add(1)
	s.Add(2)
	if !s.Contains(1) {
		t.Error("missing 1")
	}
	if !s.Contains(2) {
		t.Error("missing 2")
	}
	if s.Contains(3) {
		t.Error("shouldn't have 3")
	}
	if s.Len() != 2 {
		t.Errorf("wrong len %d; want 2", s.Len())
	}

	more := []Int{3, 4}
	s.AddSlice(more)
	if !s.Contains(3) {
		t.Error("missing 3")
	}
	if !s.Contains(4) {
		t.Error("missing 4")
	}
	if s.Contains(5) {
		t.Error("shouldn't have 5")
	}
	if s.Len() != 4 {
		t.Errorf("wrong len %d; want 4", s.Len())
	}

	es := s.Slice()
	if len(es) != 4 {
		t.Errorf("slice has wrong len %d; want 4", len(es))
	}
	for _, e := range []Int{1, 2, 3, 4} {
		if !slices.Contains(es, e) {
			t.Errorf("slice missing %d (%#v)", e, es)
		}
	}
}

func TestHashSetEqual(t *testing.T) {
	type test struct {
		name     string
		a        *HashSet[Int]
		b        *HashSet[Int]
		expected bool
	}
	tests := []test{
		{
			"equal",
			NewHashSetOf([]Int{1, 2, 3, 4}),
			NewHashSetOf([]Int{1, 2, 3, 4}),
			true,
		},
		{
			"not equal",
			NewHashSetOf([]Int{1, 2, 3, 4}),
			NewHashSetOf([]Int{1, 2, 3, 5}),
			false,
		},
		{
			"different lengths",
			NewHashSetOf([]Int{1, 2, 3, 4, 5}),
			NewHashSetOf([]Int{1, 2, 3, 5}),
			false,
		},
	}

	for _, tt := range tests {
		if tt.a.Equal(tt.b) != tt.expected {
			t.Errorf("%s: failed", tt.name)
		}
	}
}

func TestHashSetClone(t *testing.T) {
	s := NewHashSetOf[Int]([]Int{1, 2, 3, 4, 4, 1})
	if s.Len() != 4 {
		t.Errorf("wrong len %d; want 4", s.Len())
	}
	s2 := s.Clone()
	if !s.Equal(s2) {
		t.Error("clone not equal to original")
	}
	s.Add(100)
	if s.Equal(s2) {
		t.Error("clone is not distinct from original")
	}
}
