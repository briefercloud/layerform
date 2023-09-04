.PHONY: mocks
mocks:
	rm -rf mocks && mockery --all --keeptree --with-expecter --exclude mocks --exclude cmd --dir . --recursive

.PHONY: fmt
fmt:
	goimports -local=github.com/ergomake/layerform -w cmd internal

.PHONY: lint
lint:
	go vet ./...
	staticcheck ./...

.PHONY: tidy
tidy:
	go mod tidy

TESTS = ./internal/... ./pkg/...
.PHONY: test
test:
	go test -v -race $(TESTS)

COVERPKG = ./internal/...
.PHONY: coverage
coverage:
	go test -v -race -covermode=atomic -coverprofile=cover.out -coverpkg=$(COVERPKG) $(TESTS)

.PHONY: deps
deps:
	go install golang.org/x/tools/cmd/goimports@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/vektra/mockery/v2@v2.32.2

