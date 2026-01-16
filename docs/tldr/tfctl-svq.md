# tfctl-svq

> Query Terraform state version history.

- Display state file history for current directory and include Created At information.

`tfctl svq --attrs created-at`

- Display the five most recent state file versions and include the YYYY-MM-DD portion of the Created At information.

`tfctl svq --limit 5 --attrs created-at::10`
