current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))


install:
	go build -o $$GOPATH/bin/codecrafters cmd/codecrafters/main.go

release:
	git tag v$(next_version_number)
	git push origin main v$(next_version_number)

test:
	go test -v ./...