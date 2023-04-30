package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ethClientURL             = "http://127.0.0.1:8545"            // Etherum client endpoint ("https://mainnet.infura.io/v3")
	infuraAPIKey             = "d8aa39d0112e4dcdb25e6d9deba4788f" // API-Key needed if using mainnet.infura.io/v3
	sendTxPeriod             = time.Duration(4)                   // How frequently to attempt sending local transactions to the Ethereum client
	defaultBaseGasFee uint64 = 0x7fffffffffffffff                 // This is the default baseGasFee if we can't connect to and query the Ethreum client
	logFilename       string = "txProxyLOGS.txt"                  // File to store the logs
	WarningLog        *log.Logger
	InfoLog           *log.Logger
	ErrorLog          *log.Logger
)

func init() {
	// Set up logging system
	file, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	InfoLog = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {

	eClient := getEthClient(ethClientURL)

	// Try and forward local transactions to the ethereum client every 'sendTxPeriod' seconds.
	go sendTxsLoop(eClient)

	// remote is the ethereum client we relay most requests to.
	remote, err := url.Parse(ethClientURL)
	if err != nil {
		panic(err)
	}

	// This handler intercepts certain requests (eg eth_sendRawTransaction) and relays all others to the ethereum client.
	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {

			// Get the request method so we know when to handle it locally (eg eth_sendRawTransaction).
			buf, _ := ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf)) // r.Body is consumable so we need to reset it.
			rBody := ioutil.NopCloser(bytes.NewBuffer(buf))
			req := &Request{}
			json.NewDecoder(rBody).Decode(req)

			switch req.Method {
			case "eth_sendRawTransaction":
				SendRawTransactionHandler(w, r)
			case "eth_cancelTransaction":
				CancelTransactionHandler(w, r)
			default:
				// We all other route requests to the ethereumClient, so the requesting client can get things like the chain_Id, etc.
				InfoLog.Printf("Routing request to ethereum client: %s.\n", req.Method)

				// Modify request to work with infura mainnet endpoint
				if strings.Contains(ethClientURL, "mainnet.infura.io") {
					r.Host = "mainnet.infura.io"
					r.URL.Path = infuraAPIKey
				}

				p.ServeHTTP(w, r)

			}

		}
	}

	// Set up local server to listen to and serve incoming requests
	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc("/", handler(proxy))

	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}
