// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/urfave/cli/v3"

	"github.com/tfctl/tfctl/internal/backend"
	"github.com/tfctl/tfctl/internal/config"
	"github.com/tfctl/tfctl/internal/differ"
	"github.com/tfctl/tfctl/internal/meta"
	"github.com/tfctl/tfctl/internal/output"
	"github.com/tfctl/tfctl/internal/state"
)

// sqCommandAction is the action handler for the "sq" subcommand. It reads
// Terraform state (including optional decryption), supports --tldr short-
// circuit, and emits results per common flags.
func sqCommandAction(ctx context.Context, cmd *cli.Command) error {
	m := GetMeta(cmd)
	log.Debugf("Executing action for %v", m.Args[1:])

	// Bail out early if we're just dumping tldr.
	if ShortCircuitTLDR(ctx, cmd, "sq") {
		return nil
	}

	config.Config.Namespace = "sq"

	// Figure out what type of Backend we're in.
	be, err := backend.NewBackend(ctx, *cmd)
	if err != nil {
		return err
	}
	log.Debugf("typBe: %v", be)

	// Short circuit --diff mode.
	if cmd.Bool("diff") {
		if _, ok := be.(backend.SelfDiffer); ok {
			states, diffErr := be.(backend.SelfDiffer).DiffStates(ctx, cmd)
			if diffErr != nil {
				log.Errorf("diff error: %v", diffErr)
				return diffErr
			}

			return differ.Diff(ctx, cmd, states)
		} else {
			log.Debug("Backend does not implement SelfDiffer")
		}
	}

	attrs := BuildAttrs(cmd, "!.mode", "!.type", ".resource", "id", "name")
	log.Debugf("attrs: %v", attrs)

	var doc []byte
	doc, err = be.State()
	if err != nil {
		return err
	}

	// If the state is encrypted, there's a little more work to do.
	var jsonData map[string]interface{}
	if err := json.Unmarshal(doc, &jsonData); err == nil {
		if _, exists := jsonData["encrypted_data"]; exists {
			// First, look to the flag for passphrase value.
			passphrase := cmd.String("passphrase")

			// Issue 14 - Next look in env and use it if found.
			if passphrase == "" {
				passphrase = os.Getenv("TFCTL_PASSPHRASE")
			}

			// Finally, prompt for passphrase
			if passphrase == "" {
				passphrase, _ = state.GetPassphrase()
			}

			doc, err = state.DecryptOpenTofuState(doc, passphrase)
			if err != nil {
				return fmt.Errorf("failed to decrypt: %w", err)
			}
		}
	}

	var raw bytes.Buffer
	raw.Write(doc)

	postProcess := func(dataset []map[string]interface{}) error {
		if cmd.Bool("chop") {
			chopPrefix(dataset)
		}

		return nil
	}

	output.SliceDiceSpit(raw, attrs, cmd, "", os.Stdout, postProcess)

	return nil
}

// sqCommandBuilder constructs the cli.Command for "sq", wiring metadata,
// flags, and action/validator handlers.
func sqCommandBuilder(meta meta.Meta) *cli.Command {
	return &cli.Command{
		Name:      "sq",
		Usage:     "state query",
		UsageText: "tfctl sq [RootDir] [options]",
		Metadata: map[string]any{
			"meta": meta,
		},
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:  "chop",
				Usage: "chop common resource prefix from names",
				Value: false,
			},
			&cli.BoolFlag{
				Name:    "concrete",
				Aliases: []string{"k"},
				Usage:   "only include concrete resources",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  "diff",
				Usage: "find difference between state versions",
				Value: false,
			},
			&cli.StringFlag{
				Name:   "diff_filter",
				Hidden: true,
				Value:  "check_results",
			},
			&cli.IntFlag{
				Name:   "limit",
				Hidden: true,
				Usage:  "limit state versions returned",
				Value:  99999,
			},
			&cli.BoolFlag{
				Name:  "short",
				Usage: "include full resource name paths",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "passphrase",
				Usage: "encrypted state passphrase",
			},
			&cli.StringFlag{
				Name:        "sv",
				Usage:       "state version to query",
				Value:       "0",
				HideDefault: true,
			},
			// We don't want sq to get default host and org values from the config.
			// Instead, we'll depend on the backend or, in exceptional cases, explicit
			// --host and --org flags.
			NewHostFlag("sq"),
			NewOrgFlag("sq"),
			tldrFlag,
			workspaceFlag,
		}, NewGlobalFlags("sq")...),
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// If --chop is set, --short must not be set.
			if cmd.Bool("chop") {
				_ = cmd.Set("short", "false")
			}

			return ctx, GlobalFlagsValidator(ctx, cmd)
		},
		Action: sqCommandAction,
	}
}

// chopPrefix scans all dot-delimited string values in the dataset and removes
// leading segments that are identical across all entries. Starting from
// the left, it removes each segment that matches in all entries, then
// stops when it encounters a position where segments differ. Removed
// segments are replaced with "..".
func chopPrefix(dataset []map[string]interface{}) {
	if len(dataset) == 0 {
		return
	}

	// Collect all string values by key across all entries
	// Map from key -> list of (entryIdx, segments)
	type segmentedValue struct {
		entryIdx int
		segments []string
	}

	keyValues := make(map[string][]segmentedValue)

	for entryIdx, entry := range dataset {
		for key, val := range entry {
			if str, ok := val.(string); ok {
				segments := strings.Split(str, ".")
				keyValues[key] = append(keyValues[key], segmentedValue{entryIdx: entryIdx, segments: segments})
			}
		}
	}

	// For each key, find and apply the common prefix
	for key, values := range keyValues {
		if len(values) == 0 {
			continue
		}

		// Find common leading segments for this key
		var commonCount int
		for segIdx := 0; ; segIdx++ {
			// Check if all entries have a segment at this position
			if segIdx >= len(values[0].segments) {
				break
			}

			// Get the segment value from the first entry
			expectedSeg := values[0].segments[segIdx]

			// Check if all entries have the same segment at this position
			allMatch := true
			for _, val := range values {
				if segIdx >= len(val.segments) || val.segments[segIdx] != expectedSeg {
					allMatch = false
					break
				}
			}

			if !allMatch {
				break
			}

			commonCount++
		}

		// Need at least 2 common segments to be worth chopping.
		if commonCount < 2 {
			continue
		}

		// Never chop past the second-to-last segment. Ensure at least 2
		// segments remain in all values after chopping.
		minSegments := len(values[0].segments)
		for _, val := range values {
			if len(val.segments) < minSegments {
				minSegments = len(val.segments)
			}
		}
		maxChop := minSegments - 2
		if maxChop < 2 {
			continue
		}
		if commonCount > maxChop {
			commonCount = maxChop
		}

		// Build the prefix to remove
		prefixSegs := values[0].segments[:commonCount]
		prefixToRemove := strings.Join(prefixSegs, ".") + "."

		// Remove the common prefix from all entries that have this key
		for _, val := range values {
			originalValue := strings.Join(val.segments, ".")
			if strings.HasPrefix(originalValue, prefixToRemove) {
				dataset[val.entryIdx][key] = ".." + originalValue[len(prefixToRemove):]
			}
		}
	}
}
