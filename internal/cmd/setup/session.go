package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/axiomhq/cli/internal/client"
)

const onboardingSkillURL = "https://raw.githubusercontent.com/axiomhq/skills/main/skills/axiom-setup-observability/SKILL.md"

// login is the active CLI deployment. The token is deliberately absent.
type login struct {
	Alias          string
	OrganizationID string
	URL            string
	ConfigPath     string
}

func absoluteConfigPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	return filepath.Abs(path)
}

func cliArgs(info login) []string {
	args := []string{}
	if info.ConfigPath != "" {
		args = append(args, "--config", info.ConfigPath)
	}
	if info.Alias != "" {
		args = append(args, "--deployment", info.Alias)
	}
	return args
}

func sessionDir(now time.Time) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".axiom", "setup", now.UTC().Format("20060102T150405Z")), nil
}

func writeContext(dir, repo string, info login, signals []Signal) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "context.md")
	if err := os.WriteFile(path, []byte(renderContext(repo, info, signals)), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func renderContext(repo string, info login, signals []Signal) string {
	url := info.URL
	if url == "" {
		url = client.APIURL + " (CLI default)"
	}

	var b strings.Builder
	fmt.Fprintf(&b, `# Axiom setup

You are onboarding the repository below to Axiom. Follow the axiom-setup-observability skill.
If that skill is not installed, fetch %s and follow that file instead.

Do not print or commit tokens. This machine's Axiom CLI is already logged in.
Use the deployment in this file, and follow the skill's credential rules before sending data.

## This machine

- Repository: %s
- Deployment: %s
- Organization ID: %s
- API URL: %s

`, onboardingSkillURL, repo, blank(info.Alias), blank(info.OrganizationID), url)
	args, _ := json.Marshal(cliArgs(info)) // A slice of strings is always JSON-encodable.
	fmt.Fprintf(&b, `## Axiom CLI commands

For every Axiom CLI command, insert these arguments immediately after the axiom executable.
This JSON array lists separate arguments; pass each value literally, preserving spaces in paths:

%s

Keep these arguments when running from the repository directory or any other directory.
Do not switch to the default deployment or config. If using another tool or reading credentials
directly, use the same config and deployment selected above.

## OpenTelemetry already in the repo

`, args)

	if len(signals) == 0 {
		b.WriteString("No OpenTelemetry package or OTEL_ setting was found in the usual dependency and env files.\n")
		b.WriteString("The app may still be instrumented in a language this check does not recognize. Read the repository before deciding.\n")
		return b.String()
	}
	for _, signal := range signals {
		fmt.Fprintf(&b, "- `%s`: %s\n", signal.Path, signal.Detail)
	}
	return b.String()
}

func blank(value string) string {
	if value == "" {
		return "not recorded"
	}
	return value
}

func launchPrompt(contextPath string) string {
	return fmt.Sprintf("Read the file %q and follow its instructions exactly. It is your Axiom onboarding brief, and it starts by following the axiom-setup-observability skill.", contextPath)
}
