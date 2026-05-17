//go:build tools

// Package Support 包含工具依赖项。
// 这个文件通过 import _ "package" 的方式强制在 go.mod 中保留开发工具的依赖，
// 防止执行 go mod tidy 时这些未在业务代码中直接使用的工具包被清理掉。
// 配合 //go:build tools 标签，确保该文件不会被编译进正式的业务二进制文件中。
package Support

import (
	_ "ariga.io/atlas-provider-bun/bunschema"
	_ "github.com/uptrace/bun"
	_ "github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/dialect/sqlitedialect"
)
