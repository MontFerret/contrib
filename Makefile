.PHONY: install-tools modules packages build build-cli build-packages test test-unit test-packages test-integration lint lint-packages fmt fmt-packages versions deps update-package update-ferret release-major release-minor release-patch release-pre release-pre-all release-package-major release-package-minor release-package-patch release-package-pre

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

packages:
	@./scripts/packages.sh list

build: build-cli
	@./scripts/modules.sh build $(filter-out $@,$(MAKECMDGOALS))

build-packages:
	@./scripts/packages.sh build $(filter-out $@,$(MAKECMDGOALS))

build-cli:
	go build -v -o ${DIR_BIN}/runtime \
		${DIR_TEST_CLI}/main.go

test: test-unit test-integration

test-unit:
	@./scripts/modules.sh test-unit $(filter-out $@,$(MAKECMDGOALS))

test-packages:
	@./scripts/packages.sh test-unit $(filter-out $@,$(MAKECMDGOALS))

test-integration:
	@./scripts/modules.sh test-integration $(filter-out $@,$(MAKECMDGOALS))

lint:
	@./scripts/modules.sh lint $(filter-out $@,$(MAKECMDGOALS))

lint-packages:
	@./scripts/packages.sh lint $(filter-out $@,$(MAKECMDGOALS))

fmt:
	@./scripts/modules.sh fmt $(filter-out $@,$(MAKECMDGOALS))

fmt-packages:
	@./scripts/packages.sh fmt $(filter-out $@,$(MAKECMDGOALS))

versions:
	@./scripts/modules.sh versions $(filter-out $@,$(MAKECMDGOALS))

deps:
	@./scripts/modules.sh deps $(filter-out $@,$(MAKECMDGOALS))

update-package:
	@./scripts/update-package.sh $(filter-out $@,$(MAKECMDGOALS))

update-ferret:
	@./scripts/update-ferret.sh $(filter-out $@,$(MAKECMDGOALS))

release-major:
	@./scripts/release.sh major $(filter-out $@,$(MAKECMDGOALS))

release-minor:
	@./scripts/release.sh minor $(filter-out $@,$(MAKECMDGOALS))

release-patch:
	@./scripts/release.sh patch $(filter-out $@,$(MAKECMDGOALS))

release-pre:
	@./scripts/release.sh $(filter-out $@,$(MAKECMDGOALS))

release-pre-all:
	@./scripts/release-all.sh $(filter-out $@,$(MAKECMDGOALS))

release-package-major:
	@./scripts/release-package.sh major $(filter-out $@,$(MAKECMDGOALS))

release-package-minor:
	@./scripts/release-package.sh minor $(filter-out $@,$(MAKECMDGOALS))

release-package-patch:
	@./scripts/release-package.sh patch $(filter-out $@,$(MAKECMDGOALS))

release-package-pre:
	@./scripts/release-package.sh $(filter-out $@,$(MAKECMDGOALS))

%:
	@:
