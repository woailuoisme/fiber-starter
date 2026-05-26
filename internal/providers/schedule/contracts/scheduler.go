package contracts

import (
	"fmt"
	"strings"

	queueContracts "lfiber/internal/providers/queue/contracts"
)

// Event represents a scheduled task with a cron expression
type Event struct {
	Expression     string
	Job            queueContracts.Job
	NameStr        string
	DescriptionStr string
}

// Name sets the name of the event
func (e *Event) Name(name string) *Event {
	e.NameStr = name
	return e
}

// Description sets the description of the event
func (e *Event) Description(description string) *Event {
	e.DescriptionStr = description
	return e
}

// EveryMinute schedules the job to run every minute
func (e *Event) EveryMinute() *Event {
	e.Expression = "* * * * *"
	return e
}

// EveryTwoMinutes schedules the job to run every two minutes
func (e *Event) EveryTwoMinutes() *Event {
	e.Expression = "*/2 * * * *"
	return e
}

// EveryThreeMinutes schedules the job to run every three minutes
func (e *Event) EveryThreeMinutes() *Event {
	e.Expression = "*/3 * * * *"
	return e
}

// EveryFourMinutes schedules the job to run every four minutes
func (e *Event) EveryFourMinutes() *Event {
	e.Expression = "*/4 * * * *"
	return e
}

// EveryFiveMinutes schedules the job to run every five minutes
func (e *Event) EveryFiveMinutes() *Event {
	e.Expression = "*/5 * * * *"
	return e
}

// EveryTenMinutes schedules the job to run every ten minutes
func (e *Event) EveryTenMinutes() *Event {
	e.Expression = "*/10 * * * *"
	return e
}

// EveryFifteenMinutes schedules the job to run every fifteen minutes
func (e *Event) EveryFifteenMinutes() *Event {
	e.Expression = "*/15 * * * *"
	return e
}

// EveryThirtyMinutes schedules the job to run every thirty minutes
func (e *Event) EveryThirtyMinutes() *Event {
	e.Expression = "0,30 * * * *"
	return e
}

// Hourly schedules the job to run hourly
func (e *Event) Hourly() *Event {
	e.Expression = "0 * * * *"
	return e
}

// HourlyAt schedules the job to run every hour at a specific minute
func (e *Event) HourlyAt(offset int) *Event {
	e.Expression = fmt.Sprintf("%d * * * *", offset)
	return e
}

// Daily schedules the job to run daily at midnight
func (e *Event) Daily() *Event {
	e.Expression = "0 0 * * *"
	return e
}

// DailyAt schedules the job to run daily at a specific time (HH:mm)
func (e *Event) DailyAt(timeStr string) *Event {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		e.Expression = fmt.Sprintf("%s %s * * *", parts[1], parts[0])
	}
	return e
}

// TwiceDaily schedules the job to run twice daily at hours 1 and 13
func (e *Event) TwiceDaily(first, second int) *Event {
	e.Expression = fmt.Sprintf("0 %d,%d * * *", first, second)
	return e
}

// TwiceDailyAt schedules the job to run twice daily at specific times
func (e *Event) TwiceDailyAt(first, second, minute int) *Event {
	e.Expression = fmt.Sprintf("%d %d,%d * * *", minute, first, second)
	return e
}

// Weekdays schedules the job to run on weekdays (Monday-Friday)
func (e *Event) Weekdays() *Event {
	e.Expression = "0 0 * * 1-5"
	return e
}

// Weekends schedules the job to run on weekends (Saturday-Sunday)
func (e *Event) Weekends() *Event {
	e.Expression = "0 0 * * 6,0"
	return e
}

// Mondays schedules the job to run on Mondays
func (e *Event) Mondays() *Event {
	e.Expression = "0 0 * * 1"
	return e
}

// Tuesdays schedules the job to run on Tuesdays
func (e *Event) Tuesdays() *Event {
	e.Expression = "0 0 * * 2"
	return e
}

// Wednesdays schedules the job to run on Wednesdays
func (e *Event) Wednesdays() *Event {
	e.Expression = "0 0 * * 3"
	return e
}

// Thursdays schedules the job to run on Thursdays
func (e *Event) Thursdays() *Event {
	e.Expression = "0 0 * * 4"
	return e
}

// Fridays schedules the job to run on Fridays
func (e *Event) Fridays() *Event {
	e.Expression = "0 0 * * 5"
	return e
}

// Saturdays schedules the job to run on Saturdays
func (e *Event) Saturdays() *Event {
	e.Expression = "0 0 * * 6"
	return e
}

// Sundays schedules the job to run on Sundays
func (e *Event) Sundays() *Event {
	e.Expression = "0 0 * * 0"
	return e
}

// Weekly schedules the job to run weekly on Sunday at midnight
func (e *Event) Weekly() *Event {
	e.Expression = "0 0 * * 0"
	return e
}

// WeeklyOn schedules the job to run weekly on a specific day and time
func (e *Event) WeeklyOn(day int, timeStr string) *Event {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		e.Expression = fmt.Sprintf("%s %s * * %d", parts[1], parts[0], day)
	}
	return e
}

// Monthly schedules the job to run monthly on the first day at midnight
func (e *Event) Monthly() *Event {
	e.Expression = "0 0 1 * *"
	return e
}

// MonthlyOn schedules the job to run monthly on a specific day and time
func (e *Event) MonthlyOn(day int, timeStr string) *Event {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		e.Expression = fmt.Sprintf("%s %s %d * *", parts[1], parts[0], day)
	}
	return e
}

// TwiceMonthly schedules the job to run twice monthly on the 1st and 16th
func (e *Event) TwiceMonthly(first, second int, timeStr string) *Event {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		e.Expression = fmt.Sprintf("%s %s %d,%d * *", parts[1], parts[0], first, second)
	}
	return e
}

// Quarterly schedules the job to run quarterly on the first day of each quarter
func (e *Event) Quarterly() *Event {
	e.Expression = "0 0 1 1,4,7,10 *"
	return e
}

// Yearly schedules the job to run yearly on the first day of the year
func (e *Event) Yearly() *Event {
	e.Expression = "0 0 1 1 *"
	return e
}

// YearlyOn schedules the job to run yearly on a specific month, day, and time
func (e *Event) YearlyOn(month int, day int, timeStr string) *Event {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		e.Expression = fmt.Sprintf("%s %s %d %d *", parts[1], parts[0], day, month)
	}
	return e
}

// Cron sets a custom cron expression for the event
func (e *Event) Cron(expression string) *Event {
	e.Expression = expression
	return e
}

// Scheduler defines the contract for task schedulers
type Scheduler interface {
	// Job registers a job to be scheduled
	Job(job queueContracts.Job) *Event

	// Call registers a function to be scheduled
	Call(fn func() error) *Event

	// Command registers a console command to be scheduled
	Command(command string, args ...string) *Event

	// GetEvents returns all registered scheduled events
	GetEvents() []*Event

	// Run starts the scheduler process
	Run() error

	// Stop gracefully stops the scheduler process
	Stop() error
}
