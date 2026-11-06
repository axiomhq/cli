package setup

import (
	"os"
	"path/filepath"
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

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
