package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/axiomhq/axiom-go/axiom"
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

func Test_printTraceID(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "plain error carries no trace id",
			err:  errors.New("boom"),
			want: "",
		},
		{
			name: "http error",
			err:  axiom.HTTPError{Status: 400, Message: "bad request", TraceID: "abc123"},
			want: "Trace: abc123\n",
		},
		{
			name: "http error without a trace id stays quiet",
			err:  axiom.HTTPError{Status: 400, Message: "bad request"},
			want: "",
		},
		{
			name: "wrapped http error",
			err:  fmt.Errorf("running query: %w", axiom.HTTPError{Status: 400, TraceID: "def456"}),
			want: "Trace: def456\n",
		},
		{
			name: "limit error",
			err:  axiom.LimitError{HTTPError: axiom.HTTPError{Status: 429, TraceID: "ghi789"}},
			want: "Trace: ghi789\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printTraceID(&buf, tt.err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
