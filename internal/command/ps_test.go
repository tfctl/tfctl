package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlanOutput(t *testing.T) {
	input := `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
  ~ update in-place
 <= read (data resources)

Terraform will perform the following actions:

  # module.myapp.aws_s3_bucket.bucket will be created
  + resource "aws_s3_bucket" "bucket" {
      + bucket = "my-bucket"
    }

  # aws_instance.web will be updated in-place
  ~ resource "aws_instance" "web" {
      ~ instance_type = "t2.micro" -> "t3.micro"
    }

data.aws_caller_identity.validator: Reading...
module.data.aws_caller_identity.validator: Reading...
module.foo.data.bar: Reading...

Plan: 1 to add, 1 to change, 0 to destroy.
`
	reader := strings.NewReader(input)
	resources, err := parsePlanOutput(reader)
	assert.NoError(t, err)

	expected := []PlanResource{
		{Resource: "module.myapp.aws_s3_bucket.bucket", Action: "created"},
		{Resource: "aws_instance.web", Action: "updated in-place"},
		{Resource: "data.aws_caller_identity.validator", Action: "read"},
		{Resource: "module.data.aws_caller_identity.validator", Action: "read"},
		{Resource: "module.foo.data.bar", Action: "read"},
	}

	assert.Equal(t, expected, resources)
}
