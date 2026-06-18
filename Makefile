.PHONY: install-tools modules build build-cli test test-unit test-integration lint fmt versions deps release-major release-minor release-patch release-pre

DIR_BIN = ./bin
DIR_TEST = ./tests
DIR_TEST_CLI = ${DIR_TEST}/runtime

install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest && \
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest && \
	go install golang.org/x/tools/cmd/goimports@latest && \
	go install github.com/mgechev/revive@latest

modules:
	@./scripts/modules.sh list

build: build-cli
	@./scripts/modules.sh build $(filter-out $@,$(MAKECMDGOALS))

build-cli:
	go build -v -o ${DIR_BIN}/ferret \
		${DIR_TEST_CLI}/main.go

test: test-unit test-integration

test-unit:
	@./scripts/modules.sh test-unit $(filter-out $@,$(MAKECMDGOALS))

test-integration:
	@./scripts/modules.sh test-integration $(filter-out $@,$(MAKECMDGOALS))

lint:
	@./scripts/modules.sh lint $(filter-out $@,$(MAKECMDGOALS))

fmt:
	@./scripts/modules.sh fmt $(filter-out $@,$(MAKECMDGOALS))

versions:
	@./scripts/modules.sh versions $(filter-out $@,$(MAKECMDGOALS))

deps:
	@./scripts/modules.sh deps $(filter-out $@,$(MAKECMDGOALS))

release-major:
	@./scripts/release.sh major $(filter-out $@,$(MAKECMDGOALS))

release-minor:
	@./scripts/release.sh minor $(filter-out $@,$(MAKECMDGOALS))

release-patch:
	@./scripts/release.sh patch $(filter-out $@,$(MAKECMDGOALS))

release-pre:
	@./scripts/release.sh $(filter-out $@,$(MAKECMDGOALS))

%:
	@:
