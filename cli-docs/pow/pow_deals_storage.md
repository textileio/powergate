## pow deals storage

List storage deal records

### Synopsis

List storage deal records

```
pow deals storage [flags]
```

### Options

```
      --addrs strings     limit the records to deals initiated from  the specified wallet addresses, treated as and AND operation if --cids is also provided
  -a, --ascending         sort records ascending, default is sort descending
      --cids strings      limit the records to deals for the specified data cids, treated as and AND operation if --addrs is also provided
  -h, --help              help for storage
  -f, --include-final     include final deals (default true)
  -p, --include-pending   include pending deals
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "/ip4/127.0.0.1/tcp/5002")
```

### SEE ALSO

* [pow deals](pow_deals.md)	 - Provides commands to manage storage deals

