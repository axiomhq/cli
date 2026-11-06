package setup

import (
	"strings"
	"testing"
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
