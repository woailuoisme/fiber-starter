package jobs

import (
	"context"

	mail "fiber-starter/internal/providers/mail"
	helpers "fiber-starter/internal/support"

	"go.uber.org/zap"
)

type WelcomeEmailJob struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewWelcomeEmailJob(email, name string) *WelcomeEmailJob {
	return &WelcomeEmailJob{
		Email: email,
		Name:  name,
	}
}

func (j *WelcomeEmailJob) TaskName() string {
	return "send_welcome_email"
}

func (j *WelcomeEmailJob) QueueName() string {
	return "default"
}

func (j *WelcomeEmailJob) Handle(ctx context.Context) error {
	helpers.Info("Sending welcome email to", zap.String("email", j.Email))
	return mail.Send(mail.To(j.Email).Subject("Welcome").Html("<h1>Welcome, " + j.Name + "!</h1>"))
}
