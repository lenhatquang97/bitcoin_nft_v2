
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


## Run bitcoin_nft_v2 in testnet (to be updated later)