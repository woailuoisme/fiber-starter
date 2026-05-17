package realtime

import (
	"encoding/json"
	"errors"
)

const (
	EventConnectEstablished    = "pusher:connection_established"
	EventSubscribe             = "pusher:subscribe"
	EventUnsubscribe           = "pusher:unsubscribe"
	EventPing                  = "pusher:ping"
	EventPong                  = "pusher:pong"
	EventError                 = "pusher:error"
	EventSubscriptionSucceeded = "pusher_internal:subscription_succeeded"
	EventMemberAdded           = "pusher_internal:member_added"
	EventMemberRemoved         = "pusher_internal:member_removed"
)

// Message is the Pusher-like wire format used by realtime events.
type Message struct {
	Event    string          `json:"event"`
	Channel  string          `json:"channel,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	SocketID string          `json:"socket_id,omitempty"`
}

// SubscribePayload is the payload sent by the client for subscriptions.
type SubscribePayload struct {
	Channel     string `json:"channel"`
	Auth        string `json:"auth,omitempty"`
	ChannelData string `json:"channel_data,omitempty"`
}

// ConnectionEstablishedData is sent once a socket connection is accepted.
type ConnectionEstablishedData struct {
	SocketID        string `json:"socket_id"`
	ActivityTimeout int    `json:"activity_timeout"`
}

// ErrorPayload is returned when the server needs to reject or report an error.
type ErrorPayload struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

// Envelope is the internal broadcast payload used across nodes.
type Envelope struct {
	NodeID         string          `json:"node_id"`
	Event          string          `json:"event"`
	Channel        string          `json:"channel"`
	Data           json.RawMessage `json:"data,omitempty"`
	OriginSocketID string          `json:"origin_socket_id,omitempty"`
}

func encodeJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return json.RawMessage(b)
}

func decodeSubscribePayload(data []byte) (SubscribePayload, error) {
	var payload SubscribePayload
	if len(data) == 0 {
		return payload, errors.New("empty subscribe payload")
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, err
	}
	if payload.Channel == "" {
		return payload, errors.New("missing channel")
	}
	return payload, nil
}

func encodeMessage(event string, channel string, data any, socketID string) ([]byte, error) {
	msg := Message{
		Event:    event,
		Channel:  channel,
		Data:     encodeJSON(data),
		SocketID: socketID,
	}
	return json.Marshal(msg)
}
