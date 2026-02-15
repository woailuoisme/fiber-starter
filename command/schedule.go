package command

import (
	"os"
	"os/signal"
	"syscall"

	"fiber-starter/app/helpers"
	"fiber-starter/app/schedule"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule:run",
	Short: "运行定时任务调度器",
	Long:  "启动定时任务调度器，执行所有已注册的定时任务",
	Run: func(_ *cobra.Command, _ []string) {
		runSchedule()
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}

func runSchedule() {
	// 初始化数据库连接（包含配置和日志初始化）
	if err := initDB(); err != nil {
		panic(err)
	}

	helpers.Info("正在启动定时任务调度器...")

	// 创建定时任务内核
	kernel := schedule.NewKernel()

	// 启动定时任务
	kernel.Start()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	helpers.Info("正在停止定时任务调度器...")
	kernel.Stop()
	helpers.Info("定时任务调度器已停止")
}
