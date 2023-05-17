// utils/utils.go
package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strings"
	"time"
)

func HexToBigInt(hex string) *big.Int {
	// Remove the '0x' prefix if it exists
	cleanHex := strings.TrimPrefix(hex, "0x")

	bigInt := new(big.Int)
	bigInt.SetString(cleanHex, 16)
	return bigInt
}

func HexToAscii(hexStr string) string {
	decoded, err := hex.DecodeString(strings.TrimPrefix(hexStr, "0x"))
	if err != nil {
		return ""
	}
	return string(decoded)
}

func EpochToTimeUTC(epoch int64) *big.Int {
	timestamp := time.Unix(epoch, 0).UTC()
	return big.NewInt(timestamp.Unix())
}

func GetStringOrNil(value interface{}) string {
	if value != nil {
		return value.(string)
	}
	return ""
}

func DeriveChainID(v string) (*big.Int, error) {
	vBigInt, success := new(big.Int).SetString(v, 0)
	if !success {
		return nil, fmt.Errorf("failed to parse V value")
	}

	chainID := new(big.Int).Sub(vBigInt, big.NewInt(35))
	chainID.Div(chainID, big.NewInt(2))

	return chainID, nil
}

func DerivePublicKey(r, s, v, hash string) (string, error) {
	rBigInt := HexToBigInt(r)
	sBigInt := HexToBigInt(s)
	vBigInt := HexToBigInt(v)

	signature := append(rBigInt.Bytes(), sBigInt.Bytes()...)
	signature = append(signature, vBigInt.Bytes()...)

	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}

	publicKeyECDSA, err := crypto.SigToPub(hashBytes, signature)
	if err != nil {
		return "", err
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKey := common.BytesToAddress(publicKeyBytes).Hex()

	return publicKey, nil
}
