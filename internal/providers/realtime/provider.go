package realtime

import (
	"fiber-starter/configs"
	realtimeContracts "fiber-starter/internal/providers/realtime/Contracts"
)

// RegisterRealtime initializes and returns the realtime manager as a contract.
func RegisterRealtime(cfg *configs.Config) (realtimeContracts.Manager, error) {
	if !cfg.WebSocket.Enabled {
		return nil, nil
	}
	return NewManager(cfg), nil
}
