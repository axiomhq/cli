module github.com/axiomhq/cli

go 1.15

require (
	github.com/AlecAivazis/survey/v2 v2.2.5
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/axiomhq/axiom-go v0.0.0-20201215212509-678033418d51
	github.com/briandowns/spinner v1.12.0
	github.com/cli/cli v1.4.0
	github.com/dustin/go-humanize v1.0.0
	github.com/golangci/golangci-lint v1.33.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/goreleaser/goreleaser v0.149.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d
	github.com/mitchellh/go-homedir v1.1.0
	github.com/muesli/reflow v0.2.0
	github.com/muesli/termenv v0.7.4
	github.com/nwidger/jsoncolor v0.3.0
	github.com/pelletier/go-toml v1.8.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392 // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf
	gotest.tools/gotestsum v0.6.0
)

replace github.com/pelletier/go-toml v1.8.1 => github.com/pelletier/go-toml v1.8.2-0.20201124181426-2e01f733df54
