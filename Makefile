build/server: $(shell find . -iname '*.go')
	go build -o build/server cmd/server/*.go

test:
	go test -v ./...
	go vet ./...

deploy: test build/server
	cd cmd/server && up deploy --verbose && up url

tail:
	cd cmd/server/; up logs -f
