// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/apex/log"
	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/go-tfe"
	"github.com/urfave/cli/v3"

	awsx "github.com/tfctl/tfctl/internal/aws"
	"github.com/tfctl/tfctl/internal/differ"
	"github.com/tfctl/tfctl/internal/svutil"
)

type BackendS3 struct {
	Ctx              context.Context
	Cmd              *cli.Command
	RootDir          string `json:"-" validate:"dir"`
	EnvOverride      string
	SvOverride       string
	Version          int    `json:"version" validate:"gte=3"`
	TerraformVersion string `json:"terraform_version" validate:"semver"`
	Backend          struct {
		Type   string `json:"type" validate:"eq=local"`
		Config struct {
			Bucket   string `json:"bucket"`
			Key      string `json:"key"`
			Prefix   string `json:"workspace_key_prefix"`
			Region   string `json:"region"`
			Encrypt  bool   `json:"encrypt"`
			KmsKeyID string `json:"kms_key_id"`
		} `json:"config"`
		Hash int `json:"hash"`
	} `json:"backend"`
}

func (be *BackendS3) DiffStates(ctx context.Context, cmd *cli.Command) ([][]byte, error) {
	// Fixup diffArgs
	svSpecs := []string{"CSV~1", "CSV~0"}

	diffArgs := differ.ParseDiffArgs(ctx, cmd)

	switch len(diffArgs) {
	case 0:
		// No args, so use the last two states.
	case 1:
		if strings.HasPrefix(diffArgs[0], "+") {
			// limit := 9999
			// if l, err := strconv.Atoi(diffArgs[0][1:]); err == nil {
			// 	limit = l
			// }

			stateVersionList, err := be.StateVersions( /* TODO limit */ )
			if err != nil {
				return nil, fmt.Errorf("failed to get state version list: %v", err)
			}

			selectedVersions := differ.SelectStateVersions(stateVersionList)

			log.Debugf("selectedVersions: %d", len(selectedVersions))

			if len(selectedVersions) == 0 {
				return nil, nil
			} else if len(selectedVersions) == 2 {
				svSpecs[0] = selectedVersions[1].ID
				svSpecs[1] = selectedVersions[0].ID
			}
		} else {
			svSpecs[0] = diffArgs[0]
		}
	case 2:
		svSpecs = diffArgs
	}

	states, _ := be.States(svSpecs[0], svSpecs[1])

	return states, nil
}

func (be *BackendS3) Runs() ([]*tfe.Run, error) {
	return nil, fmt.Errorf("not implemented")
}

func (be *BackendS3) State() ([]byte, error) {
	sv := be.Cmd.String("sv")
	states, err := be.States(sv)
	if err != nil {
		return nil, err
	}
	return states[0], nil
}

