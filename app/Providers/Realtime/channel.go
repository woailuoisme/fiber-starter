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

type PresenceMember struct {
	UserID   string         `json:"user_id"`
	UserInfo map[string]any `json:"user_info,omitempty"`
}

type PresenceSnapshot struct {
	Count   int              `json:"count"`
	Members []PresenceMember `json:"members"`
}

// Channel describes the classification of a realtime channel.
type Channel struct {
	Name string
	Kind ChannelKind
}

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

func (c Channel) Snapshot(members []PresenceMember) PresenceSnapshot {
	snapshot := PresenceSnapshot{Count: len(members), Members: members}
	sort.Slice(snapshot.Members, func(i, j int) bool {
		return snapshot.Members[i].UserID < snapshot.Members[j].UserID
	})
	return snapshot
}
