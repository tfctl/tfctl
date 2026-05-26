# tfctl-oq

> Query HCP/TFE organizations.
> Also available as: `tfctl org`

- Display organizations and include Created At information.

`tfctl oq --attrs created-at`

- Display organizations with "myorg" in their name.

`tfctl oq --filter 'name@myorg'`
