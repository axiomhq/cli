# Axiom CLI

[![Documentation][docs_badge]][docs]
[![Go Workflow][go_workflow_badge]][go_workflow]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]
[![Docker][docker_badge]][docker]

**The [Axiom](https://axiom.co) command-line application is a fast and
straightforward tool for interacting with [Axiom](https://axiom.co).**

<p align="center"><img src=".github/img/demo.gif?raw=true"/></p>

---

## Table of Contents

1. [Introduction](#introduction)
1. [Goal](#Goal)
1. [Installation](#installation)
1. [Usage](#usage)
1. [Documentation](#documentation)
1. [Commands](#commands)
1. [Contributing](#contributing)
1. [License](#license)

## Introduction

The official command line client for [Axiom](https://www.axiom.co/). Axiom CLI
brings the power of Axiom to the command-line. 

## Goal

The Goal of the Axiom CLI is to create, manage, build and test your Axiom
projects. 

## Installation

Installing the CLI globally provides access to the Axiom command.

### Download and install the pre-compiled binary manually

Binary releases are available on
[GitHub Releases](https://github.com/axiomhq/cli/releases/latest).

### Install using [Homebrew](https://brew.sh)

```shell
brew tap axiomhq/tap
brew install axiom
```

To update:

```shell
brew upgrade axiom
```

### Install using `go install`

```shell
go install github.com/axiomhq/cli/cmd/axiom@latest
```

### Install from source

```shell
git clone https://github.com/axiomhq/cli.git
cd cli
make install # Build and install binary into $GOPATH
```

### Run the Docker image

Docker images are available on [DockerHub][docker].

```shell
docker pull axiomhq/cli
docker run axiomhq/cli
```

### Validate installation

In all cases the installation can be validated by running `axiom -v` in the
terminal which will return the CLI version number. Example:

```shell
$ axiom -v
Axiom CLI version 1.0.0
```

### Install using [Snap](https://snapcraft.io)

```shell
sudo snap install axiom
```

To send all system logs to Axiom:

```shell
# Allow Axiom to access system logs
sudo snap connect axiom log-observer

# Configure the background service
sudo snap set axiom journald-dataset=DATASET journald-url=URL journald-token=TOKEN
```

## Usage

```shell
axiom <command>
axiom <command> <subcommand> [flags]
```

## Documentation

To learn how to log in to Axiom and start gaining instant, actionable insights,
and start storing and querying unlimited machine data, visit the
[documentation on Axiom](https://docs.axiom.co/).

For full command reference, see the list below, or visit
[cli.axiom.com](https://www.axiom.co/docs/reference/cli).

## Commands

**Core Commands**

| Commands     | Description          |
| ------------ | -------------------- |
| axiom ingest | Ingest data          |
| axiom query  | Query data using APL |
| axiom stream | Live stream data     |

**Management Commands**

| Commands                       | Description                                  |
| ------------------------------ | -------------------------------------------- |
| axiom auth login               | Login to an Axiom deployment                 |
| axiom auth logout              | Logout of an Axiom deployment                |
| axiom auth select              | Select an Axiom deployment                   |
| axiom auth status              | View authentication status                   |
| axiom auth switch-org          | Switch the organization                      |
| axiom auth update-token        | Update the token of a deloyment              |
| axiom config edit              | Edit the configuration file                  |
| axiom config get               | Get a configuration value                    |
| axiom config set               | Set a configuration value                    |
| axiom dataset create           | Create a dataset                             |
| axiom dataset delete           | Delete a dataset                             |
| axiom dataset info             | Get info about a dataset                     |
| axiom dataset list             | List all datasets                            |
| axiom dataset stats            | Get statistics about all datasets            |
| axiom dataset trim             | Trim a dataset to a given size               |
| axiom dataset update           | Update a dataset                             |
| axiom organization info        | Get info about an organization               |
| axiom organization license     | Get an organizations license                 |
| axiom organization list        | List all organizations                       |
| axiom organization keys get    | Get shared access keys of an organization    |
| axiom organization keys rotate | Rotate shared access keys of an organization |
| axiom token api create         | Create a token                               |
| axiom token api delete         | Delete a token                               |
| axiom token personal create    | Create a token                               |
| axiom token personal delete    | Delete a token                               |

**Additional Commands**

| Commands                    | Description                                     |
| --------------------------- | ----------------------------------------------- |
| axiom completion bash       | Generate shell completion script for bash       |
| axiom completion fish       | Generate shell completion script for fish       |
| axiom completion powershell | Generate shell completion script for powershell |
| axiom completion zsh        | Generate shell completion script for zsh        |
| axiom help                  | Help about any command                          |
| axiom version               | Print version                                   |

### Learn more

```shell
# To get help on any information
axiom help

# For more information about a command.
axiom <command> --help
axiom <command> <subcommand> --help
```

Read the manual at https://www.axiom.co/docs/reference/cli

## Contributing

Feel free to submit PRs or to fill issues. Every kind of help is appreciated. 

Before committing, `make` should run without any issues.

Kindly check our [Contributing](Contributing.md) guide on how to propose
bugfixes and improvements, and submitting pull requests to the project.

More information about the project layout is documented
[here](.github/project-layout.md).

## License

&copy; Axiom, Inc., 2022

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

<!-- Badges -->

[docs]: https://docs.axiom.co
[docs_badge]: https://img.shields.io/badge/docs-reference-blue.svg?style=flat-square
[go_workflow]: https://github.com/axiomhq/cli/actions/workflows/push.yml
[go_workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/cli/Push?style=flat-square&ghcache=unused
[coverage]: https://codecov.io/gh/axiomhq/cli
[coverage_badge]: https://img.shields.io/codecov/c/github/axiomhq/cli.svg?style=flat-square&ghcache=unused
[report]: https://goreportcard.com/report/github.com/axiomhq/cli
[report_badge]: https://goreportcard.com/badge/github.com/axiomhq/cli?style=flat-square&ghcache=unused
[release]: https://github.com/axiomhq/cli/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/cli.svg?style=flat-square&ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/cli.svg?color=blue&style=flat-square&ghcache=unused
[docker]: https://hub.docker.com/r/axiomhq/cli
[docker_badge]: https://img.shields.io/docker/pulls/axiomhq/cli.svg?style=flat-square&ghcache=unused
