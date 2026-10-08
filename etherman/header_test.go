package etherman

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

// Block from an Amsterdam L1: go-ethereum v1.15 does not know blockAccessListHash/slotNumber,
// so header.Hash() differs from the hash reported by the node.
const amsterdamBlockJSON = `{
  "baseFeePerGas":"0x7","blobGasUsed":"0x0","blockAccessListHash":"0x66de9db265e50af2ee0fe4bad325a892a9dfa466ac453ae45d4f029d397f42e8",
  "difficulty":"0x0","excessBlobGas":"0x0","extraData":"0x","gasLimit":"0xbebc200","gasUsed":"0x0",
  "hash":"0x2000c541efe4e6307c383c5e13a0792acf4eb2daba48c29efcd126c4374e3f88",
  "logsBloom":"0x` + zeroBloom + `",
  "miner":"0x0000000000000000000000000000000000000000","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000",
  "nonce":"0x0000000000000000","number":"0x54",
  "parentBeaconBlockRoot":"0x0000000000000000000000000000000000000000000000000000000000000000",
  "parentHash":"0x1111111111111111111111111111111111111111111111111111111111111111",
  "receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  "requestsHash":"0xe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","slotNumber":"0x54",
  "stateRoot":"0x2222222222222222222222222222222222222222222222222222222222222222","timestamp":"0x6700000a",
  "transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  "withdrawalsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  "transactions":[],"uncles":[],"withdrawals":[]
}`

const zeroBloom = "0000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000" +
	"0000000000000000000000000000000000000000000000000000000000000000"

func TestHeaderWithHashByNumberUsesNodeHash(t *testing.T) {
	var gotParams []json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req struct {
			ID     json.RawMessage   `json:"id"`
			Params []json.RawMessage `json:"params"`
		}
		require.NoError(t, json.Unmarshal(body, &req))
		gotParams = req.Params
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(req.ID) + `,"result":` + amsterdamBlockJSON + `}`))
	}))
	defer srv.Close()

	client, err := ethclient.Dial(srv.URL)
	require.NoError(t, err)

	header, hash, err := HeaderWithHashByNumber(context.Background(), client, big.NewInt(84))
	require.NoError(t, err)
	require.Equal(t, `"0x54"`, string(gotParams[0]))
	require.Equal(t, uint64(84), header.Number.Uint64())
	require.Equal(t, common.HexToHash("0x2000c541efe4e6307c383c5e13a0792acf4eb2daba48c29efcd126c4374e3f88"), hash)
	require.NotEqual(t, hash, header.Hash())
}
