## pow data get

Get data stored by the user by cid

### Synopsis

Get data stored by the user by cid

```
pow data get [cid] [output file path] [flags]
```

### Options

```
  -f, --folder                Indicates that the retrieved Cid is a folder
  -h, --help                  help for get
      --ipfsrevproxy string   Powergate IPFS reverse proxy DNS address. If port 443, is assumed is a HTTPS endpoint. (default "localhost:6002")
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           user auth token
```

### SEE ALSO

* [pow data](pow_data.md)	 - Provides commands to interact with general data APIs

