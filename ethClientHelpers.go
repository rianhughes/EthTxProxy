package main

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// sendTxsLoop calls sendTxsToNodes() every 'sendTxPeriod' seconds.
func sendTxsLoop(eClient *ethclient.Client) {

	ticker := time.NewTicker(time.Duration(sendTxPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			baseFee := getBaseFee(eClient)
			sendTxs(eClient, baseFee)
		}
	}
}

// sendTxsToNodes gets the most recent baseGasFee from the ethClient, and sends any local tranasctions that have a gasFee higher than the baseFee.
func sendTxs(eClient *ethclient.Client, baseFee uint64) {

	txPool.mu.Lock()
	defer txPool.mu.Unlock()

	for rawTx, v := range txPool.txmap {
		if v >= baseFee {
			InfoLog.Printf("---> Sending transaction (0x..%s) to ethClient, since transaction gasPrice (%d) >= baseGasPrice (%d).\n\n", rawTx[len(rawTx)-8:], v, baseFee)
			tx, _ := rawTxToTx(rawTx)
			eClient.SendTransaction(context.Background(), tx)
			delete(txPool.txmap, rawTx)
		} else {
			InfoLog.Printf("---> Not sending transaction (0x..%s) to ethClient, since transaction gasPrice (%d) < baseGasPrice (%d).\n\n", rawTx[len(rawTx)-8:], v, baseFee)
		}
	}

}

// getBaseFee queries the ethereum client for the latest baseGasFee.
// If this fails, baseGasFee is set to 'defaultBaseGasFee'
func getBaseFee(eClient *ethclient.Client) uint64 {
	header, err := eClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		WarningLog.Printf("---> getBaseFee() failed to connect to the Ethereum client. Setting baseFee = %d\n. Error :%s", defaultBaseGasFee, err)
		return defaultBaseGasFee
	}
	return header.BaseFee.Uint64()
}

func getEthClient(ethClientURL string) *ethclient.Client {
	client, err := ethclient.Dial(ethClientURL)

	if err != nil {
		WarningLog.Printf("---> getEthClient() failed to connect to the Ethereum client. %s", err)
	}
	return client
}

// rawTxToTx decodes a raw transaction string into a types.Transaction type
func rawTxToTx(rawTx string) (*types.Transaction, error) {

	tx := new(types.Transaction)
	rawTxBytes, err := hex.DecodeString(rawTx[2:])
	if err != nil {
		ErrorLog.Printf("---> rawTxToTx() failed to decode the transaction (0x..%s) : %s\n", rawTx[len(rawTx)-8:], err)
		return tx, err
	}
	tx.UnmarshalBinary(rawTxBytes)
	return tx, nil
}
