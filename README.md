# Moreping

This project aims at being a pinging (ICMP) and TCP dialing tool.

## Library
[![GoDoc](https://godoc.org/github.com/tappoz/moreping/src/moreping?status.svg)](https://godoc.org/github.com/tappoz/moreping/src/moreping)
[![Build Status](https://travis-ci.org/tappoz/moreping.svg?branch=master)](https://travis-ci.org/tappoz/moreping)

This code can be used as a normal Go library to include in your Go project.
Take a look at the tests in `src/moreping/net_dialers*` for more details.

## Command

The artifact (command) produced when running `make` shares similarities
with other commands like `ping`, `telnet` and `nmap`.

### Install

Run `make`, this will put the command you just built into `/usr/local/bin/`.
For ICMP calls make sure you are running it as `sudo`.

## Requirements

This code has been tested on:

- Go 1.8
- Linux/amd64
- godep v79

## Tests

Due to how the ICMP protocol internals along with Linux raw sockets,
the ICMP integration tests and the ICMP command executions themselves **must**
be executed as a super user.

- put this `mysudo` alias into `.bashrc` to run `ginkgo` as super user with the context of
  the normal user (e.g. `$PATH` evaluated with `$GOPATH/bin`): `alias mysudo='sudo -E env "PATH=$PATH"'`
- run the tests: `mysudo go test -v ./src/moreping/...`

The `mysudo` alias is needed because of the `$GOPATH` environment variable
which is usually specified for the standard user, not for the super user.

## Examples

Some examples of usage as a library could be found in the `example` directory.

- `go run example/mainTcp.go`
- `mysudo go run example/mainIcmp.go` (you can use only `sudo` in case the `go` command is included in its `$PATH` environment variable)
