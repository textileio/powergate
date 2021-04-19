## pow offline commp

commP calculates the piece size and cid for a CAR file

### Synopsis

commP calculates the piece-size and PieceCID for a CAR file.

This command calculates the piece-size and piece-cid (CommP) from a CAR file.
This command only makes sense to run for a CAR file, so it does some quick check if the input file *seems* to be well-formated. 
You can use the --skip-car-validation, but usually shouldn't be done unless you know what you're doing (e.g.: benchmarks, or other tests)

```
pow offline commp [path] [flags]
```

### Options

```
  -h, --help                  help for commp
      --json                  avoid pretty output and use json formatting
      --skip-car-validation   skips CAR validation when processing a path
```

### Options inherited from parent commands

```
      --serverAddress string   address of the powergate service api (default "127.0.0.1:5002")
  -t, --token string           user auth token
```

### SEE ALSO

* [pow offline](pow_offline.md)	 - Provides commands to prepare data for Filecoin onbarding

