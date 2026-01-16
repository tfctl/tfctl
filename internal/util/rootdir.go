// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"path/filepath"
	"strings"
)

// ParseRootDir parses a RootDir string and returns the absolute directory and
// any optional environment override. It returns an error if the fs entry does
// not exist, is empty or is not a directory.
func ParseRootDir(rootDir string) (string, string, error) {

	if rootDir == "" {
		return "", "", os.ErrInvalid
	}

	var dir, env string

	// First, split the path to see if there is an ::env override.
	parts := strings.Split(rootDir, "::")
	if len(parts) > 1 {
		env = parts[1]
	}

	// Now determine if the actual root directory (parts[0]) is absolute or
	// relative. If it is relative, make it absolute.
	if !strings.HasPrefix(parts[0], "/") {
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}
		dir = filepath.Join(cwd, parts[0])
	} else {
		dir = parts[0]
	}

	// If the rootDir is not a directory, return an error.
	if r, err := os.Stat(dir); err != nil {
		return "", "", err
	} else if !r.IsDir() {
		return "", "", os.ErrInvalid
	}

	return dir, env, nil
}
