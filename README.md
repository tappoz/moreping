# CLI

`cd src/cmd && go build -o moreping`

# Tests

Due to how the ICMP protocol internals along with Linux raw sockets,
the ICMP integration tests and the ICMP command executions themselves **must**
be executed as a super user.

- put this `mysudo` alias into `.bashrc` to run `ginkgo` as super user with the context of
  the normal user (e.g. `$PATH` evaluated with `$GOPATH/bin`): `alias mysudo='sudo -E env "PATH=$PATH"'`
- run the tests: `mysudo go test -v ./src/service/...`
