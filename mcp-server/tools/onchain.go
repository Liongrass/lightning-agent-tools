// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package tools

import (
	"context"
	"fmt"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OnChainService handles read-only on-chain wallet operations.
type OnChainService struct {
	LightningClient lnrpc.LightningClient
}

// NewOnChainService creates a new on-chain service for read-only operations.
func NewOnChainService(client lnrpc.LightningClient) *OnChainService {
	return &OnChainService{
		LightningClient: client,
	}
}

// ListUnspentTool returns the MCP tool definition for listing unspent
// outputs.
func (s *OnChainService) ListUnspentTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_list_unspent",
		Description: "List unspent transaction outputs (UTXOs)",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"min_confs": map[string]any{
					"type":        "number",
					"description": "Minimum confirmations required",
					"minimum":     0,
				},
				"max_confs": map[string]any{
					"type":        "number",
					"description": "Maximum confirmations to include",
					"minimum":     1,
				},
				"account": map[string]any{
					"type":        "string",
					"description": "Account name to filter UTXOs",
				},
			},
		},
	}
}

// HandleListUnspent handles the list unspent request.
func (s *OnChainService) HandleListUnspent(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	minConfs, _ := args["min_confs"].(float64)
	maxConfs, _ := args["max_confs"].(float64)
	if maxConfs == 0 {
		maxConfs = 9999999
	}
	account, _ := args["account"].(string)

	resp, err := s.LightningClient.ListUnspent(
		ctx, &lnrpc.ListUnspentRequest{
			MinConfs: int32(minConfs),
			MaxConfs: int32(maxConfs),
			Account:  account,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to list unspent: " + err.Error()), nil
	}

	utxos := make([]map[string]any, len(resp.Utxos))
	totalAmount := int64(0)

	for i, utxo := range resp.Utxos {
		totalAmount += utxo.AmountSat
		utxos[i] = map[string]any{
			"address":    utxo.Address,
			"amount_sat": utxo.AmountSat,
			"pk_script":  utxo.PkScript,
			"outpoint": fmt.Sprintf(
				"%s:%d",
				utxo.Outpoint.TxidStr,
				utxo.Outpoint.OutputIndex,
			),
			"confirmations": utxo.Confirmations,
		}
	}

	return newToolResultJSON(map[string]any{
		"utxos":            utxos,
		"total_utxos":      len(utxos),
		"total_amount_sat": totalAmount,
	}), nil
}

// GetTransactionsTool returns the MCP tool definition for listing
// transactions.
func (s *OnChainService) GetTransactionsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_get_transactions",
		Description: "Get on-chain transaction history",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"start_height": map[string]any{
					"type":        "number",
					"description": "Starting block height",
					"minimum":     0,
				},
				"end_height": map[string]any{
					"type":        "number",
					"description": "Ending block height",
					"minimum":     0,
				},
				"account": map[string]any{
					"type":        "string",
					"description": "Account name to filter transactions",
				},
			},
		},
	}
}

// HandleGetTransactions handles the get transactions request.
func (s *OnChainService) HandleGetTransactions(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	startHeight, _ := args["start_height"].(float64)
	endHeight, _ := args["end_height"].(float64)
	if endHeight == 0 {
		endHeight = -1
	}
	account, _ := args["account"].(string)

	resp, err := s.LightningClient.GetTransactions(
		ctx, &lnrpc.GetTransactionsRequest{
			StartHeight: int32(startHeight),
			EndHeight:   int32(endHeight),
			Account:     account,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to get transactions: " + err.Error()), nil
	}

	transactions := make([]map[string]any, len(resp.Transactions))
	for i, tx := range resp.Transactions {
		prevOuts := make([]map[string]any, len(tx.PreviousOutpoints))
		for j, prevOut := range tx.PreviousOutpoints {
			prevOuts[j] = map[string]any{
				"outpoint":      prevOut.Outpoint,
				"is_our_output": prevOut.IsOurOutput,
			}
		}

		transactions[i] = map[string]any{
			"tx_hash":            tx.TxHash,
			"amount":             tx.Amount,
			"num_confirmations":  tx.NumConfirmations,
			"block_hash":         tx.BlockHash,
			"block_height":       tx.BlockHeight,
			"time_stamp":         tx.TimeStamp,
			"total_fees":         tx.TotalFees,
			"raw_tx_hex":         tx.RawTxHex,
			"label":              tx.Label,
			"previous_outpoints": prevOuts,
		}
	}

	return newToolResultJSON(map[string]any{
		"transactions":       transactions,
		"total_transactions": len(transactions),
	}), nil
}

// EstimateFeesTool returns the MCP tool definition for estimating fees.
func (s *OnChainService) EstimateFeesTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "lnc_estimate_fee",
		Description: "Estimate on-chain transaction fees for a given " +
			"confirmation target",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"target_conf": map[string]any{
					"type":        "number",
					"description": "Target number of confirmations (default: 6)",
					"minimum":     1,
					"maximum":     144,
				},
			},
		},
	}
}

// HandleEstimateFee handles the estimate fee request.
func (s *OnChainService) HandleEstimateFee(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	targetConf, _ := args["target_conf"].(float64)
	if targetConf == 0 {
		targetConf = 6
	}

	resp, err := s.LightningClient.EstimateFee(
		ctx, &lnrpc.EstimateFeeRequest{
			TargetConf: int32(targetConf),
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to get fee estimate: " + err.Error()), nil
	}

	return newToolResultJSON(map[string]any{
		"fee_estimates": map[string]any{
			fmt.Sprintf("target_%d_blocks", int(targetConf)): map[string]any{
				"fee_sat":       resp.FeeSat,
				"sat_per_vbyte": resp.SatPerVbyte,
			},
		},
	}), nil
}
