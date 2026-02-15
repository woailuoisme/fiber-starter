package schedule

import (
	"fiber-starter/app/helpers"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Kernel 定时任务内核
type Kernel struct {
	cron *cron.Cron
}

// NewKernel 创建定时任务内核
func NewKernel() *Kernel {
	// 创建 cron 实例，支持秒级定时
	c := cron.New(cron.WithSeconds())

	return &Kernel{
		cron: c,
	}
}

// Schedule 注册定时任务
func (k *Kernel) Schedule() {
	// 在这里注册你的定时任务

	// 🧪 快速测试：每10秒执行一次（取消注释以测试）
	// k.Cron("*/10 * * * * *", TestScheduleTask)

	// 示例1：每分钟执行一次
	// k.EveryMinute(func() {
	// 	helpers.Info("定时任务：每分钟执行")
	// 	ExampleTask()
	// })

	// 示例2：每5分钟清理临时文件
	// k.EveryFiveMinutes(CleanupTask)

	// 示例3：每小时执行一次
	// k.Hourly(func() {
	// 	helpers.Info("定时任务：每小时执行")
	// })

	// 示例4：每天凌晨2点备份数据库
	// k.DailyAt("02:00", BackupDatabaseTask)

	// 示例5：每周一早上9点生成报表
	// k.Cron("0 0 9 * * 1", GenerateReportTask)

	// 示例6：工作日每天早上8点发送邮件
	// k.Cron("0 0 8 * * 1-5", SendEmailTask)

	// 更多任务可以在这里添加...
	// 提示：取消注释上面的示例任务，或添加你自己的任务
}

// Start 启动定时任务
func (k *Kernel) Start() {
	k.Schedule()
	k.cron.Start()
	helpers.Info("定时任务已启动")
}

// Stop 停止定时任务
func (k *Kernel) Stop() {
	k.cron.Stop()
	helpers.Info("定时任务已停止")
}

// Cron 使用 cron 表达式注册任务
// 格式：秒 分 时 日 月 周
// 示例：
//   - "0 30 * * * *"  每小时的30分执行
//   - "0 0 1 * * *"   每天凌晨1点执行
//   - "0 0 0 * * 0"   每周日凌晨执行
func (k *Kernel) Cron(spec string, cmd func()) {
	_, err := k.cron.AddFunc(spec, cmd)
	if err != nil {
		helpers.Error("注册定时任务失败", zap.String("spec", spec), zap.Error(err))
	}
}

// EverySecond 每秒执行
func (k *Kernel) EverySecond(cmd func()) {
	k.Cron("* * * * * *", cmd)
}

// EveryMinute 每分钟执行
func (k *Kernel) EveryMinute(cmd func()) {
	k.Cron("0 * * * * *", cmd)
}

// EveryFiveMinutes 每5分钟执行
func (k *Kernel) EveryFiveMinutes(cmd func()) {
	k.Cron("0 */5 * * * *", cmd)
}

// EveryTenMinutes 每10分钟执行
func (k *Kernel) EveryTenMinutes(cmd func()) {
	k.Cron("0 */10 * * * *", cmd)
}

// EveryFifteenMinutes 每15分钟执行
func (k *Kernel) EveryFifteenMinutes(cmd func()) {
	k.Cron("0 */15 * * * *", cmd)
}

// EveryThirtyMinutes 每30分钟执行
func (k *Kernel) EveryThirtyMinutes(cmd func()) {
	k.Cron("0 */30 * * * *", cmd)
}

// Hourly 每小时执行
func (k *Kernel) Hourly(cmd func()) {
	k.Cron("0 0 * * * *", cmd)
}

// HourlyAt 每小时的指定分钟执行
func (k *Kernel) HourlyAt(minute int, cmd func()) {
	k.Cron("0 "+string(rune(minute))+" * * * *", cmd)
}

// Daily 每天凌晨执行
func (k *Kernel) Daily(cmd func()) {
	k.Cron("0 0 0 * * *", cmd)
}

// DailyAt 每天指定时间执行
// 示例：DailyAt("13:30", func() {})
func (k *Kernel) DailyAt(time string, cmd func()) {
	k.Cron("0 "+time+":00 * * *", cmd)
}

// Weekly 每周日凌晨执行
func (k *Kernel) Weekly(cmd func()) {
	k.Cron("0 0 0 * * 0", cmd)
}

// Monthly 每月1号凌晨执行
func (k *Kernel) Monthly(cmd func()) {
	k.Cron("0 0 0 1 * *", cmd)
}

// Yearly 每年1月1号凌晨执行
func (k *Kernel) Yearly(cmd func()) {
	k.Cron("0 0 0 1 1 *", cmd)
}
