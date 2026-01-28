# APL Operators Reference

**Documentation:** https://axiom.co/docs/apl/tabular-operators

## Tabular Operators

### Filtering

#### where
Filter rows based on condition.

```apl
| where status >= 500
| where ['service.name'] == "api-gateway"
| where message contains "error"
| where _time >= ago(1h)
```

#### search
Full-text search across all fields. **Use sparingly - expensive.**

```apl
search "error" or "exception"
search in (['logs']) "timeout"
```

### Projection

#### project
Select and rename specific fields.

```apl
| project _time, status, message
| project timestamp=_time, code=status
```

#### project-away
Remove specific fields.

```apl
| project-away internal_id, debug_info
```

#### project-rename
Rename fields without changing selection.

```apl
| project-rename responseTime=['duration'], path=['url']
```

#### extend
Add calculated fields.

```apl
| extend duration_ms = duration / 1000000
| extend is_error = status >= 400
| extend full_name = strcat(first_name, " ", last_name)
```

### Aggregation

#### summarize
Aggregate data with grouping.

```apl
// Count by field
| summarize count() by status

// Multiple aggregations
| summarize
    total = count(),
    errors = countif(status >= 500),
    avg_duration = avg(duration)
  by ['service.name']

// Time bucketing
| summarize count() by bin(_time, 5m)
| summarize count() by bin_auto(_time)

// Percentiles
| summarize
    p50 = percentile(duration, 50),
    p95 = percentile(duration, 95),
    p99 = percentile(duration, 99)
  by endpoint
```

### Sorting & Limiting

#### sort / order
Sort results.

```apl
| sort by _time desc
| sort by count_ desc, name asc
| order by duration desc  // alias for sort
```

#### top
Get top N by field.

```apl
| top 10 by count_
| top 5 by duration desc
```

#### limit / take
Limit result count.

```apl
| limit 100
| take 50  // alias for limit
```

### Joining

#### join
Combine datasets.

```apl
| join kind=inner (
    ['other-dataset'] | where _time >= ago(1h)
  ) on user_id

// Join kinds: inner, leftouter, rightouter, fullouter, leftanti, rightanti
```

#### union
Combine multiple datasets.

```apl
union ['logs-app-1'], ['logs-app-2']
union ['logs-*']  // wildcard
```

### Data Shaping

#### mv-expand
Expand arrays into rows.

```apl
| mv-expand tag = tags
| mv-expand item = parse_json(items)
```

#### parse
Extract fields from strings.

```apl
| parse message with * "user=" user " action=" action
| parse-kv message as (duration:long, error:string) with (pair_delimiter=",")
```

#### getschema
Show dataset schema.

```apl
['dataset'] | getschema
```

## Aggregation Functions

| Function | Description | Example |
|----------|-------------|---------|
| `count()` | Count rows | `summarize count()` |
| `countif(cond)` | Conditional count | `countif(status >= 500)` |
| `dcount(field)` | Distinct count | `dcount(user_id)` |
| `sum(field)` | Sum values | `sum(bytes)` |
| `avg(field)` | Average | `avg(duration)` |
| `min(field)` | Minimum | `min(_time)` |
| `max(field)` | Maximum | `max(duration)` |
| `percentile(field, n)` | Nth percentile | `percentile(duration, 95)` |
| `stdev(field)` | Standard deviation | `stdev(response_time)` |
| `variance(field)` | Variance | `variance(latency)` |

## Set Functions

| Function | Description | Example |
|----------|-------------|---------|
| `make_set(field)` | Unique values as array | `make_set(['service.name'])` |
| `make_list(field)` | All values as array | `make_list(error_code)` |
| `set_intersect(a, b)` | Common elements | `set_intersect(tags1, tags2)` |
| `array_length(arr)` | Array size | `array_length(services)` |

## Special Aggregations

#### arg_min / arg_max
Get field value at min/max of another field.

```apl
// Get operation name at earliest timestamp
| summarize first_op = arg_min(_time, name) by trace_id

// Get slowest operation name
| summarize slowest = arg_max(duration, name) by trace_id
```

#### histogram
Create histogram buckets.

```apl
| summarize histogram(duration, 100) by endpoint
```

#### topk
Top K values with counts.

```apl
| summarize topk(status, 5)
```
