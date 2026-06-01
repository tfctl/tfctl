// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package main

import (
	"reflect"
	"testing"

	"github.com/tfctl/tfctl/internal/config"
)

func TestNormalizeCommandAlias(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "summarize becomes ps",
			args:     []string{"tfctl", "summarize", "-"},
			expected: []string{"tfctl", "ps", "-"},
		},
		{
			name:     "ps stays ps",
			args:     []string{"tfctl", "ps", "-"},
			expected: []string{"tfctl", "ps", "-"},
		},
		{
			name:     "empty args unchanged",
			args:     nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCommandAlias(append([]string(nil), tt.args...))

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("normalizeCommandAlias(%v) = %v, want %v", tt.args, got, tt.expected)
			}
		})
	}
}

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
		expanded = append(expanded, splitFields(entry)...)
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

func TestShouldSkipExplicitSetToken(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		tokenIdx int
		want     bool
	}{
		{
			name:     "skip for long attrs flag",
			args:     []string{"tfctl", "sq", "--attrs", "@full"},
			tokenIdx: 3,
			want:     true,
		},
		{
			name:     "skip for short attrs flag",
			args:     []string{"tfctl", "sq", "-a", "@full"},
			tokenIdx: 3,
			want:     true,
		},
		{
			name:     "skip for long filter flag",
			args:     []string{"tfctl", "sq", "--filter", "@pink"},
			tokenIdx: 3,
			want:     true,
		},
		{
			name:     "skip for short filter flag",
			args:     []string{"tfctl", "sq", "-f", "@pink"},
			tokenIdx: 3,
			want:     true,
		},
		{
			name:     "skip for jq flag",
			args:     []string{"tfctl", "sq", "--jq", "@query1"},
			tokenIdx: 3,
			want:     true,
		},
		{
			name:     "do not skip standalone explicit set",
			args:     []string{"tfctl", "sq", "@full"},
			tokenIdx: 2,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipExplicitSetToken(tt.args, tt.tokenIdx)
			if got != tt.want {
				t.Errorf("shouldSkipExplicitSetToken(%v, %d) = %v, want %v", tt.args, tt.tokenIdx, got, tt.want)
			}
		})
	}
}

func TestProcessCommandArgs_PsSkipsExplicitSet(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"ps": map[string]interface{}{
				"full": []interface{}{"--attrs .resource,.action"},
			},
		},
	}

	args := []string{"tfctl", "ps", "@full"}
	got := processCommandArgs(args)
	want := []string{"tfctl", "ps", "-", "@full"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("processCommandArgs(%v) = %v, want %v", args, got, want)
	}
}

func TestProcessCommandArgs_SummarizeMatchesPs(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"ps": map[string]interface{}{
				"defaults": []interface{}{"--output text"},
			},
		},
	}

	psArgs := processCommandArgs([]string{"tfctl", "ps"})
	summarizeArgs := processCommandArgs(normalizeCommandAlias([]string{"tfctl", "summarize"}))

	if !reflect.DeepEqual(summarizeArgs, psArgs) {
		t.Errorf("summarize args = %v, want %v", summarizeArgs, psArgs)
	}
}

