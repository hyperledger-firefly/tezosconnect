package tezos

import (
	"errors"
	"testing"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

func TestGetAddressBalanceOK(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetHeadBlock", mock.Anything).Return(&rpc.Block{
		Hash: tezos.MustParseBlockHash("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg"),
	}, nil)
	mRPC.On("GetContractBalance", mock.Anything, mock.Anything, mock.Anything).
		Return(tezos.NewZ(999), nil)

	req := ffcapi.AddressBalanceRequest{
		Address: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
	}
	res, _, err := c.AddressBalance(ctx, &req)
	assert.NoError(t, err)
	assert.Equal(t, int64(999), res.Balance.Int64())
}

func TestGetAddressWrongAddress(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.AddressBalanceRequest{
		Address: "wrong",
	}
	_, reason, err := c.AddressBalance(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
}

func TestGetAddressBalanceGetHeadBlockErr(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetHeadBlock", mock.Anything).Return(nil, errors.New("err"))

	req := ffcapi.AddressBalanceRequest{
		Address: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
	}
	_, _, err := c.AddressBalance(ctx, &req)
	assert.Error(t, err)
}

func TestGetAddressBalanceGetContractBalanceErr(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetHeadBlock", mock.Anything).Return(&rpc.Block{
		Hash: tezos.MustParseBlockHash("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg"),
	}, nil)
	mRPC.On("GetContractBalance", mock.Anything, mock.Anything, mock.Anything).
		Return(tezos.Zero, errors.New("err"))

	req := ffcapi.AddressBalanceRequest{
		Address: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
	}
	_, _, err := c.AddressBalance(ctx, &req)
	assert.Error(t, err)
}
