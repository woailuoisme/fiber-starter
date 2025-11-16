package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 代表基础命令，当不带任何子命令调用时执行
var rootCmd = &cobra.Command{
	Use:   "fiber-starter",
	Short: "Fiber Starter 应用程序命令行工具",
	Long: `Fiber Starter 是一个基于 Go Fiber 框架的应用程序启动器。
这个命令行工具提供了各种实用功能，包括密钥生成、数据库管理等。`,
}

// Execute 添加所有子命令到根命令并设置标志
// 这由 main.main() 调用。它只需要对 rootCmd 调用一次。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "执行命令时出错: '%s'", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件 (默认是 .env)")
}
