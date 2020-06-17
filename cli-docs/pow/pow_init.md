## pow init

Initializes a config file with the provided values or defaults

### Synopsis

Initializes a config file with the provided values or defaults

```
pow init [flags]
```

### Options

```
  -a, --address string    Sets the default wallet address for the config file (default "<a-wallet-address>")
  -d, --duration int      Sets the default duration for the config file (default 1000000)
  -h, --help              help for init
  -m, --maxPrice int      Sets the default maxPrice for the config file (default 100)
  -p, --pieceSize int     Sets the default pieceSize for the config file (default 1024)
  -w, --wallets strings   Sets the wallet address list for the config file
```

### Options inherited from parent commands

```
      --config string          config file (default is $HOME/.powergate.yaml)
      --serverAddress string   address of the powergate service api (default "/ip4/127.0.0.1/tcp/5002")
```

### SEE ALSO

* [pow](pow.md)	 - A client for storage and retreival of powergate data

