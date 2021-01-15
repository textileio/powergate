## pow storage-jobs list

List storage jobs according to query flag options.

### Synopsis

List storage jobs according to query flag options.

```
pow storage-jobs list [flags]
```

### Options

```
  -a, --ascending       sort results ascending by time
  -c, --cid string      return results only for the specified cid
  -h, --help            help for list
  -l, --limit uint      limit the number of results returned
  -s, --select string   return only results using the specified selector: all, queued, executing, final (default "all")
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           user auth token
```

### SEE ALSO

* [pow storage-jobs](pow_storage-jobs.md)	 - Provides commands to query for storage jobs in various states

