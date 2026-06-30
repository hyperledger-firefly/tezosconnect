package tezos

import (
	"errors"
	"testing"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

func TestGetBlockInfoByNumberOK(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlock", mock.Anything, mock.MatchedBy(
		func(blockNumber *fftypes.FFBigInt) bool {
			return blockNumber.String() == "12345"
		})).
		Return(&rpc.Block{
			Hash: tezos.MustParseBlockHash("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg"),
			Header: rpc.BlockHeader{
				Predecessor: tezos.MustParseBlockHash("BLc1BjmZ7WevMoMoj8jxh4k2wLoRqoMUxjrQuDmKzAsApfRRjFL"),
				Level:       12345,
			},
			Operations: [][]*rpc.Operation{
				{}, // consensus operations
				{}, // voting operations
				{}, // anonymous operations
				{
					{
						Hash: tezos.MustParseOpHash("op13B8GtoK1UAx8p67L8y4huPqjay9yzRrpSFLTiY9kjJsrF5uV"),
					},
				}, // manager operations
			},
		}, nil).
		Twice() // two cache misses and a hit

	req := &ffcapi.BlockInfoByNumberRequest{
		BlockNumber: fftypes.NewFFBigInt(12345),
	}
	res, reason, err := c.BlockInfoByNumber(ctx, req)
	assert.NoError(t, err)
	assert.Empty(t, reason)

	assert.Equal(t, "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg", res.BlockHash)
	assert.Equal(t, "BLc1BjmZ7WevMoMoj8jxh4k2wLoRqoMUxjrQuDmKzAsApfRRjFL", res.ParentHash)
	assert.Equal(t, int64(12345), res.BlockNumber.Int64())

	res, reason, err = c.BlockInfoByNumber(ctx, req) // cached
	assert.NoError(t, err)
	assert.Equal(t, "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg", res.BlockHash)
	assert.Equal(t, ffcapi.ErrorReason(""), reason)

	req.ExpectedParentHash = "BMWDjzorc6GFb2DnengeB2TRikAENukebRwubnu6ghfZceicmig"
	res, reason, err = c.BlockInfoByNumber(ctx, req) // cache miss
	assert.NoError(t, err)
	assert.Equal(t, "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg", res.BlockHash)
	assert.Equal(t, ffcapi.ErrorReason(""), reason)
}

func TestGetBlockInfoByNumberBlockNotFoundError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlock", mock.Anything, mock.Anything).
		Return(nil, errors.New("status 404"))

	req := &ffcapi.BlockInfoByNumberRequest{
		BlockNumber: fftypes.NewFFBigInt(1),
	}
	res, reason, err := c.BlockInfoByNumber(ctx, req)
	assert.Regexp(t, "FF23011", err)
	assert.Equal(t, ffcapi.ErrorReasonNotFound, reason)
	assert.Nil(t, res)
}

func TestGetBlockInfoByNumberNotFound(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlock", mock.Anything, mock.Anything).
		Return(nil, nil)

	req := &ffcapi.BlockInfoByNumberRequest{
		BlockNumber: fftypes.NewFFBigInt(12345),
	}
	res, reason, err := c.BlockInfoByNumber(ctx, req)
	assert.Regexp(t, "FF23011", err)
	assert.Equal(t, ffcapi.ErrorReasonNotFound, reason)
	assert.Nil(t, res)

}

func TestGetBlockInfoByNumberFail(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlock", mock.Anything, mock.Anything).
		Return(nil, errors.New("err"))

	req := &ffcapi.BlockInfoByNumberRequest{
		BlockNumber: fftypes.NewFFBigInt(1),
	}
	res, reason, err := c.BlockInfoByNumber(ctx, req)
	assert.Error(t, err)
	assert.Empty(t, reason)
	assert.Nil(t, res)
}

func TestGetBlockInfoByHashOK(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlock", mock.Anything, mock.Anything).
		Return(&rpc.Block{
			Hash: tezos.MustParseBlockHash("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg"),
			Header: rpc.BlockHeader{
				Predecessor: tezos.MustParseBlockHash("BLc1BjmZ7WevMoMoj8jxh4k2wLoRqoMUxjrQuDmKzAsApfRRjFL"),
				Level:       12345,
			},
			Operations: [][]*rpc.Operation{
				{}, // consensus operations
				{}, // voting operations
				{}, // anonymous operations
				{
					{
						Hash: tezos.MustParseOpHash("op13B8GtoK1UAx8p67L8y4huPqjay9yzRrpSFLTiY9kjJsrF5uV"),
					},
				}, // manager operations
			},
		}, nil).
		Once()

	req := &ffcapi.BlockInfoByHashRequest{
		BlockHash: "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg",
	}
	res, reason, err := c.BlockInfoByHash(ctx, req)
	assert.NoError(t, err)
	assert.Empty(t, reason)

	assert.Equal(t, "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg", res.BlockHash)
	assert.Equal(t, "BLc1BjmZ7WevMoMoj8jxh4k2wLoRqoMUxjrQuDmKzAsApfRRjFL", res.ParentHash)
	assert.Equal(t, int64(12345), res.BlockNumber.Int64())

	res, reason, err = c.BlockInfoByHash(ctx, req) // cached
	assert.NoError(t, err)
	assert.Empty(t, reason)
	assert.Equal(t, "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg", res.BlockHash)
}

func TestGetBlockInfoByHashParseBlockHashFail(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.BlockInfoByHashRequest{
		BlockHash: "wrong",
	}
	res, reason, err := c.BlockInfoByHash(ctx, req)
	assert.Error(t, err)
	assert.Empty(t, reason)
	assert.Nil(t, res)
}

func TestGetBlockInfoByHashNotFound(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlock", mock.Anything, mock.Anything).
		Return(nil, nil)

	req := &ffcapi.BlockInfoByHashRequest{
		BlockHash: "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg",
	}
	res, reason, err := c.BlockInfoByHash(ctx, req)
	assert.Regexp(t, "FF23011", err)
	assert.Equal(t, ffcapi.ErrorReasonNotFound, reason)
	assert.Nil(t, res)
}

func TestGetBlockInfoByHashFail(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	mRPC.On("GetBlock", mock.Anything, mock.Anything).
		Return(nil, errors.New("err"))

	req := &ffcapi.BlockInfoByHashRequest{
		BlockHash: "BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg",
	}
	res, reason, err := c.BlockInfoByHash(ctx, req)
	assert.Error(t, err)
	assert.Empty(t, reason)
	assert.Nil(t, res)
}
