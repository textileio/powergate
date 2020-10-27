## pow config remove

Removes a Cid from being tracked as an active storage

### Synopsis

Removes a Cid from being tracked as an active storage. The Cid should have both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage

```
pow config remove [cid] [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           storage profile auth token
```

### SEE ALSO

* [pow config](pow_config.md)	 - Provides commands to interact with cid storage configs

