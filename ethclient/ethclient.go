// ethclient/ethclient.go
package ethclient

import (
	"bytes"
	"encoding/json"
	"evm-importer/db"
	"evm-importer/models"
	"evm-importer/utils"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/big"
	"net/http"
	"runtime/debug"
	"time"
)

var subID string

func subscribeToNewBlockHeaders(c *websocket.Conn) {
	req := models.JSONRPCRequest{
		ID:      1,
		JsonRPC: "2.0",
		Method:  "eth_subscribe",
		Params:  []interface{}{"newHeads"},
	}

	err := c.WriteJSON(req)
	if err != nil {
		log.Println("write:", err)
	}
}
func processBlockHeader(message string, chConn *db.Clickhouse, httpUrl string) {
	var rawResponse models.RawBlockHeaderResponse
	err := json.Unmarshal([]byte(message), &rawResponse)
	if err != nil {
		log.Println("Unmarshal:", err)
		return
	}

	if rawResponse.Method == "eth_subscription" && rawResponse.Params != nil {
		chainId, err := requestChainId(httpUrl)
		if err != nil {
			log.Println("Error requesting chainId:", err)
			return
		}

		blockHeader, err := models.UnmarshalBlockHeader(rawResponse.Params.Result, chainId)
		if err != nil {
			log.Println("Unmarshal BlockHeader:", err)
			return
		}
		log.Printf("Parsed block header: %+v\n", blockHeader)

		// Save the block header to the Clickhouse database
		err = chConn.SaveBlockHeader(blockHeader)
		if err != nil {
			log.Println("Save block header error:", err)
			debug.PrintStack() // Print stack trace
		}
		go fetchBlockByHash(httpUrl, blockHeader.Hash, chConn)
	}
}

func NewClient(url string, httpUrl string, chConn *db.Clickhouse) error {
	// Declare a variable to hold the WebSocket connection
	var c *websocket.Conn
	// Declare a variable to hold any errors
	var err error
	// Declare a variable to hold the message
	var message []byte

	// Define a function to handle connecting (or reconnecting) to the WebSocket
	reconnect := func() {
		for {
			c, _, err = websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				log.Println("Error connecting to WebSocket:", err)
				log.Println("Retrying...")
				time.Sleep(5 * time.Second)
			} else {
				// If the connection was successful, subscribe to new block headers
				subscribeToNewBlockHeaders(c)
				log.Println("Connected to node, processing block headers...")
				break // Break the loop if connection is successful
			}
		}
	}

	// Call the reconnect function to initiate the first connection
	reconnect()

	// If there was an error connecting, return the error
	if err != nil {
		return err
	}

	for {
		_, message, err = c.ReadMessage()

		// If there was an error reading a message, check if it was due to the WebSocket closing
		if err != nil {
			log.Println("read:", err)

			// Check if the error is a CloseError, indicating the WebSocket closed
			if _, ok := err.(*websocket.CloseError); ok {
				log.Println("WebSocket closed, attempting to reconnect...")

				// Wait for 1 seconds before trying to reconnect
				time.Sleep(1 * time.Second)

				// Attempt to reconnect
				reconnect()

				// If there was an error reconnecting, return the error
				if err != nil {
					return err
				}

				// Skip processing the message for this iteration
				continue
			}

			// If the error was not a CloseError, return the error
			return err
		}

		// Process the received block header only if there were no errors
		processBlockHeader(string(message), chConn, httpUrl)
	}
}

func fetchBlockByHash(url string, blockHash string, chConn *db.Clickhouse) {
	reqBody := fmt.Sprintf(`
	{
		"jsonrpc": "2.0",
		"id": 2,
		"method": "eth_getBlockByHash",
		"params": ["%s", true]
	}
	`, blockHash)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		log.Println("Error fetching block by hash:", err)
		return
	}
	defer resp.Body.Close()

	err = processBlockByHashResponse(resp, chConn, url)
	if err != nil {
		log.Println("Error processing block by hash response:", err)
	}
}

func processBlockByHashResponse(response *http.Response, chConn *db.Clickhouse, httpURL string) error {
	var rawResponse map[string]interface{}
	err := json.NewDecoder(response.Body).Decode(&rawResponse)
	if err != nil {
		return err
	}

	result := rawResponse["result"].(map[string]interface{})
	blockTimestamp := utils.HexToBigInt(utils.GetStringOrNil(result["timestamp"]))
	rawTransactions, ok := result["transactions"].([]interface{})
	if !ok {
		log.Println("No transactions found in the block")
		return nil
	}

	processedTransactions := make([]*models.Transaction, len(rawTransactions))

	for i, rawTransaction := range rawTransactions {
		rawTransactionBytes, err := json.Marshal(rawTransaction)
		if err != nil {
			return err
		}

		txHash := utils.GetStringOrNil(rawTransaction.(map[string]interface{})["hash"])
		txReceipt, err := fetchTransactionReceipt(httpURL, txHash)
		//log.Println("Fetched transaction receipt: %+v\n", txReceipt)
		if err != nil {
			log.Println("Error fetching transaction receipt:", err)
			return err
		}

		tx, err := models.UnmarshalTransaction(rawTransactionBytes, blockTimestamp, txReceipt)
		//log.Printf("Unmarshalled transaction: %+v\n", tx)
		if err != nil {
			return err
		}

		processedTransactions[i] = tx
	}

	return chConn.SaveTransactions(processedTransactions)
}
func requestChainId(httpUrl string) (*big.Int, error) {
	requestBody, err := json.Marshal(models.JSONRPCRequest{
		ID:      1,
		JsonRPC: "2.0",
		Method:  "eth_chainId",
		Params:  []interface{}{},
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(httpUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response map[string]json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	var chainIdHex string
	err = json.Unmarshal(response["result"], &chainIdHex)
	if err != nil {
		return nil, err
	}

	chainId := utils.HexToBigInt(chainIdHex)

	return chainId, nil
}

func fetchTransactionReceipt(url string, txHash string) (map[string]interface{}, error) {
	reqBody := fmt.Sprintf(`
	{
		"jsonrpc": "2.0",
		"id": 3,
		"method": "eth_getTransactionReceipt",
		"params": ["%s"]
	}
	`, txHash)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&rawResponse)
	if err != nil {
		return nil, err
	}

	result, ok := rawResponse["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to fetch transaction receipt")
	}

	return result, nil
}

func fetchTransactionTrace(url string, txHash string) (map[string]interface{}, error) {
	reqBody := fmt.Sprintf(`
	{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "debug_traceTransaction",
		"params": ["%s"]
	}
	`, txHash)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&rawResponse)
	if err != nil {
		return nil, err
	}

	result, ok := rawResponse["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to fetch transaction trace")
	}

	return result, nil
}