func TestExpandPresetSegments(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"attrs": map[string]interface{}{
					"set1": "arn,name",
					"set2": []interface{}{"created-at", "workspace"},
				},
			},
		},
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single preset from string",
			input: "@set1",
			want:  "arn,name",
		},
		{
			name:  "mixed literal and preset",
			input: "created-at,@set1",
			want:  "created-at,arn,name",
		},
		{
			name:  "multiple presets",
			input: "@set1,@set2",
			want:  "arn,name,created-at,workspace",
		},
		{
			name:  "multiple presets with literal",
			input: "@set1,@set2,created-at",
			want:  "arn,name,created-at,workspace,created-at",
		},
		{
			name:  "single preset from slice",
			input: "@set2",
			want:  "created-at,workspace",
		},
		{
			name:  "unknown preset preserved",
			input: "@unknown",
			want:  "@unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPresetSegments(tt.input, "presets.attrs", ",")
			if got != tt.want {
				t.Errorf("expandPresetSegments(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandFlagValuePresets(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"attrs": map[string]interface{}{
					"set1": "arn,name",
					"set2": []interface{}{"created-at", "workspace"},
					"set3": "workspace,",
				},
			},
		},
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "long flag with separate value",
			args: []string{"tfctl", "mq", "--attrs", "@set1"},
			want: []string{"tfctl", "mq", "--attrs", "arn,name"},
		},
		{
			name: "short flag with separate value",
			args: []string{"tfctl", "mq", "-a", "@set1"},
			want: []string{"tfctl", "mq", "-a", "arn,name"},
		},
		{
			name: "long flag equals syntax",
			args: []string{"tfctl", "mq", "--attrs=@set1"},
			want: []string{"tfctl", "mq", "--attrs=arn,name"},
		},
		{
			name: "mixed literal and preset",
			args: []string{"tfctl", "mq", "--attrs", "created-at,@set1"},
			want: []string{"tfctl", "mq", "--attrs", "created-at,arn,name"},
		},
		{
			name: "unknown preset unchanged",
			args: []string{"tfctl", "mq", "--attrs", "@unknown"},
			want: []string{"tfctl", "mq", "--attrs", "@unknown"},
		},
		{
			name: "multiple presets",
			args: []string{"tfctl", "mq", "--attrs", "@set1,@set2"},
			want: []string{"tfctl", "mq", "--attrs", "arn,name,created-at,workspace"},
		},
		{
			name: "multiple presets with literal",
			args: []string{"tfctl", "mq", "--attrs", "@set1,@set2,created-at"},
			want: []string{"tfctl", "mq", "--attrs", "arn,name,created-at,workspace,created-at"},
		},
		{
			name: "multiple presets with transform spec",
			args: []string{"tfctl", "mq", "--attrs", "@set1,name::U"},
			want: []string{"tfctl", "mq", "--attrs", "arn,name,name::U"},
		},
		{
			// The trailing comma should be kept here as the stripping of it doesn't
			// occur until after preset expansion.
			name: "preset with trailing comma maintained",
			args: []string{"tfctl", "mq", "--attrs", "@set3"},
			want: []string{"tfctl", "mq", "--attrs", "workspace,"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandFlagValuePresets(
				append([]string(nil), tt.args...),
				"--attrs",
				"-a",
				"presets.attrs",
				",",
			)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandFlagValuePresets(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestExpandFlagValuePresets_Filter(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"filters": map[string]interface{}{
					"set1": "_organization@prod",
					"set2": []interface{}{"_workspace@dev", "status@applied"},
				},
			},
		},
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "long flag with separate value",
			args: []string{"tfctl", "mq", "--filter", "@set1"},
			want: []string{"tfctl", "mq", "--filter", "_organization@prod"},
		},
		{
			name: "short flag with separate value",
			args: []string{"tfctl", "mq", "-f", "@set1"},
			want: []string{"tfctl", "mq", "-f", "_organization@prod"},
		},
		{
			name: "long flag equals syntax",
			args: []string{"tfctl", "mq", "--filter=@set1"},
			want: []string{"tfctl", "mq", "--filter=_organization@prod"},
		},
		{
			name: "mixed literal and preset",
			args: []string{"tfctl", "mq", "--filter", "name@prod,@set1"},
			want: []string{"tfctl", "mq", "--filter", "name@prod,_organization@prod"},
		},
		{
			name: "multiple presets",
			args: []string{"tfctl", "mq", "--filter", "@set1,@set2"},
			want: []string{"tfctl", "mq", "--filter", "_organization@prod,_workspace@dev,status@applied"},
		},
		{
			name: "unknown preset unchanged",
			args: []string{"tfctl", "mq", "--filter", "@unknown"},
			want: []string{"tfctl", "mq", "--filter", "@unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandFlagValuePresets(
				append([]string(nil), tt.args...),
				"--filter",
				"-f",
				"presets.filters",
				",",
			)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandFlagValuePresets(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestExpandFlagValuePresets_Filter_CustomDelimiter(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	t.Setenv("TFCTL_FILTER_DELIM", "|")

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"filters": map[string]interface{}{
					"set1": "status@applied",
				},
			},
		},
	}

	got := expandFlagValuePresets(
		[]string{"tfctl", "mq", "--filter", "name@prod|@set1"},
		"--filter",
		"-f",
		"presets.filters",
		filterDelimiter(),
	)
	want := []string{"tfctl", "mq", "--filter", "name@prod|status@applied"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("custom delimiter filter expansion got %v, want %v", got, want)
	}
}

