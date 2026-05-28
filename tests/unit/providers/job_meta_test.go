package providers_test

import (
	"testing"

	queueContracts "lfiber/internal/providers/queue/contracts"

	"github.com/stretchr/testify/assert"
)

func TestJobMeta_ReturnsConfiguredTaskAndQueue(t *testing.T) {
	meta := queueContracts.NewJobMeta("user_import", "low")

	assert.Equal(t, "user_import", meta.TaskName())
	assert.Equal(t, "low", meta.QueueName())
}

func TestJobMeta_DefaultsEmptyQueueToDefault(t *testing.T) {
	meta := queueContracts.NewJobMeta("user_export", "")

	assert.Equal(t, "user_export", meta.TaskName())
	assert.Equal(t, "default", meta.QueueName())
}
