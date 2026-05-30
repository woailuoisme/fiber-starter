package realtime

import (
	"encoding/json"

	"lfiber/pkg/realtime/internal/pusher"
)

const (
	EventConnectEstablished    = pusher.EventConnectEstablished
	EventSubscribe             = pusher.EventSubscribe
	EventUnsubscribe           = pusher.EventUnsubscribe
	EventPing                  = pusher.EventPing
	EventPong                  = pusher.EventPong
	EventError                 = pusher.EventError
	EventSubscriptionSucceeded = pusher.EventSubscriptionSucceeded
	EventMemberAdded           = pusher.EventMemberAdded
	EventMemberRemoved         = pusher.EventMemberRemoved
)

type (
	Message                   = pusher.Message
	SubscribePayload          = pusher.SubscribePayload
	ConnectionEstablishedData = pusher.ConnectionEstablishedData
	ErrorPayload              = pusher.ErrorPayload
	Envelope                  = pusher.Envelope
)

func encodeJSON(v any) json.RawMessage {
	return pusher.EncodeJSON(v)
}

func encodePusherData(v any) json.RawMessage {
	return pusher.EncodeData(v)
}

func encodePusherRawData(raw json.RawMessage) json.RawMessage {
	return pusher.EncodeRawData(raw)
}
