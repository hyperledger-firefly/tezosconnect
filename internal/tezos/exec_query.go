package tezos

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger-firefly/firefly-tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/trilitech/tzgo/micheline"
	"github.com/trilitech/tzgo/rpc"
	"github.com/trilitech/tzgo/tezos"
)

// QueryInvoke executes a method on a blockchain smart contract, which might execute Smart Contract code, but does not affect the blockchain state.
func (c *tezosConnector) QueryInvoke(ctx context.Context, req *ffcapi.QueryInvokeRequest) (*ffcapi.QueryInvokeResponse, ffcapi.ErrorReason, error) {
	if req == nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, errors.New("request is not defined")
	}

	params, err := c.prepareInputParams(ctx, &req.TransactionInput)
	if err != nil {
		return nil, ffcapi.ErrorReasonInvalidInputs, err
	}

	resp, err := c.runView(ctx, params.Entrypoint, req.From, req.To, params.Value)
	if err != nil {
		return nil, ffcapi.ErrorReasonTransactionReverted, err
	}
	return &ffcapi.QueryInvokeResponse{
		Outputs: convertRunViewResponseToOutputs(resp),
	}, "", nil
}

func (c *tezosConnector) runView(ctx context.Context, entrypoint, addrFrom, addrTo string, args micheline.Prim) (rpc.RunViewResponse, error) {
	toAddress, err := tezos.ParseAddress(addrTo)
	if err != nil {
		return rpc.RunViewResponse{}, i18n.NewError(ctx, msgs.MsgInvalidToAddress, addrTo, err)
	}

	fromAddress, err := tezos.ParseAddress(addrFrom)
	if err != nil {
		return rpc.RunViewResponse{}, i18n.NewError(ctx, msgs.MsgInvalidFromAddress, addrFrom, err)
	}

	req := rpc.RunViewRequest{
		Contract:     toAddress,
		View:         entrypoint,
		Input:        args,
		Source:       fromAddress,
		Payer:        fromAddress,
		UnlimitedGas: true,
		Mode:         "Readable",
	}

	var res rpc.RunViewResponse
	err = c.client.RunView(ctx, rpc.Head, &req, &res)
	if err != nil {
		return rpc.RunViewResponse{}, err
	}
	return res, nil
}

func convertRunViewResponseToOutputs(resp rpc.RunViewResponse) *fftypes.JSONAny {
	var res interface{}
	if resp.Data.LooksLikeMap() {
		resultMap := make([]map[string]interface{}, len(resp.Data.Args))
		for i, arg := range resp.Data.Args {
			mapEntry := make(map[string]interface{}, 2)
			key := arg.Args[0]
			value := arg.Args[1]
			mapEntry["key"] = key.Value(key.OpCode)
			mapEntry["value"] = value.Value(value.OpCode)
			resultMap[i] = mapEntry
		}
		res = resultMap
	} else {
		res = resp.Data.Value(resp.Data.OpCode)
	}

	outputs, _ := json.Marshal(res)
	if val, ok := res.(string); ok {
		if values := strings.Split(val, ","); len(values) > 1 {
			outputs, _ = json.Marshal(values)
		}
	}
	return fftypes.JSONAnyPtrBytes(outputs)
}
