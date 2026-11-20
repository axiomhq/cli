package setup

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxManifestBytes     = 1 << 20
	openTelemetryPackage = "OpenTelemetry package"
	openTelemetryNeedle  = "opentelemetry"
)

// skipDirs are dependency and build trees. A hit in there is not this app's setup.
var skipDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "vendor": {}, "dist": {}, "build": {},
	".next": {}, "target": {}, "venv": {}, ".venv": {}, "__pycache__": {},
}

type manifestRule struct {
	match   func(name string) bool
	needles []string
	detail  string
}

func exactName(names ...string) func(string) bool {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[strings.ToLower(name)] = struct{}{}
	}
	return func(name string) bool {
		_, ok := set[strings.ToLower(name)]
		return ok
	}
}

func suffix(ext string) func(string) bool {
	return func(name string) bool {
		return strings.HasSuffix(strings.ToLower(name), ext)
	}
}

func envFile(name string) bool {
	lower := strings.ToLower(name)
	return lower == ".env" || strings.HasPrefix(lower, ".env.")
}

// Rules stay on dependency manifests and env files. Source files are left for the agent.
var manifestRules = []manifestRule{
	{exactName("package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock", "bun.lock"), []string{"@opentelemetry/", openTelemetryNeedle}, openTelemetryPackage},
	{exactName("go.mod", "go.sum"), []string{"go.opentelemetry.io/"}, "OpenTelemetry module"},
	{exactName("requirements.txt", "pyproject.toml", "poetry.lock", "Pipfile", "Pipfile.lock", "uv.lock"), []string{openTelemetryNeedle}, openTelemetryPackage},
	{exactName("Gemfile", "Gemfile.lock"), []string{openTelemetryNeedle}, "OpenTelemetry gem"},
	{exactName("pom.xml", "build.gradle", "build.gradle.kts"), []string{openTelemetryNeedle}, "OpenTelemetry dependency"},
	{exactName("Cargo.toml", "Cargo.lock"), []string{openTelemetryNeedle}, "OpenTelemetry crate"},
	{exactName("composer.json", "composer.lock"), []string{"open-telemetry", openTelemetryNeedle}, openTelemetryPackage},
	{suffix(".csproj"), []string{openTelemetryNeedle}, openTelemetryPackage},
	{suffix(".fsproj"), []string{openTelemetryNeedle}, openTelemetryPackage},
	{envFile, []string{"OTEL_"}, "OTEL_ setting"},
}

// Signal records dependency, setting, and exporter evidence from one file.
type Signal struct {
	Path                  string
	Detail                string
	AxiomExporterSettings []string
}

// Scan walks root and reports OpenTelemetry mentions and literal Axiom exporter URLs.
// It only knows common dependency files, so a miss is not proof that the app is uninstrumented.
func Scan(root string) ([]Signal, error) {
	var signals []Signal
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if entry.IsDir() {
			if path != root {
				if _, skip := skipDirs[name]; skip {
					return filepath.SkipDir
				}
			}
			return nil
		}

		rule, ok := ruleFor(name)
		if !ok {
			return nil
		}
		data, err := readManifest(path)
		if err != nil {
			return err
		}
		var detail string
		lower := strings.ToLower(string(data))
		for _, needle := range rule.needles {
			if strings.Contains(lower, strings.ToLower(needle)) {
				detail = rule.detail
				if needle != "OTEL_" {
					detail += " (" + needle + ")"
				}
				break
			}
		}
		var exporters []string
		if envFile(name) {
			exporters = axiomExporterSettings(string(data))
		}
		if detail == "" && len(exporters) == 0 {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		signals = append(signals, Signal{Path: rel, Detail: detail, AxiomExporterSettings: exporters})
		return nil
	})
	return signals, err
}

func ruleFor(name string) (manifestRule, bool) {
	for _, rule := range manifestRules {
		if rule.match(name) {
			return rule, true
		}
	}
	return manifestRule{}, false
}

func readManifest(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > maxManifestBytes {
		return nil, nil
	}

	return io.ReadAll(io.LimitReader(f, maxManifestBytes))
}
