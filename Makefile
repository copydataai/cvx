.PHONY: test build install check schema

test:
	go test ./...

build:
	go build -o bin/cvx ./cmd/cvx

install:
	go install ./cmd/cvx

check:
	go run ./cmd/cvx check --variant variants/yc-founder-engineer.yaml --save-history=false

schema:
	go run ./cmd/cvx schema
	go run ./cmd/cvx schema check
