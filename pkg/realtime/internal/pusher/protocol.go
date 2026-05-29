package pusher

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

// Message 对应 Pusher 兼容的前端 WebSocket 帧报文格式
type Message struct {
	Event    string          `json:"event"`
	Channel  string          `json:"channel,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	SocketID string          `json:"socket_id,omitempty"`
}

// SubscribePayload 客户端发起通道订阅时的载荷体
type SubscribePayload struct {
	Channel     string `json:"channel"`
	Auth        string `json:"auth,omitempty"`
	ChannelData string `json:"channel_data,omitempty"`
}

// ConnectionEstablishedData 握手连接成功后下发给客户端的 SocketID 和心跳超时设定
type ConnectionEstablishedData struct {
	SocketID        string `json:"socket_id"`
	ActivityTimeout int    `json:"activity_timeout"`
}

// ErrorPayload 通讯发生异常或鉴权被拒时下发的信息体
type ErrorPayload struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

// User 抽象业务层用户实体，仅包含实时通信需要的最小集。
type User struct {
	ID   string         `json:"id"`
	Info map[string]any `json:"info,omitempty"`
}

// Envelope 跨服务节点进行 Redis 广播的分发信封
type Envelope struct {
	NodeID         string          `json:"node_id"`
	Event          string          `json:"event"`
	Channel        string          `json:"channel"`
	Data           json.RawMessage `json:"data,omitempty"`
	OriginSocketID string          `json:"origin_socket_id,omitempty"`
}

func EncodeJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return json.RawMessage(b)
}

func EncodeData(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		raw = []byte(`null`)
	}
	quoted, err := json.Marshal(string(raw))
	if err != nil {
		return json.RawMessage(`"null"`)
	}
	return json.RawMessage(quoted)
}

func EncodeRawData(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		raw = json.RawMessage(`null`)
	}
	quoted, err := json.Marshal(string(raw))
	if err != nil {
		return json.RawMessage(`"null"`)
	}
	return json.RawMessage(quoted)
}

func DecodeSubscribePayload(data []byte) (SubscribePayload, error) {
	var payload SubscribePayload
	if len(data) == 0 {
		return payload, errors.New("empty subscribe payload")
	}
	var encoded string
	if err := json.Unmarshal(data, &encoded); err == nil {
		data = []byte(encoded)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return payload, err
	}
	if payload.Channel == "" {
		return payload, errors.New("missing channel")
	}
	return payload, nil
}
