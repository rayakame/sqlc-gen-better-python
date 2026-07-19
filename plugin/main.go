package main

import (
	python "github.com/rayakame/sqlc-gen-better-python/internal"
	"github.com/sqlc-dev/plugin-sdk-go/codegen"
)

func main() {
	codegen.Run(python.Handler)
}
