# tfctl Filters

The `--filter` flag narrows query results using a small, expressive syntax. This page documents the syntax and gives examples.

## Overview

- The default delimiter between multiple filters is a comma (`,`).
- You can override the delimiter by setting the environment variable `TFCTL_FILTER_DELIM`.
- Each filter has the form: `key<operand>target` where `<operand>` is one of the supported operators described below.
- Prefixing an operand with `!` negates the operator (for example `name!@prod` means "name does NOT contain 'prod'").
- Filters referencing attribute keys that start with `_` are treated as server-side API filters and are ignored by the local filter engine.

## Supported operands

- `=`  : exact match (string equality)
- `~`  : case-insensitive equality
- `^`  : starts-with
- `>`  : lexicographic greater-than
- `<`  : lexicographic less-than
- `@`  : contains / membership (works for substrings and collections)
- `/`  : regular expression match

Negation is expressed by placing `!` before the operand.

## Special filters

- `hungarian` : Detects resources following Hungarian notation naming conventions (applies to `sq` command only)
  - Example: `tfctl sq --filter hungarian=true` — shows only resources with names following Hungarian notation
  - Supported values: `true`, `false`, or bare `hungarian` (equivalent to `hungarian=true`)
  - The filter analyzes resource type and name to detect common prefix patterns (e.g., `s3Bucket`, `ec2Instance`, `iamRole`)
  - Case-insensitive and supports underscores and dashes in names

## Delimiter and quoting

If a target contains the delimiter character, quote the whole filter or choose a different delimiter with `TFCTL_FILTER_DELIM`.

## Behavior notes and edge cases

- If a filter references an attribute name that doesn't exist in the discovered attribute list, the filter check will fail and an error will be logged.
- For `@` (contains), the implementation supports strings, slices, and maps:
  - For strings it does a substring test.
  - For slices it tests for membership.
  - For maps it checks for a key.
- For `/` (regex), the pattern uses Go's `regexp` package syntax. Invalid regular expressions will log an error and exclude the item from results.
- Filters are evaluated before attribute transformations are applied (so transformations in `--attrs` won't affect filter matching).
- When using `sq` with `--concrete`, tfctl automatically appends `mode=managed` to the filter set.

## Examples

**Basic filtering:**
```bash
# Simple contains
tfctl oq --filter 'name@prod'

# Negation (not contains)
tfctl oq --filter 'name!@prod'

# Exact match
tfctl wq --filter 'status=applied'

# Case-insensitive equality
tfctl oq --filter 'email~admin'

# Starts with
tfctl wq --filter 'name^prod'
```

**Regular expression filtering:**
```bash
# Find workspaces matching a naming pattern (prod-001, prod-002, etc.)
tfctl wq --filter 'name/^prod-\d{3}$'

# Find resources with version numbers
tfctl sq --filter 'version/^\d+\.\d+\.\d+$'

# Match email addresses in organization attributes
tfctl oq --filter 'email/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
```

**Complex filtering:**
```bash
# Multiple filters (comma-delimited)
tfctl oq --filter 'name@prod,created-at>2024-01-01'

# Find items that do not have a given tag
tfctl mq --filter 'tags!@deprecated'

# Combine different operators
tfctl wq --filter 'name^prod,status=applied,!description='

# Find resources with Hungarian notation naming (sq only)
tfctl sq --filter 'hungarian=true'

# Find resources NOT using Hungarian notation (sq only)
tfctl sq --filter 'hungarian=false'
```


For implementation details, see the `FilterDataset` and `BuildFilters` functions in the project source (internal/output/filters.go).
