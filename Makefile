.PHONY: install-tools modules build test lint fmt release-major release-minor release-patch

install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest && \
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest && \
	go install golang.org/x/tools/cmd/goimports@latest && \
	go install github.com/mgechev/revive@latest

modules:
	@./scripts/modules.sh list

build:
	@./scripts/modules.sh build $(filter-out $@,$(MAKECMDGOALS))

test: test-unit test-integration

test-unit:
	@./scripts/modules.sh test-unit $(filter-out $@,$(MAKECMDGOALS))

test-integration:
	@./scripts/modules.sh test-integration $(filter-out $@,$(MAKECMDGOALS))

lint:
	@./scripts/modules.sh lint $(filter-out $@,$(MAKECMDGOALS))

fmt:
	@./scripts/modules.sh fmt $(filter-out $@,$(MAKECMDGOALS))

release-major:
	@./scripts/release.sh major $(filter-out $@,$(MAKECMDGOALS))

release-minor:
	@./scripts/release.sh minor $(filter-out $@,$(MAKECMDGOALS))

release-patch:
	@./scripts/release.sh patch $(filter-out $@,$(MAKECMDGOALS))

%:
	@: