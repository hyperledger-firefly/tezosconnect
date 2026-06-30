package tezos

import (
	"context"
	"testing"

	"github.com/hyperledger-firefly/tezosconnect/mocks/tzrpcbackendmocks"
	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newTestConnector(t *testing.T) (context.Context, *tezosConnector, *tzrpcbackendmocks.RpcClient, func()) {
	mRPC := &tzrpcbackendmocks.RpcClient{}
	config.RootConfigReset()
	conf := config.RootSection("unittest")
	InitConfig(conf)
	conf.AddKnownKey(BlockchainRPC, "https://ghostnet.example.com")
	logrus.SetLevel(logrus.DebugLevel)
	ctx, done := context.WithCancel(context.Background())
	cc, err := NewTezosConnector(ctx, conf)
	assert.NoError(t, err)
	c := cc.(*tezosConnector)
	c.client = mRPC

	return ctx, c, mRPC, func() {
		done()
		mRPC.AssertExpectations(t)
	}
}

func TestConnectorInit(t *testing.T) {
	config.RootConfigReset()
	conf := config.RootSection("unittest")
	InitConfig(conf)

	conf.Set(BlockchainRPC, "")
	cc, err := NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23051", err)
	assert.Nil(t, cc)

	conf.Set(BlockchainRPC, "https://rpc.ghostnet.teztnets.com")
	conf.Set(EventsCatchupThreshold, 1)
	conf.Set(EventsCatchupPageSize, 500)

	cc, err = NewTezosConnector(context.Background(), conf)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), cc.(*tezosConnector).catchupThreshold) // set to page size

	conf.Set(BlockchainRPC, "wrong rpc")
	cc, err = NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23052", err)
	assert.Nil(t, cc)

	conf.Set(ConfigDataFormat, "map")
	conf.Set(BlockCacheSize, "-1")
	cc, err = NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23040", err)
	assert.Nil(t, cc)

	conf.Set(BlockCacheSize, "1")
	conf.Set(TxCacheSize, "-1")
	cc, err = NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23040", err)
	assert.Nil(t, cc)
}
