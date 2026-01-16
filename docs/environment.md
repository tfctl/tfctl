# Environment Variables

tfctl supports several environment variables for configuration and runtime behavior.

## Configuration

### `TFCTL_CFG_FILE`

The full path to a tfctl configuration file in YAML format.

**Usage:**
```bash
export TFCTL_CFG_FILE=$HOME/.config/tfctl/config.yaml
tfctl oq
```

**Behavior:**
- If set, tfctl uses this as the configuration file path.
- The file must exist and be a regular file (not a directory).
- If the file is not found or cannot be parsed, tfctl will error.
- If not set, tfctl looks for `tfctl.yaml` in the standard OS-specific user config directory (e.g., `$HOME/.config/tfctl/tfctl.yaml` on Linux, `$HOME/Library/Application Support/tfctl/tfctl.yaml` on macOS).

**Example configuration file:**
```yaml
# ~/.config/tfctl/config.yaml
cache:
  clean: 24      # Purge cache files older than 24 hours
  dir: ""        # Use default cache location

backend:
  s3:
    region: us-east-1
    bucket: my-terraform-state

org: my-org      # Default organization for queries
```

## Caching

### `TFCTL_CACHE`

Controls whether tfctl caches query results. Caching is enabled by default.

**Valid values:**
- Not set or empty: Caching enabled.
- `0` or `false`: Caching disabled.
- Any other value: Caching enabled.

**Usage:**
```bash
# Disable caching for this invocation
export TFCTL_CACHE=0
tfctl sq

# Re-enable caching
unset TFCTL_CACHE
tfctl sq
```

**Behavior:**
- When disabled, query results are not cached and existing cache entries are not used.
- Cache cleanup operations are still executed when caching is disabled.

### `TFCTL_CACHE_DIR`

Specifies a custom directory for storing cached query results.

**Usage:**
```bash
# Use a custom cache directory
export TFCTL_CACHE_DIR=/mnt/fast-storage/tfctl-cache
tfctl sq

# Use default cache directory
unset TFCTL_CACHE_DIR
tfctl sq
```

**Behavior:**
- If set and non-empty, tfctl uses this directory for all cache files.
- If not set or empty, tfctl uses the OS-specific user cache directory (e.g., `$HOME/.cache/tfctl` on Linux, `$HOME/Library/Caches/tfctl` on macOS).
- The directory is created automatically if it doesn't exist.
- Cache files are stored with permissions `0600` (user read/write only).

**Precedence:**
1. `TFCTL_CACHE_DIR` (if set and non-empty)
2. `$XDG_CACHE_HOME/tfctl` (if `XDG_CACHE_HOME` is set)
3. `$HOME/.cache/tfctl` (Linux/Unix default)
4. `$HOME/Library/Caches/tfctl` (macOS default)
5. `%LOCALAPPDATA%\tfctl\cache` (Windows default)

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

### Override all configuration

```bash
# Use specific config and cache, disable caching
TFCTL_CFG_FILE=/etc/tfctl/production.yaml \
TFCTL_CACHE=0 \
tfctl pq --sort created-at
```
