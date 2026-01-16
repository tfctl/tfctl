// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/hashicorp/go-tfe"
	"github.com/urfave/cli/v3"

	"github.com/tfctl/tfctl/internal/backend/cloud"
	"github.com/tfctl/tfctl/internal/backend/local"
	"github.com/tfctl/tfctl/internal/backend/remote"
	"github.com/tfctl/tfctl/internal/backend/s3"
	"github.com/tfctl/tfctl/internal/meta"
)

// Type holds common backend resolution context and flags.
type Type struct {
	Ctx         context.Context
	Cmd         *cli.Command
	RootDir     string `json:"-" validate:"dir"`
	EnvOverride string
	// Version          int    `json:"version" validate:"gte=4"`
	// TerraformVersion string `json:"terraform_version" validate:"semver"`
}

// Backend abstracts Terraform/OpenTofu backend interactions needed by the
// application.
type Backend interface {
	Runs() ([]*tfe.Run, error)
	// State() returns the CSV~0 state document.
	State() ([]byte, error)
	// States() returns the state documents specified by the specs.
	States(...string) ([][]byte, error)
	// StateVersions accepts an optional augmenter function to apply
	// server-side filters. Only remote backends use this; local and S3 ignore it.
	StateVersions(augmenter ...func(context.Context, *cli.Command, *tfe.StateVersionListOptions) error) ([]*tfe.StateVersion, error)
	String() string
	Type() (string, error)
}

// SelfDiffer is implemented by backends that can diff state snapshots without
// an external differ.
type SelfDiffer interface {
	DiffStates(ctx context.Context, cmd *cli.Command) ([][]byte, error)
}

// NewBackend returns the appropriate Backend implementation for the working
// directory represented by the resolved root dir in command metadata.
func NewBackend(ctx context.Context, cmd cli.Command) (Backend, error) {
	meta := cmd.Metadata["meta"].(meta.Meta)
	log.Debugf("NewBackend: meta: %v", meta)

	cFile, cErr := os.Stat(filepath.Join(meta.RootDir, ".terraform", "terraform.tfstate"))
	sFile, sErr := os.Stat(filepath.Join(meta.RootDir, "terraform.tfstate"))
	eFile, eErr := os.Stat(filepath.Join(meta.RootDir, ".terraform", "environment"))
	_, _, _ = cFile, sFile, eFile // HACK

	// Maybe we're in a non-sq command and just need a naked remote. This will be
	// when c, s and e are all in error meaning none of them exist.
	if cErr != nil && sErr != nil && eErr != nil {
		return remote.NewBackendRemote(ctx, &cmd, remote.BuckNaked())
	}

	// If terraform.tfstate exists but .terraform/terraform.tfstate doesn't,
	// infer local backend. This is an empty terraform.backend {} block use case.
	if cErr != nil && sErr == nil {
		return local.NewBackendLocal(ctx, &cmd,
			local.FromRootDir(meta.RootDir),
			local.WithEnvOverride(meta.Env),
		)
	}

	// If .terraform/terraform.tfstate and terraform.tfstate don't exist but
	// .terraform/environment does, we're in a local backend with multi-workspace
	// configuration. The environment file points to the workspace directory.
	if cErr != nil && sErr != nil && eErr == nil {
		return local.NewBackendLocal(ctx, &cmd,
			local.FromRootDir(meta.RootDir),
			local.WithEnvOverride(meta.Env),
		)
	}

	// Peek at the backend type so we can switch on it.
	// TODO We're double reading the file. Once in peek() and once in the New().
	typ, err := peek(meta)
	if err != nil {
		return nil, err
	}

	var result Backend
	switch typ {
	case "cloud":
		var beCloud *cloud.BackendCloud
		beCloud, err = cloud.NewBackendCloud(ctx, &cmd,
			cloud.FromRootDir(meta.RootDir),
			cloud.WithEnvOverride(meta.Env),
		)
		// Preserve prior behavior: return transformed backend alongside any error
		result = beCloud.Transform2Remote(ctx, &cmd)
	case "local":
		result, err = local.NewBackendLocal(ctx, &cmd,
			local.FromRootDir(meta.RootDir),
			local.WithEnvOverride(meta.Env),
		)
	case "remote":
		result, err = remote.NewBackendRemote(ctx, &cmd,
			remote.FromRootDir(meta.RootDir),
			remote.WithEnvOverride(meta.Env),
			remote.WithSvOverride(),
		)
	case "s3":
		result, err = s3.NewBackendS3(ctx, &cmd,
			s3.FromRootDir(meta.RootDir),
			s3.WithEnvOverride(meta.Env),
			s3.WithSvOverride(),
		)
	default:
		return nil, fmt.Errorf("unknown type %s: %w", typ, err)
	}

	return result, err
}

// peek returns the backend type by reading the local terraform state file.
func peek(meta meta.Meta) (string, error) {
	raw, err := os.ReadFile(filepath.Join(meta.RootDir, ".terraform", "terraform.tfstate"))
	if err != nil {
		return "", err
	}

	var peeker map[string]json.RawMessage
	if err := json.Unmarshal(raw, &peeker); err != nil {
		return "", fmt.Errorf("can't peek: %w", err)
	}

	if err := json.Unmarshal(peeker["backend"], &peeker); err != nil {
		return "", fmt.Errorf("can't peek: %w", err)
	}

	var typ string
	if err := json.Unmarshal(peeker["type"], &typ); err != nil {
		return "", fmt.Errorf("can't peek: %w", err)
	}
	log.Debugf("type: %s", typ)

	return typ, nil
}
