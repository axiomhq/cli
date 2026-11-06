package setup

import "testing"

func TestResolveAgent(t *testing.T) {
	claude := agent{id: "claude", command: "claude", label: "Claude Code"}
	codex := agent{id: "codex", command: "codex", label: "Codex"}

	tests := []struct {
		name      string
		installed []agent
		requested string
		wantID    string
		wantAsk   bool
		wantErr   bool
	}{
		{name: "only claude", installed: []agent{claude}, wantID: "claude"},
		{name: "requested codex", installed: []agent{claude, codex}, requested: "codex", wantID: "codex"},
		{name: "both installed", installed: []agent{claude, codex}, wantAsk: true},
		{name: "none installed"},
		{name: "requested missing", installed: []agent{claude}, requested: "codex", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ask, err := resolveAgent(tt.installed, tt.requested)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveAgent() error = %v, want error %t", err, tt.wantErr)
			}
			if ask != tt.wantAsk {
				t.Fatalf("resolveAgent() ask = %t, want %t", ask, tt.wantAsk)
			}
			if got.id != tt.wantID {
				t.Fatalf("resolveAgent() id = %q, want %q", got.id, tt.wantID)
			}
		})
	}
}
