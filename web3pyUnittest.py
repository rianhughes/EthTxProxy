import web3 as w3
import requests 
import json 
import unittest
from time import sleep
import socket

def rpc_call(method, args):
        url = "http://"+txProxyServerIP+":"+str(txProxyServerPort)
        headers = {'content-type': 'application/json'}
        payload = {
            "method": method,
            "params": [args],
            "jsonrpc": "2.0",
            "id": 1,
        }
        response = requests.post(url, data=json.dumps(payload), headers=headers)
        return response
    
def create_tx_raw_tx(gasPrice):
        testAcnt = w3.eth.Account.create()
        tx = w3.eth.Account.sign_transaction(dict(
            nonce=0,
            maxFeePerGas=gasPrice,
            maxPriorityFeePerGas=2000000000,
            gas=2000000000,
            to='0xeDa9907446e305aC327DC6891263251144f8AA1f',
            value=12345,
            data=b'',
            chainId=1,
                ),
            testAcnt.privateKey,
            )
        return tx, tx.rawTransaction.hex()

class Test(unittest.TestCase):      
    
    # Test if the sendRawTransaction method works with correct input
    def test_0_sendRawTransaction(self):
        tx, rawTxHex = create_tx_raw_tx(1000000000000)
        response = rpc_call('eth_sendRawTransaction', rawTxHex).json()
        self.assertEqual(response['result'], tx.hash.hex())
    
    
    # Test if the sendRawTransaction method works with incorrect raw-transaction
    def test_1_sendRawTransaction(self):
        rawTxHex = "0xNonsense"
        response = rpc_call('eth_sendRawTransaction', rawTxHex).json()
        self.assertEqual(response['result'],"Failed to decode transaction in eth_sendRawTransaction.")
    
    # Test that a user can cancel a transaction that does exist
    def test_2_cancelTransaction(self):
        tx, rawTxHex = create_tx_raw_tx(1000000000000)
        response = rpc_call('eth_sendRawTransaction', rawTxHex).json()
        self.assertEqual(response['result'], tx.hash.hex())
        
        # Cancel the transaction that exists
        response = rpc_call('eth_cancelTransaction', rawTxHex).json()
        self.assertEqual(response['result'], 'Transaction cancelled.')
        
    # Test if the CancelTransaction method returns the correct error 
    # when attempting to cancel a transaction that doesn't exist.
    def test_3_cancelTransaction_error(self):
        _, rawTxHex = create_tx_raw_tx(1000000000000)
        response = rpc_call('eth_cancelTransaction', rawTxHex).json()
        self.assertEqual(response['error'], "Failed to cancel transaction: it does not exist in local TxPool.")
        
        
    # This tests that the server works when given an incorrect method
    def test_4_server_method_error(self):
        rawTxHex = "0x"
        response = rpc_call('0xMethodError', rawTxHex)
        try:
            # Ganache gives a status_code=200, and error message
            response = response.json()
            self.assertIsNotNone(response['error'])    
        except:
            # Infura and no-ethereum-client give an error code
            self.assertNotEqual(response.status_code,200)
        
        
        
    # # Test that the server deletes the transaction after sending it to the network
    # # Note: This test requires waiting long enough (<2*sendTxPeriod) for the server to send the tx
    def test_5_cancelTransaction_post_send(self):
        # Above gasLimit so it always gets sent on for this unittest
        tx, rawTxHex = create_tx_raw_tx(9223372036854775808)  
        response = rpc_call('eth_sendRawTransaction', rawTxHex).json()
        self.assertEqual(response['result'], tx.hash.hex())
        
        # Wait for the server to send the tranasction on to the network, then try and cancel it
        sleep(9)
        response = rpc_call('eth_cancelTransaction', rawTxHex).json()
        self.assertEqual(response['error'], "Failed to cancel transaction: it does not exist in local TxPool.")
        
    


txProxyServerIP = "127.0.0.1"
txProxyServerPort = 8000

if __name__ == '__main__':                   
        
    try:
        # Test if the transaction-proxy server is running before executing the unittest
        s = socket.socket()
        s.connect((txProxyServerIP, txProxyServerPort))      
        
        # Send a transaction to the transaction-proxy server
        x, rawTxHex = create_tx_raw_tx(0)
        response = rpc_call('eth_sendRawTransaction', rawTxHex).json()
        print("transaction-proxy server response to eth_sendRawTransaction:\n",response)
        
        # Perform unittests
        unittest.main(warnings='ignore')
        
    except socket.error as exc:
        print("Failed to make the RPC-call. Are you sure the transaction-proxy server is running? Execute 'go run .' an try again. (%s)"%exc)
