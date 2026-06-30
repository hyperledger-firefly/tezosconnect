package tezos

import (
	"context"
	"strconv"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger-firefly/tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

// BlockInfoByNumber gets block information from the specified position (block number/index) in the canonical chain currently known to the local node
func (c *tezosConnector) BlockInfoByNumber(ctx context.Context, req *ffcapi.BlockInfoByNumberRequest) (*ffcapi.BlockInfoByNumberResponse, ffcapi.ErrorReason, error) {
	blockInfo, reason, err := c.getBlockInfoByNumber(ctx, req.BlockNumber.Int64(), true, req.ExpectedParentHash)
	if err != nil {
		return nil, reason, err
	}
	if blockInfo == nil {
		return nil, ffcapi.ErrorReasonNotFound, i18n.NewError(ctx, msgs.MsgBlockNotAvailable)
	}

	res := &ffcapi.BlockInfoByNumberResponse{}
	transformBlockInfo(blockInfo, &res.BlockInfo)
	return res, "", nil
}

func (c *tezosConnector) getBlockInfoByNumber(ctx context.Context, blockNumber int64, allowCache bool, expectedHashStr string) (*rpc.Block, ffcapi.ErrorReason, error) {
	var blockInfo *rpc.Block
	var err error

	if allowCache {
		cached, ok := c.blockCache.Get(strconv.FormatInt(blockNumber, 10))
		if ok {
			blockInfo = cached.(*rpc.Block)
			if expectedHashStr != "" && blockInfo.Header.Predecessor.String() != expectedHashStr {
				log.L(ctx).Debugf("Block cache miss for block %d due to mismatched parent hash expected=%s found=%s", blockNumber, expectedHashStr, blockInfo.Header.Predecessor)
				blockInfo = nil
			}
		}
	}

	if blockInfo == nil {
		blockInfo, err = c.client.GetBlock(ctx, fftypes.NewFFBigInt(blockNumber))
		if err != nil {
			if mapError(blockRPCMethods, err) == ffcapi.ErrorReasonNotFound {
				log.L(ctx).Debugf("Received error signifying 'block not found': '%s'", err.Error())
				return nil, ffcapi.ErrorReasonNotFound, i18n.NewError(ctx, msgs.MsgBlockNotAvailable)
			}
			return nil, ffcapi.ErrorReason(""), err
		}
		if blockInfo == nil {
			return nil, ffcapi.ErrorReason(""), err
		}
		c.addToBlockCache(blockInfo)
	}

	return blockInfo, "", nil
}

// BlockInfoByHash gets block information using the hash of the block
func (c *tezosConnector) BlockInfoByHash(ctx context.Context, req *ffcapi.BlockInfoByHashRequest) (*ffcapi.BlockInfoByHashResponse, ffcapi.ErrorReason, error) {
	blockInfo, err := c.getBlockInfoByHash(ctx, req.BlockHash)
	if err != nil {
		return nil, ffcapi.ErrorReason(""), err
	}
	if blockInfo == nil {
		return nil, ffcapi.ErrorReasonNotFound, i18n.NewError(ctx, msgs.MsgBlockNotAvailable)
	}

	res := &ffcapi.BlockInfoByHashResponse{}
	transformBlockInfo(blockInfo, &res.BlockInfo)
	return res, "", nil
}

func (c *tezosConnector) getBlockInfoByHash(ctx context.Context, hashString string) (*rpc.Block, error) {
	var blockInfo *rpc.Block
	cached, ok := c.blockCache.Get(hashString)
	if ok {
		blockInfo = cached.(*rpc.Block)
	}

	if blockInfo == nil {
		blockHash, err := tezos.ParseBlockHash(hashString)
		if err != nil {
			return nil, err
		}

		blockInfo, err = c.client.GetBlock(ctx, blockHash)
		if err != nil {
			return nil, err
		}
		if blockInfo == nil {
			return nil, nil
		}
		c.addToBlockCache(blockInfo)
	}

	return blockInfo, nil
}

func (c *tezosConnector) addToBlockCache(blockInfo *rpc.Block) {
	c.blockCache.Add(blockInfo.Hash.String(), blockInfo)
	c.blockCache.Add(strconv.Itoa(int(blockInfo.Header.Level)), blockInfo)
}

func transformBlockInfo(bi *rpc.Block, t *ffcapi.BlockInfo) {
	t.BlockNumber = fftypes.NewFFBigInt(bi.Header.Level)
	t.BlockHash = bi.Hash.String()
	t.ParentHash = bi.Header.Predecessor.String()
	stringHashes := make([]string, 0)
	// We take only 'Manager' operations that enable end-users to interact with the Tezos blockchain
	// e.g., transferring funds or calling smart contracts
	// see https://tezos.gitlab.io/active/blocks_ops.html#manager-operations-mumbai
	for _, tx := range bi.Operations[3] {
		stringHashes = append(stringHashes, tx.Hash.String())
	}
	t.TransactionHashes = stringHashes
}
