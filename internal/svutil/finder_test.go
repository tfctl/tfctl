// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package svutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

// makeStateVersions creates a test slice of StateVersions for testing.
func makeStateVersions() []*tfe.StateVersion {
	return []*tfe.StateVersion{
		{
			ID:              "sv-001",
			Serial:          100,
			JSONDownloadURL: "https://example.com/sv-001.json",
		},
		{
			ID:              "sv-002",
			Serial:          101,
			JSONDownloadURL: "https://example.com/sv-002.json",
		},
		{
			ID:              "sv-003",
			Serial:          102,
			JSONDownloadURL: "https://example.com/sv-003.json",
		},
		{
			ID:              "sv-alpha-001",
			Serial:          103,
			JSONDownloadURL: "https://example.com/sv-alpha-001.json",
		},
	}
}

func TestResolve(t *testing.T) {
	versions := makeStateVersions()

	tests := []struct {
		name      string
		versions  []*tfe.StateVersion
		specs     []string
		wantCount int
		wantIDs   []string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "no specs defaults to CSV~0",
			versions:  versions,
			specs:     []string{},
			wantCount: 1,
			wantIDs:   []string{"sv-001"},
			wantErr:   false,
		},
		{
			name:      "single CSV spec",
			versions:  versions,
			specs:     []string{"CSV~0"},
			wantCount: 1,
			wantIDs:   []string{"sv-001"},
			wantErr:   false,
		},
		{
			name:      "multiple CSV specs",
			versions:  versions,
			specs:     []string{"CSV~1", "CSV~3"},
			wantCount: 2,
			wantIDs:   []string{"sv-002", "sv-alpha-001"},
			wantErr:   false,
		},
		{
			name:      "CSV spec with lowercase",
			versions:  versions,
			specs:     []string{"csv~0"},
			wantCount: 1,
			wantIDs:   []string{"sv-001"},
			wantErr:   false,
		},
		{
			name:      "CSV spec with mixed case",
			versions:  versions,
			specs:     []string{"CsV~2"},
			wantCount: 1,
			wantIDs:   []string{"sv-003"},
			wantErr:   false,
		},
		{
			name:      "invalid CSV spec format",
			versions:  versions,
			specs:     []string{"CSV~1~2"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "invalid CSV spec format",
		},
		{
			name:      "CSV spec with non-numeric index",
			versions:  versions,
			specs:     []string{"CSV~abc"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "invalid CSV index",
		},
		{
			name:      "CSV spec index out of range",
			versions:  versions,
			specs:     []string{"CSV~99"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "out of range",
		},
		{
			name:      "serial number lookup",
			versions:  versions,
			specs:     []string{"101"},
			wantCount: 1,
			wantIDs:   []string{"sv-002"},
			wantErr:   false,
		},
		{
			name:      "multiple serial lookups",
			versions:  versions,
			specs:     []string{"100", "102"},
			wantCount: 2,
			wantIDs:   []string{"sv-001", "sv-003"},
			wantErr:   false,
		},
		{
			name:      "serial number not found",
			versions:  versions,
			specs:     []string{"999"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "failed to find state version with serial",
		},
		{
			name:      "ID prefix match",
			versions:  versions,
			specs:     []string{"sv-00"},
			wantCount: 1,
			wantIDs:   []string{"sv-001"},
			wantErr:   false,
		},
		{
			name:      "ID prefix match with longer prefix",
			versions:  versions,
			specs:     []string{"sv-002"},
			wantCount: 1,
			wantIDs:   []string{"sv-002"},
			wantErr:   false,
		},
		{
			name:      "ID prefix match ambiguous",
			versions:  versions,
			specs:     []string{"sv-"},
			wantCount: 1,
			wantIDs:   []string{"sv-001"},
			wantErr:   false,
		},
		{
			name:      "ID prefix not found",
			versions:  versions,
			specs:     []string{"sv-xyz"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "failed to find state version with ID prefix",
		},
		{
			name:      "relative index positive zeros",
			versions:  versions,
			specs:     []string{"0"},
			wantCount: 1,
			wantIDs:   []string{"sv-001"},
			wantErr:   false,
		},
		{
			name:      "relative index negative",
			versions:  versions,
			specs:     []string{"-1"},
			wantCount: 1,
			wantIDs:   []string{"sv-002"},
			wantErr:   false,
		},
		{
			name:      "relative index negative out of range",
			versions:  versions,
			specs:     []string{"-99"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "out of range",
		},
		{
			name:      "empty versions list with CSV spec",
			versions:  []*tfe.StateVersion{},
			specs:     []string{"CSV~0"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "out of range",
		},
		{
			name:      "single version in list",
			versions:  []*tfe.StateVersion{versions[0]},
			specs:     []string{"CSV~0"},
			wantCount: 1,
			wantIDs:   []string{"sv-001"},
			wantErr:   false,
		},
		{
			name:      "single version out of range",
			versions:  []*tfe.StateVersion{versions[0]},
			specs:     []string{"CSV~1"},
			wantCount: 0,
			wantErr:   true,
			errMsg:    "out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.versions, tt.specs...)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err, "unexpected error")
				assert.NotNil(t, got)
				assert.Len(t, got, tt.wantCount, "result count mismatch")
				for i, id := range tt.wantIDs {
					assert.Equal(t, id, got[i].ID, "ID mismatch at index %d", i)
				}
			}
		})
	}
}

func TestResolveCSVSpec(t *testing.T) {
	versions := makeStateVersions()

	tests := []struct {
		name     string
		spec     string
		versions []*tfe.StateVersion
		wantID   string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid index 0",
			spec:     "CSV~0",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "valid index 2",
			spec:     "CSV~2",
			versions: versions,
			wantID:   "sv-003",
			wantErr:  false,
		},
		{
			name:     "index out of range",
			spec:     "CSV~100",
			versions: versions,
			wantErr:  true,
			errMsg:   "out of range",
		},
		{
			name:     "missing tilde",
			spec:     "CSV0",
			versions: versions,
			wantErr:  true,
			errMsg:   "invalid CSV spec format",
		},
		{
			name:     "non-numeric index",
			spec:     "CSV~abc",
			versions: versions,
			wantErr:  true,
			errMsg:   "invalid CSV index",
		},
		{
			name:     "negative index",
			spec:     "CSV~-1",
			versions: versions,
			wantErr:  true,
			errMsg:   "out of range",
		},
		{
			name:     "multiple tildes",
			spec:     "CSV~1~2",
			versions: versions,
			wantErr:  true,
			errMsg:   "invalid CSV spec format",
		},
		{
			name:     "empty versions list",
			spec:     "CSV~0",
			versions: []*tfe.StateVersion{},
			wantErr:  true,
			errMsg:   "out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCSVSpec(tt.spec, tt.versions)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantID, got.ID)
			}
		})
	}
}

func TestResolveNumericSpec(t *testing.T) {
	versions := makeStateVersions()

	tests := []struct {
		name     string
		spec     string
		versions []*tfe.StateVersion
		wantID   string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "zero means index 0",
			spec:     "0",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "negative means relative index",
			spec:     "-1",
			versions: versions,
			wantID:   "sv-002",
			wantErr:  false,
		},
		{
			name:     "negative index out of range",
			spec:     "-99",
			versions: versions,
			wantErr:  true,
			errMsg:   "out of range",
		},
		{
			name:     "positive number is serial lookup",
			spec:     "101",
			versions: versions,
			wantID:   "sv-002",
			wantErr:  false,
		},
		{
			name:     "serial lookup first match",
			spec:     "100",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "serial not found",
			versions: versions,
			spec:     "999",
			wantErr:  true,
			errMsg:   "failed to find state version with serial",
		},
		{
			name:     "negative with empty list",
			spec:     "-1",
			versions: []*tfe.StateVersion{},
			wantErr:  true,
			errMsg:   "out of range",
		},
		{
			name:     "positive serial with empty list",
			spec:     "100",
			versions: []*tfe.StateVersion{},
			wantErr:  true,
			errMsg:   "failed to find state version with serial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveNumericSpec(tt.spec, tt.versions)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantID, got.ID)
			}
		})
	}
}

