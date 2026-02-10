// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

// Package testutils provides testing utilities and mock implementations for
// the MCP LNC server.
package testutils

import (
	"context"
	"testing"

	"github.com/lightninglabs/lightning-agent-kit/lightning-mcp-server/internal/interfaces"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// MockLightningClient is a mock implementation of the LightningClient
// interface for testing.
type MockLightningClient struct {
	mock.Mock
}

// Compile-time check that MockLightningClient implements
// interfaces.LightningClient.
var _ interfaces.LightningClient = (*MockLightningClient)(nil)

// GetInfo mocks the GetInfo method.
func (m *MockLightningClient) GetInfo(ctx context.Context,
	req *lnrpc.GetInfoRequest) (*lnrpc.GetInfoResponse, error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.GetInfoResponse), args.Error(1)
}

// WalletBalance mocks the WalletBalance method.
func (m *MockLightningClient) WalletBalance(ctx context.Context,
	req *lnrpc.WalletBalanceRequest) (*lnrpc.WalletBalanceResponse,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.WalletBalanceResponse), args.Error(1)
}

// ChannelBalance mocks the ChannelBalance method.
func (m *MockLightningClient) ChannelBalance(ctx context.Context,
	req *lnrpc.ChannelBalanceRequest) (*lnrpc.ChannelBalanceResponse,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.ChannelBalanceResponse), args.Error(1)
}

// ListChannels mocks the ListChannels method.
func (m *MockLightningClient) ListChannels(ctx context.Context,
	req *lnrpc.ListChannelsRequest) (*lnrpc.ListChannelsResponse,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.ListChannelsResponse), args.Error(1)
}

// PendingChannels mocks the PendingChannels method.
func (m *MockLightningClient) PendingChannels(ctx context.Context,
	req *lnrpc.PendingChannelsRequest) (
	*lnrpc.PendingChannelsResponse, error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.PendingChannelsResponse), args.Error(1)
}

// DecodePayReq mocks the DecodePayReq method.
func (m *MockLightningClient) DecodePayReq(ctx context.Context,
	req *lnrpc.PayReqString) (*lnrpc.PayReq, error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.PayReq), args.Error(1)
}

// ListInvoices mocks the ListInvoices method.
func (m *MockLightningClient) ListInvoices(ctx context.Context,
	req *lnrpc.ListInvoiceRequest) (*lnrpc.ListInvoiceResponse,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.ListInvoiceResponse), args.Error(1)
}

// LookupInvoice mocks the LookupInvoice method.
func (m *MockLightningClient) LookupInvoice(ctx context.Context,
	req *lnrpc.PaymentHash) (*lnrpc.Invoice, error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.Invoice), args.Error(1)
}

// ListPayments mocks the ListPayments method.
func (m *MockLightningClient) ListPayments(ctx context.Context,
	req *lnrpc.ListPaymentsRequest) (*lnrpc.ListPaymentsResponse,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.ListPaymentsResponse), args.Error(1)
}

// ListPeers mocks the ListPeers method.
func (m *MockLightningClient) ListPeers(ctx context.Context,
	req *lnrpc.ListPeersRequest) (*lnrpc.ListPeersResponse, error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.ListPeersResponse), args.Error(1)
}

// DescribeGraph mocks the DescribeGraph method.
func (m *MockLightningClient) DescribeGraph(ctx context.Context,
	req *lnrpc.ChannelGraphRequest) (*lnrpc.ChannelGraph, error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.ChannelGraph), args.Error(1)
}

// GetNodeInfo mocks the GetNodeInfo method.
func (m *MockLightningClient) GetNodeInfo(ctx context.Context,
	req *lnrpc.NodeInfoRequest) (*lnrpc.NodeInfo, error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.NodeInfo), args.Error(1)
}

// GetTransactions mocks the GetTransactions method.
func (m *MockLightningClient) GetTransactions(ctx context.Context,
	req *lnrpc.GetTransactionsRequest) (*lnrpc.TransactionDetails,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.TransactionDetails), args.Error(1)
}

// ListUnspent mocks the ListUnspent method.
func (m *MockLightningClient) ListUnspent(ctx context.Context,
	req *lnrpc.ListUnspentRequest) (*lnrpc.ListUnspentResponse,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.ListUnspentResponse), args.Error(1)
}

