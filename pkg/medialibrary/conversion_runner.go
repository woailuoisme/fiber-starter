package medialibrary

import (
	"context"
	"fmt"
	"strings"
)

const (
	ConversionModeSync  = "sync"
	ConversionModeQueue = "queue"

	DerivedMediaStatusPending   = "pending"
	DerivedMediaStatusCompleted = "completed"
	DerivedMediaStatusFailed    = "failed"
)

type ConversionResult struct {
	Path     string `json:"path"`
	Size     int    `json:"size"`
	MimeType string `json:"mime_type"`
	Status   string `json:"status"`
}

type ConversionRunner struct {
	service *Service
}

func NewConversionRunner(service *Service) *ConversionRunner {
	return &ConversionRunner{service: service}
}

func (r *ConversionRunner) Generate(ctx context.Context, media *Media, rules []*ConversionRule) (map[string]any, string, error) {
	_ = ctx
	results := make(map[string]any)
	if len(rules) == 0 || !strings.HasPrefix(media.MimeType, "image/") {
		return results, DerivedMediaStatusCompleted, nil
	}

	disk := r.service.storage.Disk(media.Disk)
	originalPath := r.service.pathGenerator.OriginalPath(media.UUID, media.FileName)
	original, err := disk.Get(originalPath)
	if err != nil {
		return results, DerivedMediaStatusFailed, fmt.Errorf("%w: read original media: %v", ErrConversionFailed, err)
	}

	status := DerivedMediaStatusCompleted
	var firstErr error
	for _, rule := range rules {
		convData, actualFormat, err := performImageConversion(original, rule)
		if err != nil {
			status = DerivedMediaStatusFailed
			results[rule.Name] = ConversionResult{Status: DerivedMediaStatusFailed}
			if firstErr == nil {
				firstErr = fmt.Errorf("%w: %s: %v", ErrConversionFailed, rule.Name, err)
			}
			continue
		}

		convPath := r.service.pathGenerator.ConversionPath(media.UUID, media.FileName, rule.Name, actualFormat)
		if err := disk.Put(convPath, convData); err != nil {
			status = DerivedMediaStatusFailed
			results[rule.Name] = ConversionResult{Status: DerivedMediaStatusFailed}
			if firstErr == nil {
				firstErr = fmt.Errorf("%w: write %s: %v", ErrConversionFailed, rule.Name, err)
			}
			continue
		}

		results[rule.Name] = ConversionResult{
			Path:     convPath,
			Size:     len(convData),
			MimeType: "image/" + actualFormat,
			Status:   DerivedMediaStatusCompleted,
		}
	}

	return results, status, firstErr
}
