// Package schedule defines application scheduled tasks
package schedule

import (
	"context"
	"fmt"
	"sync"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// scheduledTask represents a registered scheduled task
type scheduledTask struct {
	spec    string
	task    *asynq.Task
	entryID string
}

// Kernel Scheduled task kernel
type Kernel struct {
	scheduler *asynq.Scheduler
	client    *asynq.Client
	config    *config.Config
	tasks     []*scheduledTask
	mu        sync.Mutex
}

// NewKernel Create scheduled task kernel
func NewKernel() *Kernel {
	cfg := config.GlobalConfig
	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			panic(fmt.Sprintf("Failed to load config: %v", err))
		}
	}
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB + 2,
	}

	return &Kernel{
		scheduler: asynq.NewScheduler(redisOpt, nil),
		client:    asynq.NewClient(redisOpt),
		config:    cfg,
		tasks:     make([]*scheduledTask, 0),
	}
}

// Schedule Register scheduled tasks
func (k *Kernel) Schedule() {
	// Register your scheduled tasks here

	// 🧪 Quick test: run every 10 seconds (uncomment to test)
	// k.Cron("*/10 * * * * *", TestScheduleTask)

	// Example 1: run every minute
	// k.EveryMinute(func() {
	// 	helpers.Info("Scheduled task: run every minute")
	// 	ExampleTask()
	// })

	// Example 2: clean up temporary files every 5 minutes
	// k.EveryFiveMinutes(CleanupTask)

	// Example 3: run every hour
	// k.Hourly(func() {
	// 	helpers.Info("Scheduled task: run every hour")
	// })

	// Example 4: backup database at 2 AM daily
	// k.DailyAt("02:00", BackupDatabaseTask)

	// Example 5: generate report at 9 AM every Monday
	// k.Cron("0 0 9 * * 1", GenerateReportTask)

	// Example 6: send email at 8 AM every weekday
	// k.Cron("0 0 8 * * 1-5", SendEmailTask)

	// Add more tasks here...
	// Tip: uncomment the example tasks above, or add your own tasks
}

// Start Start scheduled tasks
func (k *Kernel) Start() {
	k.Schedule()
	helpers.Info("Scheduled tasks registered", zap.Int("count", len(k.tasks)))

	go func() {
		if err := k.scheduler.Run(); err != nil {
			helpers.Error("Scheduler failed to run", zap.Error(err))
		}
	}()

	helpers.Info("Scheduled tasks started")
}

// Stop Stop scheduled tasks
func (k *Kernel) Stop() {
	k.mu.Lock()
	defer k.mu.Unlock()

	for _, task := range k.tasks {
		if task.entryID != "" {
			if err := k.scheduler.Unregister(task.entryID); err != nil {
				helpers.Error("Failed to unregister task", zap.String("spec", task.spec), zap.Error(err))
			}
		}
	}

	k.scheduler.Shutdown()
	_ = k.client.Close()
	helpers.Info("Scheduled tasks stopped")
}

// Cron Register task using cron expression
// Format: second minute hour day month weekday
// Examples:
//   - "0 30 * * * *"  Run at 30 minutes of every hour
//   - "0 0 1 * * *"   Run at 1 AM every day
//   - "0 0 0 * * 0"   Run at midnight every Sunday
func (k *Kernel) Cron(spec string, cmd func()) {
	k.mu.Lock()
	defer k.mu.Unlock()

	taskName := fmt.Sprintf("scheduled_task_%p", cmd)
	task := asynq.NewTask(taskName, nil)

	entryID, err := k.scheduler.Register(spec, task)
	if err != nil {
		helpers.Error("Failed to register scheduled task", zap.String("spec", spec), zap.Error(err))
		return
	}

	scheduledTask := &scheduledTask{
		spec:    spec,
		task:    task,
		entryID: entryID,
	}
	k.tasks = append(k.tasks, scheduledTask)

	helpers.Info("Registered scheduled task", zap.String("spec", spec), zap.String("entryID", entryID))
}

// EverySecond Run every second
func (k *Kernel) EverySecond(cmd func()) {
	k.Cron("* * * * * *", cmd)
}

// EveryMinute Run every minute
func (k *Kernel) EveryMinute(cmd func()) {
	k.Cron("0 * * * * *", cmd)
}

// EveryFiveMinutes Run every 5 minutes
func (k *Kernel) EveryFiveMinutes(cmd func()) {
	k.Cron("0 */5 * * * *", cmd)
}

// EveryTenMinutes Run every 10 minutes
func (k *Kernel) EveryTenMinutes(cmd func()) {
	k.Cron("0 */10 * * * *", cmd)
}

// EveryFifteenMinutes Run every 15 minutes
func (k *Kernel) EveryFifteenMinutes(cmd func()) {
	k.Cron("0 */15 * * * *", cmd)
}

// EveryThirtyMinutes Run every 30 minutes
func (k *Kernel) EveryThirtyMinutes(cmd func()) {
	k.Cron("0 */30 * * * *", cmd)
}

// Hourly Run every hour
func (k *Kernel) Hourly(cmd func()) {
	k.Cron("0 0 * * * *", cmd)
}

// HourlyAt Run at specified minute of every hour
func (k *Kernel) HourlyAt(minute int, cmd func()) {
	k.Cron(fmt.Sprintf("0 %d * * * *", minute), cmd)
}

// Daily Run at midnight every day
func (k *Kernel) Daily(cmd func()) {
	k.Cron("0 0 0 * * *", cmd)
}

// DailyAt Run at specified time every day
// Example: DailyAt("13:30", func() {})
func (k *Kernel) DailyAt(time string, cmd func()) {
	k.Cron(fmt.Sprintf("0 %s:00 * * *", time), cmd)
}

// Weekly Run at midnight every Sunday
func (k *Kernel) Weekly(cmd func()) {
	k.Cron("0 0 0 * * 0", cmd)
}

// Monthly Run at midnight on the 1st of every month
func (k *Kernel) Monthly(cmd func()) {
	k.Cron("0 0 0 1 * *", cmd)
}

// Yearly Run at midnight on January 1st every year
func (k *Kernel) Yearly(cmd func()) {
	k.Cron("0 0 0 1 1 *", cmd)
}

// RegisterScheduledTaskHandlers Register all scheduled task handlers with asynq mux
func RegisterScheduledTaskHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc("scheduled_task_*", handleScheduledTask)
}

// handleScheduledTask Generic handler for scheduled tasks
func handleScheduledTask(_ context.Context, t *asynq.Task) error {
	helpers.Info("Executing scheduled task", zap.String("type", t.Type()))
	return nil
}
