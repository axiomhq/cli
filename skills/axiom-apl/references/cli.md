# Axiom CLI Reference

**Documentation:** https://axiom.co/docs/reference/cli

## Query Execution

```bash
axiom query "<APL>" [flags]
```

### Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--start-time` | Query start time | `-1h`, `2024-01-15` |
| `--end-time` | Query end time (default: now) | `-30m`, `2024-01-16` |
| `-f, --format` | Output format | `json`, `table` (default) |

### Time Range Formats

```bash
# Relative (recommended)
--start-time -1h              # 1 hour ago
--start-time -24h             # 24 hours ago
--start-time -7d              # 7 days ago

# Absolute
--start-time 2024-01-15       # Date
--start-time 2024-01-15T10:00:00Z  # ISO timestamp
```

### Output Formats

| Format | Flag | Use Case |
|--------|------|----------|
| Table | `-f table` | Human-readable (default) |
| JSON | `-f json` | Parsing, scripting |

### Examples

```bash
# Simple query
axiom query "['logs'] | limit 10" --start-time -1h

# JSON output for parsing
axiom query "['logs'] | summarize count() by status" -f json --start-time -1h

# Specific time range
axiom query "['logs'] | where error == true" --start-time -24h --end-time -1h
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
# Stream live data (no filtering - use query for filtered results)
axiom stream <dataset>
axiom stream logs
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AXIOM_TOKEN` | API token (overrides config) |
| `AXIOM_ORG_ID` | Organization ID |
| `AXIOM_URL` | API URL (for self-hosted) |

## Query Best Practices

1. **Always specify time range** - Use `--start-time` to limit data scanned
2. **Use JSON for parsing** - `-f json` for programmatic access
3. **Prefer aggregations** - Reduce data returned when possible
4. **Limit results** - Use `| limit N` during exploration
