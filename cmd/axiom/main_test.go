package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func Test_printError(t *testing.T) {
	cmd := &cobra.Command{}

	type args struct {
		err error
		cmd *cobra.Command
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "generic error",
			args: args{
				err: errors.New("the app exploded"),
				cmd: nil,
			},
			want: "Error: the app exploded\n",
		},
		{
			name: "Cobra flag error",
			args: args{
				err: cmdutil.NewFlagErrorf("unknown flag --foo"),
				cmd: cmd,
			},
			want: "Error: unknown flag --foo\n\nUsage:\n",
		},
		{
			name: "unknown Cobra command error",
			args: args{
				err: errors.New("unknown command foo"),
				cmd: cmd,
			},
			want: "Error: unknown command foo\n\nUsage:\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printError(&buf, tt.args.err, tt.args.cmd)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
