// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

// Package interfaces defines the core interfaces for the MCP LNC server. It
// enables loose coupling, dependency injection, and easier testing.
package interfaces

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Logger defines the logging interface used throughout the application. It
// makes swapping implementations or providing mocks straightforward.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
}

// LightningClient defines the read-only interface for Lightning Network
// operations. Only query methods are included; write operations (payments,
// channel opens, etc.) are intentionally excluded.
type LightningClient interface {
	GetInfo(ctx context.Context,
		req *lnrpc.GetInfoRequest) (
		*lnrpc.GetInfoResponse, error)

	WalletBalance(ctx context.Context,
		req *lnrpc.WalletBalanceRequest) (
		*lnrpc.WalletBalanceResponse, error)

	ChannelBalance(ctx context.Context,
		req *lnrpc.ChannelBalanceRequest) (
		*lnrpc.ChannelBalanceResponse, error)

	ListChannels(ctx context.Context,
		req *lnrpc.ListChannelsRequest) (
		*lnrpc.ListChannelsResponse, error)

	PendingChannels(ctx context.Context,
		req *lnrpc.PendingChannelsRequest) (
		*lnrpc.PendingChannelsResponse, error)

	DecodePayReq(ctx context.Context,
		req *lnrpc.PayReqString) (*lnrpc.PayReq, error)

	ListInvoices(ctx context.Context,
		req *lnrpc.ListInvoiceRequest) (
		*lnrpc.ListInvoiceResponse, error)

	LookupInvoice(ctx context.Context,
		req *lnrpc.PaymentHash) (*lnrpc.Invoice, error)

	ListPayments(ctx context.Context,
		req *lnrpc.ListPaymentsRequest) (
		*lnrpc.ListPaymentsResponse, error)

	ListPeers(ctx context.Context,
		req *lnrpc.ListPeersRequest) (
		*lnrpc.ListPeersResponse, error)

	DescribeGraph(ctx context.Context,
		req *lnrpc.ChannelGraphRequest) (
		*lnrpc.ChannelGraph, error)

	GetNodeInfo(ctx context.Context,
		req *lnrpc.NodeInfoRequest) (*lnrpc.NodeInfo, error)

	GetTransactions(ctx context.Context,
		req *lnrpc.GetTransactionsRequest) (
		*lnrpc.TransactionDetails, error)

	ListUnspent(ctx context.Context,
		req *lnrpc.ListUnspentRequest) (
		*lnrpc.ListUnspentResponse, error)

	EstimateFee(ctx context.Context,
		req *lnrpc.EstimateFeeRequest) (
		*lnrpc.EstimateFeeResponse, error)
}

// ConnectionCallback defines the callback function type for LNC connections.
type ConnectionCallback func(conn *grpc.ClientConn)

// Service defines the interface that all MCP tool services must implement.
type Service interface {
	// Name returns the service name for logging and identification.
	Name() string

	// Tools returns the MCP tools provided by this service.
	Tools() []ServiceTool
}

// ToolHandler defines the function signature for MCP tool handlers.
type ToolHandler = mcp.ToolHandler

// ServiceTool represents an MCP tool with its handler.
type ServiceTool struct {
	Tool    *mcp.Tool
	Handler ToolHandler
}

// ServiceManager defines the interface for managing all services.
type ServiceManager interface {
	// RegisterServices registers all services with their tools.
	RegisterServices(mcpServer MCPServer) error

	// UpdateConnection updates all services with a new Lightning
	// connection.
	UpdateConnection(client LightningClient)

	// Shutdown gracefully shuts down all services.
	Shutdown() error
}

// MCPServer defines the interface for the MCP server operations we need.
// This allows us to easily mock the MCP server for testing.
type MCPServer interface {
	AddTool(tool *mcp.Tool, handler mcp.ToolHandler)
}

// Daemon defines the interface for the main daemon operations.
type Daemon interface {
	Start() error
	Stop()
}

// Server defines the interface for the MCP server operations.
type Server interface {
	Start() error
	Stop(ctx context.Context) error
}
