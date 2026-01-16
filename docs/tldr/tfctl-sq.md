# tfctl-sq

> Query Terraform state files.

- Display state file in current directory and include Created At information.

`tfctl sq --attrs created-at`

- Display state file in current directory and include column headings.

`tfctl sq --titles`

- Display only concrete resources with "vpc" in their type or name.

`tfctl sq --concrete --filter 'resource@vpc'`

- Display the third most recent state file version in JSON format..

`tfctl sq --sv -3 --output json`
