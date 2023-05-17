// models/transaction.go
package models

import (
	"encoding/json"
	"evm-importer/utils"
	"math/big"
	"time"
)

type Transaction struct {
	ChainId          *big.Int
	Timestamp        time.Time
	Hash             string
	BlockNumber      *big.Int
	TransactionIndex *big.Int
	From             string
	To               string
	Nonce            *big.Int
	Value            *big.Int
	GasPrice         *big.Int
	Gas              *big.Int
	GasUsed          *big.Int
	Input            string
	V                string
	R                string
	S                string
	BlockHash        string
	Status           *big.Int
}

func UnmarshalTransaction(data []byte, blockTimestamp *big.Int, txReceipt map[string]interface{}) (*Transaction, error) {
	var rawTransaction map[string]interface{}
	err := json.Unmarshal(data, &rawTransaction)
	if err != nil {
		return nil, err
	}

	tx := &Transaction{
		ChainId:          utils.HexToBigInt(utils.GetStringOrNil(rawTransaction["chainId"])),
		Hash:             utils.GetStringOrNil(rawTransaction["hash"]),
		Nonce:            utils.HexToBigInt(utils.GetStringOrNil(rawTransaction["nonce"])),
		BlockHash:        utils.GetStringOrNil(rawTransaction["blockHash"]),
		BlockNumber:      utils.HexToBigInt(utils.GetStringOrNil(rawTransaction["blockNumber"])),
		TransactionIndex: utils.HexToBigInt(utils.GetStringOrNil(rawTransaction["transactionIndex"])),
		From:             utils.GetStringOrNil(rawTransaction["from"]),
		To:               utils.GetStringOrNil(rawTransaction["to"]),
		Value:            utils.HexToBigInt(utils.GetStringOrNil(rawTransaction["value"])),
		GasPrice:         utils.HexToBigInt(utils.GetStringOrNil(rawTransaction["gasPrice"])),
		Gas:              utils.HexToBigInt(utils.GetStringOrNil(rawTransaction["gas"])),
		GasUsed:          utils.HexToBigInt(utils.GetStringOrNil(txReceipt["gasUsed"])),
		Status:           utils.HexToBigInt(utils.GetStringOrNil(txReceipt["status"])),
		Input:            utils.GetStringOrNil(rawTransaction["input"]),
		V:                utils.GetStringOrNil(rawTransaction["v"]),
		R:                utils.GetStringOrNil(rawTransaction["r"]),
		S:                utils.GetStringOrNil(rawTransaction["s"]),
	}

	// Set the Timestamp field to the block timestamp
	tx.Timestamp = time.Unix(blockTimestamp.Int64(), 0).UTC()

	return tx, nil
}
