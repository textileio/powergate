## pow

A client for storage and retreival of powergate data

### Synopsis

A client for storage and retreival of powergate data

```
pow [flags]
```

### Options

```
  -h, --help                   help for pow
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           storage profile auth token
  -v, --version                display version information for pow and the connected server
```

### SEE ALSO

* [pow admin](pow_admin.md)	 - Provides admin commands
* [pow apply-config](pow_apply-config.md)	 - Apply the default or provided storage config to the specified cid
* [pow deals](pow_deals.md)	 - Provides commands to view Filecoin deal information
* [pow default-config](pow_default-config.md)	 - Returns the default storage config
* [pow get](pow_get.md)	 - Get data by cid from the storage profile
* [pow id](pow_id.md)	 - Returns the storage profile id
* [pow info](pow_info.md)	 - Get information about the current storate state of a cid
* [pow log](pow_log.md)	 - Display logs for specified cid
* [pow remove](pow_remove.md)	 - Removes a Cid from being tracked as an active storage
* [pow replace](pow_replace.md)	 - Applies a StorageConfig for c2 equal to that of c1, and removes c1
* [pow set-default-config](pow_set-default-config.md)	 - Sets the default storage config from stdin or a file
* [pow stage](pow_stage.md)	 - Temporarily stage data in the Hot layer in preparation for applying a cid storage config
* [pow storage-jobs](pow_storage-jobs.md)	 - Provides commands to query for storage jobs in various states
* [pow version](pow_version.md)	 - Display version information for pow and the connected server
* [pow wallet](pow_wallet.md)	 - Provides commands about filecoin wallets

