.PHONY: run build test clean

run:
	go run ./cmd/pms-go/main.go

build:
	go build ./cmd/pms-go/main.go

test:
	go test ./...

clean:
	rm -rf bin/
