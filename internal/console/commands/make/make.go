package make

import (
	"github.com/spf13/cobra"
)

// Commands 导出 make 命令组下的所有命令列表。
// 为什么这样做：统一入口集中暴露，方便主 Kernel 在根命令中集中装载。
func Commands() []*cobra.Command {
	return []*cobra.Command{
		MakeExportCommand(),
		MakeImportCommand(),
	}
}
