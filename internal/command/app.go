// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/tfctl/tfctl/internal/config"
	"github.com/tfctl/tfctl/internal/meta"
	"github.com/tfctl/tfctl/internal/util"
)

func InitApp(ctx context.Context, args []string) (*cli.Command, error) {

	// Save the CWD at startup and then defer restoring it so we're tidy.
	sd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(sd); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to restore directory: %v\n", err)
		}
	}()

	// The arg[1] immediately following the binary (arg[0]) is the tfctl
	// subcommand and also represents the namespace key to be used when retrieving
	// config values. arg[1] could be -h/--help, so ignore it if it appears to be
	// a flag.
	var ns string
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		ns = args[1]
	}

	// allow short if-style local cfg; no actual outer cfg
	cfg2, _ := config.Load(ns) //nolint
	meta := meta.Meta{
		Args:        args,
		Config:      cfg2,
		Context:     ctx,
		StartingDir: sd,
	}

	// See if the arg immediately following the command might be a directory.
	// This is determined by whether or not it begins with - or --.  If it does,
	// it's a flag and the CWD directory is the starting directory.  If it's not,
	// we assume we have a directory spec of some sort and need to parse it more.
	// Special-case the 'completion' and 'ps' commands which take a plain
	// positional argument (e.g., 'bash' or 'zsh' for completion, plan file
	// for ps).
	if (ns != "completion" && ns != "ps") && len(args) > 2 && !strings.HasPrefix(args[2], "-") {
		if wd, env, err := util.ParseRootDir(args[2]); err == nil {
			meta.RootDir = wd
			meta.Env = env
		} else {
			return nil, fmt.Errorf("failed to parse rootDir (%s): %w", args[2], err)
		}
	} else {
		meta.RootDir = sd
	}

	app := &cli.Command{
		Name:  "tfctl",
		Usage: "Terraform Control",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "version",
				Aliases:     []string{"v"},
				Usage:       "tfctl version info",
				HideDefault: true,
			},
		},
	}

	app.Commands = append(app.Commands,
		mqCommandBuilder(meta),
		oqCommandBuilder(meta),
		pqCommandBuilder(meta),
		psCommandBuilder(meta),
		rqCommandBuilder(meta),
		siCommandBuilder(meta),
		sqCommandBuilder(meta),
		svqCommandBuilder(meta),
		wqCommandBuilder(meta),
		completionCommandBuilder(meta),
	)

	// Make sure flags are sorted for the --help text.
	for _, cmd := range app.Commands {
		sort.Slice(cmd.Flags, func(i, j int) bool {
			return cmd.Flags[i].Names()[0] < cmd.Flags[j].Names()[0]
		})
	}

	return app, nil
}
