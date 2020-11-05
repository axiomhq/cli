module github.com/axiomhq/cli

go 1.15

require (
	axicode.axiom.co/watchmakers/axiomdb v1.2.0
	github.com/AlecAivazis/survey/v2 v2.2.0
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/briandowns/spinner v1.11.1
	github.com/cli/cli v1.2.0
	github.com/golangci/golangci-lint v1.32.2
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/goreleaser/goreleaser v0.146.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/go-homedir v1.1.0
	github.com/muesli/reflow v0.2.0
	github.com/muesli/termenv v0.7.4
	github.com/nwidger/jsoncolor v0.3.0
	github.com/pelletier/go-toml v1.8.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/sys v0.0.0-20201101102859-da207088b7d1 // indirect
	golang.org/x/text v0.3.4 // indirect
	gotest.tools/gotestsum v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.0.0+incompatible
	github.com/json-iterator/go => github.com/mhr3/jsoniter v1.1.11-0.20200909125010-fb9b85012bdc
)
