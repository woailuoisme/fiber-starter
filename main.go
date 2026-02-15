//go:build !cli
// +build !cli

// Package main 是应用程序的入口点
package main

import "fiber-starter/bootstrap"

func main() {
	bootstrap.App()
}
