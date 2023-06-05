
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
```
btcd --testnet -u DeW+bgKg011pJHZnaBvgv/lMRks= -P wD9aohGo2f5LwVg7fdj1ntHQcfY=
```
### Run btcwallet in testnet
```
btcwallet --testnet --rpcconnect=localhost:18334 -u DeW+bgKg011pJHZnaBvgv/lMRks= -P wD9aohGo2f5LwVg7fdj1ntHQcfY=
```
### Get default address account
```
btcctl --testnet --wallet --rpcuser="4bmeiF7E3ny8cGf8Ok6QJZy/0pk=" --rpcpass="2oljjSoRFzC5Go7hCGDID6xWi+c=" getaccountaddress default
```

[//]: # (rpcuser=DeW+bgKg011pJHZnaBvgv/lMRks=)

[//]: # (rpcpass=wD9aohGo2f5LwVg7fdj1ntHQcfY=)

```
btcctl --testnet --wallet --rpcuser="DeW+bgKg011pJHZnaBvgv/lMRks=" --rpcpass="wD9aohGo2f5LwVg7fdj1ntHQcfY=" getaccountaddress default
```

### Kiet Seed

```
b3f0670e72bd7e636b94fc64b6d764e32e369ff86ba55269130c22b91ffeb343
```

### Kiet Address Simnet
```
SfF7WYPTkHnjx1jKweNYAoFGnhZH1Q2291
```

btcwallet --simnet --username=youruser --password=SomeDecentp4ssw0rd --create
12345
12345
n
n
OK
