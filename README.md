
## Run bitcoin_nft_v2 in simnet
> Note: You need to change mining address and you need to set wallet passphrase to be 12345.
### Run btcd in simnet
```
btcd --simnet --rpcuser=youruser --rpcpass=SomeDecentp4ssw0rd --miningaddr=SeZdpbs8WBuPHMZETPWajMeXZt1xzCJNAJ --txindex
```

### Run btcwallet
```
btcwallet --simnet --username=youruser --password=SomeDecentp4ssw0rd
```


### Top up coins in simnet
```
btcctl --simnet --wallet --rpcuser=youruser --rpcpass=SomeDecentp4ssw0rd generate 300
```


## Run bitcoin_nft_v2 in testnet
> Note: You need to set wallet passphrase to be 12345.

### Run btcd in testnet
btcd --testnet -u 4bmeiF7E3ny8cGf8Ok6QJZy/0pk= -P 2oljjSoRFzC5Go7hCGDID6xWi+c=

### Run btcwallet in testnet
btcwallet --testnet --rpcconnect=localhost:8334 -u 4bmeiF7E3ny8cGf8Ok6QJZy/0pk= -P 2oljjSoRFzC5Go7hCGDID6xWi+c=

### Get default address account
btcctl --testnet --wallet --rpcuser="4bmeiF7E3ny8cGf8Ok6QJZy/0pk=" --rpcpass="2oljjSoRFzC5Go7hCGDID6xWi+c=" getaccountaddress default