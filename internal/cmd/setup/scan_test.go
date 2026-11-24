package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"dependencies":{"@opentelemetry/sdk-node":"1.0.0"}}`)
	writeFile(t, filepath.Join(root, "go.mod"), "module example\nrequire go.opentelemetry.io/otel v1.0.0\n")
	writeFile(t, filepath.Join(root, "requirements.txt"), "flask\nopentelemetry-sdk\n")
	writeFile(t, filepath.Join(root, ".env.example"), "OTEL_SERVICE_NAME=api\n")
	writeFile(t, filepath.Join(root, "node_modules", "pkg", "package.json"), `{"dependencies":{"@opentelemetry/api":"1.0.0"}}`)
	writeFile(t, filepath.Join(root, "README.md"), "this mentions opentelemetry but is not a manifest")

	signals, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, signal := range signals {
		got[signal.Path] = signal.Detail
	}
	for _, path := range []string{"package.json", "go.mod", "requirements.txt", ".env.example"} {
		if _, ok := got[path]; !ok {
			t.Errorf("Scan() missing %s in %#v", path, got)
		}
	}
	if _, ok := got[filepath.Join("node_modules", "pkg", "package.json")]; ok {
		t.Errorf("Scan() reported a dependency inside node_modules: %#v", got)
	}
	if _, ok := got["README.md"]; ok {
		t.Errorf("Scan() reported a non-manifest file: %#v", got)
	}
}

func TestScanEmpty(t *testing.T) {
	signals, err := Scan(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 0 {
		t.Fatalf("Scan() = %#v, want none", signals)
	}
}

func TestScanAxiomExporterURLs(t *testing.T) {
	tests := []struct {
		name, file, contents, setting string
	}{
		{"otlp", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co/v1/traces", "OTEL_EXPORTER_OTLP_ENDPOINT"},
		{"signal specific", ".env.local", "export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT='https://eu-central-1.aws.edge.axiom.co/v1/traces'", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"},
		{"logs", ".env.example", "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT=https://api.axiom.co", "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"},
		{"metrics", ".env", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=https://api.axiom.co", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"},
		{"whitespace and comment", ".env", " export\tOTEL_EXPORTER_OTLP_ENDPOINT = \"https://api.axiom.co\" # comment", "OTEL_EXPORTER_OTLP_ENDPOINT"},
		{"unquoted comment", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co # comment", "OTEL_EXPORTER_OTLP_ENDPOINT"},
		{"custom exporter left to agent", ".env.example", "CUSTOM_EXPORTER_URL=https://api.axiom.co", ""},
		{"manifest config left to agent", "package.json", `{"config":{"otlpEndpoint":"https://api.axiom.co"}}`, ""},
		{"OTLP manifest key left to agent", "package.json", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co", ""},
		{"assignment inside text", ".env", `NOTES="example, OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co"`, ""},
		{"assignment inside multiline text", ".env", "NOTES=\"example\nOTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co\n\"", ""},
		{"assignment after multiline text", ".env", "NOTES='example\nignored\n'\nOTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co", "OTEL_EXPORTER_OTLP_ENDPOINT"},
		{"unclosed quote", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=\"https://api.axiom.co", ""},
		{"quoted trailing text", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=\"https://api.axiom.co\" invalid", ""},
		{"comma in hostname", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://axiom.co,example.com", ""},
		{"comma in path", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co/a,b", "OTEL_EXPORTER_OTLP_ENDPOINT"},
		{"shell command", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co; echo example", ""},
		{"variable in URL", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co/${SUFFIX}", ""},
		{"uppercase host", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://API.AXIOM.CO:443", "OTEL_EXPORTER_OTLP_ENDPOINT"},
		{"other backend", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://otel.example.com", ""},
		{"hostname suffix", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co.example.com", ""},
		{"hostname prefix", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://notaxiom.co", ""},
		{"userinfo", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co@evil.example.com", ""},
		{"path mention", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=https://example.com/axiom.co", ""},
		{"management URL", ".env", "AXIOM_URL=https://api.axiom.co", ""},
		{"unrelated URL", ".env", "APP_URL=https://app.axiom.co", ""},
		{"comment", ".env", "# OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co", ""},
		{"variable", ".env", "OTEL_EXPORTER_OTLP_ENDPOINT=${AXIOM_ENDPOINT}", ""},
		{"dependency only", "package.json", `{"dependencies":{"@opentelemetry/sdk-node":"1.0.0"}}`, ""},
		{"skipped tree", filepath.Join("node_modules", "pkg", ".env"), "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co", ""},
		{"unscanned file", "README.md", "OTEL_EXPORTER_OTLP_ENDPOINT=https://api.axiom.co", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, tt.file), tt.contents)
			signals, err := Scan(root)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, signal := range signals {
				got = append(got, signal.AxiomExporterSettings...)
			}
			if strings.Join(got, ",") != tt.setting {
				t.Fatalf("Axiom exporter settings = %q, want %q", got, tt.setting)
			}
			body := renderContext(root, login{}, signals)
			if tt.setting == "" && !strings.Contains(body, "destination is unknown") {
				t.Fatal("context inferred a destination without an exporter match")
			}
			if tt.setting != "" && !strings.Contains(body, "`"+tt.setting+"` has a URL literal") {
				t.Fatal("context omitted the Axiom exporter finding")
			}
		})
	}
}

func TestScanExporterCredentialsStayOutOfContext(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".env"), "OTEL_EXPORTER_OTLP_ENDPOINT=https://user:secret-password@api.axiom.co/v1/traces?token=secret-query\nOTEL_EXPORTER_OTLP_HEADERS=Authorization=Bearer secret-header")
	signals, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 1 || len(signals[0].AxiomExporterSettings) != 1 {
		t.Fatalf("expected one Axiom exporter finding, got %#v", signals)
	}
	body := renderContext(root, login{}, signals)
	for _, secret := range []string{"secret-password", "secret-query", "secret-header", "https://user:"} {
		if strings.Contains(body, secret) {
			t.Fatalf("exporter credential leaked: %s", secret)
		}
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
