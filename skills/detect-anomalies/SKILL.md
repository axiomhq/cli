---
name: detect-anomalies
description: Detect anomalies in Axiom datasets using statistical analysis. Use when looking for unusual patterns, volume spikes, outliers, or new error types in observability data.
compatibility: Requires authenticated Axiom CLI (axiom)
user-invocable: true
context: fork
allowed-tools: Bash(axiom query:*), Bash(axiom dataset list:*), Read, Grep, Glob
---

# Anomaly Detection

Detect anomalies in Axiom datasets by comparing recent patterns to historical baselines using statistical analysis.

## Arguments

When invoked with a dataset name (e.g., `/detect-anomalies logs`), it's available as `$ARGUMENTS`.

## Prerequisites

Statistical anomaly detection requires sufficient data:

- **Minimum data points**: Z-score and standard deviation need ≥30 samples per bucket for statistical significance
- **Historical baseline**: At least 24 hours of data for meaningful comparison (methods use 25h lookback)
- **Consistent ingestion**: Gaps in data collection will skew baselines

If these aren't met, results may be misleading. Consider using simpler threshold-based alerting instead.

## Schema Discovery

**Always verify field names first:**

```bash
axiom query "['<dataset>'] | getschema" --start-time -1h
```

## Anomaly Detection Methods

### 1. Volume Anomaly Detection

Compare recent volume to baseline:

**Calculate baseline (past 24h excluding last hour):**

```bash
axiom query "['<dataset>']
| where _time between (ago(25h) .. ago(1h))
| summarize count() by bin(_time, 1h)
| summarize
    avg_hourly = avg(count_),
    stdev_hourly = stdev(count_)" --start-time -25h -f json
```

**Check recent volume:**

```bash
axiom query "['<dataset>']
| where _time >= ago(1h)
| summarize
    current_count = count(),
    current_hour = min(_time)" --start-time -1h -f json
```

**Z-score calculation:**
- `z_score = (current - avg) / stdev`
- `|z_score| > 2` indicates anomaly

### 2. New Value Detection

Find values that appeared recently but weren't seen before:

```bash
axiom query "['<dataset>']
| where _time >= ago(1h)
| summarize by error_code
| join kind=leftanti (
    ['<dataset>']
    | where _time between (ago(25h) .. ago(1h))
    | summarize by error_code
  ) on error_code" --start-time -25h -f json
```

Replace `error_code` with any categorical field (service, endpoint, status).

### 3. Statistical Outliers

Find values outside normal distribution:

**Calculate bounds:**

```bash
axiom query "['<dataset>']
| where _time between (ago(25h) .. ago(1h))
| summarize
    avg_val = avg(duration),
    stdev_val = stdev(duration)
| extend
    lower_bound = avg_val - 3 * stdev_val,
    upper_bound = avg_val + 3 * stdev_val" --start-time -25h -f json
```

**Find outliers:**

```bash
axiom query "['<dataset>']
| where _time >= ago(1h)
| where duration < <lower_bound> or duration > <upper_bound>
| limit 100" --start-time -1h -f json
```

### 4. Rare Event Detection

Find infrequent occurrences:

```bash
axiom query "['<dataset>']
| where _time >= ago(1h)
| summarize count() by error_message
| where count_ == 1" --start-time -1h -f json
```

### 5. Error Rate Spike

Compare error rate to baseline:

```bash
axiom query "['<dataset>']
| where _time >= ago(6h)
| summarize
    total = count(),
    errors = countif(status >= 500)
  by bin(_time, 15m)
| extend error_rate = errors * 100.0 / total
| sort by _time asc" --start-time -6h -f json
```

### 6. Latency Degradation

Track percentile changes:

```bash
axiom query "['<dataset>']
| where _time >= ago(6h)
| summarize
    p50 = percentile(duration, 50),
    p95 = percentile(duration, 95),
    p99 = percentile(duration, 99)
  by bin(_time, 15m)
| sort by _time asc" --start-time -6h -f json
```

## Anomaly Categories

| Type | Detection Method | Indicates |
|------|------------------|-----------|
| **Volume Spike** | Z-score on count | Traffic surge, attack, incident |
| **Volume Drop** | Z-score on count | Outage, data collection issue |
| **New Values** | Left anti-join | New errors, new services |
| **Statistical Outlier** | 3-sigma rule | Extreme performance issue |
| **Rare Events** | Count = 1 | Unusual conditions |
| **Error Spike** | Error rate increase | Service degradation |
| **Latency Spike** | Percentile increase | Performance issue |

## Output Format

```markdown
## Anomaly Report: <dataset>

### Summary
- Analysis period: <timeframe>
- Anomalies found: <count>

### Volume Anomalies
| Time | Count | Expected | Z-Score |
|------|-------|----------|---------|
| ... | ... | ... | ... |

### New Values
- Field: `error_code`
- New values: `TIMEOUT_ERROR`, `CONNECTION_REFUSED`

### Statistical Outliers
- Field: `duration`
- Outliers: <count> events above <threshold>

### Error Rate
- Baseline: X%
- Current: Y%
- Change: +Z%

### Recommendations
1. <Investigation action>
2. <Monitoring suggestion>
```

## Investigation Priority

1. **Assess impact** - Is this affecting users?
2. **Correlate timing** - What changed when anomaly started?
3. **Check related systems** - Shared dependencies?
4. **Verify data quality** - Is it a real issue or data problem?

## When NOT to Use

- **Insufficient data**: Z-score needs ≥30 data points; new datasets lack meaningful baselines
- **Known thresholds**: If you have specific SLOs (e.g., "p99 < 500ms"), use direct threshold queries
- **Real-time alerting**: Use Axiom Monitors for continuous anomaly detection, not ad-hoc analysis
- **Single data point**: Anomaly detection compares against distributions, not individual values

## APL Reference

For query syntax, invoke the `axiom-apl` skill which provides anomaly detection patterns and function documentation.
