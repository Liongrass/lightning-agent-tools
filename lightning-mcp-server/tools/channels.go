// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package tools

import (
	"context"
	"strconv"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ChannelService handles Lightning channel operations.
type ChannelService struct {
	LightningClient lnrpc.LightningClient
}

// NewChannelService creates a new channel service.
func NewChannelService(client lnrpc.LightningClient) *ChannelService {
	return &ChannelService{
		LightningClient: client,
	}
}

// ListChannelsTool returns the MCP tool definition for listing channels.
func (s *ChannelService) ListChannelsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_list_channels",
		Description: "List all Lightning channels with detailed information",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"active_only": map[string]any{
					"type":        "boolean",
					"description": "Only return active channels",
				},
				"inactive_only": map[string]any{
					"type":        "boolean",
					"description": "Only return inactive channels",
				},
				"public_only": map[string]any{
					"type":        "boolean",
					"description": "Only return public channels",
				},
				"private_only": map[string]any{
					"type":        "boolean",
					"description": "Only return private channels",
				},
			},
		},
	}
}

// HandleListChannels handles the list channels request.
func (s *ChannelService) HandleListChannels(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	activeOnly, _ := args["active_only"].(bool)
	inactiveOnly, _ := args["inactive_only"].(bool)
	publicOnly, _ := args["public_only"].(bool)
	privateOnly, _ := args["private_only"].(bool)

	channels, err := s.LightningClient.ListChannels(
		ctx, &lnrpc.ListChannelsRequest{
			ActiveOnly:   activeOnly,
			InactiveOnly: inactiveOnly,
			PublicOnly:   publicOnly,
			PrivateOnly:  privateOnly,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to list channels: " + err.Error()), nil
	}

	channelList := make([]map[string]any, len(channels.Channels))
	for i, ch := range channels.Channels {
		entry := map[string]any{
			"active":                  ch.Active,
			"remote_pubkey":           ch.RemotePubkey,
			"channel_point":           ch.ChannelPoint,
			"chan_id":                 strconv.FormatUint(ch.ChanId, 10),
			"capacity":                ch.Capacity,
			"local_balance":           ch.LocalBalance,
			"remote_balance":          ch.RemoteBalance,
			"commit_fee":              ch.CommitFee,
			"commit_weight":           ch.CommitWeight,
			"fee_per_kw":              ch.FeePerKw,
			"unsettled_balance":       ch.UnsettledBalance,
			"total_satoshis_sent":     ch.TotalSatoshisSent,
			"total_satoshis_received": ch.TotalSatoshisReceived,
			"num_updates":             ch.NumUpdates,
			"pending_htlcs":           len(ch.PendingHtlcs),
			"private":                 ch.Private,
			"initiator":               ch.Initiator,
			"chan_status_flags":       ch.ChanStatusFlags,
		}

		if c := ch.GetLocalConstraints(); c != nil {
			entry["local_constraints"] = constraintsToMap(c)
		}
		if c := ch.GetRemoteConstraints(); c != nil {
			entry["remote_constraints"] = constraintsToMap(c)
		}

		channelList[i] = entry
	}

	return newToolResultJSON(map[string]any{
		"channels":       channelList,
		"total_channels": len(channelList),
	}), nil
}

// PendingChannelsTool returns the MCP tool definition for listing pending
// channels.
func (s *ChannelService) PendingChannelsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_pending_channels",
		Description: "List all pending Lightning channels",
		InputSchema: ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}
}

// HandlePendingChannels handles the pending channels request.
func (s *ChannelService) HandlePendingChannels(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	pending, err := s.LightningClient.PendingChannels(
		ctx, &lnrpc.PendingChannelsRequest{},
	)
	if err != nil {
		return newToolResultError(
			"Failed to get pending channels: " + err.Error()), nil
	}

	return newToolResultJSON(map[string]any{
		"pending_open_channels": formatPendingOpenChannels(
			pending.PendingOpenChannels,
		),
		"pending_force_closing_channels": formatPendingForceClosingChannels(
			pending.PendingForceClosingChannels,
		),
		"waiting_close_channels": formatWaitingCloseChannels(
			pending.WaitingCloseChannels,
		),
		"total_limbo_balance": pending.TotalLimboBalance,
	}), nil
}

// constraintsToMap converts channel constraints to a map for JSON output.
func constraintsToMap(c *lnrpc.ChannelConstraints) map[string]any {
	if c == nil {
		return nil
	}

	return map[string]any{
		"csv_delay":            c.CsvDelay,
		"chan_reserve_sat":     c.ChanReserveSat,
		"dust_limit_sat":       c.DustLimitSat,
		"max_pending_amt_msat": c.MaxPendingAmtMsat,
		"min_htlc_msat":        c.MinHtlcMsat,
		"max_accepted_htlcs":   c.MaxAcceptedHtlcs,
	}
}

// formatPendingOpenChannels formats pending open channel data.
func formatPendingOpenChannels(
	channels []*lnrpc.PendingChannelsResponse_PendingOpenChannel,
) []map[string]any {
	result := make([]map[string]any, len(channels))
	for i, ch := range channels {
		result[i] = map[string]any{
			"channel":       formatPendingChannel(ch.Channel),
			"commit_fee":    ch.CommitFee,
			"commit_weight": ch.CommitWeight,
			"fee_per_kw":    ch.FeePerKw,
		}
	}
	return result
}

// formatPendingForceClosingChannels formats force closing channel data.
func formatPendingForceClosingChannels(
	channels []*lnrpc.PendingChannelsResponse_ForceClosedChannel,
) []map[string]any {
	result := make([]map[string]any, len(channels))
	for i, ch := range channels {
		result[i] = map[string]any{
			"channel":             formatPendingChannel(ch.Channel),
			"closing_txid":        ch.ClosingTxid,
			"limbo_balance":       ch.LimboBalance,
			"maturity_height":     ch.MaturityHeight,
			"blocks_til_maturity": ch.BlocksTilMaturity,
			"recovered_balance":   ch.RecoveredBalance,
		}
	}
	return result
}

// formatWaitingCloseChannels formats waiting close channel data.
func formatWaitingCloseChannels(
	channels []*lnrpc.PendingChannelsResponse_WaitingCloseChannel,
) []map[string]any {
	result := make([]map[string]any, len(channels))
	for i, ch := range channels {
		result[i] = map[string]any{
			"channel":       formatPendingChannel(ch.Channel),
			"limbo_balance": ch.LimboBalance,
		}
	}
	return result
}

// formatPendingChannel formats a single pending channel.
func formatPendingChannel(
	ch *lnrpc.PendingChannelsResponse_PendingChannel,
) map[string]any {
	return map[string]any{
		"remote_node_pub": ch.RemoteNodePub,
		"channel_point":   ch.ChannelPoint,
		"capacity":        ch.Capacity,
		"local_balance":   ch.LocalBalance,
		"remote_balance":  ch.RemoteBalance,
	}
}
