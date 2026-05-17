package Contracts

import (
	models "fiber-starter/app/Models"

	fiber "github.com/gofiber/fiber/v3"
)

// Guard defines the contract for an authentication guard (similar to Laravel's Guard)
type Guard interface {
	// Check determines if the current user is authenticated
	Check(c fiber.Ctx) bool

	// Guest determines if the current user is a guest
	Guest(c fiber.Ctx) bool

	// User retrieves the currently authenticated user
	User(c fiber.Ctx) *models.User

	// Id retrieves the ID for the currently authenticated user
	Id(c fiber.Ctx) int64

	// SetUser sets the current user for the given context
	SetUser(c fiber.Ctx, user *models.User)

	// Validate validates a user's credentials without logging them in
	Validate(credentials map[string]string) bool

	// Attempt attempts to authenticate a user using the given credentials
	Attempt(c fiber.Ctx, credentials map[string]string) bool

	// Login logs the given user instance into the application
	Login(c fiber.Ctx, user *models.User) error

	// LoginUsingId logs the user with the given ID into the application
	LoginUsingId(c fiber.Ctx, id int64) error

	// Logout logs the user out of the application
	Logout(c fiber.Ctx) error
}
