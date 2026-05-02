// Package main is the entry point for the HTTP server application.
package main

import (
	helpers "fiber-starter/app/Support"
	bootstrap "fiber-starter/bootstrap"

	"go.uber.org/zap"
)

func main() {
	if err := bootstrap.App(); err != nil {
		helpers.Fatal("server_bootstrap_failed", zap.Error(err))
	}
}
