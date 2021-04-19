## pow offline car

car generates a CAR file from the data

### Synopsis

Generates a CAR file from the data source. This data-source can be a file/folder path or a Cid.

If a file/folder path is provided, this command will DAGify the data and generate the CAR file.
If a Cid is provided, an extra --ipfs-api flag should be provided to connect to the IPFS node that contains this Cid data.

```
pow offline car [path | cid] [output path] [flags]
```

### Options

```
  -h, --help              help for car
      --ipfs-api string   IPFS HTTP API multiaddress that stores the cid (only for Cid processing instead of file/folder path)
      --quiet             avoid pretty output
      --tmpdir string     path of folder where a temporal blockstore is created for processing data (default "/tmp")
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           user auth token
```

### SEE ALSO

* [pow offline](pow_offline.md)	 - Provides commands to prepare data for Filecoin onbarding

