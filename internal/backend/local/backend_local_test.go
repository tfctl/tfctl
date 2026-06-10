// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// fakeFile describes a single file to materialize inside a fake IAC root.
type fakeFile struct {
	rel    string        // path relative to the root dir
	body   string        // file contents
	modAge time.Duration // age relative to a fixed base; larger == older
}

// newFakeRoot builds a temporary IAC root populated with the given files and
// returns its absolute path. Mod times are set explicitly so that StateVersions'
// descending-mod-time sort is deterministic regardless of write order.
func newFakeRoot(t *testing.T, files ...fakeFile) string {
	t.Helper()
	root := t.TempDir()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for _, f := range files {
		p := filepath.Join(root, f.rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o750))
		require.NoError(t, os.WriteFile(p, []byte(f.body), 0o600))
		mod := base.Add(-f.modAge)
		require.NoError(t, os.Chtimes(p, mod, mod))
	}
	return root
}

// stateBody returns a minimal but valid state document carrying the given serial.
func stateBody(serial int64) string {
	return fmt.Sprintf(`{"version":4,"serial":%d,"terraform_version":"1.5.0"}`, serial)
}

func TestStateVersions(t *testing.T) {
	t.Run("sorted by mod time descending with metadata", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: stateBody(7), modAge: 0},
			fakeFile{rel: "terraform.tfstate.backup", body: stateBody(6), modAge: time.Hour},
		)
		be := &BackendLocal{RootDir: root}

		versions, err := be.StateVersions()
		require.NoError(t, err)
		require.Len(t, versions, 2)

		// Newest first.
		assert.Equal(t, "terraform.tfstate", versions[0].ID)
		assert.Equal(t, int64(7), versions[0].Serial)
		assert.Equal(t, filepath.Join(root, "terraform.tfstate"), versions[0].JSONDownloadURL)

		assert.Equal(t, "terraform.tfstate.backup", versions[1].ID)
		assert.Equal(t, int64(6), versions[1].Serial)
		assert.Equal(t, filepath.Join(root, "terraform.tfstate.backup"), versions[1].JSONDownloadURL)
	})

	t.Run("malformed state file is skipped not fatal", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: stateBody(3), modAge: 0},
			fakeFile{rel: "terraform.tfstate.backup", body: "this is not json", modAge: time.Hour},
		)
		be := &BackendLocal{RootDir: root}

		versions, err := be.StateVersions()
		require.NoError(t, err)
		require.Len(t, versions, 1)
		assert.Equal(t, "terraform.tfstate", versions[0].ID)
	})

	t.Run("ignores files not matching the glob", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: stateBody(1), modAge: 0},
			fakeFile{rel: "other.tfstate", body: stateBody(2), modAge: 0},
			fakeFile{rel: "notes.txt", body: "hello", modAge: 0},
		)
		be := &BackendLocal{RootDir: root}

		versions, err := be.StateVersions()
		require.NoError(t, err)
		require.Len(t, versions, 1)
		assert.Equal(t, "terraform.tfstate", versions[0].ID)
	})

	t.Run("empty root yields no versions and no error", func(t *testing.T) {
		be := &BackendLocal{RootDir: t.TempDir()}

		versions, err := be.StateVersions()
		require.NoError(t, err)
		assert.Empty(t, versions)
	})
}

func TestStateVersionsEnvOverride(t *testing.T) {
	t.Run("environment file selects workspace dir", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: ".terraform/environment", body: "prod"},
			fakeFile{rel: "terraform.tfstate", body: stateBody(1), modAge: 0},
			fakeFile{rel: "terraform.tfstate.d/prod/terraform.tfstate", body: stateBody(42), modAge: 0},
		)
		be := &BackendLocal{RootDir: root}

		versions, err := be.StateVersions()
		require.NoError(t, err)
		require.Len(t, versions, 1)
		// The workspace file, not the root-level one.
		assert.Equal(t, int64(42), versions[0].Serial)
		assert.Equal(t, "prod", be.EnvOverride)
	})

	t.Run("environment file is whitespace trimmed", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: ".terraform/environment", body: "  prod\n"},
			fakeFile{rel: "terraform.tfstate.d/prod/terraform.tfstate", body: stateBody(9), modAge: 0},
		)
		be := &BackendLocal{RootDir: root}

		versions, err := be.StateVersions()
		require.NoError(t, err)
		require.Len(t, versions, 1)
		assert.Equal(t, "prod", be.EnvOverride)
		assert.Equal(t, int64(9), versions[0].Serial)
	})

	t.Run("explicit EnvOverride wins over environment file", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: ".terraform/environment", body: "prod"},
			fakeFile{rel: "terraform.tfstate.d/prod/terraform.tfstate", body: stateBody(1), modAge: 0},
			fakeFile{rel: "terraform.tfstate.d/dev/terraform.tfstate", body: stateBody(2), modAge: 0},
		)
		be := &BackendLocal{RootDir: root, EnvOverride: "dev"}

		versions, err := be.StateVersions()
		require.NoError(t, err)
		require.Len(t, versions, 1)
		assert.Equal(t, "dev", be.EnvOverride)
		assert.Equal(t, int64(2), versions[0].Serial)
	})
}

