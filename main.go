//go:build !cli
// +build !cli

package main

import "fiber-starter/bootstrap"

func main() {
	bootstrap.App()
}
