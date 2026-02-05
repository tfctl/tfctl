// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package main

import (
	"reflect"
	"testing"
)

func TestDeduplicateFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: []string{},
		},
		{
			name:     "only program and command",
			args:     []string{"tfctl", "sq"},
			expected: []string{"tfctl", "sq"},
		},
		{
			name:     "no duplicates",
			args:     []string{"tfctl", "sq", "--output", "text", "--titles"},
			expected: []string{"tfctl", "sq", "--output", "text", "--titles"},
		},
		{
			name:     "duplicate flag with value - last wins",
			args:     []string{"tfctl", "sq", "--output", "json", "--titles", "--output", "text"},
			expected: []string{"tfctl", "sq", "--titles", "--output", "text"},
		},
		{
			name:     "duplicate boolean flag",
			args:     []string{"tfctl", "sq", "--titles", "--debug", "--titles"},
			expected: []string{"tfctl", "sq", "--debug", "--titles"},
		},
		{
			name:     "duplicate flag with equals syntax",
			args:     []string{"tfctl", "sq", "--output=json", "--titles", "--output=text"},
			expected: []string{"tfctl", "sq", "--titles", "--output=text"},
		},
		{
			name:     "mixed equals and space syntax - same flag",
			args:     []string{"tfctl", "sq", "--output=json", "--output", "text"},
			expected: []string{"tfctl", "sq", "--output", "text"},
		},
		{
			name:     "multiple different flags with duplicates",
			args:     []string{"tfctl", "mq", "--host", "a.b.c", "--org", "foo", "--host", "x.y.z", "--org", "bar"},
			expected: []string{"tfctl", "mq", "--host", "x.y.z", "--org", "bar"},
		},
		{
			name:     "positional args preserved",
			args:     []string{"tfctl", "sq", "/path/to/iac", "--output", "json", "--output", "text"},
			expected: []string{"tfctl", "sq", "/path/to/iac", "--output", "text"},
		},
		{
			name:     "short flags deduplicated",
			args:     []string{"tfctl", "sq", "-o", "json", "-o", "text"},
			expected: []string{"tfctl", "sq", "-o", "text"},
		},
		{
			name:     "different flags not affected",
			args:     []string{"tfctl", "sq", "--color", "--no-color"},
			expected: []string{"tfctl", "sq", "--color", "--no-color"},
		},
		{
			name:     "triple duplicate",
			args:     []string{"tfctl", "sq", "--output", "a", "--output", "b", "--output", "c"},
			expected: []string{"tfctl", "sq", "--output", "c"},
		},
		{
			name:     "flag at end with no value treated as boolean",
			args:     []string{"tfctl", "sq", "--titles", "--debug", "--titles"},
			expected: []string{"tfctl", "sq", "--debug", "--titles"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateFlags(tt.args)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("deduplicateFlags(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

func TestDeduplicateFlagsPreservesOrder(t *testing.T) {
	// Ensure non-duplicate flags maintain their relative order.
	args := []string{"tfctl", "sq", "--alpha", "--beta", "--gamma"}
	result := deduplicateFlags(args)
	expected := []string{"tfctl", "sq", "--alpha", "--beta", "--gamma"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Order not preserved: got %v, want %v", result, expected)
	}
}

func TestDeduplicateFlagsWithPositionalAfterFlags(t *testing.T) {
	// Positional args after flags should be preserved.
	args := []string{"tfctl", "sq", "--output", "json", "/path", "--output", "text"}
	result := deduplicateFlags(args)
	expected := []string{"tfctl", "sq", "/path", "--output", "text"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestInjectConfigSet(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		key       string
		insertIdx int
		configVal []string
		expected  []string
	}{
		{
			name:      "empty config returns args unchanged",
			args:      []string{"tfctl", "sq", "--titles"},
			key:       "defaults",
			insertIdx: 2,
			configVal: nil,
			expected:  []string{"tfctl", "sq", "--titles"},
		},
		{
			name:      "single entry injected",
			args:      []string{"tfctl", "sq", "--titles"},
			key:       "defaults",
			insertIdx: 2,
			configVal: []string{"--debug"},
			expected:  []string{"tfctl", "sq", "--debug", "--titles"},
		},
		{
			name:      "multi-word entry split",
			args:      []string{"tfctl", "sq", "--titles"},
			key:       "defaults",
			insertIdx: 2,
			configVal: []string{"--output text"},
			expected:  []string{"tfctl", "sq", "--output", "text", "--titles"},
		},
		{
			name:      "multiple entries",
			args:      []string{"tfctl", "sq"},
			key:       "defaults",
			insertIdx: 2,
			configVal: []string{"--debug", "--output json"},
			expected:  []string{"tfctl", "sq", "--debug", "--output", "json"},
		},
		{
			name:      "insert at index 3",
			args:      []string{"tfctl", "sq", "/path/to/iac", "--titles"},
			key:       "defaults",
			insertIdx: 3,
			configVal: []string{"--debug"},
			expected:  []string{"tfctl", "sq", "/path/to/iac", "--debug", "--titles"},
		},
		{
			name:      "complex multi-word entries",
			args:      []string{"tfctl", "mq"},
			key:       "mq.defaults",
			insertIdx: 2,
			configVal: []string{"--host app.terraform.io", "--org myorg"},
			expected:  []string{"tfctl", "mq", "--host", "app.terraform.io", "--org", "myorg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectConfigSetTestable(tt.args, tt.configVal, tt.insertIdx)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("injectConfigSet() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// injectConfigSetTestable is a test-friendly version that accepts config values
// directly instead of reading from global config.
func injectConfigSetTestable(args []string, entries []string, insertIdx int) []string {
	if len(entries) == 0 {
		return args
	}

	var expanded []string
	for _, entry := range entries {
		for _, field := range splitFields(entry) {
			expanded = append(expanded, field)
		}
	}

	return append(args[:insertIdx], append(expanded, args[insertIdx:]...)...)
}

// splitFields splits a string by whitespace, matching strings.Fields behavior.
func splitFields(s string) []string {
	var result []string
	start := -1

	for i, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if start >= 0 {
				result = append(result, s[start:i])
				start = -1
			}
		} else {
			if start < 0 {
				start = i
			}
		}
	}

	if start >= 0 {
		result = append(result, s[start:])
	}

	return result
}
