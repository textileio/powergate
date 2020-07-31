## pow ffs storage

List storage deal records for an FFS instance

### Synopsis

List storage deal records for an FFS instance

```
pow ffs storage [flags]
```

### Options

```
      --addrs strings     limit the records to deals initiated from  the specified wallet addresses, treated as and AND operation if --cids is also provided
  -a, --ascending         sort records ascending, default is sort descending
      --cids strings      limit the records to deals for the specified data cids, treated as and AND operation if --addrs is also provided
  -h, --help              help for storage
  -f, --include-final     include final deals
  -p, --include-pending   include pending deals
  -t, --token string      token of the request
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
```

### SEE ALSO

* [pow ffs](pow_ffs.md)	 - Provides commands to manage ffs

