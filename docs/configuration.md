# Configuration

tfctl configuration is made in a YAML file. Environment variables can override certain settings at runtime.

## Configuration File

tfctl reads its configuration from a YAML file. The file location is determined by the following precedence:

1. `TFCTL_CFG_FILE` environment variable (if set and non-empty)
2. OS-specific user config directory:
   - Linux/Unix: `$HOME/.config/tfctl/tfctl.yaml`
   - macOS: `$HOME/Library/Application Support/tfctl/tfctl.yaml`
   - Windows: `%APPDATA%\tfctl\tfctl.yaml`

If the specified file is not found or cannot be parsed, tfctl will error.

### Configuration Structure

See [tfctl.yaml](tfctl.yaml) for a complete reference of all available configuration options, including command-specific settings and defaults.

## Environment Variable Overrides

The following environment variables override configuration file settings at runtime:

### `TFCTL_CFG_FILE`

Specifies the full path to a tfctl configuration file in YAML format.

**Usage:**
```bash
export TFCTL_CFG_FILE=$HOME/.config/tfctl/prod.yaml
tfctl sq
```

### `TFCTL_CACHE`

Controls whether tfctl caches query results. Caching is enabled by default.

**Valid values:**
- Not set or empty: Caching enabled.
- `0` or `false`: Caching disabled.
- Any other value: Caching enabled.

**Usage:**
```bash
# Disable caching for this invocation
TFCTL_CACHE=0 tfctl sq

# Re-enable caching
unset TFCTL_CACHE
```

### `TFCTL_CACHE_DIR`

Specifies a custom directory for storing cached query results.

**Usage:**
```bash
export TFCTL_CACHE_DIR=/mnt/fast-storage/tfctl-cache
tfctl sq
```

**Precedence:**
1. `TFCTL_CACHE_DIR` (if set and non-empty)
2. OS-specific user cache directory (see Configuration File section above)

## Examples

### Use a custom config file and cache directory

```bash
export TFCTL_CFG_FILE=$HOME/.tfctl-prod.yaml
export TFCTL_CACHE_DIR=$HOME/.cache/tfctl-prod
tfctl oq
```

### Disable caching for a single command

```bash
TFCTL_CACHE=0 tfctl sq --attrs arn
```

### Use a shared cache on a network drive

```bash
export TFCTL_CACHE_DIR=/mnt/shared/tfctl-cache
tfctl wq
```

### Override all settings

```bash
TFCTL_CFG_FILE=/etc/tfctl/production.yaml \
TFCTL_CACHE=0 \
tfctl pq --sort created-at
```