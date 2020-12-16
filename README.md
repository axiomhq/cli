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

**The [Axiom](https://axiom.co) command-line application is a fast and straightforward package for interacting with [Axiom](https://axiom.co) from your command line.**

<p align="center"><img src=".github/img/demo.gif?raw=true"/></p>

---

## Table of Contents

1. [Introduction](#introduction)
1. [Goal](#Goal)
1. [Usage](#usage)
1. [Installation](#installation)
1. [Documentation](#documentation)
1. [Contributing](#contributing)
1. [License](#license)

## Introduction

The official command line client for [Axiom](https://www.axiom.co/). Axiom CLI brings the power of Axiom to the command-line. 

## Goal
The Goal of the Axiom CLI is to create, manage, build and test your Axiom projects. 

## Usage

```shell
$ axiom <command> 
$ axiom <command> <subcommand> [flags]

# Run `help` for detailed information about CLI commands
Help on flags and commands:
axiom <command> help
```
------------

## Installation

Installing the CLI globally provides access to the Axiom command.

Axiom Cli requires [Go](https://golang.org/dl/) version 1.11 or above. 

#### Download and install the pre-compiled binary manually

Binary releases are available on [GitHub Releases][2].

  [2]: https://github.com/axiomhq/cli/releases/latest

#### Install using [Homebrew](https://brew.sh)

```shell
$ brew tap axiomhq/tap
$ brew install axiom
```
#### Install using `go get`

With a working Go installation (>=1.15), run:

```shell
$ go get -u github.com/axiomhq/cli/cmd/axiom
```
**Go 1.11 and higher should be sufficient enough to use `go get` but it is not 
guaranteed that the source code does not use more recent additions to the
standard library which break building.**

-----------------

#### Install from source

This project uses native [go mod](https://golang.org/cmd/go/#hdr-Module_maintenance) support and requires a working Go 1.15
installation.

```shell
$ git clone https://github.com/axiomhq/cli.git
$ cd cli
$ make install # Build and install binary into $GOPATH
```
---------------

#### Validate installation

In all cases the installation can be validated by running `axiom -v` in the
terminal:

```shell
Axiom CLI version 0.1.0
```
----------------------

## Documentation

To learn how to log in to Axiom and start gaining instant, actionable insights, and start storing and querying unlimited machine data, visit the [documentation on Axiom](https://docs.axiom.co/)

For full command reference, see the list below, or visit cli.axiom.com. 

## GOPATH

Make sure your PATH includes the `$GOPATH/bin` directory so your commands can be easily used:

```shell
export PATH=$PATH:$GOPATH/bin
```
## Contributing

Feel free to submit PRs or to fill issues. Every kind of help is appreciated. 

Before committing, `make` should run without any issues.

Kindly check our [Contributing](https://github.com/axiomhq/cli/blob/documentation/Contributing.md) guide on how to propose bugfixes and improvements, and submitting pull requests to the project.

More information about the project layout is documented
[here](https://github.com/axiomhq/cli/blob/documentation/.github/project-layout.md)

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
