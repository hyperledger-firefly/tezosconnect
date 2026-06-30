package tezos

import (
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
	"testing"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
)

func TestDeployContractPrepare(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()
	mRPC.On("GetBlockHash", ctx, mock.Anything, mock.Anything).
		Return(tezos.BlockHash{}, nil)
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

	resp, reason, err := c.DeployContractPrepare(ctx, &ffcapi.ContractDeployPrepareRequest{
		Contract: fftypes.JSONAnyPtr("{\"code\":[{\"args\":[{\"prim\":\"string\"}],\"prim\":\"parameter\"},{\"args\":[{\"prim\":\"string\"}],\"prim\":\"storage\"},{\"args\":[[{\"prim\":\"CAR\"},{\"args\":[{\"prim\":\"operation\"}],\"prim\":\"NIL\"},{\"prim\":\"PAIR\"}]],\"prim\":\"code\"}],\"storage\":{\"string\":\"hello\"}}"),
	})

	assert.NotNil(t, resp)
	assert.Equal(t, reason, ffcapi.ErrorReason(""))
	assert.NoError(t, err)
}

func TestDeployContractPrepareGetContractExtError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()
	mRPC.On("GetBlockHash", ctx, mock.Anything, mock.Anything).
		Return(tezos.BlockHash{}, nil)
	mRPC.On("GetContractExt", ctx, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
			Manager: "edpkv89Jj4aVWetK69CWm5ss1LayvK8dQoiFz7p995y1k3E8CZwqJ6",
		}, assert.AnError)

	resp, reason, err := c.DeployContractPrepare(ctx, &ffcapi.ContractDeployPrepareRequest{
		Contract: fftypes.JSONAnyPtr("{\"code\":[{\"args\":[{\"prim\":\"string\"}],\"prim\":\"parameter\"},{\"args\":[{\"prim\":\"string\"}],\"prim\":\"storage\"},{\"args\":[[{\"prim\":\"CAR\"},{\"args\":[{\"prim\":\"operation\"}],\"prim\":\"NIL\"},{\"prim\":\"PAIR\"}]],\"prim\":\"code\"}],\"storage\":{\"string\":\"hello\"}}"),
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
}

func TestDeployContractPrepareSimulateError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()
	mRPC.On("GetBlockHash", ctx, mock.Anything, mock.Anything).
		Return(tezos.BlockHash{}, nil)
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
		}, assert.AnError)

	resp, reason, err := c.DeployContractPrepare(ctx, &ffcapi.ContractDeployPrepareRequest{
		Contract: fftypes.JSONAnyPtr("{\"code\":[{\"args\":[{\"prim\":\"string\"}],\"prim\":\"parameter\"},{\"args\":[{\"prim\":\"string\"}],\"prim\":\"storage\"},{\"args\":[[{\"prim\":\"CAR\"},{\"args\":[{\"prim\":\"operation\"}],\"prim\":\"NIL\"},{\"prim\":\"PAIR\"}]],\"prim\":\"code\"}],\"storage\":{\"string\":\"hello\"}}"),
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReason(""))
}

func TestDeployContractPrepareMisingContractError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	resp, reason, err := c.DeployContractPrepare(ctx, &ffcapi.ContractDeployPrepareRequest{})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
}

func TestDeployContractPrepareParseAddressError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	resp, reason, err := c.DeployContractPrepare(ctx, &ffcapi.ContractDeployPrepareRequest{
		TransactionHeaders: ffcapi.TransactionHeaders{
			From: "wrong",
		},
		Contract: fftypes.JSONAnyPtr("{\"code\":[{\"args\":[{\"prim\":\"string\"}],\"prim\":\"parameter\"},{\"args\":[{\"prim\":\"string\"}],\"prim\":\"storage\"},{\"args\":[[{\"prim\":\"CAR\"},{\"args\":[{\"prim\":\"operation\"}],\"prim\":\"NIL\"},{\"prim\":\"PAIR\"}]],\"prim\":\"code\"}],\"storage\":{\"string\":\"hello\"}}"),
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
}
