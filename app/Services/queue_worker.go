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

	q.mu.Lock()
	defer q.mu.Unlock()

	return q.shutdownLocked()
}

func (q *queueService) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.isRunning = false
	return q.shutdownLocked()
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

func (q *queueService) shutdownLocked() error {
	if q.server != nil {
		q.server.Shutdown()
		q.server = nil
	}
	if q.client != nil {
		err := q.client.Close()
		q.client = nil
		return err
	}
	return nil
}
