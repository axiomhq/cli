# Axiom CLI

[![Documentation][docs_badge]][docs]
[![Go Workflow][go_workflow_badge]][go_workflow]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

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
$ brew tap axiomhq/tap
$ brew install axiom
```

### Install using `go get`

```shell
$ go get -u github.com/axiomhq/cli/cmd/axiom
```

### Install from source

This project uses native
[go mod](https://golang.org/cmd/go/#hdr-Module_maintenance) support.

```shell
$ git clone https://github.com/axiomhq/cli.git
$ cd cli
$ make install # Build and install binary into $GOPATH
```

### Validate installation

In all cases the installation can be validated by running `axiom -v` in the
terminal:

```shell
Axiom CLI version 0.1.0
```
## Usage

```shell
$ axiom <command> 
$ axiom <command> <subcommand> [flags]

# Run `help` for detailed information about commands
$ axiom help <command>
```

## Documentation

To learn how to log in to Axiom and start gaining instant, actionable insights,
and start storing and querying unlimited machine data, visit the
[documentation on Axiom](https://docs.axiom.co/).

For full command reference, see the list below, or visit
[cli.axiom.com](https://docs.axiom.co/getting-started/index.html).

## Commands

**Core Commands**

| Commands     | Description      |
| ------------ | ---------------- |
| axiom ingest | Ingest data      |
| axiom stream | Live stream data |

**Management Commands**

| Commands              | Description                    |
| --------------------- | ------------------------------ |
| axiom auth            | Manage Authentication State    |
| axiom config          | Manage Configuration           |
| axiom dataset         | Manage datasets                |

**Additional Commands**

| Commands                    | Description                                     |
| --------------------------- | ----------------------------------------------- |
| axiom auth login            | Login to an Axiom deployment                    |
| axiom auth status           | View authentication status                      |
| axiom auth select           | Select an Axiom deployment                      |
| axiom auth logout           | Logout of an Axiom deployment                   |
| axiom config get            | Get a configuration value                       |
| axiom config set            | Set a configuration value                       |
| axiom config edit           | Edit the configuration file                     |
| axiom dataset create        | Create a dataset                                |
| axiom dataset list          | List all datasets                               |
| axiom dataset info          | Get info about a dataset                        |
| axiom dataset update        | Update a dataset                                |
| axiom dataset delete        | Delete a dataset                                |
| axiom dataset stats         | Get statistics about all datasets               |
| axiom completion bash       | Generate shell completion script for bash       |
| axiom completion fish       | Generate shell completion script for fish       |
| axiom completion powershell | Generate shell completion script for powershell |
| axiom completion zsh        | Generate shell completion script for zsh        |

### LEARN MORE

```shell
# To get help on any information
$ axiom help

# For more information about a command.
$ axiom <command> --help
$ axiom <command> <subcommand> --help

Read the manual at https://docs.axiom.co/cli
```

## Contributing

Feel free to submit PRs or to fill issues. Every kind of help is appreciated. 

Before committing, `make` should run without any issues.

Kindly check our [Contributing](Contributing.md) guide on how to propose
bugfixes and improvements, and submitting pull requests to the project.

More information about the project layout is documented
[here](/.github/project-layout.md)

## License

&copy; Axiom, Inc., 2020

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

[![License Status][license_status_badge]][license_status]

<!-- Badges -->

[docs]: https://docs.axiom.co
[docs_badge]: https://img.shields.io/badge/docs-reference-blue.svg?style=flat-square
[godoc]: https://github.com/axiomhq/cli/axiom
[godoc_badge]: https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square&ghcache=unused
[go_workflow]: https://github.com/axiomhq/cli/actions?query=workflow%3Ago
[go_workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/cli/go?style=flat-square&ghcache=unused
[coverage]: https://codecov.io/gh/axiomhq/cli
[coverage_badge]: https://img.shields.io/codecov/c/github/axiomhq/cli.svg?style=flat-square&ghcache=unused
[report]: https://goreportcard.com/report/github.com/axiomhq/cli
[report_badge]: https://goreportcard.com/badge/github.com/axiomhq/cli?style=flat-square&ghcache=unused
[release]: https://github.com/axiomhq/cli/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/cli.svg?style=flat-square&ghcache=unused
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/cli.svg?color=blue&style=flat-square&ghcache=unused
[license_status]: https://app.fossa.com/projects/git%2Bgithub.com%2Faxiomhq%2Fcli
[license_status_badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Faxiomhq%2Fcli.svg?type=large&ghcache=unused
