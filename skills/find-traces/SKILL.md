---
name: find-traces
description: Analyze OpenTelemetry distributed traces from Axiom. Use when investigating a trace ID, finding traces by criteria (errors, latency, service), or debugging distributed system issues.
compatibility: Requires authenticated Axiom CLI (axiom)
user-invocable: true
context: fork
allowed-tools: Bash(axiom query:*), Bash(axiom dataset list:*), Read, Grep, Glob
---

# Trace Analysis

Analyze OpenTelemetry distributed traces to identify errors, latency issues, and root causes.

## Arguments

When invoked with a trace ID (e.g., `/find-traces abc123...`), it's available as `$ARGUMENTS`.

## Trace Dataset Discovery

First, find trace datasets:

```bash
axiom dataset list -f json
```

Look for datasets containing trace data (often named `*traces*`, `*spans*`, or `otel-*`).

## Schema Discovery

**Always verify field names first:**

```bash
axiom query "['<trace-dataset>'] | getschema" --start-time -1h
```

## Common Operations

### Get Trace by ID

```bash
axiom query "['<dataset>']
| where trace_id == '<TRACE_ID>'
| sort by _time asc
| limit 100" --start-time -1h -f json
```

### Find Error Traces

```bash
axiom query "['<dataset>']
| where _time >= ago(1h)
| where error == true
| extend error = coalesce(ensure_field(\"error\", typeof(bool)), false)
| summarize
    start_time = min(_time),
    total_duration = max(duration),
    span_count = count(),
    error_count = countif(error),
    services = make_set(['service.name']),
    root_operation = arg_min(_time, name)
  by trace_id
| sort by start_time desc
| limit 20" --start-time -1h -f json
```

### Find Slow Traces

```bash
axiom query "['<dataset>']
| where _time >= ago(1h)
| where duration >= 1000000000
| summarize
    start_time = min(_time),
    total_duration = max(duration),
    span_count = count(),
    services = make_set(['service.name'])
  by trace_id
| sort by total_duration desc
| limit 20" --start-time -1h -f json
```

### Find Traces by Service

```bash
axiom query "['<dataset>']
| where _time >= ago(1h)
| where ['service.name'] == '<SERVICE>'
| summarize
    start_time = min(_time),
    total_duration = max(duration),
    span_count = count(),
    error_count = countif(error == true)
  by trace_id
| sort by start_time desc
| limit 20" --start-time -1h -f json
```

### Error Spans in Trace

```bash
axiom query "['<dataset>']
| where trace_id == '<TRACE_ID>'
| where error == true
| project _time, ['service.name'], name, duration, ['status.message']" --start-time -1h -f json
```

### Critical Path Analysis

```bash
axiom query "['<dataset>']
| where trace_id == '<TRACE_ID>'
| project span_id, parent_span_id, ['service.name'], name, duration, error
| sort by duration desc" --start-time -1h -f json
```

## OTel Field Reference

| Field | Bracket? | Description |
|-------|----------|-------------|
| `trace_id` | No | 32-char trace identifier |
| `span_id` | No | 16-char span identifier |
| `parent_span_id` | No | Parent span (empty for root) |
| `name` | No | Operation name |
| `duration` | No | Duration in **nanoseconds** |
| `kind` | No | CLIENT, SERVER, INTERNAL, PRODUCER, CONSUMER |
| `error` | No | Boolean error flag |
| `['service.name']` | Yes | Service identifier |
| `['status.code']` | Yes | OK, ERROR, or nil |
| `['status.message']` | Yes | Error description |
| `['scope.name']` | Yes | Instrumentation library |

## Duration Conversion

OTel durations are in **nanoseconds**:

| Human | Nanoseconds | Filter |
|-------|-------------|--------|
| 1 ms | 1,000,000 | `duration >= 1000000` |
| 100 ms | 100,000,000 | `duration >= 100000000` |
| 1 s | 1,000,000,000 | `duration >= 1000000000` |

Convert for display:
```apl
| extend duration_ms = duration / 1000000.0
```

## Custom Attributes

Non-standard span attributes are stored in `attributes.custom` map:

```apl
// Filter by custom attribute
| where ['attributes.custom']['user_id'] == "123"

// Aggregation requires explicit cast
| summarize count() by tostring(['attributes.custom']['tenant'])
```

Without `tostring()`, aggregations fail with "grouping by field of type unknown".

## Codebase Correlation

When working in a repository that matches the traced service, correlate trace data with source code to identify root causes.

### Mapping Trace Data to Code

1. **Extract package/module path from `['scope.name']`**
   - Contains the instrumentation library or package path
   - Strip the module prefix to get the local path
   - Example: `github.com/org/repo/pkg/auth` → `pkg/auth`

2. **Find code from operation name**
   - The `name` field often contains function names or HTTP routes
   - Search the codebase for matching handlers, functions, or endpoints

3. **Trace the call chain**
   - Follow parent-child span relationships
   - Map each span to its corresponding code location
   - Identify where errors originate and propagate

**Note:** Codebase correlation is optional. Proceed with trace-only analysis if code is unavailable or doesn't match the traced services.

## Output Format

When analyzing a trace, provide:

```markdown
## Trace Summary
- **Trace ID:** <id>
- **Duration:** <human-readable>
- **Services:** <list>
- **Outcome:** success/failure

## Sequence of Events
1. <Service> - <operation> (<duration>)
2. <Service> - <operation> (<duration>) ⚠️ ERROR
...

## Error Analysis
<What failed, when, why>

## Root Cause
<Deepest error and explanation>

## Codebase Locations (if applicable)
- **Service:** <service.name>
- **Package:** <scope.name>
- **Files:** <specific files to investigate>

## Recommended Actions
1. <Specific action>
2. <What to investigate next>
```

## When NOT to Use

- **Metrics analysis**: Traces are for request flow; use logs/metrics skills for aggregated performance data
- **Non-OTel data**: This skill assumes OpenTelemetry field conventions (trace_id, span_id, etc.)
- **Known trace structure**: If you already have the query, run it directly without invoking this skill
- **Alerting on trace patterns**: Use Axiom Monitors for continuous alerting

## APL Reference

For query syntax, invoke the `axiom-apl` skill which provides trace analysis patterns and duration unit guidance.
