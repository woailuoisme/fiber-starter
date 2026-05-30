package realtime

import "lfiber/pkg/realtime/internal/pusher"

type ChannelKind = pusher.ChannelKind

const (
	ChannelPublic   = pusher.ChannelPublic
	ChannelPrivate  = pusher.ChannelPrivate
	ChannelPresence = pusher.ChannelPresence
)

type (
	PresenceMember   = pusher.PresenceMember
	PresenceSnapshot = pusher.PresenceSnapshot
	Channel          = pusher.Channel
)

func ParseChannel(name string) (Channel, error) {
	return pusher.ParseChannel(name)
}
