package command

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"fiber-starter/app/helpers"
	"fiber-starter/app/schedule"
	"fiber-starter/config"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule:run",
	Short: "运行定时任务调度器",
	Long:  "启动定时任务调度器，执行所有已注册的定时任务",
	Run: func(cmd *cobra.Command, args []string) {
		runSchedule()
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}

func runSchedule() {
	// 初始化配置
	if err := config.Init(); err != nil {
		panic(err)
	}

	// 初始化日志
	if err := helpers.Init(); err != nil {
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
