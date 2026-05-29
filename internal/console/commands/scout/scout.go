package scout

import (
	"github.com/spf13/cobra"
)

// Commands 导出 scout 命令组下的所有命令列表。
// 为什么这样做：统一入口，方便主 Kernel 在根命令中集中装载。
func Commands() []*cobra.Command {
	return []*cobra.Command{
		ScoutImportCommand(),
		ScoutFlushCommand(),
		ScoutDeleteIndexCommand(),
	}
}
