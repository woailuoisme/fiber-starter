package realtime

import (
	"lfiber/configs"
	realtimeContracts "lfiber/internal/providers/realtime/contracts"
)

// RegisterRealtime initializes and returns the realtime manager as a contract.
func RegisterRealtime(cfg *configs.Config) (realtimeContracts.Manager, error) {
	if !cfg.WebSocket.Enabled {
		return nil, nil
	}
	return NewManager(cfg), nil
}
