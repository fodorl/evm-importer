// main.go
package main

import (
	"evm-importer/config"
	"evm-importer/db"
	"evm-importer/ethclient"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	db, err := db.NewClickhouseConnection(cfg.Clickhouse)

	if err != nil {
		log.Fatal("Error connecting to Clickhouse:", err)
	}
	log.Println("Connected to Clickhouse")

	err = ethclient.NewClient(cfg.NodeURL, cfg.HTTPURL, db)
	if err != nil {
		log.Fatal("Error initializing Ethereum client:", err)
	}

}
