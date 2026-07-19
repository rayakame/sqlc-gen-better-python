-include .env
export

.PHONY: pipelines tests fmt lint lint-fix changelog
.DEFAULT_GOAL := pipelines

pipelines:
	make lint-fix
	make fmt
	make lint

tests:
	go test -shuffle=on ./...

fmt:
	golangci-lint fmt

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --no-config --default=none --fix -E godot,intrange,misspell,nlreturn,perfsprint,tagalign

changelog:
	go tool changie new
