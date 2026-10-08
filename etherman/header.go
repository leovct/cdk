package etherman

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type headerByNumberClient interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

type rpcClientProvider interface {
	Client() *rpc.Client
}

// HeaderWithHashByNumber returns the header together with the hash reported by the node.
// header.Hash() is recomputed from the fields go-ethereum knows about, so it is wrong
// for blocks of forks it does not support yet (e.g. Amsterdam).
func HeaderWithHashByNumber(
	ctx context.Context, client headerByNumberClient, number *big.Int,
) (*types.Header, common.Hash, error) {
	provider, ok := client.(rpcClientProvider)
	if !ok {
		header, err := client.HeaderByNumber(ctx, number)
		if err != nil {
			return nil, common.Hash{}, err
		}
		return header, header.Hash(), nil
	}

	var raw json.RawMessage
	if err := provider.Client().CallContext(ctx, &raw, "eth_getBlockByNumber", toBlockNumArg(number), false); err != nil {
		return nil, common.Hash{}, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, common.Hash{}, ethereum.NotFound
	}
	var header types.Header
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, common.Hash{}, err
	}
	var withHash struct {
		Hash common.Hash `json:"hash"`
	}
	if err := json.Unmarshal(raw, &withHash); err != nil {
		return nil, common.Hash{}, err
	}
	return &header, withHash.Hash, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	if number.IsInt64() {
		return rpc.BlockNumber(number.Int64()).String()
	}
	return "<invalid " + number.String() + ">"
}
