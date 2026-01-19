---
name: explore-dataset
description: Explore an Axiom dataset to understand its schema, fields, volume, and patterns. Use when discovering a new dataset, investigating data structure, or understanding what data is available.
compatibility: Requires authenticated Axiom CLI (axiom)
user-invocable: true
context: fork
allowed-tools: Bash(axiom query:*), Bash(axiom dataset list:*), Read, Grep, Glob
---

# Dataset Exploration

Systematically explore an Axiom dataset to understand its structure, content, and potential use cases.

## Arguments

When invoked with a dataset name (e.g., `/explore-dataset logs`), the name is available as `$ARGUMENTS`.

## Exploration Protocol

### 1. List Available Datasets

If no dataset specified, list what's available:

```bash
axiom dataset list -f json
```

### 2. Schema Discovery

**Always start here.** Discover actual field names and types:

```bash
axiom query "['<dataset>'] | getschema" --start-time -1h
```

Identify:
- Field names and types
- Dotted fields requiring bracket notation
- Timestamp fields
- Key dimensions (service, status, level)

**OTel trace data:** If schema contains `trace_id`, `span_id`, `attributes.*`, note that:
- Service fields are promoted: use `['service.name']` not `['resource.service.name']`
- Custom attributes: `['attributes.custom']['field']` with `tostring()` for aggregations
- See `axiom-apl` skill's [OTel reference](../axiom-apl/references/otel.md) for field mappings

### 3. Sample Data

Examine actual values:

```bash
axiom query "['<dataset>'] | limit 10" --start-time -1h -f json
```

Look for:
- Data structure and relationships
- Field value formats
- Data quality issues

### 4. Volume Analysis

Understand data volume patterns:

```bash
axiom query "['<dataset>'] | summarize count() by bin(_time, 1h) | sort by _time asc" --start-time -24h
```

Analyze:
- Event volume over time
- Data freshness
- Collection gaps

### 5. Categorical Field Analysis

For each key categorical field (status, level, service):

```bash
axiom query "['<dataset>'] | summarize count() by <field> | top 20 by count_" --start-time -1h
```

Identify:
- Value distributions
- Cardinality
- Key dimensions for filtering

### 6. Numerical Field Statistics

For numeric fields (duration, bytes, count):

```bash
axiom query "['<dataset>'] | summarize count(), min(<field>), max(<field>), avg(<field>), percentiles(<field>, 50, 95, 99)" --start-time -1h
```

### 7. Error Pattern Detection

Search for error indicators:

```bash
axiom query "search in (['<dataset>']) 'error' or 'fail' or 'exception' | limit 20" --start-time -1h
```

## Output Format

Provide a summary including:

```markdown
## Dataset Summary: <name>

### Purpose
<What system generated this data, what it represents>

### Key Fields
| Field | Type | Description |
|-------|------|-------------|
| ... | ... | ... |

### Volume
- Events per hour: ~X
- Data freshness: last event at X

### Key Dimensions
- `status`: 200, 400, 500, ...
- `service.name`: api, web, worker, ...

### Recommended Queries
<Common queries for this dataset>

### Monitoring Opportunities
<What could be alerted on>
```

## When NOT to Use

- **Known datasets**: If you already understand the schema, skip exploration and query directly
- **Quick field check**: Use `getschema` directly for single field lookups
- **Production queries**: Exploration uses expensive operations (`search`); extract patterns then optimize
- **Repeated analysis**: Once explored, document findings and reuse—don't re-explore

## APL Reference

For query syntax, invoke the `axiom-apl` skill which provides comprehensive documentation on operators, functions, and patterns.
