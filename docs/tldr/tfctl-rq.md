# tfctl-rq

> Query HCP/TFE runs for the given workspace.

- Display runs and include Created At and Status information.

`tfctl rq --attrs created-at,status`

- Display errored runs in the "prod" workspace of the "hr" org.

`tfctl rq --org hr --workspace prod --filter 'status@errored'`
