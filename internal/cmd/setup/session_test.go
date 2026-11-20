package setup

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestRenderContextOmitsToken(t *testing.T) {
	body := renderContext("/repo", login{
		Alias:          "prod",
		OrganizationID: "org-1",
		URL:            "https://api.axiom.co",
	}, []Signal{{Path: "package.json", Detail: "OpenTelemetry package (@opentelemetry/)"}})

	if strings.Contains(body, "xaat-") {
		t.Fatalf("context leaked a token:\n%s", body)
	}
	for _, want := range []string{"/repo", "prod", "org-1", "https://api.axiom.co", "package.json", onboardingSkillURL} {
		if !strings.Contains(body, want) {
			t.Errorf("context missing %q:\n%s", want, body)
		}
	}
}

func TestContextPreservesCLISelection(t *testing.T) {
	configPath, err := absoluteConfigPath(filepath.Join("config with spaces", "work.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(configPath) {
		t.Fatalf("config path must survive a change of working directory: %q", configPath)
	}
	info := login{Alias: "work", ConfigPath: configPath}
	body := renderContext("/another/repo", info, nil)

	// Parse the handoff exactly as an agent would: decode the JSON argv, then
	// pass the individual arguments to the CLI flag parser.
	start := strings.Index(body, "[\"")
	if start < 0 {
		t.Fatal("context has no CLI arguments")
	}
	end := strings.IndexByte(body[start:], '\n')
	var args []string
	if err := json.Unmarshal([]byte(body[start:start+end]), &args); err != nil {
		t.Fatal(err)
	}
	flags := pflag.NewFlagSet("axiom", pflag.ContinueOnError)
	var gotConfig, gotDeployment string
	flags.StringVar(&gotConfig, "config", "", "")
	flags.StringVar(&gotDeployment, "deployment", "personal", "")
	if err := flags.Parse(args); err != nil {
		t.Fatal(err)
	}
	if gotConfig != configPath || gotDeployment != "work" {
		t.Fatalf("handoff selected config %q, deployment %q", gotConfig, gotDeployment)
	}
}

func TestCLIArgsWithoutCustomConfig(t *testing.T) {
	// A missing default config is valid when credentials come from the environment.
	// Do not force --config in that case.
	path, err := absoluteConfigPath("")
	if err != nil || path != "" {
		t.Fatalf("absoluteConfigPath() = %q, %v", path, err)
	}
	for _, info := range []login{{}, {Alias: "work"}} {
		for _, arg := range cliArgs(info) {
			if arg == "--config" {
				t.Fatal("default configuration should not require an explicit file")
			}
		}
	}
}

func TestRenderContextSaysWhenNothingMatched(t *testing.T) {
	body := renderContext("/repo", login{}, nil)
	if !strings.Contains(body, "No OpenTelemetry package") {
		t.Fatalf("context = %s", body)
	}
	if !strings.Contains(body, "not recorded") {
		t.Fatalf("context = %s", body)
	}
}

func TestLaunchPromptNamesTheContextFile(t *testing.T) {
	prompt := launchPrompt("/tmp/context.md")
	if !strings.Contains(prompt, "/tmp/context.md") {
		t.Fatalf("prompt = %s", prompt)
	}
}
