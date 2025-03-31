package aiosqlite

import (
	"github.com/rayakame/sqlc-gen-better-python/internal/codegen"
	"github.com/rayakame/sqlc-gen-better-python/internal/core"
	"github.com/sqlc-dev/plugin-sdk-go/metadata"
)

func BuildPyQueryFunc(*core.Importer, *core.Query, *codegen.IndentStringBuilder) error {
	return nil
}

func AcceptedDriverCMDs() []string {
	return []string{
		metadata.CmdExec,
		metadata.CmdExecResult,
		metadata.CmdExecLastId,
		metadata.CmdExecRows,
		metadata.CmdOne,
		metadata.CmdMany,
	}
}
