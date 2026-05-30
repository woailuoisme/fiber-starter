package providers_test

import (
	"testing"

	"lfiber/configs"
	schedule "lfiber/internal/providers/schedule"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleProvider_Manager(t *testing.T) {
	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	manager := schedule.NewManager(cfg)
	require.NotNil(t, manager)

	t.Run("DefaultScheduler", func(t *testing.T) {
		scheduler, err := manager.Scheduler()
		require.NoError(t, err)
		assert.NotNil(t, scheduler)
	})

	t.Run("SchedulerCaching", func(t *testing.T) {
		s1, err := manager.Scheduler()
		require.NoError(t, err)
		s2, err := manager.Scheduler()
		require.NoError(t, err)
		assert.Equal(t, s1, s2, "Scheduler instances should be cached")
	})
}

func TestScheduleProvider_Operations(t *testing.T) {
	cfg, _, _ := configs.LoadConfig()
	s, err := schedule.NewManager(cfg).Scheduler()
	require.NoError(t, err)

	t.Run("EventCreation", func(t *testing.T) {
		// Call
		e1 := s.Call(func() error { return nil })
		assert.NotNil(t, e1)
		e1.EveryMinute().Name("test_call")

		// Command
		e2 := s.Command("test:cmd", "--force")
		assert.NotNil(t, e2)
		e2.Hourly()

		// Job
		e3 := s.Job(&mockJob{})
		assert.NotNil(t, e3)
		e3.Daily()
	})

	t.Run("CronExpressions", func(t *testing.T) {
		e := s.Call(func() error { return nil })
		e.Cron("*/5 * * * *")
		// Internal state check if possible, or just ensuring no panic
	})
}
