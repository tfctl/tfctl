# tfctl Attributes

The `--attrs` flag is one of tfctl's most powerful features, allowing you to specify exactly which data fields to extract and how to format them. Understanding the `--attrs` syntax enables you to create precise results tailored to your needs.

## Prerequisites

To effectively use `--attrs`, you should be familiar with:
- **Terraform/OpenTofu state file structure** - Understanding how resources and their attributes are organized
- **JSON path notation** - How to navigate nested JSON objects using dot notation
- **Terraform API schemas** - The structure of data returned by HCP Terraform/Enterprise APIs

Use the `--schema` flag with any command (except `sq`) to explore available attributes.

## Basic Syntax

The `--attrs` flag accepts a comma-separated list of attribute specifications:

```
tfctl command --attrs spec1,spec2,spec3
```

Each specification has the format:
```
json_path:output_name:transform_spec
```

Where:
- `json_path` - The JSON path to extract (required)
- `output_name` - Column name in output (optional)
- `transform_spec` - Data transformation rules (optional)

## JSON Path Extraction

### Root vs Attributes Paths

**Attributes path (default):**
```sh
# Extracts from .attributes.email
tfctl oq --attrs email
```

**Root path (starts with `.`):**
```sh
# Extracts from .id (root level)
tfctl oq --attrs .id
```

**Nested paths:**
```sh
# Deep nested extraction
tfctl wq --attrs vcs-repo.identifier
tfctl pq --attrs permissions.can-create-workspaces
```

### Examples

**Organizations (`oq`):**
```sh
# Basic attributes
tfctl oq --attrs name,email,created-at

# Root level data
tfctl oq --attrs .id,.type

# Mixed paths
tfctl oq --attrs .id,email,external-id
```

**Workspaces (`wq`):**
```sh
# Workspace details
tfctl wq --attrs name,terraform-version,working-directory

# VCS information
tfctl wq --attrs vcs-repo.identifier,vcs-repo.branch

# Permissions and settings
tfctl wq --attrs auto-apply,queue-all-runs
```

## Output Naming

Control how columns appear in your output:

```sh
# Default: uses last segment of JSON path
tfctl oq --attrs created-at
# Output column: "created-at"

# Custom name: specify after colon
tfctl oq --attrs created-at:Created
# Output column: "Created"

# Multiple custom names
tfctl oq --attrs email:Admin,created-at:Date
# Output columns: "Admin", "Date"
```

## Data Transformations

Transform data as it's extracted using transformation specifications.

### Case Transformations

```sh
# Convert to uppercase
tfctl oq --attrs name::U

# Convert to lowercase
tfctl oq --attrs name::L

# Mixed transformations
tfctl oq --attrs name::U,email::L
```

### Length Transformations

```sh
# Truncate to first N characters
tfctl oq --attrs name::10
# "my-long-organization" → "my-long-or"

# Compress long strings (show beginning and end)
tfctl oq --attrs name::-8
# "my-long-organization" → "my-l..on"
```

### Time Transformations

```sh
# Convert UTC to local timezone (requires TZ environment variable)
tfctl oq --attrs created-at::t

# Example with TZ set:
export TZ="America/New_York"
tfctl oq --attrs created-at::t
# "2023-01-15T10:30:00Z" → "2023-01-15T05:30:00EST"
```

### Combined Transformations

```sh
# Multiple transforms applied in sequence
tfctl oq --attrs name::U10        # Uppercase, then truncate to 10 chars
tfctl oq --attrs created-at::tL    # Convert timezone, then lowercase
tfctl oq --attrs email:Admin:L15   # Custom name, lowercase, truncate to 15
```

## Advanced Usage

### Exclusion

Exclude attributes from processing (useful for filtering/sorting only):

```sh
# Include in processing but not in output
tfctl oq --attrs 'name,email' --filter 'name@prod'
```

### Global Transformations

Apply transformations to all attributes:

```sh
# Make all output uppercase
tfctl oq --attrs '*::U,name,email,created-at'
```

### Schema Discovery

Explore available attributes before crafting queries:

```sh
# Show all available attributes for organizations
tfctl oq --schema

# Common output:
# created-at
# email
# external-id
# name
# permissions
# collaborator-auth-policy
# ...
```

## Practical Examples

### Audit Report
```sh
# Create detailed org audit with custom formatting
tfctl oq --attrs name:Organization:U,email:Admin,created-at:Created::t \
  --output json > org_audit.json
```

### Resource Inventory
```sh
# Workspace inventory with VCS info
tfctl wq --attrs name::20,terraform-version:TF_Ver::U,vcs-repo.identifier:Repo::-30
```

### Filtered Extraction
```sh
# Get production workspace details
tfctl wq --filter 'name@prod' \
  --attrs name:Workspace,working-directory:Path,auto-apply:Auto::U
```

### State Analysis
```sh
# Analyze state file resources
tfctl sq --attrs type::15,name::25,provider::10
```

## Tips and Best Practices

1. **Start with `--schema`** - Always explore available attributes first.
2. **Use meaningful output names** - Make reports self-documenting.
3. **Test transformations** - Try transform specs on sample data first.
4. **Combine with filtering** - Use `--attrs` and `--filter` together for powerful queries.
5. **Consider output format** - JSON output preserves full data for further processing.

## Error Handling

**Invalid paths** - tfctl will show empty values for non-existent paths
**Invalid transforms** - Bad transformation specs are ignored
**Type mismatches** - Only string values can be transformed; others pass through unchanged

Understanding these attribute extraction patterns unlocks tfctl's full querying power, enabling you to extract exactly the data you need in the format you want.
