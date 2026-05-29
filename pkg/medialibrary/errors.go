package medialibrary

import "errors"

var (
	ErrNoFileContent      = errors.New("media library: no file content")
	ErrFileTooLarge       = errors.New("media library: file too large")
	ErrMimeNotAllowed     = errors.New("media library: mime type not allowed")
	ErrFileNameNotAllowed = errors.New("media library: file name not allowed")
	ErrDiskWriteFailed    = errors.New("media library: disk write failed")
	ErrRecordCreateFailed = errors.New("media library: record create failed")
	ErrConversionFailed   = errors.New("media library: conversion failed")
)
