package tezos

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/codec"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

func TestTransactionSendSuccess(t *testing.T) {
	// Set up tezos connector mocks
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"signature\":\"sigWetzF5zVM2qdYt8QToj7e5cNBm9neiPRc3rpePBDrr8N1brFbErv2YfXMSoSgemJ8AwZcLfmkBDg78bmUEzF1sf1YotnS\"}"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	mRPC.On("Broadcast", ctx, mock.Anything).
		Return(tezos.OpHash([]byte("oovD5cUigLGLT6kGDqsLMyF2sc3MLyfYhJWRymCPxUKEx3vtQ5v")), nil)

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	resp, _, err := c.TransactionSend(ctx, req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestTransactionSendDecodeStrError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionSendRequest{
		TransactionData: "1",
	}
	res, reason, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
	assert.Nil(t, res)
}

func TestTransactionSendDecodeOpError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionSendRequest{}
	res, reason, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
	assert.Nil(t, res)
}

func TestTransactionSendWrongFromAddressError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "wrong",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	res, _, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23019", err)
	assert.Nil(t, res)
}

func TestTransactionSendGetContractExtError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(nil, errors.New("error"))

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	_, _, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
}

func TestTransactionSendBroadcastError(t *testing.T) {
	// Set up tezos connector mocks
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"signature\":\"sigWetzF5zVM2qdYt8QToj7e5cNBm9neiPRc3rpePBDrr8N1brFbErv2YfXMSoSgemJ8AwZcLfmkBDg78bmUEzF1sf1YotnS\"}"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	mRPC.On("Broadcast", ctx, mock.Anything).
		Return(nil, errors.New("error"))

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	resp, _, err := c.TransactionSend(ctx, req)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestTransactionSendSignTxError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	req := &ffcapi.TransactionSendRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
		},
		TransactionData: "424d426559724d4a704c577271437337555463466155514365574271736a434c6c00889816a17ae688c971be1ad34bfe1990f8fa5e0f000b0000000130a980e6e41028da2cacfca4ddefea252d18bed900ffff05706175736500000002030a",
	}
	_, _, err := c.TransactionSend(ctx, req)
	assert.Error(t, err)
}

func Test_signTxRemotelyNilOperationError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	err := c.signTxRemotely(ctx, nil)
	assert.Error(t, err)
}

func Test_signTxRemotelyNilContextError(t *testing.T) {
	_, c, _, done := newTestConnector(t)
	defer done()

	op := codec.NewOp()
	op.WithSource(tezos.MustParseAddress("tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN"))

	err := c.signTxRemotely(nil, op)
	assert.Error(t, err)
}

func Test_signTxRemotelyHttpError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	op := codec.NewOp()
	op.WithSource(tezos.MustParseAddress("tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN"))

	err := c.signTxRemotely(ctx, op)
	assert.Error(t, err)
}

func Test_signTxRemotelyHttpWrongStatusError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	op := codec.NewOp()
	op.WithSource(tezos.MustParseAddress("tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN"))

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	err := c.signTxRemotely(ctx, op)
	assert.Error(t, err)
}

func Test_signTxRemotelyUnmarshalRespError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	op := codec.NewOp()
	op.WithSource(tezos.MustParseAddress("tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN"))

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(nil)
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	err := c.signTxRemotely(ctx, op)
	assert.Error(t, err)
}

func Test_signTxRemotelyUnmarshalSignatureError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	op := codec.NewOp()
	op.WithSource(tezos.MustParseAddress("tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN"))

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"wrong\":\"sigWetzF5zVM2qdYt8QToj7e5cNBm9neiPRc3rpePBDrr8N1brFbErv2YfXMSoSgemJ8AwZcLfmkBDg78bmUEzF1sf1YotnS\"}"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	err := c.signTxRemotely(ctx, op)
	assert.Error(t, err)
}
