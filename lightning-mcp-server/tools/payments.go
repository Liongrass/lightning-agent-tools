// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package tools

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PaymentService handles read-only Lightning payment operations.
type PaymentService struct {
	LightningClient lnrpc.LightningClient
}

// NewPaymentService creates a new payment service for read-only operations.
func NewPaymentService(
	lightningClient lnrpc.LightningClient) *PaymentService {
	return &PaymentService{
		LightningClient: lightningClient,
	}
}

// ListPaymentsTool returns the MCP tool definition for listing payments.
func (s *PaymentService) ListPaymentsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_list_payments",
		Description: "List historical Lightning payments made by this node",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"include_incomplete": map[string]any{
					"type":        "boolean",
					"description": "Include incomplete/failed payments",
				},
				"index_offset": map[string]any{
					"type":        "number",
					"description": "Start index for pagination",
					"minimum":     0,
				},
				"max_payments": map[string]any{
					"type":        "number",
					"description": "Maximum number of payments to return",
					"minimum":     1,
					"maximum":     1000,
				},
				"reversed": map[string]any{
					"type":        "boolean",
					"description": "Return payments in reverse chronological order",
				},
			},
		},
	}
}

// HandleListPayments handles the list payments request.
func (s *PaymentService) HandleListPayments(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	includeIncomplete, _ := args["include_incomplete"].(bool)
	indexOffset, _ := args["index_offset"].(float64)
	maxPayments, _ := args["max_payments"].(float64)
	if maxPayments == 0 {
		maxPayments = 100
	}
	reversed, _ := args["reversed"].(bool)

	resp, err := s.LightningClient.ListPayments(
		ctx, &lnrpc.ListPaymentsRequest{
			IncludeIncomplete: includeIncomplete,
			IndexOffset:       uint64(indexOffset),
			MaxPayments:       uint64(maxPayments),
			Reversed:          reversed,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to list payments: " + err.Error()), nil
	}

	paymentList := make([]map[string]any, len(resp.Payments))
	for i, p := range resp.Payments {
		paymentList[i] = map[string]any{
			"payment_hash":     p.PaymentHash,
			"value_sat":        p.ValueSat,
			"value_msat":       p.ValueMsat,
			"payment_preimage": p.PaymentPreimage,
			"payment_request":  p.PaymentRequest,
			"status":           p.Status.String(),
			"fee_sat":          p.FeeSat,
			"fee_msat":         p.FeeMsat,
			"creation_time_ns": p.CreationTimeNs,
			"payment_index":    p.PaymentIndex,
			"failure_reason":   p.FailureReason.String(),
			"htlc_count":       len(p.Htlcs),
		}
	}

	return newToolResultJSON(map[string]any{
		"payments":           paymentList,
		"first_index_offset": resp.FirstIndexOffset,
		"last_index_offset":  resp.LastIndexOffset,
		"total_payments":     len(paymentList),
	}), nil
}

// TrackPaymentTool returns the MCP tool definition for tracking a payment.
func (s *PaymentService) TrackPaymentTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_track_payment",
		Description: "Track the status of a Lightning payment by its hash",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"payment_hash": map[string]any{
					"type":        "string",
					"description": "Payment hash to track (hex encoded)",
					"pattern":     "^[0-9a-fA-F]{64}$",
				},
			},
			Required: []string{"payment_hash"},
		},
	}
}

// HandleTrackPayment handles the track payment request. It searches for the
// payment by hash in the payment history.
func (s *PaymentService) HandleTrackPayment(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	paymentHash, ok := args["payment_hash"].(string)
	if !ok {
		return newToolResultError("payment_hash is required"), nil
	}

	if len(paymentHash) != 64 {
		return newToolResultError(
			"payment_hash must be a 64-character hex string"), nil
	}

	// Search for the specific payment by listing with the hash filter.
	resp, err := s.LightningClient.ListPayments(
		ctx, &lnrpc.ListPaymentsRequest{
			IncludeIncomplete: true,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to fetch payments: " + err.Error()), nil
	}

	for _, p := range resp.Payments {
		if p.PaymentHash == paymentHash {
			return newToolResultJSON(map[string]any{
				"found":            true,
				"payment_hash":     p.PaymentHash,
				"status":           p.Status.String(),
				"value_sat":        p.ValueSat,
				"fee_sat":          p.FeeSat,
				"creation_time_ns": p.CreationTimeNs,
				"payment_preimage": p.PaymentPreimage,
				"failure_reason":   p.FailureReason.String(),
			}), nil
		}
	}

	return newToolResultJSON(map[string]any{
		"found":   false,
		"message": "Payment not found",
	}), nil
}
