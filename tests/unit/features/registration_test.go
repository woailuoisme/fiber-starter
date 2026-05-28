package features_test

import (
	"testing"

	"lfiber/internal/features/monitoring"
	"lfiber/internal/features/user"
	queueContracts "lfiber/internal/providers/queue/contracts"
	scheduleContracts "lfiber/internal/providers/schedule/contracts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingQueueRegistry struct {
	jobs []queueContracts.Job
}

func (r *recordingQueueRegistry) Job(job queueContracts.Job) {
	r.jobs = append(r.jobs, job)
}

type recordingScheduleRegistry struct {
	events []*scheduleContracts.Event
}

func (r *recordingScheduleRegistry) Job(job queueContracts.Job) *scheduleContracts.Event {
	event := &scheduleContracts.Event{Job: job}
	r.events = append(r.events, event)
	return event
}

func (r *recordingScheduleRegistry) Call(func() error) *scheduleContracts.Event {
	event := &scheduleContracts.Event{}
	r.events = append(r.events, event)
	return event
}

func (r *recordingScheduleRegistry) Command(string, ...string) *scheduleContracts.Event {
	event := &scheduleContracts.Event{}
	r.events = append(r.events, event)
	return event
}

func TestUserRegisterJobs_RegistersFeatureJobs(t *testing.T) {
	registry := &recordingQueueRegistry{}

	user.RegisterJobs(registry)

	require.Len(t, registry.jobs, 3)
	assert.Equal(t, "send_welcome_email", registry.jobs[0].TaskName())
	assert.Equal(t, "user_import", registry.jobs[1].TaskName())
	assert.Equal(t, "user_export", registry.jobs[2].TaskName())
}

func TestMonitoringRegisterJobs_RegistersFeatureJobs(t *testing.T) {
	registry := &recordingQueueRegistry{}

	monitoring.RegisterJobs(registry)

	require.Len(t, registry.jobs, 1)
	assert.Equal(t, "cleanup_temp_files", registry.jobs[0].TaskName())
	assert.Equal(t, "low", registry.jobs[0].QueueName())
}

func TestMonitoringSchedule_RegistersCleanupTask(t *testing.T) {
	registry := &recordingScheduleRegistry{}

	monitoring.Schedule(registry)

	require.Len(t, registry.events, 1)
	assert.Equal(t, "cleanup_temp_files", registry.events[0].Job.TaskName())
	assert.Equal(t, "* * * * *", registry.events[0].Expression)
}
