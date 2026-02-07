// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

// Package client provides Lightning Network client wrappers that implement
// our defined interfaces.
package client

import (
	"context"

	"github.com/lightninglabs/lightning-agent-kit/mcp-server/internal/interfaces"
	"github.com/lightningnetwork/lnd/lnrpc"
)

// lightningClientWrapper wraps the LND Lightning client to implement
// our read-only LightningClient interface.
type lightningClientWrapper struct {
	client lnrpc.LightningClient
}

// NewLightningClient creates a new Lightning client wrapper.
func NewLightningClient(
	client lnrpc.LightningClient,
) interfaces.LightningClient {
	return &lightningClientWrapper{client: client}
}

func (w *lightningClientWrapper) GetInfo(ctx context.Context,
	req *lnrpc.GetInfoRequest) (*lnrpc.GetInfoResponse, error) {
	return w.client.GetInfo(ctx, req)
}

func (w *lightningClientWrapper) WalletBalance(ctx context.Context,
	req *lnrpc.WalletBalanceRequest) (
	*lnrpc.WalletBalanceResponse, error) {
	return w.client.WalletBalance(ctx, req)
}

func (w *lightningClientWrapper) ChannelBalance(ctx context.Context,
	req *lnrpc.ChannelBalanceRequest) (
	*lnrpc.ChannelBalanceResponse, error) {
	return w.client.ChannelBalance(ctx, req)
}

func (w *lightningClientWrapper) ListChannels(ctx context.Context,
	req *lnrpc.ListChannelsRequest) (
	*lnrpc.ListChannelsResponse, error) {
	return w.client.ListChannels(ctx, req)
}

func (w *lightningClientWrapper) PendingChannels(ctx context.Context,
	req *lnrpc.PendingChannelsRequest) (
	*lnrpc.PendingChannelsResponse, error) {
	return w.client.PendingChannels(ctx, req)
}

func (w *lightningClientWrapper) DecodePayReq(ctx context.Context,
	req *lnrpc.PayReqString) (*lnrpc.PayReq, error) {
	return w.client.DecodePayReq(ctx, req)
}

func (w *lightningClientWrapper) ListInvoices(ctx context.Context,
	req *lnrpc.ListInvoiceRequest) (
	*lnrpc.ListInvoiceResponse, error) {
	return w.client.ListInvoices(ctx, req)
}

func (w *lightningClientWrapper) LookupInvoice(ctx context.Context,
	req *lnrpc.PaymentHash) (*lnrpc.Invoice, error) {
	return w.client.LookupInvoice(ctx, req)
}

func (w *lightningClientWrapper) ListPayments(ctx context.Context,
	req *lnrpc.ListPaymentsRequest) (
	*lnrpc.ListPaymentsResponse, error) {
	return w.client.ListPayments(ctx, req)
}

func (w *lightningClientWrapper) ListPeers(ctx context.Context,
	req *lnrpc.ListPeersRequest) (
	*lnrpc.ListPeersResponse, error) {
	return w.client.ListPeers(ctx, req)
}

func (w *lightningClientWrapper) DescribeGraph(ctx context.Context,
	req *lnrpc.ChannelGraphRequest) (*lnrpc.ChannelGraph, error) {
	return w.client.DescribeGraph(ctx, req)
}

func (w *lightningClientWrapper) GetNodeInfo(ctx context.Context,
	req *lnrpc.NodeInfoRequest) (*lnrpc.NodeInfo, error) {
	return w.client.GetNodeInfo(ctx, req)
}

func (w *lightningClientWrapper) GetTransactions(ctx context.Context,
	req *lnrpc.GetTransactionsRequest) (
	*lnrpc.TransactionDetails, error) {
	return w.client.GetTransactions(ctx, req)
}

func (w *lightningClientWrapper) ListUnspent(ctx context.Context,
	req *lnrpc.ListUnspentRequest) (
	*lnrpc.ListUnspentResponse, error) {
	return w.client.ListUnspent(ctx, req)
}

func (w *lightningClientWrapper) EstimateFee(ctx context.Context,
	req *lnrpc.EstimateFeeRequest) (
	*lnrpc.EstimateFeeResponse, error) {
	return w.client.EstimateFee(ctx, req)
}
