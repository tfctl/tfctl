// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"context"

	"github.com/tfctl/tfctl/internal/config"
)

// RootDirSpec holds the resolved root directory and optional environment
// override used when evaluating backends.
type RootDirSpec struct {
	RootDir string
	Env     string
}

// Meta contains runtime metadata shared by commands. It carries CLI arguments,
// loaded configuration, context, the resolved root directory specification, and
// the starting working directory.
type Meta struct {
	Args    []string
	Config  config.Type
	Context context.Context
	RootDirSpec
	StartingDir string
}
