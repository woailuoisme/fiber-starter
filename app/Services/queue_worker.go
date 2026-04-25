package services

import (
	"fmt"

	helpers "fiber-starter/app/Support"

	"go.uber.org/zap"
)

func (q *queueService) StartWorker() error {
	if !q.setRunning(true) {
		return fmt.Errorf("worker process is already running")
	}

	go func() {
		if err := q.getServer().Run(q.mux); err != nil {
			helpers.LogError("Queue worker process start failed", zap.Error(err))
		}
		q.setRunning(false)
	}()

	return nil
}

func (q *queueService) RunWorker() error {
	if !q.setRunning(true) {
		return fmt.Errorf("worker process is already running")
	}

	err := q.getServer().Run(q.mux)
	q.setRunning(false)

	return err
}

func (q *queueService) StopWorker() error {
	if !q.setRunning(false) {
		return fmt.Errorf("worker process is not running")
	}

	q.getServer().Shutdown()
	_ = q.getClient().Close()
	return nil
}

func (q *queueService) Close() error {
	if q.setRunning(false) {
		q.getServer().Shutdown()
	}
	if q.client != nil {
		return q.client.Close()
	}
	return nil
}

func (q *queueService) setRunning(running bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isRunning == running {
		return false
	}

	q.isRunning = running
	return true
}
