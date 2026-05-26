package tests

import (
	"bytes"
	"testing"

	"lfiber/internal/console/ui"

	"github.com/stretchr/testify/assert"
)

func TestProgressBar_Rendering(t *testing.T) {
	var buf bytes.Buffer
	// Create a progress bar with 10 total steps
	bar := ui.NewProgressBar(&buf, 10)
	bar.SetWidth(20)

	// Since we are writing to a buffer (non-TTY), it should fall back to non-TTY mode
	bar.Advance(2)
	assert.Contains(t, buf.String(), "Progress: 2/10 (20%)")

	buf.Reset()
	bar.Finish()
	// Completing should write final state and append a newline
	assert.Contains(t, buf.String(), "Progress: 10/10 (100%)")
}
