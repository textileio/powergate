## pow admin storage-info list

Returns a list of information about all stored cids, filtered by user ids and cids if provided.

### Synopsis

Returns a list of information about all stored cids, filtered by user ids and cids if provided.

```
pow admin storage-info list [flags]
```

### Options

```
      --cids strings       filter results by provided cids.
  -h, --help               help for list
      --user-ids strings   filter results by provided user ids.
```

### Options inherited from parent commands

```
      --admin-token string     admin auth token
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           user auth token
```

### SEE ALSO

* [pow admin storage-info](pow_admin_storage-info.md)	 - Provides admin storage info commands

