install:
	go build -o $$GOPATH/bin/codecrafters cmd/codecrafters/main.go

test:
	go test -v ./...