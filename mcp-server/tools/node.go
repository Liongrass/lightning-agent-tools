// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package tools

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NodeService handles Lightning node information operations.
type NodeService struct {
	LightningClient lnrpc.LightningClient
}

// NewNodeService creates a new node service.
func NewNodeService(client lnrpc.LightningClient) *NodeService {
	return &NodeService{
		LightningClient: client,
	}
}

// GetInfoTool returns the MCP tool definition for getting node info.
func (s *NodeService) GetInfoTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "lnc_get_info",
		Description: "Get Lightning node information including version, " +
			"peers, and channels",
		InputSchema: ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}
}

// HandleGetInfo handles the node info request.
func (s *NodeService) HandleGetInfo(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	info, err := s.LightningClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return newToolResultError(
			"Failed to get node info: " + err.Error()), nil
	}

	chains := chainNetworks(info.Chains)
	primaryNetwork := ""
	if len(chains) > 0 {
		primaryNetwork = chains[0]
	}

	return newToolResultJSON(map[string]any{
		"node_id":               info.IdentityPubkey,
		"alias":                 info.Alias,
		"version":               info.Version,
		"num_peers":             info.NumPeers,
		"num_active_channels":   info.NumActiveChannels,
		"num_inactive_channels": info.NumInactiveChannels,
		"num_pending_channels":  info.NumPendingChannels,
		"synced_to_chain":       info.SyncedToChain,
		"synced_to_graph":       info.SyncedToGraph,
		"block_height":          info.BlockHeight,
		"block_hash":            info.BlockHash,
		"primary_network":       primaryNetwork,
		"chains":                chains,
	}), nil
}

// GetBalanceTool returns the MCP tool definition for getting wallet balance.
func (s *NodeService) GetBalanceTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_get_balance",
		Description: "Get on-chain wallet balance and channel balance information",
		InputSchema: ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}
}

// HandleGetBalance handles the balance request.
func (s *NodeService) HandleGetBalance(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	// Get on-chain balance.
	walletBalance, err := s.LightningClient.WalletBalance(
		ctx, &lnrpc.WalletBalanceRequest{},
	)
	if err != nil {
		return newToolResultError(
			"Failed to get wallet balance: " + err.Error()), nil
	}

	// Get channel balance.
	channelBalance, err := s.LightningClient.ChannelBalance(
		ctx, &lnrpc.ChannelBalanceRequest{},
	)
	if err != nil {
		return newToolResultError(
			"Failed to get channel balance: " + err.Error()), nil
	}

	local := safeAmount(channelBalance.GetLocalBalance())
	remote := safeAmount(channelBalance.GetRemoteBalance())
	unsettledLocal := safeAmount(channelBalance.GetUnsettledLocalBalance())
	unsettledRemote := safeAmount(
		channelBalance.GetUnsettledRemoteBalance(),
	)
	pendingLocal := safeAmount(
		channelBalance.GetPendingOpenLocalBalance(),
	)
	pendingRemote := safeAmount(
		channelBalance.GetPendingOpenRemoteBalance(),
	)

	return newToolResultJSON(map[string]any{
		"wallet_balance": map[string]any{
			"total_balance":       walletBalance.TotalBalance,
			"confirmed_balance":   walletBalance.ConfirmedBalance,
			"unconfirmed_balance": walletBalance.UnconfirmedBalance,
		},
		"channel_balance": map[string]any{
			"total_balance":               local.sat + remote.sat,
			"pending_open_balance":        pendingLocal.sat + pendingRemote.sat,
			"local_balance":               amountMap(local),
			"remote_balance":              amountMap(remote),
			"unsettled_local_balance":     amountMap(unsettledLocal),
			"unsettled_remote_balance":    amountMap(unsettledRemote),
			"pending_open_local_balance":  amountMap(pendingLocal),
			"pending_open_remote_balance": amountMap(pendingRemote),
		},
	}), nil
}

// amountMap converts a balanceBreakdown to a map for JSON serialization.
func amountMap(b balanceBreakdown) map[string]uint64 {
	return map[string]uint64{"sat": b.sat, "msat": b.msat}
}

type balanceBreakdown struct {
	sat  uint64
	msat uint64
}

func safeAmount(amount *lnrpc.Amount) balanceBreakdown {
	if amount == nil {
		return balanceBreakdown{}
	}
	return balanceBreakdown{sat: amount.Sat, msat: amount.Msat}
}

// chainNetworks extracts chain networks from Chain slice.
func chainNetworks(chains []*lnrpc.Chain) []string {
	networks := make([]string, len(chains))
	for i, chain := range chains {
		networks[i] = chain.Network
	}
	return networks
}
