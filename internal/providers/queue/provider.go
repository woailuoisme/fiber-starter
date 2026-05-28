package queue

import (
	"lfiber/configs"
	queueContracts "lfiber/internal/providers/queue/contracts"
)

// RegisterQueue initializes and returns the queue manager and the default queue.
func RegisterQueue(cfg *configs.Config) (queueContracts.Manager, queueContracts.Queue, error) {
	manager := NewManager(cfg)
	if !cfg.Queue.Enabled {
		return manager, manager.Drive("noop"), nil
	}
	return manager, manager.Drive(), nil
}
