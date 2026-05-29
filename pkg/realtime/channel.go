package realtime

import (
	"errors"
	"sort"
	"strings"
)

type ChannelKind string

const (
	ChannelPublic   ChannelKind = "public"
	ChannelPrivate  ChannelKind = "private"
	ChannelPresence ChannelKind = "presence"
)

// PresenceMember 在线成员的简短标识和扩展属性
type PresenceMember struct {
	UserID   string         `json:"user_id"`
	UserInfo map[string]any `json:"user_info,omitempty"`
}

// PresenceSnapshot 在线成员列表切片的快照，用于向刚加入通道的客户端下发
type PresenceSnapshot struct {
	Count   int              `json:"count"`
	Members []PresenceMember `json:"members"`
}

type pusherPresenceData struct {
	Presence pusherPresenceMembers `json:"presence"`
}

type pusherPresenceMembers struct {
	Count int                       `json:"count"`
	IDs   []string                  `json:"ids"`
	Hash  map[string]map[string]any `json:"hash"`
}

// Channel 实时通道的结构化定义
type Channel struct {
	Name string
	Kind ChannelKind
}

// ParseChannel 解析并校验频道名
func ParseChannel(name string) (Channel, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Channel{}, errors.New("channel is empty")
	}
	if strings.ContainsAny(name, " \t\r\n") {
		return Channel{}, errors.New("channel contains whitespace")
	}

	channel := Channel{Name: name, Kind: ChannelPublic}
	switch {
	case strings.HasPrefix(name, "presence-"):
		channel.Kind = ChannelPresence
	case strings.HasPrefix(name, "private-"):
		channel.Kind = ChannelPrivate
	}
	return channel, nil
}

func (c Channel) IsPrivate() bool {
	return c.Kind == ChannelPrivate || c.Kind == ChannelPresence
}

func (c Channel) IsPresence() bool {
	return c.Kind == ChannelPresence
}

func (c Channel) IsPublic() bool {
	return c.Kind == ChannelPublic
}

func (c Channel) Prefix() string {
	switch c.Kind {
	case ChannelPrivate:
		return "private-"
	case ChannelPresence:
		return "presence-"
	default:
		return ""
	}
}

// Snapshot 根据去重、排序后的在线成员数据生成推送快照
func (c Channel) Snapshot(members []PresenceMember) PresenceSnapshot {
	snapshot := PresenceSnapshot{Count: len(members), Members: members}
	sort.Slice(snapshot.Members, func(i, j int) bool {
		return snapshot.Members[i].UserID < snapshot.Members[j].UserID
	})
	return snapshot
}

func (s PresenceSnapshot) PusherData() pusherPresenceData {
	hash := make(map[string]map[string]any, len(s.Members))
	ids := make([]string, 0, len(s.Members))
	seen := make(map[string]struct{}, len(s.Members))

	for _, member := range s.Members {
		if member.UserID == "" {
			continue
		}
		if _, ok := seen[member.UserID]; ok {
			continue
		}
		seen[member.UserID] = struct{}{}
		ids = append(ids, member.UserID)
		if member.UserInfo == nil {
			hash[member.UserID] = map[string]any{}
		} else {
			hash[member.UserID] = member.UserInfo
		}
	}
	sort.Strings(ids)

	return pusherPresenceData{
		Presence: pusherPresenceMembers{
			Count: len(ids),
			IDs:   ids,
			Hash:  hash,
		},
	}
}
