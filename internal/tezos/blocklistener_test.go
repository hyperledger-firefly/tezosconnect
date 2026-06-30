package tezos

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

func TestBlockListenerStartGettingHighestBlockRetry(t *testing.T) {
	_, c, mRPC, done := newTestConnector(t)
	bl := c.blockListener

	mRPC.On("GetHeadBlock", mock.Anything).Return(nil, errors.New("err")).Once()
	mRPC.On("GetHeadBlock", mock.Anything).Return(
		&rpc.Block{
			Hash: tezos.MustParseBlockHash("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg"),
			Header: rpc.BlockHeader{
				Predecessor: tezos.MustParseBlockHash("BLc1BjmZ7WevMoMoj8jxh4k2wLoRqoMUxjrQuDmKzAsApfRRjFL"),
				Level:       12345,
			},
		}, nil)

	assert.Equal(t, int64(12345), bl.getHighestBlock(bl.ctx))
	done() // Stop immediately in this case, while we're in the polling interval

	<-bl.listenLoopDone

	mRPC.AssertExpectations(t)
}

func TestBlockListenerStartGettingHighestBlockFailBeforeStop(t *testing.T) {
	_, c, mRPC, done := newTestConnector(t)
	done() // Stop before we start
	bl := c.blockListener

	mRPC.On("GetHeadBlock", mock.Anything).Return(nil, errors.New("err")).Maybe()

	assert.Equal(t, int64(-1), bl.getHighestBlock(bl.ctx))

	<-bl.listenLoopDone

	mRPC.AssertExpectations(t)
}
