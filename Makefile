.PHONY: fmt fmt-check test race vet vuln legacy-hash-check fuzz bench build cli check

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

vuln:
	govulncheck ./...

legacy-hash-check:
	@if go list -deps ./... | grep -E '^crypto/(md5|sha1)$$'; then \
		echo "legacy hash dependency detected"; \
		exit 1; \
	fi

fuzz:
	go test -fuzz='^FuzzParseUUID$' -fuzztime=10s .
	go test -fuzz='^FuzzParseUUIDLoose$' -fuzztime=10s .
	go test -fuzz='^FuzzParseID$' -fuzztime=10s .
	go test -fuzz='^FuzzParseStrongID$' -fuzztime=10s .
	go test -fuzz='^FuzzParseSortableID$' -fuzztime=10s .

bench:
	go test -run '^$$' -bench . -benchmem ./...

build:
	go build ./...

cli:
	go build -o bin/quuid ./cmd/quuid

check: fmt-check test race vet legacy-hash-check build
