package tezos

import (
	"testing"
	"time"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

func TestNewBlockListenerOKWithDelay(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	c.blockListener.blockPollingInterval = 1
	mRPC.On("GetHeadBlock", mock.Anything).Return(
		&rpc.Block{
			Hash: tezos.MustParseBlockHash("BMBeYrMJpLWrqCs7UTcFaUQCeWBqsjCLejX5D8zE8m9syHqHnZg"),
			Header: rpc.BlockHeader{
				Predecessor: tezos.MustParseBlockHash("BLc1BjmZ7WevMoMoj8jxh4k2wLoRqoMUxjrQuDmKzAsApfRRjFL"),
				Level:       12345,
			},
		}, nil).Maybe()
	mRPC.On("MonitorBlockHeader", mock.Anything, mock.Anything).Return(nil)

	req := &ffcapi.NewBlockListenerRequest{
		ID:              fftypes.NewUUID(),
		ListenerContext: ctx,
		BlockListener:   make(chan<- *ffcapi.BlockHashEvent),
	}

	res, _, err := c.NewBlockListener(ctx, req)

	time.Sleep(1 * time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
