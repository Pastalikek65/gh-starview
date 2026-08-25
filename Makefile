.PHONY: test build vet clean

test:
	go test ./... -count=1 -timeout 30s

vet:
	go vet ./...

build:
	go build -ldflags="-s -w" -o gh-starview .

clean:
	rm -f gh-starview
