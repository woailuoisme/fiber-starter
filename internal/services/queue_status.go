package services

// QueueStatus describes the current state of an Asynq queue.
type QueueStatus struct {
	Name        string
	Pending     int
	Running     int
	Succeeded   int
	Failed      int
	Scheduled   int
	Retry       int
	Archived    int
	Paused      bool
	Size        int
	Processed   int
	TotalFail   int
	MemoryUsage int64
}
