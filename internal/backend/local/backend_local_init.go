// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/urfave/cli/v3"
)

type BackendLocalOption = func(ctx context.Context, cmd *cli.Command, be *BackendLocal) error

func FromRootDir(rootDir string) BackendLocalOption {
	return func(ctx context.Context, cmd *cli.Command, be *BackendLocal) error {

		// Is rootDir a relative or absolute path?
		if filepath.IsAbs(rootDir) {
			be.RootDir = rootDir
		} else {
			cwd, _ := os.Getwd()
			be.RootDir = filepath.Join(cwd, rootDir)
		}

		// Does the backend file exist in RootDir and is it valid?
		// This is a Local so we should never return an error. If there had been
		// a missing backend file, we should have already built out a default Remote
		// backend.
		return be.load(ctx, cmd)
	}
}

// NewBackendLocal returns a BackendLocal object that implements the Backend
// interface. It is load()ed from the config file found in the rootDir.
func NewBackendLocal(ctx context.Context, cmd *cli.Command, options ...BackendLocalOption) (*BackendLocal, error) {
	options = append([]BackendLocalOption{WithDefaults()}, options...)

	be := &BackendLocal{Ctx: ctx, Cmd: cmd}

	for _, opt := range options {
		if err := opt(ctx, cmd, be); err != nil {
			return nil, err
		}
	}

	return be, nil
}

func WithDefaults() BackendLocalOption {
	return func(ctx context.Context, cmd *cli.Command, be *BackendLocal) error {
		cwd, _ := os.Getwd()
		be.RootDir = cwd

		be.Version = 4
		be.TerraformVersion = "0.0.0"
		be.Backend.Type = "local"

		return nil
	}
}

func WithEnvOverride(env string) BackendLocalOption {
	return func(ctx context.Context, cmd *cli.Command, be *BackendLocal) error {
		if env != "" {
			be.EnvOverride = env
		}
		return nil
	}
}

func WithNoBackend(rootDir string) BackendLocalOption {
	return func(ctx context.Context, cmd *cli.Command, be *BackendLocal) error {
		// Is rootDir a relative or absolute path?
		if filepath.IsAbs(rootDir) {
			be.RootDir = rootDir
		} else {
			cwd, _ := os.Getwd()
			be.RootDir = filepath.Join(cwd, rootDir)
		}

		be.Version = 4
		be.TerraformVersion = "0.0.0"
		be.Backend.Type = "local"
		be.Backend.Config.Path = "terraform.tfstate"

		return nil
	}
}

// load reads the terraform config file and unmarshals it into the ConfigLocal
// struct. It is simply a convenience method to make NewConfigLocal more
// readable.
func (be *BackendLocal) load(_ context.Context, _ *cli.Command) error {
	tfFile := be.RootDir + "/.terraform/terraform.tfstate"
	data, err := os.ReadFile(tfFile)
	if err != nil {
		// Deal with a no terraform.backend {} situation. In this case, it looks
		// like we're in a real IAC root, but there is no backend config file.
		// This is a valid situation, so we just return a local backend with no
		// config.
		if os.IsNotExist(err) {
			log.Debugf("local backend config file %s does not exist, assuming no backend", tfFile)
			be.Backend.Type = "local"
			return nil
		} else {
			return fmt.Errorf("failed to read local config file: %w", err)
		}
	}

	var temp BackendLocal
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal local config file: %w", err)
	}

	if temp.Backend.Type != "local" {
		return fmt.Errorf("%w: backend type is not local: %s", errors.New("bad"), temp.Backend.Type)
	}

	be.Version = temp.Version
	be.TerraformVersion = temp.TerraformVersion
	be.Backend = temp.Backend

	return nil
}
