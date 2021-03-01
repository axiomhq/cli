module github.com/axiomhq/cli

go 1.16

require (
	github.com/AlecAivazis/survey/v2 v2.2.7
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/axiomhq/axiom-go v0.0.0-20210224120541-47ba5c96ab05
	github.com/briandowns/spinner v1.12.0
	github.com/cli/cli v1.6.1
	github.com/dustin/go-humanize v1.0.0
	github.com/gabriel-vasile/mimetype v1.1.2
	github.com/golangci/golangci-lint v1.37.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/goreleaser/goreleaser v0.157.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d
	github.com/mitchellh/go-homedir v1.1.0
	github.com/muesli/reflow v0.2.0
	github.com/muesli/termenv v0.7.4
	github.com/nwidger/jsoncolor v0.3.0
	github.com/pelletier/go-toml v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d
	gotest.tools/gotestsum v1.6.2
)

replace github.com/pelletier/go-toml v1.8.1 => github.com/pelletier/go-toml v1.8.2-0.20201124181426-2e01f733df54
