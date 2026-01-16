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
	"github.com/tfctl/tfctl/internal/util"
	"github.com/tfctl/tfctl/internal/version"
)

var ctx = context.Background()

func main() {
	os.Exit(realMain())
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
		// For ps and other commands, process @set first.
		args = processSetOnly(args)
		log.Debugf("args after set processing: args=%v", args)

		if len(args) > 1 && args[1] == "ps" {
			args = processPsArgs(args)
		} else {
			args = processOtherArgs(args)
		}
		return args
	}
}

// processPsArgs handles argument processing for the ps command.
func processPsArgs(args []string) []string {
	// Ensure the argument immediately following "ps" is "-" or an existing file.
	if len(args) == 2 || (args[2] != "-" && !isExistingFile(args[2])) {
		args = append(args[:2], append([]string{"-"}, args[2:]...)...)
	}
	return args
}

// processOtherArgs handles argument processing for other commands.
func processOtherArgs(args []string) []string {
	rootDir, _ := os.Getwd()
	if len(args) > 2 {
		if _, _, err := util.ParseRootDir(args[2]); err == nil {
			rootDir = args[2]
		}
	}
	if len(args) == 2 {
		args = append(args, rootDir)
	} else if args[2] != rootDir {
		args = append(args[:2], append([]string{rootDir}, args[2:]...)...)
	}
	return args
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

// isExistingFile checks if the given path exists and is a file.
func isExistingFile(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// processSetOnly handles the @set logic for all commands, expanding set arguments at the @set position.
func processSetOnly(args []string) []string {
	// Look for an explicit @set argument starting from index 2.
	idx := 2
	set := "defaults"
	removeIdx := -1
	for i, a := range args[idx:] {
		if strings.HasPrefix(a, "@") {
			set = a[1:]
			removeIdx = idx + i
			break
		}
	}
	if removeIdx != -1 {
		// Remove the @set argument.
		args = append(args[:removeIdx], args[removeIdx+1:]...)
		// Expand the set arguments at the removeIdx position.
		setArgs, _ := config.GetStringSlice(args[1] + "." + set)
		for _, arg := range setArgs {
			parts := strings.Fields(arg)
			args = append(args[:removeIdx], append(parts, args[removeIdx:]...)...)
			removeIdx += len(parts)
		}
	}
	return args
}
