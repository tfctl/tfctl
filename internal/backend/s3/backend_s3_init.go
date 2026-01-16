// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package s3

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

type BackendS3Option = func(ctx context.Context, cmd *cli.Command, be *BackendS3) error

func FromRootDir(rootDir string, required ...bool) BackendS3Option {
	return func(ctx context.Context, cmd *cli.Command, be *BackendS3) error {
		// Is rootDir a relative or absolute path?
		if filepath.IsAbs(rootDir) {
			be.RootDir = rootDir
		} else {
			cwd, _ := os.Getwd()
			be.RootDir = filepath.Join(cwd, rootDir)
		}

		log.Debugf("NewBackendS3 FromRootDir(): rootDir = %s", be.RootDir)

		err := be.load()

		// Return no error is required is present and false.
		if len(required) > 0 && !required[0] {
			return nil
		}
		return err
	}
}

// NewBackendS3 returns a BackendS3 object that implements the Backend
// interface. It is load()ed from the config file found in the rootDir.
func NewBackendS3(ctx context.Context, cmd *cli.Command, options ...BackendS3Option) (*BackendS3, error) {
	options = append([]BackendS3Option{WithDefaults()}, options...)

	be := &BackendS3{Ctx: ctx, Cmd: cmd}

	for _, opt := range options {
		if err := opt(ctx, cmd, be); err != nil {
			return nil, err
		}
	}

	return be, nil
}

func WithDefaults() BackendS3Option {
	return func(ctx context.Context, cmd *cli.Command, be *BackendS3) error {
		cwd, _ := os.Getwd()
		be.RootDir = cwd

		be.Version = 4
		be.TerraformVersion = "0.0.0"
		be.Backend.Type = "s3"

		log.Debugf("NewBackendCloud WithDefaults():")

		return nil
	}
}

func WithEnvOverride(env string) BackendS3Option {
	return func(ctx context.Context, cmd *cli.Command, be *BackendS3) error {
		if env != "" {
			be.EnvOverride = env
		}
		return nil
	}
}

func WithSvOverride() BackendS3Option {
	return func(ctx context.Context, cmd *cli.Command, be *BackendS3) error {
		sv := cmd.String("sv")
		if sv != "" {
			be.SvOverride = sv
		}
		return nil
	}
}

func (be *BackendS3) load() error {
	tfFile := be.RootDir + "/.terraform/terraform.tfstate"
	data, err := os.ReadFile(tfFile)
	if err != nil {
		return fmt.Errorf("failed to read local config file: %w", err)
	}

	var temp BackendS3
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal local config file: %w", err)
	}

	if temp.Backend.Type != "s3" {
		return fmt.Errorf("%w: backend type is not s3: %s", errors.New("bad"), temp.Backend.Type)
	}

	be.Version = temp.Version
	be.TerraformVersion = temp.TerraformVersion
	be.Backend = temp.Backend

	return nil
}
