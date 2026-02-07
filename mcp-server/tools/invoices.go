// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package tools

import (
	"context"
	"encoding/hex"
	"strconv"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InvoiceService handles read-only Lightning invoice operations.
type InvoiceService struct {
	LightningClient lnrpc.LightningClient
}

// NewInvoiceService creates a new invoice service for read-only operations.
func NewInvoiceService(client lnrpc.LightningClient) *InvoiceService {
	return &InvoiceService{
		LightningClient: client,
	}
}

// DecodeInvoiceTool returns the MCP tool definition for decoding invoices.
func (s *InvoiceService) DecodeInvoiceTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_decode_invoice",
		Description: "Decode a BOLT11 Lightning invoice to inspect its contents",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"invoice": map[string]any{
					"type":        "string",
					"description": "BOLT11 invoice string to decode",
					"pattern":     "^ln[a-z0-9]+$",
				},
			},
			Required: []string{"invoice"},
		},
	}
}

// HandleDecodeInvoice handles the decode invoice request.
func (s *InvoiceService) HandleDecodeInvoice(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	invoice, ok := args["invoice"].(string)
	if !ok {
		return newToolResultError("invoice is required"), nil
	}

	// Basic BOLT11 validation.
	if len(invoice) < 3 || invoice[:2] != "ln" {
		return newToolResultError(
			"invalid BOLT11 invoice format"), nil
	}

	decoded, err := s.LightningClient.DecodePayReq(
		ctx, &lnrpc.PayReqString{PayReq: invoice},
	)
	if err != nil {
		return newToolResultError(
			"Failed to decode invoice: " + err.Error()), nil
	}

	// Format route hints.
	routeHints := make([]map[string]any, len(decoded.RouteHints))
	for i, hint := range decoded.RouteHints {
		hops := make([]map[string]any, len(hint.HopHints))
		for j, hop := range hint.HopHints {
			hops[j] = map[string]any{
				"node_id":    hop.NodeId,
				"chan_id":    hop.ChanId,
				"fee_base":   hop.FeeBaseMsat,
				"fee_prop":   hop.FeeProportionalMillionths,
				"cltv_delta": hop.CltvExpiryDelta,
			}
		}
		routeHints[i] = map[string]any{"hop_hints": hops}
	}

	// Format features.
	features := make(map[string]bool)
	for k, v := range decoded.Features {
		features[strconv.FormatUint(uint64(k), 10)] = v.IsKnown
	}

	return newToolResultJSON(map[string]any{
		"destination":      decoded.Destination,
		"payment_hash":     decoded.PaymentHash,
		"amount_sats":      decoded.NumSatoshis,
		"amount_msat":      decoded.NumMsat,
		"timestamp":        decoded.Timestamp,
		"expiry":           decoded.Expiry,
		"description":      decoded.Description,
		"description_hash": decoded.DescriptionHash,
		"fallback_address": decoded.FallbackAddr,
		"cltv_expiry":      decoded.CltvExpiry,
		"route_hints":      routeHints,
		"payment_addr":     hex.EncodeToString(decoded.PaymentAddr),
		"features":         features,
	}), nil
}

// ListInvoicesTool returns the MCP tool definition for listing invoices.
func (s *InvoiceService) ListInvoicesTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_list_invoices",
		Description: "List invoices created by this Lightning node",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"pending_only": map[string]any{
					"type":        "boolean",
					"description": "Only return pending/unpaid invoices",
				},
				"index_offset": map[string]any{
					"type":        "number",
					"description": "Start index for pagination",
					"minimum":     0,
				},
				"num_max_invoices": map[string]any{
					"type":        "number",
					"description": "Maximum number of invoices to return",
					"minimum":     1,
					"maximum":     1000,
				},
				"reversed": map[string]any{
					"type":        "boolean",
					"description": "Return invoices in reverse chronological order",
				},
			},
		},
	}
}

