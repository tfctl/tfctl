# NAME

tfctl sq — State Query

# DESCRIPTION

Query Terraform state files.

# USAGE

`tfctl sq [RootDir] [options]`


# OPTIONS

| Flag | Default | Description |
|------| ------- |-------------|
| --attrs/-a string |  | Comma-separated list of attributes to include in results. [More info](#attrs)|
| --chop |  | Chop common resource prefix from names. |
| --color/-c |  | Colorize text output. |
| --concrete/-k |  | Only include concrete resources. |
| --diff [ old[:new] \| + ] | CSV~1 CSV~0 | Find differences between state versions. [More info](#diff)|
| --filter/-f string |  | Comma-separated list of filters to apply to results |
| --help |  | Show command help. |
| --host/-h string | app.terraform.io | HCP/TFE host to use. Overrides the backend. |
| --local/-l |  | Show timestamps in local time according to your OS. |
| --org string |  | Organization to use for all commands. Overrides the backend. |
| --output/-o string |  | Output format (text (default), json, yaml, csv). [More info](#output)|
| --passphrase string |  | OpenTofu encrypted state passphrase. |
| --short | false | Shorten full resource name paths. |
| --sort/-s string |  | Comma-separated list of fields to sort by. [More info](#sort)|
| --sv string | latest | State version to query.  Can be a version number, `latest`, or `earliest`. [More info](#sv)|
| --titles |  | Include column headings in text output. |
| --tldr |  | Show tldr page if a client is installed. |
| --workspace/-w string | default | Workspace to use for all commands. Overrides the backend. |



  ## attrs
  There are many possible attributes depending on the command. Common attributes include: `id`, `type`, `name`, `created-at`, `updated-at`, etc. But, each command has it's own schema. Other than the `sq` command, use the `--schema` flag to see the full list of available attributes.

There is also a feature-rich syntax for transforming attribute values. See [Attributes](../attrs.md) for details.

  ## diff
  The `old` and `new` parameters can be version numbers, `latest`, or `earliest`. If `+` is provided, compares the latest version to the previous version.

  ## output
  The `text` output format is a human-friendly table format. The `json`, `yaml`, and `csv` formats are machine-friendly and can be used for further processing.

  ## sort
  Prefix field names with `-` to sort in descending order.  For example, `--sort=-created-at,name` sorts first by Created At (newest first), then by Name (A-Z).

  ## sv
  Can also be relativative. For example, `-1` is the previous version, `-2` is two versions ago, `+1` is the next version, and so on.


## NOTES

- `sq` operates against an IaC root directory (defaults to CWD when not provided).

- When using encrypted state, `sq` will prompt for a passphrase or use `TF_VAR_passphrase`.




# EXAMPLES

**Display state file in current directory and include Created At information.**

```sh
tfctl sq --attrs created-at
```


**Display state file in current directory and include column headings.**

```sh
tfctl sq --titles
```


**Display only concrete resources with "vpc" in their type or name.**

```sh
tfctl sq --concrete --filter 'resource@vpc'
```


**Display the third most recent state file version in JSON format..**

```sh
tfctl sq --sv -3 --output json
```


