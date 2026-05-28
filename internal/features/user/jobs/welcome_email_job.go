package jobs

import (
	"context"

	mail "lfiber/internal/providers/mail"
	queueContracts "lfiber/internal/providers/queue/contracts"
	helpers "lfiber/internal/support"

	"go.uber.org/zap"
)

type WelcomeEmailJob struct {
	queueContracts.JobMeta
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewWelcomeEmailJob(email, name string) *WelcomeEmailJob {
	return &WelcomeEmailJob{
		JobMeta: queueContracts.NewJobMeta("send_welcome_email", "default"),
		Email:   email,
		Name:    name,
	}
}

func (j *WelcomeEmailJob) Handle(ctx context.Context) error {
	helpers.Info("Sending welcome email to", zap.String("email", j.Email))
	return mail.Send(mail.To(j.Email).Subject("Welcome").Html("<h1>Welcome, " + j.Name + "!</h1>"))
}