// HandleListInvoices handles the list invoices request.
func (s *InvoiceService) HandleListInvoices(ctx context.Context,
	request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := requestArguments(request)

	if s.LightningClient == nil {
		return newToolResultError(
			"Not connected to Lightning node. " +
				"Use lnc_connect first."), nil
	}

	pendingOnly, _ := args["pending_only"].(bool)
	indexOffset, _ := args["index_offset"].(float64)
	numMaxInvoices, _ := args["num_max_invoices"].(float64)
	if numMaxInvoices == 0 {
		numMaxInvoices = 100
	}
	reversed, _ := args["reversed"].(bool)

	resp, err := s.LightningClient.ListInvoices(
		ctx, &lnrpc.ListInvoiceRequest{
			PendingOnly:    pendingOnly,
			IndexOffset:    uint64(indexOffset),
			NumMaxInvoices: uint64(numMaxInvoices),
			Reversed:       reversed,
		},
	)
	if err != nil {
		return newToolResultError(
			"Failed to list invoices: " + err.Error()), nil
	}

	invoiceList := make([]map[string]any, len(resp.Invoices))
	for i, inv := range resp.Invoices {
		invoiceList[i] = map[string]any{
			"memo":            inv.Memo,
			"payment_request": inv.PaymentRequest,
			"r_hash":          hex.EncodeToString(inv.RHash),
			"value":           inv.Value,
			"value_msat":      inv.ValueMsat,
			"settled":         inv.State == lnrpc.Invoice_SETTLED,
			"creation_date":   inv.CreationDate,
			"settle_date":     inv.SettleDate,
			"expiry":          inv.Expiry,
			"cltv_expiry":     inv.CltvExpiry,
			"private":         inv.Private,
			"add_index":       inv.AddIndex,
			"settle_index":    inv.SettleIndex,
			"amt_paid_sat":    inv.AmtPaidSat,
			"amt_paid_msat":   inv.AmtPaidMsat,
			"state":           inv.State.String(),
			"is_keysend":      inv.IsKeysend,
			"payment_addr":    hex.EncodeToString(inv.PaymentAddr),
		}
	}

	return newToolResultJSON(map[string]any{
		"invoices":           invoiceList,
		"first_index_offset": resp.FirstIndexOffset,
		"last_index_offset":  resp.LastIndexOffset,
		"total_invoices":     len(invoiceList),
	}), nil
}

// LookupInvoiceTool returns the MCP tool definition for looking up a
// specific invoice.
func (s *InvoiceService) LookupInvoiceTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lnc_lookup_invoice",
		Description: "Look up a specific invoice by its payment hash",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"payment_hash": map[string]any{
					"type":        "string",
					"description": "Payment hash of the invoice (hex encoded)",
					"pattern":     "^[0-9a-fA-F]{64}$",
				},
			},
			Required: []string{"payment_hash"},
		},
	}
}

// HandleLookupInvoice handles the lookup invoice request.
func (s *InvoiceService) HandleLookupInvoice(ctx context.Context,
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

	rhashBytes, err := hex.DecodeString(paymentHash)
	if err != nil {
		return newToolResultError(
			"invalid payment_hash format"), nil
	}

	inv, err := s.LightningClient.LookupInvoice(
		ctx, &lnrpc.PaymentHash{RHash: rhashBytes},
	)
	if err != nil {
		return newToolResultError(
			"Failed to lookup invoice: " + err.Error()), nil
	}

	return newToolResultJSON(map[string]any{
		"memo":            inv.Memo,
		"payment_request": inv.PaymentRequest,
		"r_hash":          hex.EncodeToString(inv.RHash),
		"value":           inv.Value,
		"value_msat":      inv.ValueMsat,
		"settled":         inv.State == lnrpc.Invoice_SETTLED,
		"creation_date":   inv.CreationDate,
		"settle_date":     inv.SettleDate,
		"expiry":          inv.Expiry,
		"cltv_expiry":     inv.CltvExpiry,
		"private":         inv.Private,
		"add_index":       inv.AddIndex,
		"settle_index":    inv.SettleIndex,
		"amt_paid_sat":    inv.AmtPaidSat,
		"amt_paid_msat":   inv.AmtPaidMsat,
		"state":           inv.State.String(),
		"is_keysend":      inv.IsKeysend,
	}), nil
}
