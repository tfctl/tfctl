// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package svutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/go-tfe"
)

// Resolve takes a collection of StateVersions plus a spec and returns the
// StateVersions that match the specs. The spec can be in any of the formats
// shown below. The list of StateVersions is in descending serial order, which
// effectively makes it most recent first.
func Resolve(versions []*tfe.StateVersion, specs ...string) ([]*tfe.StateVersion, error) {
	var result = []*tfe.StateVersion{}

	// specs is going to be zero or more (almost certainly max=2) SV specs.  A
	// spec could be -
	//   empty  - the CSV.
	//   sv-id  - the SV with that ID.
	//   CSV~1  - the -1 SV.
	//   serial - the specific serial number.
	//   url    - the SV URL to download.
	//   file   - the SV file to read.

	// Short ciruit if no spec was provided and return the most recent.
	if len(specs) == 0 {
		specs = []string{"CSV~0"}
	}

	// Process each spec and resolve to a StateVersion.
	for _, spec := range specs {
		sv, err := resolveSpec(spec, versions)
		if err != nil {
			return nil, err
		}
		result = append(result, sv)
	}

	return result, nil
}

// resolveSpec takes a single spec string and returns the matching
// StateVersion. Specs can be:
//   - CSV~N: relative index (negative means recent)
//   - numeric serial: find SV with that serial number
//   - file path: read from local file
//   - ID prefix: find first SV matching that ID prefix
func resolveSpec(spec string, versions []*tfe.StateVersion) (*tfe.StateVersion, error) {
	switch {
	case strings.HasPrefix(strings.ToUpper(spec), "CSV~"):
		return resolveCSVSpec(spec, versions)

	case isNumeric(spec):
		return resolveNumericSpec(spec, versions)

	case isFilePath(spec):
		return resolveFileSpec(spec)

	default:
		return resolveIDSpec(spec, versions)
	}
}

// resolveCSVSpec handles CSV~N format specs.
func resolveCSVSpec(spec string, versions []*tfe.StateVersion) (*tfe.StateVersion, error) {
	parts := strings.Split(spec, "~")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid CSV spec format: %s", spec)
	}

	index, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid CSV index: %s", parts[1])
	}

	if index < 0 || index > len(versions)-1 {
		return nil, fmt.Errorf("index %d out of range for versions of length %d", index, len(versions))
	}

	return versions[index], nil
}

// resolveNumericSpec handles numeric serial or relative index specs.
func resolveNumericSpec(spec string, versions []*tfe.StateVersion) (*tfe.StateVersion, error) {
	i, _ := strconv.Atoi(spec)

	if i <= 0 {
		// <= 0 means it's a relative index into the version list
		index := -i
		if index > len(versions)-1 {
			return nil, fmt.Errorf("index %d out of range for versions of length %d", index, len(versions))
		}
		return versions[index], nil
	}

	// Otherwise it's a state serial number that we have to find.
	for _, v := range versions {
		if v.Serial == int64(i) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("failed to find state version with serial %d", i)
}

// resolveFileSpec handles file path specs.
func resolveFileSpec(spec string) (*tfe.StateVersion, error) {
	return &tfe.StateVersion{
		ID:              spec,
		Serial:          0,
		JSONDownloadURL: spec,
	}, nil
}

// resolveIDSpec handles state version ID prefix specs.
func resolveIDSpec(spec string, versions []*tfe.StateVersion) (*tfe.StateVersion, error) {
	for _, v := range versions {
		if strings.HasPrefix(v.ID, spec) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("failed to find state version with ID prefix: %s", spec)
}

// isNumeric checks if a string is a numeric value.
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// isFilePath checks if a string is a valid file path.
func isFilePath(s string) bool {
	_, err := os.Stat(s)
	return err == nil && !os.IsNotExist(err)
}
