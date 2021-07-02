## pow offline prepare

prepare generates a CAR file for data

### Synopsis

Prepares a data source generating all needed to execute an offline deal.
The data source can be a file/folder path or a Cid.

If a file/folder path is provided, this command will DAGify the data and generate the CAR file.
If a Cid is provided, an extra --ipfs-api flag should be provided to connect to the IPFS node that contains this Cid data.

This command prepares data in a more efficiently than running car+commp subcommands, since it already starts calculating CommP at the same time that the CAR file is being generated.

By default prints to stdout the generated CAR file. You can provide a second argument to
specify the output file path, or simply pipe the stdout result.

The piece-size and piece-cid are printed to stderr. For scripting usage, its recommended to use the --json flag.

```
pow offline prepare [cid | path] [output CAR file path] [flags]
```

### Options

```
      --aggregate         aggregates a folder of files
  -h, --help              help for prepare
      --ipfs-api string   IPFS HTTP API multiaddress that stores the cid (only for Cid processing instead of file/folder path)
      --json              avoid pretty output and use json formatting
      --tmpdir string     path of folder where a temporal blockstore is created for processing data (default "/tmp")
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           user auth token
```

### SEE ALSO

* [pow offline](pow_offline.md)	 - Provides commands to prepare data for Filecoin onbarding

