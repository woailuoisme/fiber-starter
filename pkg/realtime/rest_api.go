package realtime

import (
	"context"
	"crypto/hmac"
	"crypto/md5" //nolint:gosec // Pusher REST API uses body_md5 as part of its signature contract.
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"lfiber/pkg/realtime/internal/pusher"

	"github.com/gofiber/fiber/v3"
)

type triggerEventRequest struct {
	Name     string          `json:"name"`
	Channel  string          `json:"channel"`
	Channels []string        `json:"channels"`
	Data     json.RawMessage `json:"data"`
	SocketID string          `json:"socket_id"`
}

func (m *ManagerImpl) handleAPI(c fiber.Ctx) error {
	if err := m.validateRESTSignature(c); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	switch {
	case c.Method() == fiber.MethodPost && c.Params("appID") != "" && strings.HasSuffix(c.Path(), "/events"):
		return m.handleTriggerEvent(c)
	case c.Method() == fiber.MethodGet && c.Params("channel") != "" && strings.HasSuffix(c.Path(), "/users"):
		return m.handlePresenceUsers(c)
	case c.Method() == fiber.MethodGet && c.Params("channel") != "":
		return m.handleChannelInfo(c)
	case c.Method() == fiber.MethodGet:
		return m.handleChannels(c)
	default:
		return fiber.ErrNotFound
	}
}

func (m *ManagerImpl) validateRESTSignature(c fiber.Ctx) error {
	if m.cfg == nil {
		return errors.New("realtime config unavailable")
	}
	if m.cfg.AppID != "" && c.Params("appID") != "" && c.Params("appID") != m.cfg.AppID {
		return errors.New("invalid app id")
	}

	q := c.Queries()
	if q["auth_key"] != m.cfg.AppKey {
		return errors.New("invalid auth key")
	}
	if q["auth_version"] != "1.0" {
		return errors.New("invalid auth version")
	}
	if q["auth_signature"] == "" {
		return errors.New("missing auth signature")
	}

	if ts := strings.TrimSpace(q["auth_timestamp"]); ts != "" {
		parsed, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			if unix, parseErr := parseUnixTimestamp(ts); parseErr == nil {
				parsed = time.Unix(unix, 0)
			}
		}
		if !parsed.IsZero() && time.Since(parsed) > 10*time.Minute {
			return errors.New("auth timestamp expired")
		}
	}

	if bodyMD5 := strings.TrimSpace(q["body_md5"]); bodyMD5 != "" {
		sum := md5.Sum(c.Body()) //nolint:gosec // Pusher protocol specifies MD5 hash for request body verification.
		if !hmac.Equal([]byte(bodyMD5), []byte(hex.EncodeToString(sum[:]))) {
			return errors.New("invalid body md5")
		}
	}

	expected := signRESTRequest(m.cfg.AppSecret, c.Method(), c.Path(), q)
	if !hmac.Equal([]byte(expected), []byte(q["auth_signature"])) {
		return errors.New("invalid auth signature")
	}
	return nil
}

func parseUnixTimestamp(value string) (int64, error) {
	var out int64
	_, err := fmt.Sscan(value, &out)
	return out, err
}

func signRESTRequest(secret, method, path string, query map[string]string) string {
	keys := make([]string, 0, len(query))
	for key := range query {
		if key != "auth_signature" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	values := url.Values{}
	for _, key := range keys {
		values.Set(key, query[key])
	}

	stringToSign := strings.ToUpper(method) + "\n" + path + "\n" + values.Encode()
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stringToSign))
	return hex.EncodeToString(mac.Sum(nil))
}

