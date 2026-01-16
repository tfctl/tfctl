// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package cloud

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

type BackendCloudOption = func(ctx context.Context, cmd *cli.Command, be *BackendCloud) error

func FromRootDir(rootDir string, required ...bool) BackendCloudOption {
	return func(ctx context.Context, cmd *cli.Command, be *BackendCloud) error {
		// Is rootDir a relative or absolute path?
		if filepath.IsAbs(rootDir) {
			be.RootDir = rootDir
		} else {
			cwd, _ := os.Getwd()
			be.RootDir = filepath.Join(cwd, rootDir)
		}

		log.Debugf("NewBackendCloud FromRootDir(): rootDir = %s", be.RootDir)

		err := be.load(ctx, cmd)

		// Return no error is required is present and false.
		if len(required) > 0 && !required[0] {
			return nil
		}
		return err
	}
}

// NewBackendCloud returns a BackendCloud object that implements the Backend
// interface. It is load()ed from the config file found in the rootDir.
func NewBackendCloud(ctx context.Context, cmd *cli.Command, options ...BackendCloudOption) (*BackendCloud, error) {
	options = append([]BackendCloudOption{WithDefaults()}, options...)

	be := &BackendCloud{Ctx: ctx, Cmd: cmd}

	for _, opt := range options {
		if err := opt(ctx, cmd, be); err != nil {
			return nil, err
		}
	}

	return be, nil
}

func WithDefaults() BackendCloudOption {
	return func(ctx context.Context, cmd *cli.Command, be *BackendCloud) error {
		cwd, _ := os.Getwd()
		be.RootDir = cwd

		be.Version = 4
		be.TerraformVersion = "0.0.0"
		be.Backend.Type = "cloud"

		log.Debugf("NewBackendCloud WithDefaults():")

		return nil
	}
}

func WithEnvOverride(env string) BackendCloudOption {
	return func(ctx context.Context, cmd *cli.Command, be *BackendCloud) error {
		if env != "" {
			be.EnvOverride = env
		}
		return nil
	}
}

func (be *BackendCloud) load(_ context.Context, _ *cli.Command) error {
	tfFile := be.RootDir + "/.terraform/terraform.tfstate"
	data, err := os.ReadFile(tfFile)
	if err != nil {
		return fmt.Errorf("failed to read local config file: %w", err)
	}

	var temp BackendCloud
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal local config file: %w", err)
	}

	if temp.Backend.Type != "cloud" {
		return fmt.Errorf("%w: backend type is not cloud: %s", errors.New("bad"), temp.Backend.Type)
	}

	be.Version = temp.Version
	be.TerraformVersion = temp.TerraformVersion
	be.Backend = temp.Backend

	if be.Backend.Config.Token == nil {
		token, _ := be.Token()
		be.Backend.Config.Token = token
	}

	return nil
}
