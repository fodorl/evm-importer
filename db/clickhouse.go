// db/clickhouse.go
package db

import (
	"database/sql"
	"evm-importer/config"
	"evm-importer/models"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	_ "github.com/mailru/go-clickhouse"
)

type Clickhouse struct {
	db       *sql.DB
	database string
}

func NewClickhouseConnection(cfg config.ClickhouseConfig) (*Clickhouse, error) {
	connStr := fmt.Sprintf("http://%s:%s?&password=%s&read_timeout=10s&write_timeout=20s",
		cfg.Host, cfg.Port, cfg.Password)

	db, err := sql.Open("clickhouse", connStr)
	if err != nil {
		log.Println("Error executing query:", err)
		debug.PrintStack() // Print stack trace
		return nil, err
	}

	// Set the database to use
	if _, err = db.Exec(fmt.Sprintf("USE %s", cfg.Database)); err != nil {
		log.Println("Error setting database:", err)
		debug.PrintStack() // Print stack trace
		return nil, err
	}

	return &Clickhouse{db: db, database: cfg.Database}, nil

}

func (ch *Clickhouse) SaveBlockHeader(header *models.BlockHeader) error {
	query := fmt.Sprintf(`
        INSERT INTO %s.blocks (
            chain_id, timestamp, hash, number, extra_data, base_fee_per_gas,
            gas_used, gas_limit, miner, parent_hash, sha3_uncles,
            state_root, transactions_root, receipts_root, logs_bloom,
            difficulty, mix_hash, nonce, withdrawals_root
        ) VALUES (
            ?, ?, ?, ?, ?, ?,
            ?, ?, ?, ?, ?,
            ?, ?, ?, ?, ?, ?, ?,
            ?
        )
    `, ch.database)

	stmt, err := ch.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		header.ChainId.String(), header.Timestamp, header.Hash, header.Number.String(), header.ExtraData, header.BaseFeePerGas.String(),
		header.GasUsed.String(), header.GasLimit.String(), header.Miner, header.ParentHash, header.Sha3Uncles,
		header.StateRoot, header.TransactionsRoot, header.ReceiptsRoot, header.LogsBloom,
		header.Difficulty.String(), header.MixHash, header.Nonce, header.WithdrawalsRoot,
	)

	if err != nil {
		log.Println("Error executing query:", err)
		debug.PrintStack() // Print stack trace
		return err
	}

	fmt.Printf("Block header saved: %s\n", header.Hash)
	return nil
}

func (ch *Clickhouse) SaveTransactions(transactions []*models.Transaction) error {
	// Prepare the bulk insert query
	query := fmt.Sprintf(`
		INSERT INTO %s.transactions (
			chain_id, timestamp, hash, block_number, transaction_index,
			from_address, to_address, nonce, value, gas_price, gas_limit, gas_used, status,
			input, v, r, s, block_hash
		) VALUES`, ch.database)

	// Prepare the values placeholders and the values slice
	valuePlaceholders := ""
	values := make([]interface{}, 0)

	for _, t := range transactions {
		valuePlaceholders += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
		values = append(values,
			t.ChainId.String(), t.Timestamp, t.Hash, t.BlockNumber.String(), t.TransactionIndex.String(),
			t.From, t.To, t.Nonce.String(), t.Value.String(), t.GasPrice.String(), t.Gas.String(), t.GasUsed.String(),
			t.Status.String(), t.Input, t.V, t.R, t.S, t.BlockHash,
		)
	}

	// Remove the trailing comma from value placeholders
	valuePlaceholders = strings.TrimRight(valuePlaceholders, ",")

	// Append the value placeholders to the query
	query += valuePlaceholders

	// Execute the bulk insert query
	_, err := ch.db.Exec(query, values...)
	if err != nil {
		log.Println("Error executing query:", err)
		debug.PrintStack() // Print stack trace
		return err
	}

	fmt.Printf("Transactions saved: %d\n", len(transactions))
	return nil
}
