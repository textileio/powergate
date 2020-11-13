## pow deals storage

List storage deal records for the user

### Synopsis

List storage deal records for the user

```
pow deals storage [flags]
```

### Options

```
      --addrs strings     limit the records to deals initiated from  the specified wallet addresses, treated as and AND operation if --cids is also provided
  -a, --ascending         sort records ascending, default is sort descending
      --cids strings      limit the records to deals for the specified data cids, treated as and AND operation if --addrs is also provided
  -h, --help              help for storage
  -f, --include-final     include final deals
  -p, --include-pending   include pending deals
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           user auth token
```

### SEE ALSO

* [pow deals](pow_deals.md)	 - Provides commands to view Filecoin deal information