func (m *ManagerImpl) handleTriggerEvent(c fiber.Ctx) error {
	var req triggerEventRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing event name")
	}

	channels := req.Channels
	if req.Channel != "" {
		channels = append(channels, req.Channel)
	}
	if len(channels) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "missing channel")
	}
	if len(req.Data) == 0 {
		req.Data = json.RawMessage(`null`)
	}
	data := req.Data
	var encoded string
	if err := json.Unmarshal(req.Data, &encoded); err == nil {
		if json.Valid([]byte(encoded)) {
			data = json.RawMessage(encoded)
		} else {
			data = encodeJSON(encoded)
		}
	}

	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		m.publishEnvelope(pusher.Envelope{
			NodeID:         m.nodeID,
			Event:          req.Name,
			Channel:        channel,
			Data:           data,
			OriginSocketID: req.SocketID,
		})
	}

	return c.JSON(fiber.Map{})
}

func (m *ManagerImpl) handleChannels(c fiber.Ctx) error {
	filter := c.Query("filter_by_prefix")
	info := parseInfo(c.Query("info"))

	channels := m.channelState(filter, info)
	if channels == nil {
		channels = map[string]map[string]any{}
	}

	return c.JSON(fiber.Map{"channels": channels})
}

func (m *ManagerImpl) handleChannelInfo(c fiber.Ctx) error {
	channelName := c.Params("channel")
	info := parseInfo(c.Query("info"))

	count := m.hub.Count(channelName)
	resp := fiber.Map{"occupied": count > 0}
	if info["subscription_count"] {
		resp["subscription_count"] = count
	}
	if info["user_count"] {
		channel, err := ParseChannel(channelName)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if !channel.IsPresence() {
			return fiber.NewError(fiber.StatusBadRequest, "user_count is only available for presence channels")
		}
		members, err := m.presenceMembers(channelName)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		resp["user_count"] = len(uniquePresenceMembers(members))
	}

	return c.JSON(resp)
}

func (m *ManagerImpl) handlePresenceUsers(c fiber.Ctx) error {
	channelName := c.Params("channel")
	channel, err := ParseChannel(channelName)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if !channel.IsPresence() {
		return fiber.NewError(fiber.StatusBadRequest, "users are only available for presence channels")
	}

	members, err := m.presenceMembers(channelName)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	users := make([]fiber.Map, 0, len(members))
	for _, member := range uniquePresenceMembers(members) {
		users = append(users, fiber.Map{"id": member.UserID})
	}
	return c.JSON(fiber.Map{"users": users})
}

func (m *ManagerImpl) channelState(filter string, info map[string]bool) map[string]map[string]any {
	channels := m.hub.Channels()
	out := make(map[string]map[string]any, len(channels))
	for channel, count := range channels {
		if filter != "" && !strings.HasPrefix(channel, filter) {
			continue
		}
		row := map[string]any{}
		if info["subscription_count"] {
			row["subscription_count"] = count
		}
		if info["user_count"] {
			if parsed, err := ParseChannel(channel); err == nil && parsed.IsPresence() {
				if members, err := m.presenceMembers(channel); err == nil {
					row["user_count"] = len(uniquePresenceMembers(members))
				}
			}
		}
		out[channel] = row
	}
	return out
}

func (m *ManagerImpl) presenceMembers(channel string) ([]pusher.PresenceMember, error) {
	if m.presence == nil {
		return []pusher.PresenceMember{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	members, err := m.presence.Members(ctx, channel)
	if err != nil {
		return nil, err
	}
	if members == nil {
		return []pusher.PresenceMember{}, nil
	}
	return members, nil
}

func parseInfo(value string) map[string]bool {
	out := map[string]bool{}
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out[item] = true
		}
	}
	return out
}

func uniquePresenceMembers(members []pusher.PresenceMember) []pusher.PresenceMember {
	seen := map[string]pusher.PresenceMember{}
	for _, member := range members {
		if member.UserID == "" {
			continue
		}
		if _, ok := seen[member.UserID]; !ok {
			seen[member.UserID] = member
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]pusher.PresenceMember, 0, len(ids))
	for _, id := range ids {
		out = append(out, seen[id])
	}
	return out
}
