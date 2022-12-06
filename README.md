![cli: The power of Axiom on the command line](.github/images/banner-dark.svg#gh-dark-mode-only)
![cli: The power of Axiom on the command line](.github/images/banner-light.svg#gh-light-mode-only)

<div align="center">

[![Documentation][docs_badge]][docs]
[![Go Workflow][workflow_badge]][workflow]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

</div>

[Axiom](https://axiom.co) unlocks observability at any scale.

- **Ingest with ease, store without limits:** Axiom’s next-generation datastore enables ingesting petabytes of data with ultimate efficiency. Ship logs from Kubernetes, AWS, Azure, Google Cloud, DigitalOcean, Nomad, and others.
- **Query everything, all the time:** Whether DevOps, SecOps, or EverythingOps, query all your data no matter its age. No provisioning, no moving data from cold/archive to “hot”, and no worrying about slow queries. All your data, all. the. time.
- **Powerful dashboards, for continuous observability:** Build dashboards to collect related queries and present information that’s quick and easy to digest for you and your team. Dashboards can be kept private or shared with others, and are the perfect way to bring together data from different sources

For more information check out the [official documentation](https://axiom.co/docs).

## Usage

There are multiple ways you can install the CLI:

- With Homebrew: `brew install axiomhq/tap/axiom`
- Download the pre-built binary from the
  [GitHub Releases](https://github.com/axiomhq/cli/releases/latest)
- Using Go: `go install github.com/axiomhq/cli/cmd/axiom@latest`
- Use the [Docker image](https://hub.docker.com/r/axiomhq/cli): `docker run axiomhq/cli`

Run `axiom help` to get familiar with the supported commands:

```shell
The power of Axiom on the command-line.

USAGE
  axiom <command> <subcommand> [flags]

CORE COMMANDS
  ingest:      Ingest structured data
  query:       Query data using APL
  stream:      Livestream data

MANAGEMENT COMMANDS
  auth:        Manage authentication state
  config:      Manage configuration
  dataset:     Manage datasets

ADDITIONAL COMMANDS
  completion:  Generate shell completion scripts
  help:        Help about any command
  version:     Print version
  web:         Open Axiom in the browser

FLAGS
  -O, --auth-org-id string   Organization ID to use
  -T, --auth-token string    Token to use
  -U, --auth-url string      Url to use
  -C, --config string        Path to configuration file to use
  -D, --deployment string    Deployment to use
  -h, --help                 Show help for command
  -I, --insecure             Bypass certificate validation
      --no-spinner           Disable the activity indicator
  -v, --version              Show axiom version

EXAMPLES
  $ axiom auth login
  $ axiom version
  $ cat logs.json | axiom ingest my-logs

AUTHENTICATION
  See 'axiom help credentials' for help and guidance on authentication.

ENVIRONMENT VARIABLES
  See 'axiom help environment' for the list of supported environment variables.

LEARN MORE
  Use 'axiom <command> <subcommand> --help' for more information about a command.
  Read the manual at https://axiom.co/docs/reference/cli
```

### Configuration

The default configuration file is `.axiom.toml` located in the home directory.
Configuration values can also be set using flags or the environment. Flags get
precedence over environment variables which get precedence over the
configuration file values.

## License

Distributed under the [MIT License](./LICENSE).

<!-- Badges -->

[docs]: https://docs.axiom.co
[docs_badge]: https://img.shields.io/badge/docs-reference-blue.svg
[workflow]: https://github.com/axiomhq/cli/actions/workflows/push.yaml
[workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/cli/Push
[release]: https://github.com/axiomhq/cli/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/cli.svg
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/cli.svg?color=blue
