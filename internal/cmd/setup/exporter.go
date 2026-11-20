package setup

import (
	"net/url"
	"strings"
)

// Only literal OTLP endpoint assignments in env files are interpreted here.
// Other formats and variable expansion are left for the agent to inspect.
func axiomExporterSettings(data string) []string {
	var settings []string
	seen := make(map[string]bool)
	var multilineQuote byte
	for line := range strings.SplitSeq(data, "\n") {
		if multilineQuote != 0 {
			if closingQuote(line, multilineQuote) >= 0 {
				multilineQuote = 0
			}
			continue
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") || strings.HasPrefix(line, "export\t") {
			line = strings.TrimSpace(line[len("export"):])
		}
		key, raw, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value, quote, literal := envLiteral(raw)
		multilineQuote = quote
		if !literal || !otlpEndpointKey(key) {
			continue
		}
		endpoint, err := url.Parse(value)
		if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
			continue
		}
		host := strings.TrimSuffix(strings.ToLower(endpoint.Hostname()), ".")
		if (host == "axiom.co" || strings.HasSuffix(host, ".axiom.co")) && !seen[key] {
			settings = append(settings, key)
			seen[key] = true
		}
	}
	return settings
}

func otlpEndpointKey(key string) bool {
	switch key {
	case "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT":
		return true
	default:
		return false
	}
}

func envLiteral(raw string) (value string, multilineQuote byte, literal bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, false
	}
	if quote := raw[0]; quote == '\'' || quote == '"' {
		end := closingQuote(raw[1:], quote)
		if end < 0 {
			return "", quote, false
		}
		value = raw[1 : end+1]
		tail := strings.TrimSpace(raw[end+2:])
		if tail != "" && !strings.HasPrefix(tail, "#") {
			return "", 0, false
		}
	} else {
		value, _, _ = strings.Cut(raw, "#")
		value = strings.TrimSpace(value)
	}
	// Be conservative rather than evaluate shell syntax, escapes, or variables.
	return value, 0, !strings.ContainsAny(value, " \t\r\n$`\\\"'")
}

func closingQuote(value string, quote byte) int {
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\\':
			i++
		case quote:
			return i
		}
	}
	return -1
}
