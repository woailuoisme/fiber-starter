// Package kernel defines application scheduled tasks
package kernel

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type scheduledTask struct {
	spec     string
	entryID  string
	taskType string
}

type Kernel struct {
	scheduler   *asynq.Scheduler
	client      *asynq.Client
	tasks       []*scheduledTask
	mu          sync.Mutex
	catchUpOnce sync.Once
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

func NewKernel() *Kernel {
	cfg := loadConfig()
	redisOpt := newRedisOpt(cfg)

	return &Kernel{
		scheduler: asynq.NewScheduler(redisOpt, nil),
		client:    asynq.NewClient(redisOpt),
		tasks:     nil,
	}
}

func (k *Kernel) Schedule() {}

func (k *Kernel) Start() {
	k.Schedule()
	helpers.Info("Scheduled tasks registered", zap.Int("count", len(k.tasks)))
	k.catchUpOnce.Do(func() {
		k.enqueueCatchUpTasks()
	})

	go k.run()

	helpers.Info("Scheduled tasks started")
}

func (k *Kernel) run() {
	if err := k.scheduler.Run(); err != nil {
		helpers.Error("Scheduler failed to run", zap.Error(err))
	}
}

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

func (k *Kernel) Cron(spec string, cmd func()) {
	k.mu.Lock()
	defer k.mu.Unlock()

	taskType := taskTypeFor(cmd)
	task := asynq.NewTask(taskType, nil)

	entryID, err := k.scheduler.Register(spec, task)
	if err != nil {
		helpers.Error("Failed to register scheduled task", zap.String("spec", spec), zap.Error(err))
		return
	}

	scheduledTask := &scheduledTask{
		spec:     spec,
		entryID:  entryID,
		taskType: taskType,
	}
	k.tasks = append(k.tasks, scheduledTask)

	helpers.Info("Registered scheduled task", zap.String("spec", spec), zap.String("entryID", entryID))
}

func (k *Kernel) EverySecond(cmd func())      { k.Cron("* * * * * *", cmd) }
func (k *Kernel) EveryMinute(cmd func())      { k.Cron("0 * * * * *", cmd) }
func (k *Kernel) EveryFiveMinutes(cmd func()) { k.Cron("0 */5 * * * *", cmd) }
func (k *Kernel) EveryTenMinutes(cmd func())  { k.Cron("0 */10 * * * *", cmd) }
func (k *Kernel) EveryFifteenMinutes(cmd func()) {
	k.Cron("0 */15 * * * *", cmd)
}
func (k *Kernel) EveryThirtyMinutes(cmd func()) { k.Cron("0 */30 * * * *", cmd) }
func (k *Kernel) Hourly(cmd func())             { k.Cron("0 0 * * * *", cmd) }
func (k *Kernel) HourlyAt(minute int, cmd func()) {
	k.Cron(fmt.Sprintf("0 %d * * * *", minute), cmd)
}
func (k *Kernel) Daily(cmd func()) { k.Cron("0 0 0 * * *", cmd) }

func (k *Kernel) DailyAt(time string, cmd func()) {
	spec, err := cronDailyAtSpec(time)
	if err != nil {
		helpers.Error("Invalid daily schedule time", zap.String("time", time), zap.Error(err))
		return
	}

	k.Cron(spec, cmd)
}

func (k *Kernel) Weekly(cmd func())  { k.Cron("0 0 0 * * 0", cmd) }
func (k *Kernel) Monthly(cmd func()) { k.Cron("0 0 0 1 * *", cmd) }
func (k *Kernel) Yearly(cmd func())  { k.Cron("0 0 0 1 1 *", cmd) }

func RegisterScheduledTaskHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc("scheduled_task_*", handleScheduledTask)
}

func handleScheduledTask(_ context.Context, t *asynq.Task) error {
	helpers.Info("Executing scheduled task", zap.String("type", t.Type()))
	return nil
}

func taskTypeFor(cmd func()) string {
	return fmt.Sprintf("scheduled_task_%p", cmd)
}

func (k *Kernel) enqueueCatchUpTasks() {
	for _, task := range k.tasks {
		if task == nil || task.taskType == "" {
			continue
		}
		if _, err := k.client.Enqueue(
			asynq.NewTask(task.taskType, nil),
			asynq.ProcessAt(time.Now().UTC()),
			asynq.MaxRetry(3),
		); err != nil {
			helpers.Error("Failed to enqueue catch-up task", zap.String("spec", task.spec), zap.Error(err))
		}
	}
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
