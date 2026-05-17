package queue

import (
	queueContracts "fiber-starter/app/Providers/Queue/Contracts"
	"fiber-starter/configs"
)

// RegisterQueue initializes and returns the queue manager and the default queue.
func RegisterQueue(cfg *configs.Config) (queueContracts.Manager, queueContracts.Queue, error) {
	if !cfg.Queue.Enabled {
		return nil, nil, nil
	}
	manager := NewManager(cfg)
	return manager, manager.Drive(), nil
}
