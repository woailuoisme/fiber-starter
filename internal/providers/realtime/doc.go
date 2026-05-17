// Package realtime implements an embedded, Pusher-like realtime layer for Fiber.
//
// Design boundaries:
//   - socketio is used only as the transport adapter.
//   - Protocol, channel semantics, auth, presence, hub, and bus live in this package.
//   - The package intentionally stays single-package to keep the public surface small.
//
// File responsibilities:
//   - server.go: Fiber/socketio wiring and socket lifecycle registration.
//   - session.go: per-connection lifecycle, pumps, heartbeat, and write queue.
//   - hub.go: local connection registry and channel membership tracking.
//   - channel.go: public/private/presence channel classification and snapshot types.
//   - protocol.go: wire format, system events, and message encoding helpers.
//   - auth.go: private/presence auth request/response and signature helpers.
//   - bus.go and redis_bus.go: cross-instance publish/subscribe abstraction.
//   - presence_store.go: presence member storage for memory and Redis/Valkey modes.
//   - manager.go: orchestration entrypoint used by providers and routes.
package realtime
