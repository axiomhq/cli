# OpenTelemetry Field Mappings

Axiom transforms OpenTelemetry data during ingestion. Field names differ from standard OTel conventions.

## Promoted Resource Fields

These resource attributes are stored at top level (without `resource.` prefix):

**Service:** `service.name`, `service.version`, `service.namespace`, `service.instance.id`

**Telemetry SDK:** `telemetry.sdk.name`, `telemetry.sdk.version`, `telemetry.sdk.language`, `telemetry.distro.name`, `telemetry.distro.version`

All other resource attributes: `resource.<key>` or `resource.custom` map.

## Standard Span Fields

Always present in trace datasets:

| Field | Type | Notes |
|-------|------|-------|
| `trace_id` | string | 32-char hex |
| `span_id` | string | 16-char hex |
| `parent_span_id` | string | Empty for root spans |
| `name` | string | Operation name |
| `kind` | string | internal/server/client/consumer/producer |
| `duration` | int | **Nanoseconds** (not milliseconds) |
| `status.code` | string | OK, ERROR, or nil |
| `status.message` | string | Error details |

## Attribute Storage

### Semantic Convention Attributes

Standard OTel semconv attributes stored flat under `attributes.`:

```
attributes.http.request.method
attributes.http.response.status_code
attributes.db.system
attributes.db.statement
attributes.rpc.service
```

### AI Observability Attributes

GenAI and Eval attributes are stored flat (not in custom map):

```
attributes.gen_ai.system
attributes.gen_ai.request.model
attributes.gen_ai.response.model
attributes.gen_ai.usage.input_tokens
attributes.gen_ai.usage.output_tokens
attributes.eval.name
attributes.eval.score.value
```

### Custom Attributes

Non-standard attributes go to a map field:

```
attributes.custom = {"my_field": "value", "user_id": "123"}
```

**Access pattern:**

```apl
// Filter/project - works directly
| where ['attributes.custom']['user_id'] == "123"
| project ['attributes.custom']['user_id']

// Aggregation - MUST cast to string
| summarize count() by tostring(['attributes.custom']['user_id'])
```

Without `tostring()`, aggregations fail: "grouping by field of type unknown is not supported".

## Common Mistakes

| Want | Wrong | Correct |
|------|-------|---------|
| Service name | `['resource.service.name']` | `['service.name']` |
| SDK version | `['resource.telemetry.sdk.version']` | `['telemetry.sdk.version']` |
| Custom field | `['attributes.my_field']` | `['attributes.custom']['my_field']` |
| Group by custom | `by ['attributes.custom']['x']` | `by tostring(['attributes.custom']['x'])` |

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
