package setup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/cli/safeexec"
)

type agent struct {
	id      string
	command string
	label   string
}

var knownAgents = []agent{
	{id: "claude", command: "claude", label: "Claude Code"},
	{id: "codex", command: "codex", label: "Codex"},
}

func installedAgents() []agent {
	var installed []agent
	for _, candidate := range knownAgents {
		if _, err := safeexec.LookPath(candidate.command); err == nil {
			installed = append(installed, candidate)
		}
	}
	return installed
}

// resolveAgent picks from the agents that are installed.
// needPrompt is true when the user has to choose between more than one.
func resolveAgent(installed []agent, requested string) (agent, bool, error) {
	if requested != "" {
		for _, candidate := range installed {
			if candidate.id == requested {
				return candidate, false, nil
			}
		}
		return agent{}, false, fmt.Errorf("%s is not installed or was not found on PATH", requested)
	}
	switch len(installed) {
	case 0:
		return agent{}, false, nil
	case 1:
		return installed[0], false, nil
	default:
		return agent{}, true, nil
	}
}

var errAgentChoiceRequired = errors.New("both claude and codex are installed and there is no terminal to ask which to use. Re-run with --agent claude or --agent codex")

func runAgent(ctx context.Context, command, repo, prompt string) error {
	cmd := exec.CommandContext(ctx, command, prompt)
	cmd.Dir = repo
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
