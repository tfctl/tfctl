// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

// Package filters provides filtering capabilities for Terraform state resources.
//
// The package parses filter expressions to select subsets of resources based on
// attribute values. Filters are specified as key-operator-target expressions and
// can be combined using a configurable delimiter (default: comma).
//
// Operators include:
//
//   - = : exact match (supports negation with !=)
//   - ^ : prefix match (supports negation with !^)
//   - ~ : regex match (supports negation with !~)
//   - < : less than (numeric comparison)
//   - > : greater than (numeric comparison)
//   - @ : contains substring (supports negation with !@)
//   - / : JSON path match (supports negation with !/)
//
// Examples:
//
//   - "name=my-resource" : matches resources where name equals "my-resource"
//   - "type^aws_" : matches resources where type starts with "aws_"
//   - "tags~^env-" : matches resources where tags contains values matching "^env-"
//   - "count>5" : matches resources where count is greater than 5
//   - "name!@test" : matches resources where name does not contain "test"
//
// Filter Keys and Attributes:
//
// Filter keys are matched against the OutputKey of attributes (see attrs package).
// Keys prefixed with underscore (_) are reserved for Terraform API native filters
// and are silently ignored by this package.
//
// Filter Parsing:
//
// The BuildFilters function parses a comma-delimited (or custom-delimited) filter
// specification string. Invalid specifications (unsupported operands or malformed
// expressions) are logged as warnings and skipped, allowing partial filter sets
// to be processed.
//
// Filter Application:
//
// The Apply function filters a list of candidate resources, keeping only those
// that match all provided filter expressions. Attributes specified in the attrs
// parameter are used to determine which fields from the resource are included
// in the filtered result.
package filters
