current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))


install:
	sudo go build -o /usr/local/bin/codecrafters cmd/codecrafters/main.go

uninstall:
	sudo rm /usr/local/bin/codecrafters

release:
	git tag v$(next_version_number)
	git push origin main v$(next_version_number)

test:
	go test -v ./...