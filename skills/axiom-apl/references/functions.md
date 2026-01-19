# APL Functions Reference

**Documentation:** https://axiom.co/docs/apl/scalar-functions

## String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `strlen(s)` | String length | `strlen(message)` |
| `tolower(s)` | Lowercase | `tolower(method)` |
| `toupper(s)` | Uppercase | `toupper(level)` |
| `trim(s)` | Remove whitespace | `trim(name)` |
| `substring(s, start, len)` | Extract substring | `substring(id, 0, 8)` |
| `strcat(a, b, ...)` | Concatenate | `strcat(first, " ", last)` |
| `split(s, delim)` | Split to array | `split(path, "/")` |
| `replace_string(s, old, new)` | Replace literal | `replace_string(url, "http", "https")` |
| `replace_regex(s, pat, repl)` | Regex replace | `replace_regex(msg, "\\d+", "N")` |

### String Predicates

| Function | Description | Example |
|----------|-------------|---------|
| `contains` | Contains substring | `message contains "error"` |
| `contains_cs` | Case-sensitive contains | `message contains_cs "Error"` |
| `startswith` | Starts with | `url startswith "/api"` |
| `endswith` | Ends with | `file endswith ".json"` |
| `has` | Word boundary match | `tags has "production"` |
| `has_cs` | Case-sensitive word match | `tags has_cs "Production"` |
| `matches regex` | Regex match | `path matches regex @"/api/v\d+"` |

**Performance tip:** `has_cs` is 5-10x faster than `contains`. Prefer case-sensitive variants.

### Extract Functions

```apl
// Extract with regex
| extend user_id = extract("user=([^&]+)", 1, url)

// Extract all matches
| extend numbers = extract_all(@"\d+", message)
```

## DateTime Functions

| Function | Description | Example |
|----------|-------------|---------|
| `now()` | Current time | `now()` |
| `ago(timespan)` | Time in past | `ago(1h)`, `ago(7d)` |
| `datetime(s)` | Parse datetime | `datetime("2024-01-15")` |
| `todatetime(s)` | Convert to datetime | `todatetime(timestamp)` |
| `format_datetime(dt, fmt)` | Format datetime | `format_datetime(_time, "yyyy-MM-dd")` |

### DateTime Parts

| Function | Description | Example |
|----------|-------------|---------|
| `hourofday(dt)` | Hour (0-23) | `hourofday(_time)` |
| `dayofweek(dt)` | Day of week | `dayofweek(_time)` |
| `dayofmonth(dt)` | Day (1-31) | `dayofmonth(_time)` |
| `weekofyear(dt)` | Week number | `weekofyear(_time)` |
| `monthofyear(dt)` | Month (1-12) | `monthofyear(_time)` |
| `getyear(dt)` | Year | `getyear(_time)` |

### Time Arithmetic

```apl
| extend tomorrow = _time + 1d
| extend last_hour = _time - 1h
| extend duration_hours = (end_time - start_time) / 1h
```

## Type Conversion

| Function | Description | Example |
|----------|-------------|---------|
| `tostring(v)` | To string | `tostring(status)` |
| `toint(v)` | To integer | `toint(code)` |
| `tolong(v)` | To long | `tolong(bytes)` |
| `toreal(v)` | To float | `toreal(count) / toreal(total)` |
| `todouble(v)` | To double | `todouble(duration)` |
| `tobool(v)` | To boolean | `tobool(is_active)` |
| `todatetime(v)` | To datetime | `todatetime(ts)` |
| `totimespan(v)` | To timespan | `totimespan("1:30:00")` |

## Null Handling

| Function | Description | Example |
|----------|-------------|---------|
| `isnull(v)` | Is null | `isnull(error_code)` |
| `isnotnull(v)` | Is not null | `isnotnull(user_id)` |
| `isempty(v)` | Is null or empty | `isempty(message)` |
| `coalesce(a, b, ...)` | First non-null | `coalesce(name, "unknown")` |
| `iff(cond, then, else)` | Conditional | `iff(status >= 500, "error", "ok")` |

### Safe Field Access

```apl
// Ensure field exists with default type
| extend error = coalesce(ensure_field("error", typeof(bool)), false)
```

## JSON Functions

| Function | Description | Example |
|----------|-------------|---------|
| `parse_json(s)` | Parse JSON string | `parse_json(payload)` |
| `tostring(obj.field)` | Access JSON field | `tostring(data.user.name)` |
| `bag_keys(obj)` | Get object keys | `bag_keys(attributes)` |
| `pack(k1, v1, ...)` | Create object | `pack("status", status, "time", _time)` |
| `pack_all()` | Pack all fields | `pack_all()` |

```apl
// Access nested JSON
| extend props = parse_json(properties)
| where props.level == "error"

// Access dotted attribute
| extend dataset = ['attributes.custom'].dataset
```

## Mathematical Functions

| Function | Description | Example |
|----------|-------------|---------|
| `abs(n)` | Absolute value | `abs(delta)` |
| `round(n, digits)` | Round | `round(avg_latency, 2)` |
| `floor(n)` | Floor | `floor(ratio)` |
| `ceiling(n)` | Ceiling | `ceiling(count / 10.0)` |
| `log(n)` | Natural log | `log(value)` |
| `log10(n)` | Base-10 log | `log10(bytes)` |
| `pow(base, exp)` | Power | `pow(2, 10)` |
| `sqrt(n)` | Square root | `sqrt(variance)` |

## Conditional Functions

```apl
// Simple if
| extend level = iff(status >= 500, "error", "ok")

// Case/switch
| extend severity = case(
    status >= 500, "critical",
    status >= 400, "error",
    status >= 300, "warning",
    "info"
  )
```

## Array Functions

| Function | Description | Example |
|----------|-------------|---------|
| `array_length(arr)` | Array length | `array_length(tags)` |
| `array_concat(a, b)` | Concatenate | `array_concat(list1, list2)` |
| `array_slice(arr, start, end)` | Slice | `array_slice(items, 0, 5)` |
| `pack_array(a, b, ...)` | Create array | `pack_array(a, b, c)` |

```apl
// Check membership
| where url in ("login", "logout", "home")
| where status !in (200, 201, 204)
```
