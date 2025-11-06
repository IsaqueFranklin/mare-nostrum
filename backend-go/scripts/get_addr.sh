#!/bin/bash

COMPILED_PROGRAM="$1"  # Receives first argument passed from Go

# Extract Taproot address (where to send funds to lock them in this contract)
CONTRACT_ADDRESS=$(hal-simplicity simplicity info "$COMPILED_PROGRAM" | jq -r .liquid_testnet_address_unconf)

echo "$CONTRACT_ADDRESS"
#curl "https://liquidtestnet.com/faucet?address=${CONTRACT_ADDRESS}&action=lbtc"
#https://liquidtestnet.com/faucet?address=tex1pjl24xxgz58zwswqhsrj3m3f2867t3k4ews2k3wl0hf3sx8zg94qq830xyq&action=lbtc
#Sent 100000 sats to address tex1pjl24xxgz58zwswqhsrj3m3f2867t3k4ews2k3wl0hf3sx8zg94qq830xyq with transaction e400c8df78bc94b1af67691eea98964198304937be2bd3b3054f719df3204168.