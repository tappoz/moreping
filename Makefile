install: build
	@echo "\n\n\nTASK: installing the command\n"
	sudo mv src/cmd/moreping /usr/local/bin
	@echo "\n"

test:
	@echo "\n\n\nTASK: unit and integration tests (ICMP and TCP dials)\n"
	sudo -E env "PATH=$$PATH" go test -v ./src/service/...
	@echo "\n"

build: test
	@echo "\n\n\nTASK: building the artifact\n"
	cd src/cmd && go build -o moreping
	@echo "\n"
