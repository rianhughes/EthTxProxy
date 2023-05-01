
# Transaction-proxy overview

This repo solves the takehome task as described in 'Transaction_Lifecycle_Tech_Test.pdf'. 

The main golang program sets up a 'transaction-proxy' relays requests to an ethereum client, except for 'eth_sendRawTransaction' and 'eth_cancelTransaction' methods (ie it's effecitvely an intercepting reverse proxy server). These methods add/remove any incoming transactions to/from a local mempool (TxPool). Every 'sendTxPeriod' seconds we query the ethereum client for the latest baseFee ( defaulting to 'defaultBaseGasFee' if we fail to connect). If any of the local transactions has a higher gasFee (maxBaseFee), we send them to the ethereum client and remove them from the local mempool/TxPool. Otherwise we wait until either the client cancels their transaction, or the baseGasFee drops below the their specified gas price to send it to the ethereum client. The web3-library based unittests are implemented in the web3Unittest.py python file. The 'transaction-proxy' server needs to be running for the python script to work. The 'transaction-proxy' server also logs its actions in txProxyLOGS.txt (logFilename).

If connecting to an ethereum client, I recommend using something like Ganahce v7.8.0 (eg see the VSCode plug in), although this code also works with a mainnet.infura.io endpoint.

## Metamask integration

This code also works with Metamask when connected to either Ganache or an infura endpoint. To connect to Metamask, open up the Metamask settings, and add a new Network with the following settings:
 ```
    Network name: TransactionProxy1337  
    New RPC URL: http://127.0.0.1:8000    
    Chain ID: 1337 (assuming you're using Ganache, or 1 if using mainnet and infura)
    Currency Symbol: ETH  
```

Note the Chain ID will depend on the ethereum client you connect to (I'd recommend adding a different network if you plan to use both Ganache and an infura mainnet endpoint because they have different Chain IDs). Once connected to metamask you will be able to send transactions from your wallet etc.

The main problem with using a metamask wallet to interact with this 'transaction-proxy' server is that there is no way to call the eth_cancelTransaction function. Currently, Metamask does provide an option to 'Cancel' pending transactions, however, this just calls eth_sendRawTransaction with a higher gas price and with the transaction data zero'd out. Ideally, selecting the 'Cancel' option on the metamask wallet would call this servers eth_cancelTransaction function instead. A work-around that wouldn't require any changes on Metamasks side would be to detect if users have sent a new transaction using the 'Cancel' option on metamask. For example, we could check if the from and nonce fields of any incoming transactions match any existing transactions (eg by using a new map[from+nonce]rawTxHash), and cancel any matching transactions if the 'Cancel'd transactions data field has been zero'd out. 



## Requirements to use this code
Programs: Golang 1.19, Python 3.6.9, Ganache v7.8.0.   
Dependencies: The golang and python dependencies are specified in go.mod and requirements.txt.  
Ethereum-client endpoint: This program doesn't need an ethereum endpoint to operate, but the baseFee will be static. 

### Step 0:
Install the golang and python dependencies by executing:   
```
go mod tidy
pip3 install -r requirements.txt
```

### Step 1:

Setup up an ethereum client endpoint. For example, start a development chain (in a new terminal) on ganache using port 8545:
```
ganache -p 8545
```

### Step 2:

Start the transaction-proxy server (in a new terminal):   
```
go run .
```
### Step 3:

Use the web3Unittest.py file to perform client-based unit-tests of the transaction-proxy server (in a new terminal):  
```
python3 web3unittest.py
````


## Docker 

A dockerfile has been included to build a stable image. Here are some useful docker commands to get it up and running, and to interact with the image, etc:
```
docker build  --tag txproxy .
docker run  -p 8000:8000 -it txproxy /bin/bash
docker ps # To get the Container_ID
docker exec -it <Container_ID> bash
```
