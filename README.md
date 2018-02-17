# Moreping

This project aims at pinging (ICMP) and TCP dialing.
The artifact (command) produced when running `make` shares similarities
with other commands like `ping`, `telnet` and `nmap`.

This code can also be used as a normal Go library to include in your Go project.
Take a look at the tests in `src/service/net_dialers*` for more details.

## Install

Run `make`, this will put the command you just built into `/usr/local/bin/`.
For ICMP calls make sure you are running it as `sudo`.

## Tests

Due to how the ICMP protocol internals along with Linux raw sockets,
the ICMP integration tests and the ICMP command executions themselves **must**
be executed as a super user.

- put this `mysudo` alias into `.bashrc` to run `ginkgo` as super user with the context of
  the normal user (e.g. `$PATH` evaluated with `$GOPATH/bin`): `alias mysudo='sudo -E env "PATH=$PATH"'`
- run the tests: `mysudo go test -v ./src/service/...`

The `mysudo` alias is needed because of the `$GOPATH` environment variable
which is usually specified for the standard user, not for the super user.

## Examples

Some examples of usage as a library could be found in the `example` directory.

- `go run example/mainTcp.go`
- `mysudo go run example/mainIcmp.go` (you can use only `sudo` in case the `go` command is included in its `$PATH` environment variable)
