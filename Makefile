install: build
	$(call log_this,"Install the command")
	sudo mv $(OUT_PATH)/moreping /usr/local/bin
	$(call log_this,"DONE")

# sudo -E env "PATH=$$PATH" go test -v ./src/moreping/...
# otherwise the "net" package says: "operation not permitted"
test:
	$(call log_this,"Unit tests")
	go test -v $(MOREPING_PROJECT)/src/moreping/... -run TestInPlaceAvg
	$(call log_this,"Integration tests")
	go test -v $(MOREPING_PROJECT)/src/moreping/... -run TestTCP
	$(call log_this,"Sudo integration tests - ICMP")
	sudo -E env "PATH=$$PATH" go test -v $(MOREPING_PROJECT)/src/moreping/... -run TestICMP
	$(call log_this,"DONE")

build: test
	$(call log_this,"Build the artefact")
	mkdir -p $(OUT_PATH)
	cd $(MOREPING_PROJECT)/src/cmd && go build -o $(OUT_PATH)/moreping
	$(call log_this,"DONE")

cmd-test: build
	$(call log_this,"Acceptance test - call the command artefact")
	$(OUT_PATH)/moreping tcp --domain google.com --port 443
	$(call log_this,"DONE")

# docs: https://golangci-lint.run/usage/install/#linux-and-windows
install-golangci-lint:
	$(call log_this,"Install golangci-lint")
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.43.0
	$(call log_this,"Check golangci-lint")
	golangci-lint --version
	$(call log_this,"DONE")

lint:
	$(call log_this,"Lint the source code")
	golangci-lint run --config golangci-config.yml
	$(call log_this,"DONE")

# --- helpers --- #

# path details
MOREPING_PROJECT=$(PWD)
OUT_PATH=$(MOREPING_PROJECT)/out

# colors
RED=\033[1;31m
GRN=\033[1;32m
YEL=\033[1;33m
MAG=\033[1;35m
CYN=\033[1;36m
NC=\033[0m

# logging details
LOG_UTC_TIMESTAMP := $$(date -u "+%Y-%m-%d %H:%M:%S")
LOG_PREFIX=$(CYN)$(LOG_UTC_TIMESTAMP) $(NC)$(RED)Moreping $(NC)

# logging function
define log_this
	@echo "$(LOG_PREFIX)$(YEL)$1$(NC)"
endef
