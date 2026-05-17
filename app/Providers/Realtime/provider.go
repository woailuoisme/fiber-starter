package realtime

import (
	realtimeContracts "fiber-starter/app/Providers/Realtime/Contracts"
	"fiber-starter/configs"
)

// RegisterRealtime initializes and returns the realtime manager as a contract.
func RegisterRealtime(cfg *configs.Config) (realtimeContracts.Manager, error) {
	if !cfg.WebSocket.Enabled {
		return nil, nil
	}
	return NewManager(cfg), nil
}
