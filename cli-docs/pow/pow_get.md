## pow get

Get data by cid from the storage profile

### Synopsis

Get data by cid from the storage profile

```
pow get [cid] [output file path] [flags]
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
  -t, --token string           storage profile auth token
```

### SEE ALSO

* [pow](pow.md)	 - A client for storage and retreival of powergate data

