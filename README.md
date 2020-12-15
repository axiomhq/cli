# Axiom CLI

[![Documentation][docs_badge]][docs]
[![Go Workflow][go_workflow_badge]][go_workflow]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![GoDoc][godoc_badge]][godoc]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]
[![License Status][license_status_badge]][license_status]

This is the home of the Axiom Cli code.

> The [Axiom][1] command-line application is a fast and straightforward package for interacting with [Axiom](https://axiom.co) from your command line. 

  [1]: https://axiom.co

<p align="center"><img src=".github/img/demo.gif?raw=true"/></p>

---

## Table of Contents

1. [Introduction](#introduction)
1. [Goal](#Goal)
1. [Usage](#usage)
1. [Contributing](#contributing)
1. [License](#license)

## Introduction

The official command line client for [Axiom](https://www.axiom.co/). Axiom CLI brings the power of Axiom to the command-line. 

## Goal
The Goal of the Axiom CLI is to create, manage, build and test your Axiom projects. 


## Usage

Installing the CLI globally provides access to the Axiom command.

```shell
$ axiom <command> 
$ axiom <command> <subcommand> [flags]

# Run `help` for detailed information about CLI commands
axiom <command> help
```

### Installation
Axiom Cli requires [Go](https://golang.org/dl/) version 1.11 or above. 

#### Download and install the pre-compiled binary manually

Binary releases are available on [GitHub Releases][2].

  [2]: https://github.com/axiomhq/cli/releases/latest

#### Install using [Homebrew][3]

```shell
$ brew tap axiomhq/tap
$ brew install axiom
```

  [3]: https://brew.sh

#### Install using `go get`

With a working Go installation (>=1.15), run:

```shell
$ go get -u github.com/axiomhq/cli/cmd/axiom
```

Go 1.11 and higher _should_ be sufficient enough to use `go get` but it is not 
guaranteed that the source code does not use more recent additions to the
standard library which break building.

#### Install from source

This project uses native [go mod][4] support and requires a working Go 1.15
installation.

```shell
$ git clone https://github.com/axiomhq/cli.git
$ cd cli
$ make install # Build and install binary into $GOPATH
```

  [4]: https://golang.org/cmd/go/#hdr-Module_maintenance

#### Validate installation

In all cases the installation can be validated by running `axiom -v` in the
terminal:

```shell
Axiom CLI version 0.1.0
```

### Using the application

CLI usage:

```shell
$ axiom <command> <subcommand> [flags]
```

Help on flags and commands:

```shell
$ axiom --help
```

## Contributing

Feel free to submit PRs or to fill issues. Every kind of help is appreciated.

Before committing, `make` should run without any issues.

More information about the project layout is documented
[here](.github/project_layout.md).

## License

&copy; Axiom, Inc., 2020

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

[![License Status Large][license_status_large_badge]][license_status_large]

<!-- Badges -->

[docs]: https://docs.axiom.co/cli
[docs_badge]: https://img.shields.io/badge/docs-reference-blue.svg?style=flat-square
[go_workflow]: https://github.com/axiomhq/cli/actions?query=workflow%3Ago
[go_workflow_badge]: https://img.shields.io/github/workflow/status/axiomhq/cli/go?style=flat-square
[coverage]: https://codecov.io/gh/axiomhq/cli
[coverage_badge]: https://img.shields.io/codecov/c/github/axiomhq/cli.svg?style=flat-square
[report]: https://goreportcard.com/report/github.com/axiomhq/cli
[report_badge]: https://goreportcard.com/badge/github.com/axiomhq/cli?style=flat-square
[godoc]: https://github.com/axiomhq/cli
[godoc_badge]: https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square
[release]: https://github.com/axiomhq/cli/releases/latest
[release_badge]: https://img.shields.io/github/release/axiomhq/cli.svg?style=flat-square
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/github/license/axiomhq/cli.svg?color=blue&style=flat-square
[license_status]: https://app.fossa.com/projects/git%2Bgithub.com%2Faxiomhq%2Fcli?ref=badge_shield
[license_status_badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Faxiomhq%2Fcli.svg
[license_status_large]: https://app.fossa.com/projects/git%2Bgithub.com%2Faxiomhq%2Fcli?ref=badge_large
[license_status_large_badge]: https://app.fossa.com/api/projects/git%2Bgithub.com%2Faxiomhq%2Fcli.svg?type=large
