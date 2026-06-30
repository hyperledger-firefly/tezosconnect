package tezos

import (
	"errors"
	"testing"

	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/rpc"
)

func TestGetNextNonceOK(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetContractExt", mock.Anything, mock.Anything, mock.Anything).
		Return(&rpc.ContractInfo{
			Counter: 10,
		}, nil)

	req := &ffcapi.NextNonceForSignerRequest{
		Signer: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
	}
	res, reason, err := c.NextNonceForSigner(ctx, req)
	assert.NoError(t, err)
	assert.Empty(t, reason)
	assert.Equal(t, int64(11), res.Nonce.Int64())
}

func TestGetNextNonceFail(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetContractExt", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("err"))

	req := &ffcapi.NextNonceForSignerRequest{
		Signer: "tz1Y6GnVhC4EpcDDSmD3ibcC4WX6DJ4Q1QLN",
	}
	res, reason, err := c.NextNonceForSigner(ctx, req)
	assert.Error(t, err)
	assert.Empty(t, reason)
	assert.Nil(t, res)
}
