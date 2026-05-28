package artisan

import (
	"errors"

	"lfiber/internal/providers/artisan/contracts"
	"lfiber/internal/support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// Facade returns the Artisan service from the application container.
func Facade() contracts.Artisan {
	rt := appctx.App()
	if rt == nil {
		return nil
	}

	type artisanProvider interface {
		ArtisanService() contracts.Artisan
	}

	if provider, ok := rt.(artisanProvider); ok {
		return provider.ArtisanService()
	}
	return nil
}

// Call runs a console command through the application Artisan service.
func Call(command string, args ...string) (contracts.Result, error) {
	if service := Facade(); service != nil {
		return service.Call(command, args...)
	}
	return contracts.Result{ExitCode: 1}, ErrContainerNotInitialized
}

// List returns all registered console commands.
func List() []contracts.CommandInfo {
	if service := Facade(); service != nil {
		return service.List()
	}
	return nil
}

// Has reports whether a console command exists.
func Has(command string) bool {
	if service := Facade(); service != nil {
		return service.Has(command)
	}
	return false
}
