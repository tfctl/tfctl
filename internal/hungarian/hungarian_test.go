// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package hungarian

import (
	"testing"
)

func TestIsHungarian(t *testing.T) {
	tests := []struct {
		name     string
		typ      string
		resName  string
		expected bool
	}{
		// Token equality tests - name part matches type token exactly.
		{
			name:     "instance token in name",
			typ:      "aws_instance",
			resName:  "instance_prod",
			expected: true,
		},
		{
			name:     "s3 token in name",
			typ:      "aws_s3_bucket",
			resName:  "s3_logs",
			expected: true,
		},
		{
			name:     "bucket token in name",
			typ:      "aws_s3_bucket",
			resName:  "my_bucket",
			expected: true,
		},
		{
			name:     "vpc token in name",
			typ:      "aws_vpc",
			resName:  "vpc_main",
			expected: true,
		},
		{
			name:     "lambda token in name",
			typ:      "aws_lambda_function",
			resName:  "lambda_handler",
			expected: true,
		},
		{
			name:     "function token in name",
			typ:      "aws_lambda_function",
			resName:  "my_function",
			expected: true,
		},
		// Substring tests - type token appears as substring in name.
		{
			name:     "s3 as substring with underscore separator",
			typ:      "aws_s3_bucket",
			resName:  "my_s3_logs",
			expected: true,
		},
		{
			name:     "s3 as substring with dash separator",
			typ:      "aws_s3_bucket",
			resName:  "my-s3-logs",
			expected: true,
		},
		{
			name:     "vpc as substring with underscore separator",
			typ:      "aws_vpc",
			resName:  "my_vpc_main",
			expected: true,
		},
		{
			name:     "vpc as substring with dash separator",
			typ:      "aws_vpc",
			resName:  "my-vpc-main",
			expected: true,
		},
		{
			name:     "instance as substring with underscore separator",
			typ:      "aws_instance",
			resName:  "my_instance_server",
			expected: true,
		},
		{
			name:     "instance as substring with dash separator",
			typ:      "aws_instance",
			resName:  "my-instance-server",
			expected: true,
		},
		// Case insensitivity tests.
		{
			name:     "uppercase type token",
			typ:      "AWS_INSTANCE",
			resName:  "instance_prod",
			expected: true,
		},
		{
			name:     "mixed case resource name",
			typ:      "aws_instance",
			resName:  "Instance_Prod",
			expected: true,
		},
		{
			name:     "mixed case both",
			typ:      "AWS_S3_BUCKET",
			resName:  "S3_Logs",
			expected: true,
		},
		// Non-Hungarian tests.
		{
			name:     "no matching tokens",
			typ:      "aws_security_group",
			resName:  "sg_app",
			expected: false,
		},
		{
			name:     "name has no meaningful tokens",
			typ:      "aws_instance",
			resName:  "my_prod_server",
			expected: false,
		},
		{
			name:     "aws token not in specific name",
			typ:      "aws_s3_bucket",
			resName:  "my_data_store",
			expected: false,
		},
		// Edge cases.
		{
			name:     "empty type",
			typ:      "",
			resName:  "something",
			expected: false,
		},
		{
			name:     "empty name",
			typ:      "aws_instance",
			resName:  "",
			expected: false,
		},
		{
			name:     "both empty",
			typ:      "",
			resName:  "",
			expected: false,
		},
		{
			name:     "aws token in name",
			typ:      "aws_instance",
			resName:  "aws_instance",
			expected: true,
		},
		{
			name:     "aws as substring",
			typ:      "aws_instance",
			resName:  "myawsserver",
			expected: true,
		},
		// Underscore handling.
		{
			name:     "type with multiple underscores",
			typ:      "aws_s3_bucket",
			resName:  "s3bucket",
			expected: true,
		},
		{
			name:     "name with multiple underscores",
			typ:      "aws_s3_bucket",
			resName:  "my_s3_bucket_name",
			expected: true,
		},
		// Dash and other separators.
		{
			name:     "dash separated name",
			typ:      "aws_security_group",
			resName:  "security-group-main",
			expected: true,
		},
		{
			name:     "dot separated name",
			typ:      "aws_instance",
			resName:  "instance.local.prod",
			expected: true,
		},
		{
			name:     "mixed separators",
			typ:      "aws_lambda_function",
			resName:  "my-lambda.function_handler",
			expected: true,
		},
		// Special characters and edge separators.
		{
			name:     "name with numbers",
			typ:      "aws_s3_bucket",
			resName:  "s3bucket123",
			expected: true,
		},
		{
			name:     "type token is number-like",
			typ:      "test_1_resource",
			resName:  "1_something",
			expected: true,
		},
		// Real Terraform examples.
		{
			name:     "real aws_resourcegroups_group",
			typ:      "aws_resourcegroups_group",
			resName:  "lab-myapp-zakp-100",
			expected: false,
		},
		{
			name:     "real aws_s3_bucket with s3",
			typ:      "aws_s3_bucket",
			resName:  "my-s3-bucket",
			expected: true,
		},
		{
			name:     "real aws_vpc",
			typ:      "aws_vpc",
			resName:  "main_vpc",
			expected: true,
		},
		{
			name:     "real aws_security_group with group",
			typ:      "aws_security_group",
			resName:  "app_group",
			expected: true,
		},
		// Multiple tokens from type matching.
		{
			name:     "first token matches",
			typ:      "aws_s3_bucket",
			resName:  "aws_storage",
			expected: true,
		},
		{
			name:     "second token matches",
			typ:      "aws_s3_bucket",
			resName:  "my_s3_store",
			expected: true,
		},
		{
			name:     "third token matches",
			typ:      "aws_s3_bucket",
			resName:  "storage_bucket",
			expected: true,
		},
		// CamelCase tests.
		{
			name:     "s3 with camelCase s3BucketName",
			typ:      "aws_s3_bucket",
			resName:  "s3BucketName",
			expected: true,
		},
		{
			name:     "bucket with camelCase s3BucketName",
			typ:      "aws_s3_bucket",
			resName:  "s3BucketName",
			expected: true,
		},
		{
			name:     "instance with camelCase MyInstance",
			typ:      "aws_instance",
			resName:  "MyInstance",
			expected: true,
		},
		{
			name:     "instance with camelCase myInstanceServer",
			typ:      "aws_instance",
			resName:  "myInstanceServer",
			expected: true,
		},
		{
			name:     "vpc with camelCase MainVpc",
			typ:      "aws_vpc",
			resName:  "MainVpc",
			expected: true,
		},
		{
			name:     "lambda with camelCase MyLambdaFunction",
			typ:      "aws_lambda_function",
			resName:  "MyLambdaFunction",
			expected: true,
		},
		{
			name:     "no match with camelCase MyDataStore",
			typ:      "aws_security_group",
			resName:  "MyDataStore",
			expected: false,
		},
		{
			name:     "camelCase with mixed separators s3-MyBucket_name",
			typ:      "aws_s3_bucket",
			resName:  "s3-MyBucket_name",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHungarian(tt.typ, tt.resName)
			if result != tt.expected {
				t.Errorf("IsHungarian(%q, %q) = %v, expected %v",
					tt.typ, tt.resName, result, tt.expected)
			}
		})
	}
}

// BenchmarkIsHungarian benchmarks the IsHungarian function to ensure it
// performs well with typical resource names and types.
func BenchmarkIsHungarian(b *testing.B) {
	tests := []struct {
		name    string
		typ     string
		resName string
	}{
		{"simple", "aws_instance", "instance_prod"},
		{"complex", "aws_s3_bucket", "my-data-store-bucket"},
		{"many_tokens", "aws_ec2_network_interface_attachment", "network_interface_attachment_eth0"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsHungarian(tt.typ, tt.resName)
			}
		})
	}
}
