package bitcoin

import (
	"encoding/json"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

type Client struct {
	RPCClient *rpcclient.Client
}

func NewClient(host string, user string, pass string) (*Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	rpcClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		RPCClient: rpcClient,
	}, nil
}

func (c *Client) GetBlockHeightByHash(hash *chainhash.Hash) (int64, error) {
	blockVerbose, err := c.RPCClient.GetBlockVerbose(hash)
	if err != nil {
		return 0, err
	}
	return blockVerbose.Height, nil
}

func (c *Client) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	blockChainInfo, err := c.RPCClient.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}
	return blockChainInfo, nil
}

func (c *Client) GetBlockHashesByStartHeight(startHeight int64, count int64) ([]*chainhash.Hash, error) {
	var loopErr error
	blockHashes := make([]*chainhash.Hash, 0, count)
	for i := int64(0); i < count; i++ {
		blockHash, err := c.RPCClient.GetBlockHash(startHeight + i)
		if err != nil {
			loopErr = err
			break
		}
		blockHashes = append(blockHashes, blockHash)
	}
	return blockHashes, loopErr
}

func (c *Client) GetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error) {
	tx, err := c.RPCClient.GetRawTransactionVerbose(txID)
	if err != nil {
		return nil, err
	}
	return chainhash.NewHashFromStr(tx.BlockHash)
}

func (c *Client) GetTxProof(txID *chainhash.Hash, blockHash *chainhash.Hash) (string, error) {
	blockHashStr := blockHash.String()
	blockHashStrPtr := &blockHashStr
	cmd := btcjson.NewGetTxOutProofCmd([]string{txID.String()}, blockHashStrPtr)

	responseChannel := c.RPCClient.SendCmd(cmd)

	response, err := rpcclient.ReceiveFuture(responseChannel)
	if err != nil {
		return "", err
	}

	var proof string

	err = json.Unmarshal(response, &proof)
	if err != nil {
		return "", err
	}

	return proof, nil
}

func (c *Client) ShutdownRPCClient() {
	c.RPCClient.Shutdown()
}