func TestProcessCommandArgs_ExpandsAttrsPreset(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"attrs": map[string]interface{}{
					"set1": "arn,name",
				},
			},
		},
	}

	got := processCommandArgs([]string{"tfctl", "mq", "--attrs", "created-at,@set1"})
	want := []string{"tfctl", "mq", "--attrs", "created-at,arn,name"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("processCommandArgs attrs expansion got %v, want %v", got, want)
	}
}

func TestProcessCommandArgs_ExpandsFilterPreset(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"filters": map[string]interface{}{
					"set1": "_organization@prod",
				},
			},
		},
	}

	got := processCommandArgs([]string{"tfctl", "mq", "--filter", "name@prod,@set1"})
	want := []string{"tfctl", "mq", "--filter", "name@prod,_organization@prod"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("processCommandArgs filter expansion got %v, want %v", got, want)
	}
}

func TestProcessCommandArgs_ExpandsAttrsAndFilterPresets_WithRootDir(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"attrs": map[string]interface{}{
					"set1": "region,layer",
				},
				"filters": map[string]interface{}{
					"set2": "layer=myapp",
				},
			},
		},
	}

	rootDir := t.TempDir()

	got := processCommandArgs(
		[]string{"tfctl", "sq", rootDir, "--attrs", "@set1", "--filter", "@set2"},
	)
	want := []string{
		"tfctl",
		"sq",
		rootDir,
		"--attrs",
		"region,layer",
		"--filter",
		"layer=myapp",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("processCommandArgs attrs+filter expansion got %v, want %v", got, want)
	}
}

func TestExpandFlagSingleValuePreset(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"jq": map[string]interface{}{
					"query1": `.updated > "2025-06-01"`,
					"query2": `.name | contains("cloud")`,
				},
			},
		},
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "long flag with separate value",
			args: []string{"tfctl", "wq", "--jq", "@query1"},
			want: []string{"tfctl", "wq", "--jq", `.updated > "2025-06-01"`},
		},
		{
			name: "long flag equals syntax",
			args: []string{"tfctl", "wq", `--jq=@query2`},
			want: []string{"tfctl", "wq", `--jq=.name | contains("cloud")`},
		},
		{
			name: "non preset value unchanged",
			args: []string{"tfctl", "wq", `--jq=.name | contains("cloud")`},
			want: []string{"tfctl", "wq", `--jq=.name | contains("cloud")`},
		},
		{
			name: "unknown preset unchanged",
			args: []string{"tfctl", "wq", "--jq", "@unknown"},
			want: []string{"tfctl", "wq", "--jq", "@unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandFlagSingleValuePreset(
				append([]string(nil), tt.args...),
				"--jq",
				"",
				"presets.jq",
			)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandFlagSingleValuePreset(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestProcessCommandArgs_ExpandsJQPreset(t *testing.T) {
	originalConfig := config.Config
	t.Cleanup(func() {
		config.Config = originalConfig
	})

	config.Config = config.Type{
		Data: map[string]interface{}{
			"presets": map[string]interface{}{
				"jq": map[string]interface{}{
					"query1": `.updated > "2025-06-01"`,
				},
			},
		},
	}

	got := processCommandArgs([]string{"tfctl", "wq", "--jq", "@query1"})
	want := []string{"tfctl", "wq", "--jq", `.updated > "2025-06-01"`}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("processCommandArgs jq expansion got %v, want %v", got, want)
	}
}
