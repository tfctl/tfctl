// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

//go:build integration
// +build integration

package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_S3PutGetDelete verifies real S3 bucket operations using
// configured AWS credentials. Requires AWS_ACCESS_KEY_ID and
// AWS_SECRET_ACCESS_KEY environment variables to be set.
func TestIntegration_S3PutGetDelete(t *testing.T) {
	ctx := context.Background()

	// Load AWS config using default credential chain (env vars, config
	// files, IMDS, etc.)
	cfg, err := LoadAWSConfig(ctx, WithRegion("us-east-1"))
	require.NoError(t, err)

	client := s3v2.NewFromConfig(cfg)

	bucketName := fmt.Sprintf("tfctlgo-test-%d", time.Now().UnixNano())
	testKey := "test-object.txt"
	testData := []byte("Hello from AWS!")

	// Create bucket
	_, err = client.CreateBucket(ctx, &s3v2.CreateBucketInput{
		Bucket: awsv2.String(bucketName),
	})
	require.NoError(t, err)
	defer func() {
		// Clean up: delete bucket
		client.DeleteBucket(ctx, &s3v2.DeleteBucketInput{
			Bucket: awsv2.String(bucketName),
		})
	}()

	// Put object
	_, err = client.PutObject(ctx, &s3v2.PutObjectInput{
		Bucket: awsv2.String(bucketName),
		Key:    awsv2.String(testKey),
		Body:   bytes.NewReader(testData),
	})
	require.NoError(t, err)

	// Get object
	result, err := client.GetObject(ctx, &s3v2.GetObjectInput{
		Bucket: awsv2.String(bucketName),
		Key:    awsv2.String(testKey),
	})
	require.NoError(t, err)

	// Verify content
	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, testData, body)
	result.Body.Close()

	// Delete object
	_, err = client.DeleteObject(ctx, &s3v2.DeleteObjectInput{
		Bucket: awsv2.String(bucketName),
		Key:    awsv2.String(testKey),
	})
	require.NoError(t, err)

	// Verify deletion
	_, err = client.GetObject(ctx, &s3v2.GetObjectInput{
		Bucket: awsv2.String(bucketName),
		Key:    awsv2.String(testKey),
	})
	assert.Error(t, err)
}

// TestIntegration_S3ListObjects verifies S3 ListObjectsV2 operation with
// real AWS credentials.
func TestIntegration_S3ListObjects(t *testing.T) {
	ctx := context.Background()

	cfg, err := LoadAWSConfig(ctx, WithRegion("us-east-1"))
	require.NoError(t, err)

	client := s3v2.NewFromConfig(cfg)

	bucketName := fmt.Sprintf("tfctlgo-list-%d", time.Now().UnixNano())

	// Create bucket
	_, err = client.CreateBucket(ctx, &s3v2.CreateBucketInput{
		Bucket: awsv2.String(bucketName),
	})
	require.NoError(t, err)
	defer func() {
		client.DeleteBucket(ctx, &s3v2.DeleteBucketInput{
			Bucket: awsv2.String(bucketName),
		})
	}()

	// Put multiple objects
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("object-%d.txt", i)
		_, err := client.PutObject(ctx, &s3v2.PutObjectInput{
			Bucket: awsv2.String(bucketName),
			Key:    awsv2.String(key),
			Body: bytes.NewReader(
				[]byte(fmt.Sprintf("content-%d", i)),
			),
		})
		require.NoError(t, err)
	}

	// List objects
	result, err := client.ListObjectsV2(ctx, &s3v2.ListObjectsV2Input{
		Bucket: awsv2.String(bucketName),
	})
	require.NoError(t, err)

	assert.NotNil(t, result.Contents)
	assert.Equal(t, 3, len(result.Contents))
}

// TestIntegration_S3HeadBucket verifies S3 HeadBucket operation with real
// AWS credentials.
func TestIntegration_S3HeadBucket(t *testing.T) {
	ctx := context.Background()

	cfg, err := LoadAWSConfig(ctx, WithRegion("us-east-1"))
	require.NoError(t, err)

	client := s3v2.NewFromConfig(cfg)

	bucketName := fmt.Sprintf("tfctlgo-head-%d", time.Now().UnixNano())

	// Create bucket
	_, err = client.CreateBucket(ctx, &s3v2.CreateBucketInput{
		Bucket: awsv2.String(bucketName),
	})
	require.NoError(t, err)
	defer func() {
		client.DeleteBucket(ctx, &s3v2.DeleteBucketInput{
			Bucket: awsv2.String(bucketName),
		})
	}()

	// Head bucket
	_, err = client.HeadBucket(ctx, &s3v2.HeadBucketInput{
		Bucket: awsv2.String(bucketName),
	})
	assert.NoError(t, err)

	// Head non-existent bucket
	nonexistentBucket := fmt.Sprintf("nonexistent-%d", time.Now().UnixNano())
	_, err = client.HeadBucket(ctx, &s3v2.HeadBucketInput{
		Bucket: awsv2.String(nonexistentBucket),
	})
	assert.Error(t, err)
}

// TestIntegration_S3CopyObject verifies S3 CopyObject operation with real
// AWS credentials.
func TestIntegration_S3CopyObject(t *testing.T) {
	ctx := context.Background()

	cfg, err := LoadAWSConfig(ctx, WithRegion("us-east-1"))
	require.NoError(t, err)

	client := s3v2.NewFromConfig(cfg)

	bucketName := fmt.Sprintf("tfctlgo-copy-%d", time.Now().UnixNano())
	sourceKey := "source.txt"
	destKey := "destination.txt"

	// Create bucket
	_, err = client.CreateBucket(ctx, &s3v2.CreateBucketInput{
		Bucket: awsv2.String(bucketName),
	})
	require.NoError(t, err)
	defer func() {
		client.DeleteBucket(ctx, &s3v2.DeleteBucketInput{
			Bucket: awsv2.String(bucketName),
		})
	}()

	// Put source object
	_, err = client.PutObject(ctx, &s3v2.PutObjectInput{
		Bucket: awsv2.String(bucketName),
		Key:    awsv2.String(sourceKey),
		Body:   bytes.NewReader([]byte("source content")),
	})
	require.NoError(t, err)

	// Copy object
	copySource := bucketName + "/" + sourceKey
	_, err = client.CopyObject(ctx, &s3v2.CopyObjectInput{
		Bucket:     awsv2.String(bucketName),
		Key:        awsv2.String(destKey),
		CopySource: awsv2.String(copySource),
	})
	require.NoError(t, err)

	// Verify destination exists
	result, err := client.GetObject(ctx, &s3v2.GetObjectInput{
		Bucket: awsv2.String(bucketName),
		Key:    awsv2.String(destKey),
	})
	require.NoError(t, err)
	result.Body.Close()
}

// TestIntegration_S3MultiRegionConfig verifies config with different
// region settings and client creation.
func TestIntegration_S3MultiRegionConfig(t *testing.T) {
	ctx := context.Background()
	testRegions := []string{"us-east-1", "eu-west-1", "ap-southeast-1"}

	for _, testRegion := range testRegions {
		t.Run(fmt.Sprintf("region-%s", testRegion), func(t *testing.T) {
			cfg, err := LoadAWSConfig(ctx, WithRegion(testRegion))
			require.NoError(t, err)

			client := NewS3(cfg)

			// Client should be created successfully
			assert.NotNil(t, client)
			assert.Equal(t, testRegion, cfg.Region)
		})
	}
}
