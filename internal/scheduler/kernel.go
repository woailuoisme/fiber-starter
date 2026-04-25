// Package schedule defines application scheduled tasks
package schedule

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// scheduledTask represents a registered scheduled task
type scheduledTask struct {
	spec    string
	entryID string
}

// Kernel Scheduled task kernel
type Kernel struct {
	scheduler *asynq.Scheduler
	client    *asynq.Client
	tasks     []*scheduledTask
	mu        sync.Mutex
}

func loadConfig() *config.Config {
	cfg := config.GlobalConfig
	if cfg != nil {
		return cfg
	}

	loaded, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	return loaded
}

func newRedisOpt(cfg *config.Config) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB + 2,
	}
}

// NewKernel Create scheduled task kernel
func NewKernel() *Kernel {
	cfg := loadConfig()
	redisOpt := newRedisOpt(cfg)

	return &Kernel{
		scheduler: asynq.NewScheduler(redisOpt, nil),
		client:    asynq.NewClient(redisOpt),
		tasks:     nil,
	}
}

// Schedule Register scheduled tasks
func (k *Kernel) Schedule() {
}

// Start scheduled tasks
func (k *Kernel) Start() {
	k.Schedule()
	helpers.Info("Scheduled tasks registered", zap.Int("count", len(k.tasks)))

	go k.run()

	helpers.Info("Scheduled tasks started")
}

func (k *Kernel) run() {
	if err := k.scheduler.Run(); err != nil {
		helpers.Error("Scheduler failed to run", zap.Error(err))
	}
}

// Stop scheduled tasks
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

	task := asynq.NewTask(taskTypeFor(cmd), nil)

	entryID, err := k.scheduler.Register(spec, task)
	if err != nil {
		helpers.Error("Failed to register scheduled task", zap.String("spec", spec), zap.Error(err))
		return
	}

	scheduledTask := &scheduledTask{
		spec:    spec,
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
	spec, err := cronDailyAtSpec(time)
	if err != nil {
		helpers.Error("Invalid daily schedule time", zap.String("time", time), zap.Error(err))
		return
	}

	k.Cron(spec, cmd)
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

func taskTypeFor(cmd func()) string {
	return fmt.Sprintf("scheduled_task_%p", cmd)
}

func cronDailyAtSpec(value string) (string, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("expected HH:MM, got %q", value)
	}

	hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return "", fmt.Errorf("invalid hour %q: %w", parts[0], err)
	}

	minute, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", fmt.Errorf("invalid minute %q: %w", parts[1], err)
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return "", fmt.Errorf("time out of range: %02d:%02d", hour, minute)
	}

	return fmt.Sprintf("0 %d %d * * *", minute, hour), nil
}
