// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"

	"github.com/urfave/cli/v3"
)

// Augmenter[T] is a callback function that customizes options before
// each API call. It receives the context, command, and a pointer to the
// options object, allowing mutation of options based on command flags or
// other context. Return an error to abort pagination.
type Augmenter[T any] func(
	context.Context,
	*cli.Command,
	*T,
) error
