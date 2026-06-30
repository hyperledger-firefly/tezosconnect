package tezos

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/codec"
	"github.com/trilitech/tzgo/contract"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

func TestTransactionPrepareSuccess(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
				},
			},
		}, nil)

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	res, reason, err := c.TransactionPrepare(ctx, req)
	assert.NoError(t, err)
	assert.Empty(t, reason)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.TransactionData)
}

func TestTransactionPrepareWithRevealSuccess(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"public_key\":\"edpkvHVuLHkr5eDiTtQKyUPqgYVAk3Sy4m7qBD8r6abemHkZsMU5Kh\"}"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Reveal{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
				},
			},
		}, nil)

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	res, reason, err := c.TransactionPrepare(ctx, req)
	assert.NoError(t, err)
	assert.Empty(t, reason)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.TransactionData)
}

func TestTransactionPrepareWrongParamsError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("wrong"),
			},
		},
	}
	_, reason, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23014", err)
	assert.Equal(t, ffcapi.ErrorReasonInvalidInputs, reason)
}

func TestTransactionPrepareWrongToAddressError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "wrong",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	_, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23020", err)
}

func TestTransactionPrepareWrongFromAddressError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "wrong",
				To:   "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	_, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Regexp(t, "FF23019", err)
}

func TestTransactionPrepareGetContractExtError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(nil, errors.New("error"))

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	_, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
}

func TestTransactionPrepareSimulateError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).Return(nil, errors.New("error"))

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}

	_, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
}

func TestTransactionPrepareWrongSimulateStatusError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Errors: []rpc.OperationError{
											{
												GenericError: rpc.GenericError{
													ID:   "error id: script_rejected",
													Kind: "error: script_rejected",
												},
												Raw: json.RawMessage{},
											},
										},
										Status: tezos.OpStatusFailed,
									},
								},
							},
						},
					},
				},
			},
		}, nil)

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}

	_, reason, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonTransactionReverted)
}

func Test_estimateAndAssignTxCostIgnoreLimitsOk(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
				},
			},
		}, nil)

	op := codec.NewOp()
	txArgs := contract.TxArgs{}
	op.WithContents(txArgs.Encode())

	opts := &rpc.DefaultOptions
	opts.IgnoreLimits = true

	_, err := c.estimateAndAssignTxCost(ctx, op, opts)
	assert.NoError(t, err)
}

func Test_estimateAndAssignExceedMaxLimitError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("Simulate", ctx, mock.Anything, mock.Anything).
		Return(&rpc.Receipt{
			Op: &rpc.Operation{
				Contents: []rpc.TypedOperation{
					rpc.Transaction{
						Manager: rpc.Manager{
							Generic: rpc.Generic{
								Metadata: rpc.OperationMetadata{
									Result: rpc.OperationResult{
										Status: tezos.OpStatusApplied,
									},
								},
							},
						},
					},
				},
			},
		}, nil)

	op := codec.NewOp()
	txArgs := contract.TxArgs{}
	op.WithContents(txArgs.Encode())
	op.WithLimits([]tezos.Limits{
		{
			Fee: 100,
		},
	}, 100)

	opts := &rpc.DefaultOptions
	opts.MaxFee = 1

	_, err := c.estimateAndAssignTxCost(ctx, op, opts)
	assert.Error(t, err)
}

func TestTransactionPrepareWithRevealEmptyServerError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlockHash", ctx, mock.Anything).
		Return(tezos.NewBlockHash([]byte("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg")), nil)

	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, nil)

	req := &ffcapi.TransactionPrepareRequest{
		TransactionInput: ffcapi.TransactionInput{
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
				To:   "KT1D254HTPKq5GZNVcF73XBinG9BLybHqu8s",
			},
			Method: fftypes.JSONAnyPtr("\"pause\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("{\"entrypoint\":\"pause\",\"value\":{\"prim\":\"True\"}}"),
			},
		},
	}
	resp, _, err := c.TransactionPrepare(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_getNetworkParamsByName(t *testing.T) {
	params := getNetworkParamsByName("ghostnet")
	assert.Equal(t, params, tezos.GhostnetParams)

	params = getNetworkParamsByName("parisnet")
	assert.Equal(t, params, tezos.ParisnetParams)

	params = getNetworkParamsByName("default")
	assert.Equal(t, params, tezos.DefaultParams)
}

func Test_getPubKeyFromSignatoryInvalidRespKeySuccess(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"public_key\":\"edpkvHVuLHkr5eDiTtQKyUPqgYVAk3Sy4m7qBD8r6abemHkZsMU5Kh\"}"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	resp, err := c.getPubKeyFromSignatory(ctx, "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN")
	assert.NotNil(t, resp)
	assert.NoError(t, err)
}

func Test_getPubKeyFromSignatoryyNilContextError(t *testing.T) {
	_, c, _, done := newTestConnector(t)
	defer done()

	_, err := c.getPubKeyFromSignatory(nil, "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN")
	assert.Error(t, err)
}

func Test_getPubKeyFromSignatoryHttpError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	_, err := c.getPubKeyFromSignatory(ctx, "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN")
	assert.Error(t, err)
}

func Test_getPubKeyFromSignatoryHttpWrongStatusError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	_, err := c.getPubKeyFromSignatory(ctx, "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN")
	assert.Error(t, err)
}

func Test_getPubKeyFromSignatoryUnmarshalRespError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(nil)
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	_, err := c.getPubKeyFromSignatory(ctx, "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN")
	assert.Error(t, err)
}

func Test_getPubKeyFromSignatoryInvalidRespKeyError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	// Set up http mocks
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"public_key\":\"invalid\"}"))
	}))
	defer svr.Close()
	c.signatoryURL = svr.URL

	_, err := c.getPubKeyFromSignatory(ctx, "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN")
	assert.Error(t, err)
}
