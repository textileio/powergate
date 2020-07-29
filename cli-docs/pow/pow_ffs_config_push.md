## pow ffs config push

Add data to FFS via cid

### Synopsis

Add data to FFS via a cid already in IPFS

```
pow ffs config push [cid] [flags]
```

### Options

```
  -c, --conf string    Optional path to a file containing storage config json, falls back to stdin, uses FFS default by default
  -h, --help           help for push
  -o, --override       If set, override any pre-existing storage configuration for the cid
  -t, --token string   FFS access token
  -w, --watch          Watch the progress of the resulting job
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
```

### SEE ALSO

* [pow ffs config](pow_ffs_config.md)	 - Provides commands to manage storage configuration

