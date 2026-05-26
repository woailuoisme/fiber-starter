package queue

import (
	"lfiber/configs"
	queueContracts "lfiber/internal/providers/queue/contracts"
)

// RegisterQueue initializes and returns the queue manager and the default queue.
func RegisterQueue(cfg *configs.Config) (queueContracts.Manager, queueContracts.Queue, error) {
	if !cfg.Queue.Enabled {
		return nil, nil, nil
	}
	manager := NewManager(cfg)
	return manager, manager.Drive(), nil
}
