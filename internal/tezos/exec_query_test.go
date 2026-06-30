package tezos

import (
	"math/big"
	"os"
	"testing"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trilitech/tzgo/micheline"
	"github.com/trilitech/tzgo/rpc"
)

func TestQueryInvokeSuccess(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	req := &ffcapi.QueryInvokeRequest{
		TransactionInput: ffcapi.TransactionInput{
			Method: fftypes.JSONAnyPtr("\"simple_view\""),
		},
	}
	res := rpc.RunViewResponse{
		Data: micheline.Prim{
			Type:   micheline.PrimString,
			String: "3",
		},
	}
	mRPC.On("RunView", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(3).(*rpc.RunViewResponse)
		*arg = res
	})

	resp, reason, err := c.QueryInvoke(ctx, req)

	assert.NotNil(t, resp)
	assert.Equal(t, resp.Outputs.String(), `"3"`)
	assert.Empty(t, reason)
	assert.NoError(t, err)
}

func TestQueryInvokeSuccessArray(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	req := &ffcapi.QueryInvokeRequest{
		TransactionInput: ffcapi.TransactionInput{
			Method: fftypes.JSONAnyPtr("\"simple_view\""),
		},
	}
	res := rpc.RunViewResponse{
		Data: micheline.Prim{
			Type:   micheline.PrimSequence,
			OpCode: micheline.D_PAIR,
			Args: []micheline.Prim{
				{Type: micheline.PrimString, String: "str"},
				{Type: micheline.PrimInt, Int: big.NewInt(1)},
			},
		},
	}
	mRPC.On("RunView", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(3).(*rpc.RunViewResponse)
		*arg = res
	})

	resp, reason, err := c.QueryInvoke(ctx, req)

	assert.NotNil(t, resp)
	assert.Equal(t, resp.Outputs.String(), `["str","1"]`)
	assert.Empty(t, reason)
	assert.NoError(t, err)
}

func TestQueryInvokeSuccessMap(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	req := &ffcapi.QueryInvokeRequest{
		TransactionInput: ffcapi.TransactionInput{
			Method: fftypes.JSONAnyPtr("\"simple_view\""),
		},
	}
	res := rpc.RunViewResponse{
		Data: micheline.Prim{
			Type: micheline.PrimSequence,
			Args: micheline.PrimList{
				{
					OpCode: micheline.D_ELT,
					Args: []micheline.Prim{
						{Type: micheline.PrimString, String: "str1"},
						{Type: micheline.PrimInt, Int: big.NewInt(1)},
					},
				},
				{
					OpCode: micheline.D_ELT,
					Args: []micheline.Prim{
						{Type: micheline.PrimString, String: "str2"},
						{Type: micheline.PrimInt, Int: big.NewInt(2)},
					},
				},
			},
		},
	}
	mRPC.On("RunView", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(3).(*rpc.RunViewResponse)
		*arg = res
	})

	resp, reason, err := c.QueryInvoke(ctx, req)

	assert.NotNil(t, resp)
	assert.Equal(t, resp.Outputs.String(), `[{"key":"str1","value":"1"},{"key":"str2","value":"2"}]`)
	assert.Empty(t, reason)
	assert.NoError(t, err)
}

func TestQueryInvokeRunViewError(t *testing.T) {
	ctx, c, mRPC, done := newTestConnector(t)
	defer done()

	req := &ffcapi.QueryInvokeRequest{
		TransactionInput: ffcapi.TransactionInput{
			Method: fftypes.JSONAnyPtr("\"simple_view\""),
		},
	}
	mRPC.On("RunView", ctx, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	resp, reason, err := c.QueryInvoke(ctx, req)

	assert.Nil(t, resp)
	assert.Equal(t, reason, ffcapi.ErrorReasonTransactionReverted)
	assert.Error(t, err)
}

func TestQueryInvokeWrongParamsError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	os.Setenv("ENV", "test")
	req := &ffcapi.QueryInvokeRequest{
		TransactionInput: ffcapi.TransactionInput{
			Method: fftypes.JSONAnyPtr("\"simple_view\""),
			Params: []*fftypes.JSONAny{
				fftypes.JSONAnyPtr("wrong"),
			},
		},
	}

	resp, reason, err := c.QueryInvoke(ctx, req)

	assert.Nil(t, resp)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
	assert.Error(t, err)
}

func TestQueryInvokeParseAddressToError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.QueryInvokeRequest{
		TransactionInput: ffcapi.TransactionInput{
			Method: fftypes.JSONAnyPtr("\"simple_view\""),
			TransactionHeaders: ffcapi.TransactionHeaders{
				To: "t......",
			},
		},
	}

	resp, reason, err := c.QueryInvoke(ctx, req)

	assert.Nil(t, resp)
	assert.Equal(t, reason, ffcapi.ErrorReasonTransactionReverted)
	assert.Error(t, err)
}

func TestQueryInvokeParseAddressFromError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	req := &ffcapi.QueryInvokeRequest{
		TransactionInput: ffcapi.TransactionInput{
			Method: fftypes.JSONAnyPtr("\"simple_view\""),
			TransactionHeaders: ffcapi.TransactionHeaders{
				From: "t......",
			},
		},
	}

	resp, reason, err := c.QueryInvoke(ctx, req)

	assert.Nil(t, resp)
	assert.Equal(t, reason, ffcapi.ErrorReasonTransactionReverted)
	assert.Error(t, err)
}

func TestQueryInvokeRequestNotDefinedError(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	resp, reason, err := c.QueryInvoke(ctx, nil)

	assert.Nil(t, resp)
	assert.Equal(t, reason, ffcapi.ErrorReasonInvalidInputs)
	assert.Error(t, err)
}
