package realtime

import "lfiber/pkg/realtime/internal/pusher"

type ChannelKind = pusher.ChannelKind

const (
	ChannelPublic   = pusher.ChannelPublic
	ChannelPrivate  = pusher.ChannelPrivate
	ChannelPresence = pusher.ChannelPresence
)

type PresenceMember = pusher.PresenceMember
type PresenceSnapshot = pusher.PresenceSnapshot
type Channel = pusher.Channel

func ParseChannel(name string) (Channel, error) {
	return pusher.ParseChannel(name)
}
