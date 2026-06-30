package tezos

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-common/pkg/retry"
	"github.com/hyperledger-firefly/firefly-tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/trilitech/tzgo/rpc"
)

type tezosConnector struct {
	catchupPageSize            int64
	catchupThreshold           int64
	checkpointBlockGap         int64
	retry                      *retry.Retry
	eventBlockTimestamps       bool
	blockListener              *blockListener
	eventFilterPollingInterval time.Duration

	client       rpc.RpcClient
	networkName  string
	signatoryURL string

	mux          sync.Mutex
	eventStreams map[fftypes.UUID]*eventStream
	blockCache   *lru.Cache
	txCache      *lru.Cache
}

func NewTezosConnector(ctx context.Context, conf config.Section) (cc ffcapi.API, err error) {
	c := &tezosConnector{
		eventStreams:               make(map[fftypes.UUID]*eventStream),
		catchupPageSize:            conf.GetInt64(EventsCatchupPageSize),
		catchupThreshold:           conf.GetInt64(EventsCatchupThreshold),
		checkpointBlockGap:         conf.GetInt64(EventsCheckpointBlockGap),
		eventBlockTimestamps:       conf.GetBool(EventsBlockTimestamps),
		eventFilterPollingInterval: conf.GetDuration(EventsFilterPollingInterval),
		retry: &retry.Retry{
			InitialDelay: conf.GetDuration(RetryInitDelay),
			MaximumDelay: conf.GetDuration(RetryMaxDelay),
			Factor:       conf.GetFloat64(RetryFactor),
		},
	}
	if c.catchupThreshold < c.catchupPageSize {
		log.L(ctx).Warnf("Catchup threshold %d must be at least as large as the catchup page size %d (overridden to %d)", c.catchupThreshold, c.catchupPageSize, c.catchupPageSize)
		c.catchupThreshold = c.catchupPageSize
	}
	c.blockCache, err = lru.New(conf.GetInt(BlockCacheSize))
	if err != nil {
		return nil, i18n.WrapError(ctx, err, msgs.MsgCacheInitFail, "block")
	}

	c.txCache, err = lru.New(conf.GetInt(TxCacheSize))
	if err != nil {
		return nil, i18n.WrapError(ctx, err, msgs.MsgCacheInitFail, "transaction")
	}

	rpcClientURL := conf.GetString(BlockchainRPC)
	if rpcClientURL == "" {
		return nil, i18n.WrapError(ctx, err, msgs.MsgMissingRPCUrl)
	}
	c.client, err = rpc.NewClient(rpcClientURL, nil)
	if err != nil {
		return nil, i18n.WrapError(ctx, err, msgs.MsgFailedRPCInitialization)
	}
	c.networkName = conf.GetString(BlockchainNetwork)

	// service for tx signing
	c.signatoryURL = conf.GetString(BlockchainSignatory)

	c.blockListener = newBlockListener(ctx, c, conf)

	return c, nil
}
