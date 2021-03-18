module github.com/axiomhq/cli

go 1.16

require (
	github.com/AlecAivazis/survey/v2 v2.2.9
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/axiomhq/axiom-go v0.0.0-20210312122006-3294c6b958f9
	github.com/axiomhq/pkg v0.0.0-20210318171555-dc26762456be
	github.com/briandowns/spinner v1.12.0
	github.com/cli/cli v1.7.0
	github.com/dustin/go-humanize v1.0.0
	github.com/golangci/golangci-lint v1.38.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/goreleaser/goreleaser v0.159.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d
	github.com/mitchellh/go-homedir v1.1.0
	github.com/muesli/reflow v0.2.0
	github.com/muesli/termenv v0.8.0
	github.com/nwidger/jsoncolor v0.3.0
	github.com/pelletier/go-toml v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6
	gotest.tools/gotestsum v1.6.2
)

replace github.com/pelletier/go-toml v1.8.1 => github.com/pelletier/go-toml v1.8.2-0.20201124181426-2e01f733df54
