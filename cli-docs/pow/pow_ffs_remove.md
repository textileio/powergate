## pow ffs remove

Removes a Cid from being tracked as an active storage

### Synopsis

Removes a Cid from being tracked as an active storage. The Cid should have both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage

```
pow ffs remove [cid] [flags]
```

### Options

```
  -h, --help           help for remove
  -t, --token string   FFS access token
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
```

### SEE ALSO

* [pow ffs](pow_ffs.md)	 - Provides commands to manage ffs

