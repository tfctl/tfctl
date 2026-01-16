// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"errors"
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
)

// ErrorContext carries input context for improving API error messages.
type ErrorContext struct {
	Host      string
	Org       string
	Workspace string
	Operation string // e.g., "list state versions", "read workspace"
	Resource  string // e.g., "organization", "workspace"
}

// FriendlyTFE wraps a TFE error with a contextual, user-friendly message while
// preserving the original error for further inspection via errors.Is/As.
func FriendlyTFE(err error, ctx ErrorContext) error {
	if err == nil {
		return nil
	}

	host := nonEmpty(ctx.Host, "<unknown>")

	// Map well-known go-tfe sentinel errors to friendly text with pseudo-status.
	switch {
	case errors.Is(err, tfe.ErrUnauthorized):
		// No need to append the original sentinel here (would read as "unauthorized").
		return fmt.Errorf("%s on %s: authentication failed (401). Set %s or TF_TOKEN",
			nonEmpty(ctx.Operation, "request"), host, hostEnvKey(ctx.Host))

	case errors.Is(err, tfe.ErrResourceNotFound):
		if ctx.Workspace != "" {
			return fmt.Errorf("%s: workspace %q not found in organization %q on %s (404)",
				nonEmpty(ctx.Operation, "request"), ctx.Workspace, nonEmpty(ctx.Org, "<unknown>"), host)
		}
		return fmt.Errorf("%s: organization %q not found on %s (404)",
			nonEmpty(ctx.Operation, "request"), nonEmpty(ctx.Org, "<unknown>"), host)
	}

	// Unknown error: provide generic context and wrap
	return fmt.Errorf("%s on %s for org=%q workspace=%q: %w",
		nonEmpty(ctx.Operation, "request"), host, ctx.Org, ctx.Workspace, err)
}

func hostEnvKey(host string) string {
	if host == "" {
		return ""
	}
	// mirrors Token() env var construction logic: dots to underscores
	// e.g., app.terraform.io -> app_terraform_io
	key := "TF_TOKEN_" + replaceDots(host)
	return key
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func replaceDots(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			b[i] = '_'
		} else {
			b[i] = s[i]
		}
	}
	return string(b)
}
