.PHONY: fmt
fmt:
	goimports -local=github.com/ergomake/layerform -w client cmd internal

.PHONY: lint
lint:
	go vet ./...
	staticcheck ./...

.PHONY: tidy
tidy:
	go mod tidy

TESTS = ./internal/... ./client/...
.PHONY: test
test:
	go test -v -race $(TESTS)

COVERPKG = ./internal/...,./client/...
.PHONY: coverage
coverage:
	go test -v -race -covermode=atomic -coverprofile cover.out -coverpkg $(COVERPKG) $(TESTS)

.PHONY: deps
deps:
	go install golang.org/x/tools/cmd/goimports@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