func TestResolveFileSpec(t *testing.T) {
	// Get absolute path to testdata directory
	absStateFile, err := filepath.Abs(filepath.Join("testdata", "state.json"))
	assert.NoError(t, err)

	tests := []struct {
		name    string
		spec    string
		wantURL string
		wantErr bool
	}{
		{
			name:    "valid file path",
			spec:    absStateFile,
			wantURL: absStateFile,
			wantErr: false,
		},
		{
			name:    "file path becomes ID",
			spec:    absStateFile,
			wantURL: absStateFile,
			wantErr: false,
		},
		{
			name:    "nonexistent file path still accepted",
			spec:    "/nonexistent/file/path.json",
			wantURL: "/nonexistent/file/path.json",
			wantErr: false,
		},
		{
			name:    "relative path to nonexistent still accepted",
			spec:    "testdata/nonexistent.json",
			wantURL: "testdata/nonexistent.json",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveFileSpec(tt.spec)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.spec, got.ID)
				assert.Equal(t, tt.wantURL, got.JSONDownloadURL)
				assert.Equal(t, int64(0), got.Serial)
			}
		})
	}
}

func TestResolveIDSpec(t *testing.T) {
	versions := makeStateVersions()

	tests := []struct {
		name     string
		spec     string
		versions []*tfe.StateVersion
		wantID   string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "exact ID match",
			spec:     "sv-001",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "prefix match",
			spec:     "sv-00",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "partial prefix match",
			spec:     "sv-",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "longer prefix match",
			spec:     "sv-002",
			versions: versions,
			wantID:   "sv-002",
			wantErr:  false,
		},
		{
			name:     "alpha prefix match",
			spec:     "sv-alpha",
			versions: versions,
			wantID:   "sv-alpha-001",
			wantErr:  false,
		},
		{
			name:     "ID not found",
			spec:     "sv-xyz",
			versions: versions,
			wantErr:  true,
			errMsg:   "failed to find state version with ID prefix",
		},
		{
			name:     "single character prefix",
			spec:     "s",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "empty spec",
			spec:     "",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "empty versions list",
			spec:     "sv-001",
			versions: []*tfe.StateVersion{},
			wantErr:  true,
			errMsg:   "failed to find state version with ID prefix",
		},
		{
			name:     "case sensitive match",
			spec:     "SV",
			versions: versions,
			wantErr:  true,
			errMsg:   "failed to find state version with ID prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveIDSpec(tt.spec, tt.versions)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantID, got.ID)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "positive number",
			s:    "123",
			want: true,
		},
		{
			name: "zero",
			s:    "0",
			want: true,
		},
		{
			name: "negative number",
			s:    "-1",
			want: true,
		},
		{
			name: "large number",
			s:    "9999999999",
			want: true,
		},
		{
			name: "non-numeric string",
			s:    "abc",
			want: false,
		},
		{
			name: "alphanumeric",
			s:    "123abc",
			want: false,
		},
		{
			name: "float",
			s:    "123.45",
			want: false,
		},
		{
			name: "empty string",
			s:    "",
			want: false,
		},
		{
			name: "whitespace",
			s:    " 123 ",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNumeric(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsFilePath(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "svutil-test-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "valid file path",
			s:    tmpFile.Name(),
			want: true,
		},
		{
			name: "nonexistent file",
			s:    "/nonexistent/file/path.json",
			want: false,
		},
		{
			name: "empty string",
			s:    "",
			want: false,
		},
		{
			name: "relative path to nonexistent",
			s:    "testdata/nonexistent.json",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFilePath(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveSpec(t *testing.T) {
	versions := makeStateVersions()
	tmpFile, err := os.CreateTemp("", "svutil-resolve-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name     string
		spec     string
		versions []*tfe.StateVersion
		wantID   string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "CSV spec dispatch",
			spec:     "CSV~1",
			versions: versions,
			wantID:   "sv-002",
			wantErr:  false,
		},
		{
			name:     "numeric spec dispatch",
			spec:     "101",
			versions: versions,
			wantID:   "sv-002",
			wantErr:  false,
		},
		{
			name:     "file path spec dispatch",
			spec:     tmpFile.Name(),
			versions: versions,
			wantID:   tmpFile.Name(),
			wantErr:  false,
		},
		{
			name:     "ID prefix spec dispatch",
			spec:     "sv-001",
			versions: versions,
			wantID:   "sv-001",
			wantErr:  false,
		},
		{
			name:     "invalid CSV spec",
			spec:     "CSV~invalid",
			versions: versions,
			wantErr:  true,
			errMsg:   "invalid CSV index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSpec(tt.spec, tt.versions)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantID, got.ID)
			}
		})
	}
}