// EstimateFee mocks the EstimateFee method.
func (m *MockLightningClient) EstimateFee(ctx context.Context,
	req *lnrpc.EstimateFeeRequest) (*lnrpc.EstimateFeeResponse,
	error) {
	args := m.Mock.Called(ctx, req)
	return args.Get(0).(*lnrpc.EstimateFeeResponse), args.Error(1)
}

// MockLogger is a mock implementation of the Logger interface for testing.
type MockLogger struct {
	mock.Mock
}

// Debug mocks the Debug method.
func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	args := []any{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Mock.Called(args...)
}

// Info mocks the Info method.
func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	args := []any{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Mock.Called(args...)
}

// Warn mocks the Warn method.
func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	args := []any{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Mock.Called(args...)
}

// Error mocks the Error method.
func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	args := []any{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Mock.Called(args...)
}

// Fatal mocks the Fatal method.
func (m *MockLogger) Fatal(msg string, fields ...zap.Field) {
	args := []any{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Mock.Called(args...)
}

// With mocks the With method.
func (m *MockLogger) With(fields ...zap.Field) interfaces.Logger {
	args := []any{}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Mock.Called(args...)
	return args[0].(interfaces.Logger)
}

// TestLogger creates a test logger for use in tests.
func TestLogger(t *testing.T) *zap.Logger {
	return zaptest.NewLogger(t)
}

// CreateMockPayReq creates a mock PayReq for testing invoice decoding.
func CreateMockPayReq(amount int64, memo string) *lnrpc.PayReq {
	return &lnrpc.PayReq{
		Destination:     "mock_destination_pubkey_66_chars_long_hex_encoded_exactly",
		PaymentHash:     "mock_payment_hash_64_chars_long_hex_encoded_exactly_here",
		NumSatoshis:     amount,
		Timestamp:       1692633600, // Fixed timestamp for testing
		Expiry:          3600,       // 1 hour
		Description:     memo,
		DescriptionHash: "",
		FallbackAddr:    "",
		CltvExpiry:      40,
		RouteHints:      []*lnrpc.RouteHint{},
		PaymentAddr:     []byte("mock_payment_addr_32_bytes_long_ok"),
		NumMsat:         amount * 1000,
	}
}

// CreateMockGetInfoResponse creates a mock GetInfoResponse for testing.
func CreateMockGetInfoResponse() *lnrpc.GetInfoResponse {
	return &lnrpc.GetInfoResponse{
		Version:             "0.17.0-beta commit=v0.17.0-beta",
		CommitHash:          "mock_commit_hash",
		IdentityPubkey:      "mock_identity_pubkey_66_chars_long_hex_encoded_exactly",
		Alias:               "MockTestNode",
		Color:               "#3399ff",
		NumPendingChannels:  0,
		NumActiveChannels:   2,
		NumInactiveChannels: 0,
		NumPeers:            2,
		BlockHeight:         800000,
		BlockHash:           "mock_block_hash_64_chars_long_hex_encoded_exactly_here",
		BestHeaderTimestamp: 1692633600,
		SyncedToChain:       true,
		SyncedToGraph:       true,
		Testnet:             true,
		Chains: []*lnrpc.Chain{
			{
				Chain:   "bitcoin",
				Network: "testnet",
			},
		},
		Uris: []string{
			"mock_identity_pubkey@localhost:9735",
		},
		Features: map[uint32]*lnrpc.Feature{
			0: {Name: "data-loss-protect", IsRequired: true, IsKnown: true},
			5: {Name: "upfront-shutdown-script", IsRequired: false, IsKnown: true},
		},
	}
}

// AssertNoError is a test helper that fails the test if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError is a test helper that fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// MockMCPServer is a mock implementation of the MCP server for testing.
type MockMCPServer struct {
	mock.Mock
	tools map[string]any // Store registered tools for verification
}

// NewMockMCPServer creates a new mock MCP server.
func NewMockMCPServer() *MockMCPServer {
	return &MockMCPServer{
		tools: make(map[string]any),
	}
}

// AddTool mocks the AddTool method and stores the tool for verification.
func (m *MockMCPServer) AddTool(tool any, handler any) {
	m.Mock.Called(tool, handler)
	// Store tool for verification in tests
	if t, ok := tool.(interface{ GetName() string }); ok {
		m.tools[t.GetName()] = tool
	}
}

// GetRegisteredTools returns all registered tools for test verification.
func (m *MockMCPServer) GetRegisteredTools() map[string]any {
	return m.tools
}
