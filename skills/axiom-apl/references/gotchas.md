# Common APL Gotchas

Mistakes that cause errors or unexpected results, and how to fix them.

## Bracket Notation

### Problem: Dotted field names without brackets

```apl
// WRONG - syntax error or wrong field
| where service.name == "api"
| project status.code, service.name

// CORRECT - bracket notation required
| where ['service.name'] == "api"
| project ['status.code'], ['service.name']
```

**Rule:** Any field with dots, dashes, or spaces needs `['field.name']` notation.

### When to use brackets

| Field | Bracket Required | Example |
|-------|------------------|---------|
| `trace_id` | No | `trace_id == "abc"` |
| `service.name` | **Yes** | `['service.name'] == "api"` |
| `http-status` | **Yes** | `['http-status'] >= 500` |
| `my field` | **Yes** | `['my field']` |

## Type Mismatches

### Problem: Comparing wrong types

```apl
// WRONG - comparing string to integer
| where ['status.code'] == 2

// CORRECT - status.code is a string
| where ['status.code'] == "ERROR"
```

### Problem: Arithmetic on strings

```apl
// WRONG - duration might be stored as string
| extend duration_ms = duration / 1000

// CORRECT - convert first
| extend duration_ms = tolong(duration) / 1000
```

**Rule:** Use `getschema` to check field types before operating on them.

## Duration Units

### Problem: Assuming wrong time units

OTel traces store duration in **nanoseconds**, not milliseconds.

```apl
// WRONG - looking for 1ms threshold
| where duration > 1

// CORRECT - 1ms = 1,000,000 nanoseconds
| where duration > 1000000

// BETTER - explicit conversion
| extend duration_ms = duration / 1000000.0
| where duration_ms > 1
```

**Conversion table:**

| Target | Divisor | Example |
|--------|---------|---------|
| Nanoseconds | 1 | `duration` |
| Microseconds | 1,000 | `duration / 1000` |
| Milliseconds | 1,000,000 | `duration / 1000000` |
| Seconds | 1,000,000,000 | `duration / 1000000000` |

## Time Filtering

### Problem: Missing time filter

```apl
// WRONG - scans entire dataset (expensive!)
['logs']
| where status >= 500

// CORRECT - always filter by time first
['logs']
| where _time >= ago(1h)
| where status >= 500
```

**Rule:** Always include time filter. Put it first for best performance.

### Problem: Wrong time syntax

```apl
// WRONG - missing ago()
| where _time >= 1h

// CORRECT - use ago() for relative time
| where _time >= ago(1h)

// ALSO CORRECT - explicit datetime
| where _time >= datetime(2024-01-15)
```

## Null Handling

### Problem: Null comparisons fail silently

```apl
// WRONG - rows with null error field are excluded
| where error == true

// CORRECT - handle nulls explicitly
| where coalesce(error, false) == true

// OR use ensure_field
| extend error = coalesce(ensure_field("error", typeof(bool)), false)
| where error == true
```

## Aggregation Issues

### Problem: Counting vs distinct counting

```apl
// Counts all rows
| summarize count() by user_id

// Counts unique values
| summarize dcount(user_id)
```

### Problem: Missing group-by context

```apl
// WRONG - avg over entire result
| extend avg_duration = avg(duration)

// CORRECT - use summarize for aggregations
| summarize avg_duration = avg(duration) by service
```

## Search Performance

### Problem: Using expensive search

```apl
// SLOW - scans all fields
search "error"

// FASTER - target specific field
| where message contains "error"

// FASTEST - case-sensitive word match
| where message has_cs "error"
```

**Performance ranking:**
1. `has_cs` (fastest) - case-sensitive word boundary
2. `has` - word boundary match
3. `contains_cs` - case-sensitive substring
4. `contains` - substring match
5. `search` (slowest) - full-text all fields

## Result Limits

### Problem: Unbounded queries

```apl
// WRONG - could return millions of rows
['logs']
| where _time >= ago(24h)

// CORRECT - always limit or aggregate
['logs']
| where _time >= ago(24h)
| limit 1000

// OR aggregate
['logs']
| where _time >= ago(24h)
| summarize count() by status
```

**Rule:** Either `limit` results or `summarize` to aggregate.

## Query Structure

### Problem: Wrong operator order

```apl
// WRONG - project before where loses fields
['logs']
| project _time, message
| where status >= 500  // status no longer exists!

// CORRECT - filter before projecting
['logs']
| where status >= 500
| project _time, message
```

**Recommended order:**
1. `where _time` - Time filter first
2. `where` - Other filters
3. `extend` - Add computed fields
4. `summarize` - Aggregate
5. `project` - Select final fields
6. `sort` - Order results
7. `limit` - Restrict count

## CLI Issues

### Problem: Quoting in shell

```bash
# WRONG - shell interprets special chars
axiom query ['logs'] | where status >= 500

# CORRECT - quote the entire query
axiom query "['logs'] | where status >= 500"

# For complex queries, use heredoc
axiom query "$(cat <<'EOF'
['logs']
| where _time >= ago(1h)
| where message contains "error"
| limit 100
EOF
)"
```

### Problem: Missing time range

```bash
# WRONG - uses default (might be too broad)
axiom query "['logs'] | limit 10"

# CORRECT - explicit time range
axiom query "['logs'] | limit 10" --start-time -1h
```

## OTel Field Mismatches

### Problem: Using standard OTel field paths

Axiom promotes some resource attributes to top level.

```apl
// WRONG - Axiom promotes these fields
| where ['resource.service.name'] == "api"
| where ['resource.telemetry.sdk.version'] == "1.0"

// CORRECT - No resource. prefix for promoted fields
| where ['service.name'] == "api"
| where ['telemetry.sdk.version'] == "1.0"
```

**Rule:** Check `getschema` output. See [otel.md](otel.md) for complete field mappings.

### Problem: Custom attribute access

Non-semconv attributes are stored in a map field.

```apl
// WRONG - field doesn't exist at this path
| where ['attributes.my_field'] == "value"

// CORRECT - access via custom map
| where ['attributes.custom']['my_field'] == "value"

// WRONG - aggregation without cast fails
| summarize count() by ['attributes.custom']['field']

// CORRECT - explicit cast required for aggregations
| summarize count() by tostring(['attributes.custom']['field'])
```
