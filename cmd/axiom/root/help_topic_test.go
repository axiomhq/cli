// Derived from https://github.com/cli/cli/blob/trunk/pkg/cmd/root/help_topic_test.go
package root

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/cli/internal/testutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func Test_newHelpTopic(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		args     []string
		flags    []string
		wantsErr bool
	}{
		{
			name:     "valid topic",
			topic:    "environment",
			args:     []string{},
			flags:    []string{},
			wantsErr: false,
		},
		{
			name:     "invalid topic",
			topic:    "invalid",
			args:     []string{},
			flags:    []string{},
			wantsErr: false,
		},
		{
			name:     "more than zero args",
			topic:    "environment",
			args:     []string{"invalid"},
			flags:    []string{},
			wantsErr: true,
		},
		{
			name:     "more than zero flags",
			topic:    "environment",
			args:     []string{},
			flags:    []string{"--invalid"},
			wantsErr: true,
		},
		{
			name:     "help arg",
			topic:    "environment",
			args:     []string{"help"},
			flags:    []string{},
			wantsErr: true,
		},
		{
			name:     "help flag",
			topic:    "environment",
			args:     []string{},
			flags:    []string{"--help"},
			wantsErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newHelpTopic(terminal.TestIO(), tt.topic)
			cmd.SetArgs(append(tt.args, tt.flags...))

			testutil.SetupCmd(cmd)

			if _, err := cmd.ExecuteC(); tt.wantsErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
