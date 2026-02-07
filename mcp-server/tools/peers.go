// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package tools

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PeerService handles read-only Lightning peer operations.
type PeerService struct {
	LightningClient lnrpc.LightningClient
}

// NewPeerService creates a new peer service for read-only operations.
func NewPeerService(client lnrpc.LightningClient) *PeerService {
	return &PeerService{
		LightningClient: client,
	}
}

// ListPeersTool returns the MCP tool definition for listing peers.
func (s *PeerService) ListPeersTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "lnc_list_peers",
		Description: "List all connected Lightning Network peers with " +
			"detailed connection information",
		InputSchema: ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}
}

// HandleListPeers handles the list peers request.
func (s *PeerService) HandleListPeers(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	peers, err := s.LightningClient.ListPeers(
		ctx, &lnrpc.ListPeersRequest{},
	)
	if err != nil {
		return newToolResultError(
			"Failed to list peers: " + err.Error()), nil
	}

	peerList := make([]map[string]any, len(peers.Peers))
	for i, peer := range peers.Peers {
		features := make([]map[string]any, 0, len(peer.Features))
		for featureKey, feature := range peer.Features {
			features = append(features, map[string]any{
				"feature":     featureKey,
				"name":        feature.Name,
				"is_required": feature.IsRequired,
				"is_known":    feature.IsKnown,
			})
		}

		peerList[i] = map[string]any{
			"pub_key":    peer.PubKey,
			"address":    peer.Address,
			"bytes_sent": peer.BytesSent,
			"bytes_recv": peer.BytesRecv,
			"sat_sent":   peer.SatSent,
			"sat_recv":   peer.SatRecv,
			"inbound":    peer.Inbound,
			"ping_time":  peer.PingTime,
			"sync_type":  peer.SyncType.String(),
			"features":   features,
			"errors":     formatPeerErrors(peer.Errors),
			"flap_count": peer.FlapCount,
		}
	}

	return newToolResultJSON(map[string]any{
		"peers":       peerList,
		"total_peers": len(peerList),
	}), nil
}

// DescribeGraphTool returns the MCP tool definition for getting network
// graph information.
func (s *PeerService) DescribeGraphTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "lnc_describe_graph",
		Description: "Get Lightning Network graph information including " +
			"nodes and channels",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"include_unannounced": map[string]any{
					"type":        "boolean",
					"description": "Include unannounced channels in the graph",
				},
			},
		},
	}
}

// HandleDescribeGraph handles the describe graph request.
func (s *PeerService) HandleDescribeGraph(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	includeUnannounced, _ := args["include_unannounced"].(bool)

	graph, err := s.LightningClient.DescribeGraph(
		ctx, &lnrpc.ChannelGraphRequest{
			IncludeUnannounced: includeUnannounced,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to describe graph: " + err.Error()), nil
	}

	// Sample first few nodes and edges to avoid overwhelming output.
	const maxSamples = 5

	sampleNodes := make([]map[string]any, 0, maxSamples)
	for i, node := range graph.Nodes {
		if i >= maxSamples {
			break
		}
		addresses := make([]string, len(node.Addresses))
		for j, addr := range node.Addresses {
			addresses[j] = addr.Addr
		}
		sampleNodes = append(sampleNodes, map[string]any{
			"pub_key":   node.PubKey,
			"alias":     node.Alias,
			"addresses": addresses,
			"color":     node.Color,
		})
	}

	sampleEdges := make([]map[string]any, 0, maxSamples)
	for i, edge := range graph.Edges {
		if i >= maxSamples {
			break
		}
		sampleEdges = append(sampleEdges, map[string]any{
			"channel_id": edge.ChannelId,
			"chan_point": edge.ChanPoint,
			"node1_pub":  edge.Node1Pub,
			"node2_pub":  edge.Node2Pub,
			"capacity":   edge.Capacity,
		})
	}

	return newToolResultJSON(map[string]any{
		"total_nodes":         len(graph.Nodes),
		"total_edges":         len(graph.Edges),
		"include_unannounced": includeUnannounced,
		"sample_nodes":        sampleNodes,
		"sample_edges":        sampleEdges,
	}), nil
}

// GetNodeInfoTool returns the MCP tool definition for getting specific node
// information.
func (s *PeerService) GetNodeInfoTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "lnc_get_node_info",
		Description: "Get detailed information about a specific " +
			"Lightning Network node",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"pub_key": map[string]any{
					"type":        "string",
					"description": "Public key of the node to get info for (hex encoded)",
					"pattern":     "^[0-9a-fA-F]{66}$",
				},
				"include_channels": map[string]any{
					"type":        "boolean",
					"description": "Include the node's channels in the response",
				},
			},
			Required: []string{"pub_key"},
		},
	}
}

// HandleGetNodeInfo handles the get node info request.
func (s *PeerService) HandleGetNodeInfo(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	pubKey, ok := args["pub_key"].(string)
	if !ok {
		return newToolResultError("pub_key is required"), nil
	}

	includeChannels, _ := args["include_channels"].(bool)

	nodeInfo, err := s.LightningClient.GetNodeInfo(
		ctx, &lnrpc.NodeInfoRequest{
			PubKey:          pubKey,
			IncludeChannels: includeChannels,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to get node info: " + err.Error()), nil
	}

	addresses := make([]string, len(nodeInfo.Node.Addresses))
	for i, addr := range nodeInfo.Node.Addresses {
		addresses[i] = addr.Addr
	}

	nodeData := map[string]any{
		"pub_key":        nodeInfo.Node.PubKey,
		"alias":          nodeInfo.Node.Alias,
		"addresses":      addresses,
		"color":          nodeInfo.Node.Color,
		"num_channels":   nodeInfo.NumChannels,
		"total_capacity": nodeInfo.TotalCapacity,
	}

	if includeChannels && len(nodeInfo.Channels) > 0 {
		channels := make([]map[string]any, len(nodeInfo.Channels))
		for i, ch := range nodeInfo.Channels {
			channels[i] = map[string]any{
				"channel_id": ch.ChannelId,
				"chan_point": ch.ChanPoint,
				"node1_pub":  ch.Node1Pub,
				"node2_pub":  ch.Node2Pub,
				"capacity":   ch.Capacity,
			}
		}
		nodeData["channels"] = channels
	}

	return newToolResultJSON(nodeData), nil
}

// formatPeerErrors formats peer error information for JSON output.
func formatPeerErrors(
	errors []*lnrpc.TimestampedError) []map[string]any {
	result := make([]map[string]any, len(errors))
	for i, e := range errors {
		result[i] = map[string]any{
			"error":     e.Error,
			"timestamp": e.Timestamp,
		}
	}
	return result
}
