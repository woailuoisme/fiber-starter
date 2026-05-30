package realtime

import (
	"encoding/json"
	"errors"
	"fmt"

	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/ui"
	"lfiber/pkg/realtime"

	"github.com/spf13/cobra"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{
		broadcastCommand(),
		cleanupCommand(),
		statusCommand(),
	}
}

func broadcastCommand() *cobra.Command {
	var channel, event, dataStr string

	cmd := &cobra.Command{
		Use:     "realtime:broadcast",
		Short:   "Broadcast an event payload to a specific realtime channel",
		GroupID: "system",
		RunE: func(cmd *cobra.Command, args []string) error {
			if channel == "" || event == "" {
				return errors.New("missing channel or event parameter")
			}

			rt, err := commandutil.BuildRuntime()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(rt) }()

			if rt.Realtime == nil {
				return errors.New("realtime service is not registered/enabled")
			}

			var payload any
			if dataStr != "" {
				var jsonVal any
				if err := json.Unmarshal([]byte(dataStr), &jsonVal); err == nil {
					payload = jsonVal
				} else {
					payload = dataStr
				}
			}

			if err := rt.Realtime.Dispatch(channel, event, payload); err != nil {
				ui.Error(cmd.OutOrStderr(), "Failed to broadcast message: %v", err)
				return err
			}

			ui.Success(cmd.OutOrStdout(), "Broadcast sent successfully [Channel: %s, Event: %s]", channel, event)
			return nil
		},
	}

	cmd.Flags().StringVar(&channel, "channel", "", "The destination channel name (required)")
	cmd.Flags().StringVar(&event, "event", "", "The event signature/name (required)")
	cmd.Flags().StringVar(&dataStr, "data", "", "The event data string payload (JSON supported)")
	_ = cmd.MarkFlagRequired("channel")
	_ = cmd.MarkFlagRequired("event")

	return cmd
}

func cleanupCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "realtime:cleanup",
		Short:   "N/A: Presence lifecycle is fully managed by Centrifugo",
		GroupID: "system",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Centrifugo 原生管理 Presence TTL，Fiber 侧无需手动清理
			ui.Info(cmd.OutOrStdout(), "Presence cleanup is handled by Centrifugo natively (via presence_ttl config).")
			ui.Info(cmd.OutOrStdout(), "No manual cleanup is required on the application side.")
			return nil
		},
	}
}

func statusCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "realtime:status",
		Short:   "Display current realtime cluster status and online channels information",
		GroupID: "system",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := commandutil.BuildRuntime()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(rt) }()

			if rt.Realtime == nil {
				return errors.New("realtime service is not registered/enabled")
			}

			mgr, ok := rt.Realtime.(*realtime.ManagerImpl)
			if !ok {
				return errors.New("realtime provider is not a *realtime.ManagerImpl instance")
			}

			ctx := cmd.Context()
			nodeID := mgr.GetNodeID()

			ui.Info(cmd.OutOrStdout(), "--- Realtime Status ---")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current Instance Node ID : %s\n", ui.Highlight(cmd.OutOrStdout(), nodeID))

			rdb := mgr.Config().RedisClient
			if rdb != nil {
				busPrefix := mgr.Config().RedisPrefix
				if busPrefix == "" {
					busPrefix = "realtime"
				}

				nodeKeys, err := rdb.Keys(ctx, fmt.Sprintf("%s:node:*:heartbeat", busPrefix)).Result()
				if err == nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Active Cluster Node Count: %d\n", len(nodeKeys))
					for _, key := range nodeKeys {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", key)
					}
				}

				presenceKeys, err := rdb.Keys(ctx, fmt.Sprintf("%s:presence:*", busPrefix)).Result()
				if err == nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Presence Channels Count  : %d\n", len(presenceKeys))
					for _, pKey := range presenceKeys {
						total, _ := rdb.HLen(ctx, pKey).Result()
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  - %s (Members: %d)\n", pKey, total)
					}
				}
			} else {
				ui.Warning(cmd.OutOrStdout(), "Bus mode is Memory. Clustering statistics are unavailable.")
			}

			return nil
		},
	}
}
