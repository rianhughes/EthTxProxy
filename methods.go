package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// TxPool stores the rawTransaction types in-memory using a map.
// The keys are the rawTransaction string, and the values are the rawTransactions gasFee.
type TxPool struct {
	txmap map[string]uint64
	mu    sync.Mutex
}

// txPool is the local in-memory transaction database
var txPool = TxPool{txmap: make(map[string]uint64)}

type Request struct {
	Id      int      `json:"id"`
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

type Response struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Error   string `json:"error"`
}

// SendRawTransaction takes a rawTransaciton and adds it to the local in-memory txPool database.
func SendRawTransactionHandler(w http.ResponseWriter, r *http.Request) {

	var resp Response

	// Try and decode the transaction
	req := &Request{}
	json.NewDecoder(r.Body).Decode(req)
	tx, err := rawTxToTx(req.Params[0])

	if err != nil {
		InfoLog.Printf("---> Not adding transaction (0x..%s) to local Txpool: Failed to decode.", req.Params[0][len(req.Params[0])-8:])
		resp = Response{Id: req.Id, Jsonrpc: req.Jsonrpc, Result: "Failed to decode transaction in eth_sendRawTransaction.", Error: err.Error()}
	} else {
		txPool.mu.Lock()
		defer txPool.mu.Unlock()
		InfoLog.Printf("---> Adding transaction (0x..%s) to local Txpool.\n", req.Params[0][len(req.Params[0])-8:])
		txPool.txmap[req.Params[0]] = tx.GasPrice().Uint64()
		resp = Response{Result: tx.Hash().String(), Id: req.Id, Jsonrpc: req.Jsonrpc}
	}

	json.NewEncoder(w).Encode(resp)
}

// Eth_cancelTransaction takes a rawTransaction and removes it from the local in-memory txPool database.
func CancelTransactionHandler(w http.ResponseWriter, r *http.Request) {

	var resp Response

	// Try and decode the transaction
	req := &Request{}
	json.NewDecoder(r.Body).Decode(req)

	txPool.mu.Lock()
	defer txPool.mu.Unlock()

	if _, ok := txPool.txmap[req.Params[0]]; ok {
		InfoLog.Printf("---> Transaction (0x..%s) cancelled.\n", req.Params[0][len(req.Params[0])-8:])
		delete(txPool.txmap, req.Params[0])
		resp = Response{Result: "Transaction cancelled.", Id: req.Id, Jsonrpc: req.Jsonrpc}
	} else {
		InfoLog.Printf("---> Transaction (0x..%s) not cancelled: it does not exist in local TxPool.\n", req.Params[0][len(req.Params[0])-8:])
		resp = Response{Id: req.Id, Jsonrpc: req.Jsonrpc, Error: "Failed to cancel transaction: it does not exist in local TxPool."}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}
