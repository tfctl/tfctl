// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package filters

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"

	"github.com/tfctl/tfctl/internal/attrs"
)

//go:embed testdata/*.yaml
var testDataFS embed.FS

// testBuildFiltersCase represents a single test case for TestBuildFilters.
type testBuildFiltersCase struct {
	Name      string   `yaml:"name"`
	Spec      string   `yaml:"spec"`
	Delimiter string   `yaml:"delimiter"`
	Want      []Filter `yaml:"want"`
	WantCount int      `yaml:"wantCount"`
}

// testCheckStringOperandCase represents a single test case for
// TestCheckStringOperand.
type testCheckStringOperandCase struct {
	Name   string `yaml:"name"`
	Value  string `yaml:"value"`
	Filter Filter `yaml:"filter"`
	Want   bool   `yaml:"want"`
}

// testCheckNumericOperandCase represents a single test case for
// TestCheckNumericOperand.
type testCheckNumericOperandCase struct {
	Name   string  `yaml:"name"`
	Value  float64 `yaml:"value"`
	Filter Filter  `yaml:"filter"`
	Want   bool    `yaml:"want"`
}

// testCheckContainsOperandCase represents a single test case for
// TestCheckContainsOperand.
type testCheckContainsOperandCase struct {
	Name   string      `yaml:"name"`
	Value  interface{} `yaml:"value"`
	Filter Filter      `yaml:"filter"`
	Want   bool        `yaml:"want"`
}

// testToFloat64Case represents a single test case for TestToFloat64.
type testToFloat64Case struct {
	Name      string      `yaml:"name"`
	Value     interface{} `yaml:"value"`
	Want      float64     `yaml:"want"`
	WantOk    bool        `yaml:"wantOk"`
	ValueType string      `yaml:"valueType"`
}

// testApplyFiltersCase represents a single test case for TestApplyFilters.
type testApplyFiltersCase struct {
	Name    string   `yaml:"name"`
	Filters []Filter `yaml:"filters"`
	Want    bool     `yaml:"want"`
}

// testFilterDatasetCase represents a single test case for TestFilterDataset.
type testFilterDatasetCase struct {
	Name      string   `yaml:"name"`
	Spec      string   `yaml:"spec"`
	WantCount int      `yaml:"wantCount"`
	WantNames []string `yaml:"wantNames"`
}

// loadTestData loads test data from embedded YAML files.
func loadTestData(filename string, v interface{}) error {
	data, err := testDataFS.ReadFile("testdata/" + filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

func TestBuildFilters(t *testing.T) {
	var tests []testBuildFiltersCase
	require.NoError(t, loadTestData("filters_test_build_filters.yaml", &tests))

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Delimiter != "" {
				t.Setenv("TFCTL_FILTER_DELIM", tt.Delimiter)
			}

			got := BuildFilters(tt.Spec)
			assert.Len(t, got, tt.WantCount)
			if tt.Want != nil {
				for i, filter := range tt.Want {
					assert.Equal(t, filter.Key, got[i].Key)
					assert.Equal(t, filter.Operand, got[i].Operand)
					assert.Equal(t, filter.Value, got[i].Value)
					assert.Equal(t, filter.Negate, got[i].Negate)
				}
			}
		})
	}
}

func TestCheckStringOperand(t *testing.T) {
	var tests []testCheckStringOperandCase
	require.NoError(t, loadTestData("filters_test_check_string_operand.yaml", &tests))

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := checkStringOperand(tt.Value, tt.Filter)
			assert.Equal(t, tt.Want, got)
		})
	}
}

func TestCheckNumericOperand(t *testing.T) {
	var tests []testCheckNumericOperandCase
	require.NoError(t, loadTestData("filters_test_check_numeric_operand.yaml", &tests))

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := checkNumericOperand(tt.Value, tt.Filter)
			assert.Equal(t, tt.Want, got)
		})
	}
}

func TestCheckContainsOperand(t *testing.T) {
	var tests []testCheckContainsOperandCase
	require.NoError(t, loadTestData("filters_test_check_contains_operand.yaml", &tests))

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := checkContainsOperand(tt.Value, tt.Filter)
			assert.Equal(t, tt.Want, got)
		})
	}
}

func TestToFloat64(t *testing.T) {
	var tests []testToFloat64Case
	require.NoError(t, loadTestData("filters_test_to_float64.yaml", &tests))

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, ok := toFloat64(tt.Value)
			assert.Equal(t, tt.WantOk, ok)
			if ok {
				assert.Equal(t, tt.Want, got)
			}
		})
	}
}

func TestApplyFilters(t *testing.T) {
	var tests []testApplyFiltersCase
	require.NoError(t, loadTestData("filters_test_apply_filters.yaml", &tests))

	testData := `
	{
		"id": "res-123",
		"name": "my-resource",
		"type": "aws_instance",
		"region": "us-east-1",
		"count": 5,
		"tags": ["prod", "web"],
		"metadata": {"env": "production"},
		"description": null,
		"nested": {"inner": "value"}
	}
	`

	attrList := attrs.AttrList{
		{Key: "name", OutputKey: "name", Include: true},
		{Key: "type", OutputKey: "type", Include: true},
		{Key: "region", OutputKey: "region", Include: true},
		{Key: "count", OutputKey: "count", Include: true},
		{Key: "description", OutputKey: "description", Include: true},
		{Key: "nested", OutputKey: "nested", Include: true},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := gjson.Parse(testData)
			got := applyFilters(result, attrList, tt.Filters)
			assert.Equal(t, tt.Want, got)
		})
	}
}

func TestFilterDataset(t *testing.T) {
	var tests []testFilterDatasetCase
	require.NoError(t, loadTestData("filters_test_filter_dataset.yaml", &tests))

	testData := `
	[
		{
			"id": "res-1",
			"name": "aws-resource-1",
			"type": "aws_instance"
		},
		{
			"id": "res-2",
			"name": "gcp-resource",
			"type": "google_instance"
		},
		{
			"id": "res-3",
			"name": "aws-resource-2",
			"type": "aws_network"
		}
	]
	`

	attrList := attrs.AttrList{
		{Key: "name", OutputKey: "name", Include: true},
		{Key: "type", OutputKey: "type", Include: true},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			candidates := gjson.Parse(testData)
			got := FilterDataset(candidates, attrList, tt.Spec)
			assert.Len(t, got, tt.WantCount)
			for i, expected := range tt.WantNames {
				assert.Equal(t, expected, got[i]["name"])
			}
		})
	}
}
