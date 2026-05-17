package drivers

import (
	"context"
	"encoding/json"
	"fmt"

	"fiber-starter/configs"
	queueContracts "fiber-starter/internal/providers/queue/contracts"
	"fiber-starter/internal/providers/schedule/contracts"
	helpers "fiber-starter/internal/support"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// AsynqScheduler implements the Scheduler interface using asynq
type AsynqScheduler struct {
	scheduler *asynq.Scheduler
	events    []*contracts.Event
	config    *configs.Config
}

// NewAsynqScheduler creates a new asynq-based scheduler
func NewAsynqScheduler(cfg *configs.Config) *AsynqScheduler {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB + 2, // Use a separate DB for scheduler if needed
	}

	return &AsynqScheduler{
		scheduler: asynq.NewScheduler(redisOpt, nil),
		config:    cfg,
	}
}

// Job registers a job to be scheduled
func (s *AsynqScheduler) Job(job queueContracts.Job) *contracts.Event {
	event := &contracts.Event{Job: job}
	s.events = append(s.events, event)
	return event
}

// Call registers a function to be scheduled
func (s *AsynqScheduler) Call(fn func() error) *contracts.Event {
	job := &closureJob{fn: fn}
	return s.Job(job)
}

// Command registers a console command to be scheduled
func (s *AsynqScheduler) Command(command string, args ...string) *contracts.Event {
	job := &commandJob{command: command, args: args}
	return s.Job(job)
}

type closureJob struct {
	fn func() error
}

func (j *closureJob) Handle(ctx context.Context) error { return j.fn() }
func (j *closureJob) TaskName() string                 { return "schedule:closure" }
func (j *closureJob) QueueName() string                { return "default" }

type commandJob struct {
	command string
	args    []string
}

func (j *commandJob) Handle(ctx context.Context) error {
	// In a real implementation, this would execute the command via os/exec or a command runner
	fmt.Printf("Executing scheduled command: %s %v\n", j.command, j.args)
	return nil
}
func (j *commandJob) TaskName() string  { return "schedule:command" }
func (j *commandJob) QueueName() string { return "default" }

// GetEvents returns all registered scheduled events
func (s *AsynqScheduler) GetEvents() []*contracts.Event {
	return s.events
}

// Run starts the scheduler and registers all events
func (s *AsynqScheduler) Run() error {
	for _, event := range s.events {
		if event.Expression == "" {
			helpers.Warn("Scheduled job has no expression, skipping", zap.String("task", event.Job.TaskName()))
			continue
		}

		payload, err := json.Marshal(event.Job)
		if err != nil {
			helpers.Error("Failed to marshal scheduled job payload", zap.String("task", event.Job.TaskName()), zap.Error(err))
			continue
		}

		task := asynq.NewTask(event.Job.TaskName(), payload)

		entryID, err := s.scheduler.Register(event.Expression, task)
		if err != nil {
			helpers.Error("Failed to register scheduled job",
				zap.String("task", event.Job.TaskName()),
				zap.String("cron", event.Expression),
				zap.Error(err))
		} else {
			helpers.Info("Registered scheduled job",
				zap.String("task", event.Job.TaskName()),
				zap.String("cron", event.Expression),
				zap.String("entryID", entryID))
		}
	}

	helpers.Info("Scheduler starting...")
	return s.scheduler.Run()
}

// Stop stops the scheduler process
func (s *AsynqScheduler) Stop() error {
	s.scheduler.Shutdown()
	helpers.Info("Scheduler stopped")
	return nil
}
