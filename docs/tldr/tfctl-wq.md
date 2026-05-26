# tfctl-wq

> Query HCP/TFE workspaces.
> Also available as: `tfctl workspace`

- Display workspaces and include Created At information.

`tfctl wq --attrs created-at`

- Display workspaces in the "hr" org with "prod" in their name.

`tfctl wq --org hr --filter 'name@prod'`
