package realtime

import (
	"errors"
	"strings"
)

type ChannelKind string

const (
	ChannelPublic   ChannelKind = "public"
	ChannelPrivate  ChannelKind = "private"
	ChannelPresence ChannelKind = "presence"
)

type Channel struct {
	Name string
	Kind ChannelKind
}

func (c Channel) IsPublic() bool {
	return c.Kind == ChannelPublic
}

func (c Channel) IsPrivate() bool {
	return c.Kind == ChannelPrivate
}

func (c Channel) IsPresence() bool {
	return c.Kind == ChannelPresence
}

type PresenceMember struct {
	UserID   string         `json:"user_id"`
	UserInfo map[string]any `json:"user_info,omitempty"`
}

type PresenceSnapshot struct {
	Channel string                    `json:"channel"`
	Members map[string]PresenceMember `json:"members"`
}

func ParseChannel(name string) (Channel, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Channel{}, errors.New("channel name cannot be empty")
	}

	kind := ChannelPublic
	if strings.HasPrefix(name, "private:") {
		kind = ChannelPrivate
	} else if strings.HasPrefix(name, "presence:") {
		kind = ChannelPresence
	}

	return Channel{
		Name: name,
		Kind: kind,
	}, nil
}
