package tezos

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hyperledger-firefly/tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/trilitech/tzgo/codec"
	"github.com/trilitech/tzgo/contract"
	"github.com/trilitech/tzgo/micheline"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

// TransactionPrepare validates transaction inputs against the supplied schema/Michelson and performs any binary serialization required (prior to signing) to encode a transaction from JSON into the native blockchain format
func (c *tezosConnector) TransactionPrepare(ctx context.Context, req *ffcapi.TransactionPrepareRequest) (res *ffcapi.TransactionPrepareResponse, reason ffcapi.ErrorReason, err error) {
	params, err := c.prepareInputParams(ctx, &req.TransactionInput)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	op, err := c.buildOp(ctx, params, req.From, req.To, req.Nonce)
	if err != nil {
		return nil, "", err
	}

	opts := &rpc.DefaultOptions
	if reason, err = c.estimateAndAssignTxCost(ctx, op, opts); err != nil {
		return nil, reason, err
	}
	log.L(ctx).Infof("Prepared transaction method=%s dataLen=%d", req.Method.String(), len(op.Bytes()))

	return &ffcapi.TransactionPrepareResponse{
		Gas:             req.Gas,
		TransactionData: hex.EncodeToString(op.Bytes()),
	}, "", nil
}

func (c *tezosConnector) estimateAndAssignTxCost(ctx context.Context, op *codec.Op, opts *rpc.CallOptions) (ffcapi.ErrorReason, error) {
	// Simulate the transaction (dry run)
	sim, reason, err := c.callTransaction(ctx, op, nil)
	if err != nil {
		return reason, err
	}

	// apply simulated cost as limits to tx list
	if !opts.IgnoreLimits {
		op.WithLimits(sim.MinLimits(), rpc.ExtraSafetyMargin)
	}

	// log info about tx costs
	costs := sim.Costs()
	for i, v := range op.Contents {
		verb := "used"
		if opts.IgnoreLimits {
			verb = "forced"
		}
		limits := v.Limits()
		log.L(ctx).Debugf("OP#%03d: %s gas_used(sim)=%d storage_used(sim)=%d storage_burn(sim)=%d alloc_burn(sim)=%d fee(%s)=%d gas_limit(%s)=%d storage_limit(%s)=%d ",
			i, v.Kind(), costs[i].GasUsed, costs[i].StorageUsed, costs[i].StorageBurn, costs[i].AllocationBurn,
			verb, limits.Fee, verb, limits.GasLimit, verb, limits.StorageLimit,
		)
	}

	// check minFee calc against maxFee if set
	if opts.MaxFee > 0 {
		if l := op.Limits(); l.Fee > opts.MaxFee {
			return "", fmt.Errorf("estimated cost %d > max %d", l.Fee, opts.MaxFee)
		}
	}

	return "", nil
}

func (c *tezosConnector) callTransaction(ctx context.Context, op *codec.Op, opts *rpc.CallOptions) (*rpc.Receipt, ffcapi.ErrorReason, error) {
	sim, err := c.client.Simulate(ctx, op, opts)
	if err != nil {
		return nil, mapError(callRPCMethods, err), err
	}
	// fail with Tezos error when simulation failed
	if !sim.IsSuccess() {
		return nil, mapError(callRPCMethods, sim.Error()), sim.Error()
	}
	return sim, "", nil
}

func (c *tezosConnector) prepareInputParams(ctx context.Context, req *ffcapi.TransactionInput) (micheline.Parameters, error) {
	var tezosParams micheline.Parameters

	for i, p := range req.Params {
		if p != nil {
			err := tezosParams.UnmarshalJSON([]byte(*p))
			if err != nil {
				return tezosParams, i18n.NewError(ctx, msgs.MsgUnmarshalParamFail, i, err)
			}
		}
	}

	return tezosParams, nil
}

func (c *tezosConnector) buildOp(ctx context.Context, params micheline.Parameters, fromString, toString string, nonce *fftypes.FFBigInt) (*codec.Op, error) {
	op := codec.NewOp()

	toAddress, err := tezos.ParseAddress(toString)
	if err != nil {
		return nil, i18n.NewError(ctx, msgs.MsgInvalidToAddress, toString, err)
	}

	txArgs := contract.TxArgs{}
	txArgs.WithParameters(params)
	txArgs.WithDestination(toAddress)
	op.WithContents(txArgs.Encode())

	err = c.completeOp(ctx, op, fromString, nonce)
	if err != nil {
		return nil, err
	}

	return op, nil
}

func (c *tezosConnector) completeOp(ctx context.Context, op *codec.Op, fromString string, nonce *fftypes.FFBigInt) error {
	fromAddress, err := tezos.ParseAddress(fromString)
	if err != nil {
		return i18n.NewError(ctx, msgs.MsgInvalidFromAddress, fromString, err)
	}
	op.WithSource(fromAddress)

	hash, _ := c.client.GetBlockHash(ctx, rpc.Head)
	op.WithBranch(hash)

	op.WithParams(getNetworkParamsByName(c.networkName))

	state, err := c.client.GetContractExt(ctx, fromAddress, rpc.Head)
	if err != nil {
		return err
	}

	mayNeedReveal := len(op.Contents) > 0 && op.Contents[0].Kind() != tezos.OpTypeReveal
	// add reveal if necessary
	if mayNeedReveal && !state.IsRevealed() {
		key, err := c.getPubKeyFromSignatory(ctx, op.Source.String())
		if err != nil {
			return err
		}

		reveal := &codec.Reveal{
			Manager: codec.Manager{
				Source: fromAddress,
			},
			PublicKey: *key,
		}
		reveal.WithLimits(rpc.DefaultRevealLimits)
		op.WithContentsFront(reveal)
	}

	// assign nonce
	nextCounter := nonce.Int64()
	// Note: there are situations when a nonce becomes obsolete after assigning it to connector.NextNonceForSigner.
	// In such cases, we update it with a more recent one.
	if nextCounter != state.Counter+1 {
		nextCounter = state.Counter + 1
	}
	for _, op := range op.Contents {
		op.WithCounter(nextCounter)
		nextCounter++
	}

	return nil
}

func getNetworkParamsByName(name string) *tezos.Params {
	switch strings.ToLower(name) {
	case "ghostnet":
		return tezos.GhostnetParams
	case "parisnet":
		return tezos.ParisnetParams
	default:
		return tezos.DefaultParams
	}
}

func (c *tezosConnector) getPubKeyFromSignatory(ctx context.Context, tezosAddress string) (*tezos.Key, error) {
	url := c.signatoryURL + "/keys/" + tezosAddress
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("signatory resp with wrong status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var pubKeyJSON struct {
		PubKey string `json:"public_key"`
	}
	err = json.Unmarshal(body, &pubKeyJSON)
	if err != nil {
		return nil, err
	}

	key, err := tezos.ParseKey(pubKeyJSON.PubKey)
	if err != nil {
		return nil, err
	}

	return &key, nil
}
