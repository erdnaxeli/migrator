.PHONY: examples/*
examples/*:
	cd $@ && rm -f db.sqlite && go run main.go

.PHONY: test
test: examples/*
	go test ./...

.PHONY: style
style:
	golangci-lint fmt ./...
	golangci-lint run ./...
