# Axiom CLI Reference

**Documentation:** https://axiom.co/docs/reference/cli

## Configuration & Deployments

The CLI reads configuration from `~/.axiom.toml`. Environment variables (`AXIOM_TOKEN`, `AXIOM_URL`, `AXIOM_ORG_ID`) override config values.

### Introspect Configuration

```bash
# Active deployment name
axiom config get active_deployment

# Deployment URL and org ID (replace <name> with deployment alias)
axiom config get "deployments.<name>.url"
axiom config get "deployments.<name>.org_id"
```

**Do NOT use `config get` to read tokens.** If you need to export credentials (e.g., to `.env` or `.envrc` files), use `eval $(axiom config export --force)` instead.

### Switch Deployments

```bash
# Per-command override (does not persist)
axiom -D staging query "['logs'] | limit 10" --start-time -1h

# Persistent switch
axiom auth select <alias>

# Switch organization within current deployment
axiom auth switch-org <org-id>
```

### Verify Connectivity

```bash
# Check all deployments
axiom auth status

# Check specific deployment
axiom auth status <alias>
```

### Self-Discovery

```bash
axiom help credentials    # Token types and authentication guidance
axiom help environment    # All supported environment variables
axiom --help              # Top-level command overview
axiom <command> --help    # Subcommand details
```

## Query Execution

```bash
axiom query "<APL>" [flags]
```

### Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--start-time` | Query start time | `-7d`, `-1h`, `2024-01-15` |
| `--end-time` | Query end time (default: now) | `-30m`, `2024-01-16` |
| `-f, --format` | Output format | `json`, `table` (default) |
| `--fail-on-empty` | Exit code 1 if no results | (flag, no value) |

### Time Range Formats

```bash
# Relative (recommended)
--start-time -20m             # 20 minutes ago
--start-time -1h              # 1 hour ago
--start-time -24h             # 24 hours ago
--start-time -7d              # 7 days ago
--start-time -2w              # 2 weeks ago

# Mixed relative
--start-time -1d12h           # 1 day and 12 hours ago

# Absolute
--start-time 2024-01-15       # Date
--start-time 2024-01-15T10:00:00Z  # ISO timestamp
```

Supported relative units: `w` (week), `d` (day), `h` (hour), `m` (minute), `s` (second), `ms`, `us`, `ns`.

### Output Formats

| Format | Flag | Use Case |
|--------|------|----------|
| Table | `-f table` | Human-readable (default) |
| JSON | `-f json` | Parsing, scripting |

`-f json` writes one JSON object per row to stdout.

### Diagnostics Go To stderr

The command writes results to stdout and diagnostics to stderr. Diagnostics
are warnings from the server, and the trace id of a failed request.

Read stdout alone when you parse the output:

```bash
# CORRECT - stdout stays valid JSON
axiom query "['logs'] | limit 10" -f json --start-time -1h 2>/dev/null | jq .

# WRONG - the warning lines break the parser
axiom query "['logs'] | limit 10" -f json --start-time -1h 2>&1 | jq .
```

Two warnings change what the result means:

| Warning | Meaning | Fix |
|---------|---------|-----|
| `! A default limit of 1000 was applied` | The result is truncated. | Add `\| limit N` or aggregate. |
| `! Results are estimated` | The counts are approximate. A `top` over a high-cardinality group does this. | Narrow the range or the group. |

### Trace ID

A failed query prints the trace id of the request on its own line:

```
Error: API error 400: ...
Trace: a182bfccdb785f26a6ae6e1ab7a9d1d9
```

Give this id to Axiom support. It identifies the request server side.

### Examples

```bash
# Simple query
axiom query "['logs'] | limit 10" --start-time -1h

# JSON output for parsing
axiom query "['logs'] | summarize count() by status" -f json --start-time -1h

# Specific time range
axiom query "['logs'] | where error == true" --start-time -24h --end-time -1h

# Query against a different deployment
axiom -D prod query "['logs'] | count" --start-time -1h
```

## Dataset Operations

```bash
# List all datasets
axiom dataset list
axiom dataset list -f json

# Aliases
axiom dataset ls
```

## Livestream

```bash
# Stream live data (no filtering; dataset argument required)
axiom stream <dataset>
axiom stream logs
```

## Query Best Practices

1. **Always specify time range** with `--start-time`. Without it the query
   covers the last 10 years, and `bin_auto` buckets by year. The default limit
   bounds the rows returned, not the rows read.
2. **Use JSON for parsing** with `-f json` for programmatic access
3. **Prefer aggregations** to reduce data volume returned
4. **Limit results** with `| limit N` during exploration
5. **Discover schema first** with `| getschema` before writing queries