func TestStates(t *testing.T) {
	t.Run("explicit ID spec returns that file body", func(t *testing.T) {
		body := stateBody(5)
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: body, modAge: 0},
			fakeFile{rel: "terraform.tfstate.backup", body: stateBody(4), modAge: time.Hour},
		)
		be := &BackendLocal{RootDir: root}

		states, err := be.States("terraform.tfstate")
		require.NoError(t, err)
		require.Len(t, states, 1)
		assert.JSONEq(t, body, string(states[0]))
	})

	t.Run("no spec defaults to most recent", func(t *testing.T) {
		newest := stateBody(5)
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: newest, modAge: 0},
			fakeFile{rel: "terraform.tfstate.backup", body: stateBody(4), modAge: time.Hour},
		)
		be := &BackendLocal{RootDir: root}

		states, err := be.States()
		require.NoError(t, err)
		require.Len(t, states, 1)
		assert.JSONEq(t, newest, string(states[0]))
	})

	t.Run("non-matching spec returns a resolve error", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: stateBody(1), modAge: 0},
		)
		be := &BackendLocal{RootDir: root}

		states, err := be.States("does-not-exist")
		require.Error(t, err)
		assert.Nil(t, states)
	})

	t.Run("file-path spec is read directly even when outside the scan", func(t *testing.T) {
		// svutil.resolveFileSpec resolves an existing path that the glob never
		// surfaced, so States must fall back to reading it from disk.
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: stateBody(1), modAge: 0},
		)
		external := stateBody(99)
		extPath := filepath.Join(root, "snapshot.json") // not a terraform.tfstate* match
		require.NoError(t, os.WriteFile(extPath, []byte(external), 0o600))
		be := &BackendLocal{RootDir: root}

		states, err := be.States(extPath)
		require.NoError(t, err)
		require.Len(t, states, 1)
		assert.JSONEq(t, external, string(states[0]))
	})
}

func TestLoad(t *testing.T) {
	validConfig := `{
		"version": 4,
		"terraform_version": "1.5.0",
		"backend": {
			"type": "local",
			"config": {"path": "terraform.tfstate", "workspace_dir": ""},
			"hash": 123
		}
	}`

	t.Run("valid local config populates fields", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: ".terraform/terraform.tfstate", body: validConfig},
		)
		be := &BackendLocal{RootDir: root}

		require.NoError(t, be.load(context.Background(), nil))
		assert.Equal(t, 4, be.Version)
		assert.Equal(t, "1.5.0", be.TerraformVersion)
		assert.Equal(t, "local", be.Backend.Type)
		assert.Equal(t, "terraform.tfstate", be.Backend.Config.Path)
	})

	t.Run("missing config file is not an error and assumes local", func(t *testing.T) {
		be := &BackendLocal{RootDir: t.TempDir()}

		require.NoError(t, be.load(context.Background(), nil))
		assert.Equal(t, "local", be.Backend.Type)
	})

	t.Run("non-local backend type is an error", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: ".terraform/terraform.tfstate", body: `{"backend":{"type":"s3"}}`},
		)
		be := &BackendLocal{RootDir: root}

		assert.Error(t, be.load(context.Background(), nil))
	})

	t.Run("malformed config json is an error", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: ".terraform/terraform.tfstate", body: "not json at all"},
		)
		be := &BackendLocal{RootDir: root}

		assert.Error(t, be.load(context.Background(), nil))
	})
}

// cmdWithSV builds a minimal cli.Command exposing an --sv flag preset to value.
func cmdWithSV(value string) *cli.Command {
	return &cli.Command{
		Name:  "local",
		Flags: []cli.Flag{&cli.StringFlag{Name: "sv", Value: value}},
	}
}

func TestState(t *testing.T) {
	t.Run("matching sv returns the resolved body", func(t *testing.T) {
		body := stateBody(8)
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: body, modAge: 0},
		)
		be := &BackendLocal{RootDir: root, Cmd: cmdWithSV("terraform.tfstate")}

		got, err := be.State()
		require.NoError(t, err)
		assert.JSONEq(t, body, string(got))
	})

	// Resolve guarantees one StateVersion per spec on success, so State's
	// states[0] indexing cannot panic: a miss surfaces as an error instead.
	// These cases lock in that invariant.
	t.Run("non-matching sv returns an error without panicking", func(t *testing.T) {
		root := newFakeRoot(t,
			fakeFile{rel: "terraform.tfstate", body: stateBody(1), modAge: 0},
		)
		be := &BackendLocal{RootDir: root, Cmd: cmdWithSV("nope")}

		assert.NotPanics(t, func() {
			got, err := be.State()
			require.Error(t, err)
			assert.Nil(t, got)
		})
	})

	t.Run("empty root with empty sv returns an error without panicking", func(t *testing.T) {
		be := &BackendLocal{RootDir: t.TempDir(), Cmd: cmdWithSV("")}

		assert.NotPanics(t, func() {
			got, err := be.State()
			require.Error(t, err)
			assert.Nil(t, got)
		})
	})
}
