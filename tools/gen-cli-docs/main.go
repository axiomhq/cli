package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"

	"github.com/axiomhq/cli/cmd/axiom/root"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

func main() {
	var flagError pflag.ErrorHandling
	cmd := pflag.NewFlagSet("", flagError)
	dir := cmd.StringP("dir", "d", "", "Output directory")
	tag := cmd.StringP("tag", "t", "", "Tag to set")
	help := cmd.BoolP("help", "h", false, "Help about any command")

	if err := cmd.Parse(os.Args); err != nil {
		os.Exit(1)
	}

	if *help {
		if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s", os.Args[0], cmd.FlagUsages()); err != nil {
			fatal(err)
		}
		os.Exit(1)
	}

	if len(*dir) == 0 {
		fatal("no dir set")
	}

	f := cmdutil.NewFactory()
	f.Config = &config.Config{}

	rootCmd := root.NewRootCmd(f)
	if err := os.MkdirAll(*dir, 0755); err != nil {
		fatal(err)
	}
	disableAutoGenTag(rootCmd)

	header := &doc.GenManHeader{
		Manual: "Axiom CLI",
	}

	if len(*tag) > 0 {
		header.Source = header.Manual + " " + *tag
	}

	if err := doc.GenManTree(rootCmd, header, *dir); err != nil {
		fatal(err)
	}
}

func fatal(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func disableAutoGenTag(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, cmd := range cmd.Commands() {
		disableAutoGenTag(cmd)
	}
}
