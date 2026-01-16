# NAME

tfctl svq — State Version Query

# DESCRIPTION

Query Terraform state version history.

# USAGE

`tfctl svq [RootDir] [options]`


# OPTIONS

| Flag | Default | Description |
|------| ------- |-------------|
| --attrs/-a string |  | Comma-separated list of attributes to include in results. [More info](#attrs)|
| --color/-c |  | Colorize text output. |
| --filter/-f string |  | Comma-separated list of filters to apply to results |
| --help |  | Show command help. |
| --host/-h string | app.terraform.io | HCP/TFE host to use. Overrides the backend. |
| --limit/-l int | unlimited | Limit the result set to `int` results. |
| --local/-l |  | Show timestamps in local time according to your OS. |
| --org string |  | Organization to use for all commands. Overrides the backend. |
| --output/-o string |  | Output format (text (default), json, yaml, csv). [More info](#output)|
| --sort/-s string |  | Comma-separated list of fields to sort by. [More info](#sort)|
| --titles |  | Include column headings in text output. |
| --tldr |  | Show tldr page if a client is installed. |
| --workspace/-w string | default | Workspace to use for all commands. Overrides the backend. |



  ## attrs
  There are many possible attributes depending on the command. Common attributes include: `id`, `type`, `name`, `created-at`, `updated-at`, etc. But, each command has it's own schema. Other than the `sq` command, use the `--schema` flag to see the full list of available attributes.

There is also a feature-rich syntax for transforming attribute values. See [Attributes](../attrs.md) for details.

  ## output
  The `text` output format is a human-friendly table format. The `json`, `yaml`, and `csv` formats are machine-friendly and can be used for further processing.

  ## sort
  Prefix field names with `-` to sort in descending order.  For example, `--sort=-created-at,name` sorts first by Created At (newest first), then by Name (A-Z).




# EXAMPLES

**Display state file history for current directory and include Created At information.**

```sh
tfctl svq --attrs created-at
```


**Display the five most recent state file versions and include the YYYY-MM-DD portion of the Created At information.**

```sh
tfctl svq --limit 5 --attrs created-at::10
```


