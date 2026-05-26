# tfctl-pq

> Query HCP/TFE projects.
> Also available as: `tfctl project`

- Display projects and include Updated At information.

`tfctl pq --attrs updated-at`

- Display projects in the "hr" org with "prod" in their name.

`tfctl pq --org hr --filter 'name@prod'`
