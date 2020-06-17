## pow ffs push

Add data to FFS via cid

### Synopsis

Add data to FFS via a cid already in IPFS

```
pow ffs push [cid] [flags]
```

### Options

```
  -c, --conf string    Optional path to a file containing cid storage config json, falls back to stdin, uses FFS default by default
  -h, --help           help for push
  -o, --override       If set, override any pre-existing cid storage configuration
  -t, --token string   FFS access token
  -w, --watch          Watch the progress of the resulting job
```

### Options inherited from parent commands

```
      --config string          config file (default is $HOME/.powergate.yaml)
      --serverAddress string   address of the powergate service api (default "/ip4/127.0.0.1/tcp/5002")
```

### SEE ALSO

* [pow ffs](pow_ffs.md)	 - Provides commands to manage ffs

