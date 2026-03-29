.PHONY: build test lint fmt release-major release-minor release-patch

build:
	@./scripts/modules.sh build $(filter-out $@,$(MAKECMDGOALS))

test:
	@./scripts/modules.sh test $(filter-out $@,$(MAKECMDGOALS))

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