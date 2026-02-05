// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tfctl/tfctl/internal/cacheutil"
	"github.com/tfctl/tfctl/internal/command"
	"github.com/tfctl/tfctl/internal/config"
	"github.com/tfctl/tfctl/internal/log"
	"github.com/tfctl/tfctl/internal/version"
)

var ctx = context.Background()

func main() {
	os.Exit(realMain())
}

func realMain() int {
	log.InitLogger()

	args := os.Args
	log.Debugf("args captured: args=%v", args)

	if handleVersion(args) {
		return 0
	}

	args = handleNakedCommand(args)

	// If --help appears anywhere, skip command processing and let the CLI handle it.
	helpFound := false
	for _, a := range args {
		if a == "--help" || a == "-h" {
			helpFound = true
			break
		}
	}

	if !helpFound {
		args = processCommandArgs(args)
	}

	return initAndRunApp(args)
}

// initAndRunApp initializes the app and runs it, returning the exit code.
func initAndRunApp(args []string) int {
	// Pre-create cache directory when caching is enabled.
	if _, ok, err := cacheutil.EnsureBaseDir(); err != nil && ok {
		fmt.Fprintln(os.Stderr, err)
		log.Debugf("cache ensure err: err=%v", err)
	}

	app, err := command.InitApp(ctx, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Debugf("app init err: err=%v", err)
		return 1
	}

	if err := app.Run(ctx, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Debugf("app run err: err=%v", err)
		return 2
	}

	return 0
}

// handleVersion checks for --version/-v and returns whether it was handled.
func handleVersion(args []string) bool {
	for _, a := range args {
		if a == "--version" || a == "-v" {
			fmt.Println(version.Version)
			return true
		}
	}
	return false
}

// handleNakedCommand appends --help if no command is provided.
func handleNakedCommand(args []string) []string {
	if len(args) <= 1 {
		return append(args, "--help")
	}
	return args
}

// processCommandArgs handles command-specific argument processing.
func processCommandArgs(args []string) []string {
	switch {
	case len(args) > 1 && args[1] == "completion":
		// Short-circuit completion: pass args directly.
		return args
	default:
		insertIdx := 2
		// If args[2] is a directory or file, that means that the IAC root has been
		// specified, so we need to insert after that.
		if len(args) > 2 {
			if isExistingFile(args[2]) {
				insertIdx = 3
			} else if stat, err := os.Stat(args[2]); err == nil && stat.IsDir() {
				insertIdx = 3
			}
		}

		// Inject sets in reverse precedence order (defaults → nostate → global)
		// so that higher-precedence sets appear later and win on duplicates.
		// After injection, we deduplicate by keeping the last occurrence of each
		// flag.
		args = injectConfigSet(args, args[1]+".defaults", insertIdx)

		if args[1] != "sq" {
			args = injectConfigSet(args, "nostate", insertIdx)
		}

		args = injectConfigSet(args, "defaults", insertIdx)
		args = injectExplicitSet(args)
		args = deduplicateFlags(args)

		log.Debugf("args after set processing: args=%v", args)

		if len(args) > 1 && args[1] == "ps" {
			args = processPsArgs(args)
		}

		return args
	}
}

// injectConfigSet retrieves the config slice for the given key, expands each
// entry by whitespace, and inserts the resulting args at the specified index.
func injectConfigSet(args []string, key string, insertIdx int) []string {
	entries, _ := config.GetStringSlice(key)
	if len(entries) == 0 {
		return args
	}

	var expanded []string
	for _, entry := range entries {
		expanded = append(expanded, strings.Fields(entry)...)
	}

	return append(args[:insertIdx], append(expanded, args[insertIdx:]...)...)
}

// deduplicateFlags removes duplicate flags from args, keeping the last
// occurrence of each flag. Flags are identified by the "--" or "-" prefix.
// For flags with values (e.g., "--output text"), we track by the flag name and
// keep the last flag+value pair.
func deduplicateFlags(args []string) []string {
	if len(args) <= 2 {
		return args
	}

	// We preserve args[0] (program) and args[1] (command), then dedupe the rest.
	prefix := args[:2]
	rest := args[2:]

	// Track the last position where each flag appears.
	flagPositions := make(map[string]int)

	for i := 0; i < len(rest); i++ {
		arg := rest[i]

		if strings.HasPrefix(arg, "-") {
			flagName := arg

			// Handle --flag=value syntax.
			if before, _, ok := strings.Cut(arg, "="); ok {
				flagName = before
			}

			flagPositions[flagName] = i
		}
	}

	// Build result, skipping flags that appear again later.
	var result []string
	for i := 0; i < len(rest); i++ {
		arg := rest[i]

		if strings.HasPrefix(arg, "-") {
			flagName := arg
			if before, _, ok := strings.Cut(arg, "="); ok {
				flagName = before
			}

			// Skip this flag if it appears again later.
			if flagPositions[flagName] > i {
				// If this flag has a separate value arg, skip that too.
				if !strings.Contains(arg, "=") && i+1 < len(rest) &&
					!strings.HasPrefix(rest[i+1], "-") {
					i++
				}

				continue
			}
		}

		result = append(result, arg)
	}

	return append(prefix, result...)
}

// processPsArgs handles argument processing for the ps command.
func processPsArgs(args []string) []string {
	// Ensure the argument immediately following "ps" is "-" or an existing file.
	if len(args) == 2 || (args[2] != "-" && !isExistingFile(args[2])) {
		args = append(args[:2], append([]string{"-"}, args[2:]...)...)
	}
	return args
}

// isExistingFile checks if the given path exists and is a file.
func isExistingFile(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// injectExplicitSet handles the @set logic for all commands, expanding set
// arguments at the @set position.
func injectExplicitSet(args []string) []string {
	// Look for an explicit @set argument starting from starting idx.
	idx := 2
	set := ""
	setIdx := len(args)

	for i, a := range args[idx:] {
		if strings.HasPrefix(a, "@") {
			set = strings.TrimPrefix(a, "@")
			setIdx = 2 + i
			args = append(args[:setIdx], args[setIdx+1:]...)
			break
		}
	}

	if set != "" {
		setArgs, _ := config.GetStringSlice(args[1] + "." + set)
		for _, arg := range setArgs {
			parts := strings.Fields(arg)
			args = append(args[:setIdx], append(parts, args[setIdx:]...)...)
			setIdx += len(parts)
		}
	}

	return args
}
