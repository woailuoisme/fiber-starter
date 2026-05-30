package realtime

import "lfiber/pkg/realtime/internal/sse"

type (
	SSEEnvelope = sse.Envelope
	SSEFrame    = sse.Frame
)

func NewSSEFrame(env Envelope) (SSEFrame, error) {
	return sse.NewFrame(env)
}
