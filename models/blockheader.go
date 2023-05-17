// models/blockheader.go
package models

import (
	"encoding/json"
	"evm-importer/utils"
	"math/big"
	"time"
)

type JSONRPCRequest struct {
	ID      int64         `json:"id"`
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RawBlockHeaderResponse struct {
	Jsonrpc string     `json:"jsonrpc"`
	ID      int64      `json:"id"`
	Method  string     `json:"method"`
	Params  *RawParams `json:"params"`
}

type RawParams struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

type BlockHeader struct {
	ChainId          *big.Int
	ParentHash       string
	Sha3Uncles       string
	Miner            string
	StateRoot        string
	TransactionsRoot string
	ReceiptsRoot     string
	LogsBloom        string
	Difficulty       *big.Int
	Number           *big.Int
	GasLimit         *big.Int
	GasUsed          *big.Int
	Timestamp        time.Time
	ExtraData        string
	MixHash          string
	Nonce            string
	BaseFeePerGas    *big.Int
	WithdrawalsRoot  string
	Hash             string
}

func UnmarshalBlockHeader(data []byte, chainId *big.Int) (*BlockHeader, error) {
	var rawHeader map[string]string
	err := json.Unmarshal(data, &rawHeader)
	if err != nil {
		return nil, err
	}

	header := &BlockHeader{
		ChainId:          chainId,
		ParentHash:       rawHeader["parentHash"],
		Sha3Uncles:       rawHeader["sha3Uncles"],
		Miner:            rawHeader["miner"],
		StateRoot:        rawHeader["stateRoot"],
		TransactionsRoot: rawHeader["transactionsRoot"],
		ReceiptsRoot:     rawHeader["receiptsRoot"],
		LogsBloom:        rawHeader["logsBloom"],

		Difficulty: utils.HexToBigInt(rawHeader["difficulty"]),
		Number:     utils.HexToBigInt(rawHeader["number"]),
		GasLimit:   utils.HexToBigInt(rawHeader["gasLimit"]),
		GasUsed:    utils.HexToBigInt(rawHeader["gasUsed"]),
		ExtraData:  utils.HexToAscii(rawHeader["extraData"]),

		MixHash:         rawHeader["mixHash"],
		Nonce:           rawHeader["nonce"],
		BaseFeePerGas:   utils.HexToBigInt(rawHeader["baseFeePerGas"]),
		WithdrawalsRoot: rawHeader["withdrawalsRoot"],
		Hash:            rawHeader["hash"],
	}

	// Convert the Timestamp field to a UTC time
	timestampStr := rawHeader["timestamp"]
	timestamp := utils.HexToBigInt(timestampStr)
	header.Timestamp = time.Unix(timestamp.Int64(), 0).UTC()

	return header, nil
}
