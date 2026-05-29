package sse

import "sync"

type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	channels map[string]map[string]*Session
}

func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*Session),
		channels: make(map[string]map[string]*Session),
	}
}

func (h *Hub) Register(session *Session) {
	if session == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[session.ID()] = session
}

func (h *Hub) Sessions() []*Session {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]*Session, 0, len(h.sessions))
	for _, session := range h.sessions {
		out = append(out, session)
	}
	return out
}

func (h *Hub) Unregister(sessionID string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	session := h.sessions[sessionID]
	if session == nil {
		return nil
	}

	delete(h.sessions, sessionID)

	joined := session.ChannelNames()
	for _, channel := range joined {
		members := h.channels[channel]
		if members == nil {
			continue
		}
		delete(members, sessionID)
		if len(members) == 0 {
			delete(h.channels, channel)
		}
	}

	return joined
}

func (h *Hub) Join(session *Session, channel string) int {
	if session == nil || channel == "" {
		return 0
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	members := h.channels[channel]
	if members == nil {
		members = make(map[string]*Session)
		h.channels[channel] = members
	}
	members[session.ID()] = session
	session.AddChannel(channel)
	return len(members)
}

func (h *Hub) Members(channel string) []*Session {
	h.mu.RLock()
	defer h.mu.RUnlock()

	members := h.channels[channel]
	if len(members) == 0 {
		return nil
	}

	out := make([]*Session, 0, len(members))
	for _, session := range members {
		out = append(out, session)
	}
	return out
}
