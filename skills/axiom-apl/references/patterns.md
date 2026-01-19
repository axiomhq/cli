# APL Query Patterns

Common query patterns organized by use case.

## Log Analysis

### Error Rate by Service

```apl
['logs']
| where _time >= ago(1h)
| summarize
    total = count(),
    errors = countif(status >= 500)
  by ['service.name']
| extend error_rate = round(errors * 100.0 / total, 2)
| where errors > 0
| sort by error_rate desc
```

### Error Timeline

```apl
['logs']
| where _time >= ago(6h)
| where status >= 500
| summarize errors = count() by bin(_time, 15m)
| sort by _time asc
```

### Top Error Messages

```apl
['logs']
| where _time >= ago(1h)
| where level == "error" or status >= 500
| summarize count() by message
| top 20 by count_
```

### Latency Percentiles by Endpoint

```apl
['logs']
| where _time >= ago(1h)
| summarize
    p50 = percentile(duration, 50),
    p90 = percentile(duration, 90),
    p95 = percentile(duration, 95),
    p99 = percentile(duration, 99)
  by endpoint
| sort by p99 desc
```

### Request Volume Over Time

```apl
['logs']
| where _time >= ago(24h)
| summarize requests = count() by bin(_time, 1h)
| sort by _time asc
```

## Trace Analysis

### Get All Spans for a Trace

```apl
['traces']
| where trace_id == "<TRACE_ID>"
| sort by _time asc
| limit 100
```

### Find Error Spans in Trace

```apl
['traces']
| where trace_id == "<TRACE_ID>"
| where error == true
| project _time, ['service.name'], name, duration, ['status.message']
```

### Find Traces by Criteria

```apl
['traces']
| where _time >= ago(1h)
| where ['service.name'] == "<SERVICE>"
| where error == true
| extend error = coalesce(ensure_field("error", typeof(bool)), false)
| summarize
    start_time = min(_time),
    total_duration = max(duration),
    span_count = count(),
    error_count = countif(error),
    services = make_set(['service.name']),
    root_operation = arg_min(_time, name)
  by trace_id
| sort by start_time desc
| limit 20
```

### Slow Traces (> 1 second)

```apl
['traces']
| where _time >= ago(1h)
| where duration >= 1000000000  // 1s in nanoseconds
| summarize
    start_time = min(_time),
    total_duration = max(duration),
    span_count = count(),
    services = make_set(['service.name'])
  by trace_id
| sort by total_duration desc
| limit 20
```

### Service Dependencies

```apl
['traces']
| where _time >= ago(1h)
| where kind == "CLIENT"
| summarize calls = count() by caller=['service.name'], callee=name
| sort by calls desc
```

## Metrics & Performance

### Service Performance Summary

```apl
['logs']
| where _time >= ago(1h)
| summarize
    requests = count(),
    errors = countif(status >= 500),
    p50_latency = percentile(duration, 50),
    p95_latency = percentile(duration, 95),
    p99_latency = percentile(duration, 99)
  by ['service.name']
| extend error_rate = round(errors * 100.0 / requests, 2)
| sort by requests desc
```

### High Latency Services

```apl
['logs']
| where _time >= ago(1h)
| summarize
    p99 = percentile(duration, 99)
  by ['service.name']
| where p99 > 1000000000  // > 1s
| extend p99_ms = p99 / 1000000.0
| sort by p99 desc
```

### Throughput Over Time

```apl
['logs']
| where _time >= ago(6h)
| summarize
    requests = count(),
    errors = countif(status >= 500)
  by bin(_time, 5m)
| extend error_rate = round(errors * 100.0 / requests, 2)
| sort by _time asc
```

## Anomaly Detection

### Volume Anomaly (Z-Score)

```apl
['logs']
| where _time >= ago(24h)
| summarize count() by bin(_time, 1h)
| extend
    avg_count = avg(count_),
    stdev_count = stdev(count_),
    z_score = (count_ - avg(count_)) / stdev(count_)
| where abs(z_score) > 2
```

### New Values Detection

```apl
// Find new error codes in last hour not seen in previous 24h
['logs']
| where _time >= ago(1h)
| summarize by error_code
| join kind=leftanti (
    ['logs']
    | where _time between (ago(25h) .. ago(1h))
    | summarize by error_code
  ) on error_code
```

### Statistical Outliers

```apl
['logs']
| where _time >= ago(1h)
| summarize
    avg_duration = avg(duration),
    stdev_duration = stdev(duration)
| extend
    lower_bound = avg_duration - 3 * stdev_duration,
    upper_bound = avg_duration + 3 * stdev_duration
```

## Data Exploration

### Schema Discovery

```apl
['dataset'] | getschema
```

### Sample Data

```apl
['dataset']
| where _time >= ago(1h)
| limit 10
```

### Field Cardinality

```apl
['dataset']
| where _time >= ago(1h)
| summarize
    total = count(),
    unique_values = dcount(field_name)
  by field_name
```

### Value Distribution

```apl
['dataset']
| where _time >= ago(1h)
| summarize count() by field_name
| top 20 by count_
```

### Time Distribution

```apl
['dataset']
| where _time >= ago(24h)
| summarize count() by bin(_time, 1h)
| sort by _time asc
```

## Cross-Dataset Correlation

### Event Correlation by Time

```apl
union
  (['dataset1'] | summarize d1_count = count() by bin(_time, 5m)),
  (['dataset2'] | summarize d2_count = count() by bin(_time, 5m))
| summarize
    dataset1 = sum(d1_count),
    dataset2 = sum(d2_count)
  by _time
| sort by _time asc
```

### Sequential Pattern Mining

```apl
['logs']
| where _time >= ago(1h)
| sort by _time asc
| extend
    next_event = next(event_type, 1),
    time_to_next = next(_time, 1) - _time
| where time_to_next < 5m
| summarize count() by pattern = strcat(event_type, " -> ", next_event)
| top 20 by count_
```

## OpenTelemetry Specific

### OTel Field Reference

| Field | Bracket Notation | Description |
|-------|------------------|-------------|
| `trace_id` | No | 32-char trace ID |
| `span_id` | No | 16-char span ID |
| `parent_span_id` | No | Parent reference |
| `name` | No | Operation name |
| `duration` | No | Duration in nanoseconds |
| `kind` | No | CLIENT/SERVER/INTERNAL |
| `error` | No | Boolean error flag |
| `['service.name']` | Yes | Service identifier |
| `['status.code']` | Yes | OK/ERROR/nil |
| `['status.message']` | Yes | Error message |
| `['scope.name']` | Yes | Instrumentation library |

### Duration Conversion

OTel durations are in **nanoseconds**:

| Human | Nanoseconds | APL Expression |
|-------|-------------|----------------|
| 1 µs | 1,000 | `duration / 1000` |
| 1 ms | 1,000,000 | `duration / 1000000` |
| 1 s | 1,000,000,000 | `duration / 1000000000` |

```apl
| extend duration_ms = duration / 1000000.0
| where duration >= 1000000000  // >= 1 second
```
