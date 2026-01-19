---
name: axiom-apl
description: APL query language reference for Axiom. Provides operators, functions, patterns, and CLI usage. Auto-invoked by specialized Axiom skills when writing or debugging APL queries.
compatibility: Requires authenticated Axiom CLI (axiom)
user-invocable: false
context: fork
allowed-tools: Bash(axiom query:*), Bash(axiom dataset list:*), Bash(axiom stream:*), Read, Grep, Glob
---

# Axiom Processing Language (APL)

APL is Axiom's query language for analyzing observability data. This skill provides comprehensive guidance for writing, debugging, and optimizing APL queries.

## Quick Reference

**Documentation:** https://axiom.co/docs/apl/introduction

**CLI usage:** See [references/cli.md](references/cli.md)

## Core Workflow

### 1. List Available Datasets

```bash
axiom dataset list -f json
```

### 2. Discover Schema (CRITICAL - Always Do First)

```apl
['<dataset>'] | getschema
```

Never guess field names. The schema shows all fields with their types.

### 3. Sample Data

```apl
['<dataset>'] | limit 10
```

### 4. Write Query

See references for operators, functions, and patterns.

## APL Syntax Essentials

### Dataset Reference

```apl
['dataset-name']           // Bracket notation (required for names with dots/dashes)
dataset_name               // Plain identifier (only for simple names)
```

### Field Reference

```apl
field_name                 // Plain field
['field.with.dots']        // Bracket notation for dotted fields
['service.name']           // OTel data (see references/otel.md for field mappings)
```

### Basic Query Structure

```apl
['dataset']
| where <condition>
| extend <new_field> = <expression>
| summarize <aggregation> by <grouping>
| project <fields>
| sort by <field> desc
| limit 100
```

## Time Handling

**Always filter by time first** - it's the most selective filter.

```apl
// Relative time
| where _time >= ago(1h)
| where _time >= ago(24h) and _time < ago(1h)

// Absolute time
| where _time >= datetime(2024-01-15T10:00:00Z)
| where _time between (datetime(2024-01-15) .. datetime(2024-01-16))
```

**Time functions:**
- `ago(timespan)` - Relative past time
- `now()` - Current time
- `datetime(string)` - Parse datetime
- `bin(_time, 5m)` - Time bucketing
- `bin_auto(_time)` - Automatic bucketing

## When NOT to Use

- **Simple field lookup**: Use `getschema` directly instead of invoking the full skill
- **Known query patterns**: If you already have a working query, don't re-invoke for syntax help
- **Real-time alerting**: Use Axiom Monitors for continuous alerting, not ad-hoc queries

## References

- **[CLI Usage](references/cli.md)** - Command flags and execution
- **[Operators](references/operators.md)** - Tabular and scalar operators
- **[Functions](references/functions.md)** - String, datetime, aggregation functions
- **[Patterns](references/patterns.md)** - Query patterns by use case
- **[Common Gotchas](references/gotchas.md)** - Mistakes and fixes
- **[OpenTelemetry](references/otel.md)** - OTel field mappings and trace patterns
