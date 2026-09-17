package cmdutil

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/pkg/terminal"
)

func TestPopulateFromArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr error
	}{
		{name: "no argument", args: nil, wantErr: ErrNoPromptArgRequired},
		{name: "empty argument", args: []string{""}, wantErr: ErrNoPromptArgRequired},
		{name: "argument", args: []string{"logs"}, want: "logs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			err := PopulateFromArgs(&Factory{IO: terminal.TestIO()}, &got)(&cobra.Command{}, tt.args)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("value = %q, want %q", got, tt.want)
			}
		})
	}
}