func (be *BackendS3) StateBody(svID string) ([]byte, error) {
	if err := PurgeCache(); err != nil {
		log.WithError(err).Warn("failed to purge cache")
	}

	if entry, ok := CacheReader(be, svID); ok {
		return entry.Data, nil
	}

	var env string
	// If there's already an envOverride (rootDir::env), use it.
	if be.EnvOverride != "" {
		env = be.EnvOverride
		// Else if we're in a prefixed workspace, get the env from the file.
	} else if be.Backend.Config.Prefix != "" {
		envData, err := os.ReadFile(filepath.Join(be.RootDir, ".terraform/environment"))
		if err == nil {
			env = string(envData)
		}
	}
	key := filepath.Join(be.Backend.Config.Prefix, env, be.Backend.Config.Key)

	// Build AWS config (inherit env; override region if provided)
	var cfgOpts []awsx.Option
	if be.Backend.Config.Region != "" {
		cfgOpts = append(cfgOpts, awsx.WithRegion(be.Backend.Config.Region))
	}
	cfg, err := awsx.LoadAWSConfig(be.Ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	svc := awsx.NewS3(cfg)
	input := &s3v2.GetObjectInput{
		Bucket:    awsv2.String(be.Backend.Config.Bucket),
		Key:       awsv2.String(key),
		VersionId: awsv2.String(svID),
	}

	result, err := svc.GetObject(be.Ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return data, nil
}

// StateVersions implements backend.Backend. It scans be.RootDir for state and
// backup files, parses them, and creates minimal tfe.StateVersion with ID as
// filename, CreatedAt from file timestamp, and Serial from the document.
func (be *BackendS3) StateVersions(augmenter ...func(context.Context, *cli.Command, *tfe.StateVersionListOptions) error) ([]*tfe.StateVersion, error) {
	var env string
	if be.EnvOverride != "" {
		env = be.EnvOverride
	} else if be.Backend.Config.Prefix != "" {
		envData, err := os.ReadFile(filepath.Join(be.RootDir, ".terraform/environment"))
		if err == nil {
			env = string(envData)
		}
	}
	prefix := filepath.Join(be.Backend.Config.Prefix, env, be.Backend.Config.Key)

	var cfgOpts []awsx.Option
	if be.Backend.Config.Region != "" {
		cfgOpts = append(cfgOpts, awsx.WithRegion(be.Backend.Config.Region))
	}
	cfg, err := awsx.LoadAWSConfig(be.Ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	svc := awsx.NewS3(cfg)
	paginator := s3v2.NewListObjectVersionsPaginator(svc, &s3v2.ListObjectVersionsInput{
		Bucket: awsv2.String(be.Backend.Config.Bucket),
		Prefix: awsv2.String(prefix),
	})
	combinedVersions := []*tfe.StateVersion{}

	var allDeleteMarkers []types.DeleteMarkerEntry
	var allVersions []types.ObjectVersion
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(be.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list object versions: %w", err)
		}
		allDeleteMarkers = append(allDeleteMarkers, page.DeleteMarkers...)
		allVersions = append(allVersions, page.Versions...)
	}
	var mostRecentDelete time.Time
	for _, d := range allDeleteMarkers {
		// This filters out tflock files. The prefix is literally a prefix so both
		// the actual state file versions and any lock files they might have, are
		// returned by the AWS API.
		if d.Key == nil || *d.Key != prefix {
			if d.Key != nil {
				log.Debugf("Throwing away delete marker %s", *d.Key)
			}
			continue
		}
		if d.LastModified != nil && d.LastModified.After(mostRecentDelete) {
			mostRecentDelete = *d.LastModified
		}
	}

	for _, v := range allVersions {
		if v.Key == nil || *v.Key != prefix {
			if v.Key != nil {
				log.Debugf("Throwing away %s", *v.Key)
			}
			continue
		}

		if v.LastModified != nil && v.LastModified.Before(mostRecentDelete) {
			continue
		}

		obj, err := svc.GetObject(be.Ctx, &s3v2.GetObjectInput{
			Bucket:    awsv2.String(be.Backend.Config.Bucket),
			Key:       awsv2.String(prefix),
			VersionId: v.VersionId,
		})
		if err != nil {
			log.WithError(err).Error("s3 get object failed")
			continue
		}

		var body []byte
		if v.VersionId == nil {
			// Shouldn't happen, but skip if no version id
			_ = obj.Body.Close()
			continue
		}
		entry, ok := CacheReader(be, *v.VersionId)
		if !ok {
			body, err = io.ReadAll(obj.Body)
			obj.Body.Close()
			if err != nil {
				continue
			}

			if err := CacheWriter(be, *v.VersionId, body); err != nil {
				log.WithError(err).Error("error writing to cache")
			}
		} else {
			body = entry.Data
		}

		var doc map[string]interface{}
		_ = json.Unmarshal(body, &doc)
		serial := doc["serial"]

		var serialInt int64
		switch s := serial.(type) {
		case float64:
			serialInt = int64(s)
		case int64:
			serialInt = s
		case int:
			serialInt = int64(s)
		default:
			serialInt = 0
		}

		// Guard against nil pointers
		if v.VersionId != nil && v.LastModified != nil {
			combinedVersions = append(combinedVersions, &tfe.StateVersion{
				ID:        *v.VersionId,
				CreatedAt: *v.LastModified,
				Serial:    serialInt,
			})
		}

	}

	sort.Slice(combinedVersions, func(i, j int) bool {
		return combinedVersions[i].CreatedAt.After(combinedVersions[j].CreatedAt)
	})

	currentVersions := []*tfe.StateVersion{}

	for _, v := range combinedVersions {
		if v.Serial == 0 {
			break
		}

		currentVersions = append(currentVersions, v)
	}

	limit := be.Cmd.Int("limit")
	if len(currentVersions) > limit {
		currentVersions = currentVersions[:limit]
	}

	return currentVersions, nil
}

func (be *BackendS3) States(specs ...string) ([][]byte, error) {
	var results [][]byte

	candidates, _ := be.StateVersions()
	versions, err := svutil.Resolve(candidates, specs...)
	if err != nil {
		return nil, err
	}
	log.Debugf("versions: %v", versions)

	// Now pound through the found versions and return each of their state bodies.
	for _, v := range versions {
		body, err := be.StateBody(v.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get state: %w", err)
		}
		results = append(results, body)
	}

	return results, nil
}

func (be *BackendS3) String() string {
	// TODO: provide a meaningful string representation if needed by callers
	return "backend-s3"
}

func (be *BackendS3) Type() (string, error) {
	return be.Backend.Type, nil
}
