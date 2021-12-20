install: build
	@echo "\n\n\nTASK: installing the command\n"
	sudo mv src/cmd/moreping /usr/local/bin
	@echo "\n"

# sudo -E env "PATH=$$PATH" go test -v ./src/moreping/...
# otherwise the "net" package says: "operation not permitted"
test:
	@echo "\n\n\nTASK: unit and integration tests\n"
	go test -v ./src/moreping/... -run TestInPlaceAvg
	go test -v ./src/moreping/... -run TestTCP
	@echo "\n\n\nTASK: ICMP sudo integration tests\n"
	sudo -E env "PATH=$$PATH" go test -v ./src/moreping/... -run TestICMP
	@echo "\n"

build: test
	@echo "\n\n\nTASK: building the artifact\n"
	mkdir -p out/
	cd src/cmd && go build -o ../../out/moreping
	@echo "\n"

cmd-test: build
	./src/cmd/moreping tcp --domain google.com --port 443