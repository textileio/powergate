# Powergate bench

Powergate `powbench` is an utility tool that allows to run benchmark scenarios against a properly configured Powergate server.

To build and install `powbench`, from the root of the Powergate repo, run:

```bash
> make build-powbench
> powbench -h
Usage of ./powbench:
      --maxParallel int    Max parallel file storage (default 1)
      --minerAddr string   Miner address to force Powergate to select for making deals (default "t01000")
      --pgAddr string      Powergate server multiaddress (default "/ip4/127.0.0.1/tcp/5002")
      --randSeed int       Random seed used to generate random samples data (default 42)
      --sampleSize int     Size of randomly generated files in bytes (default 1024)
      --totalSamples int   Total samples to run (default 3)
```

The targeted Powergate server should have enabled the auto-funding of newly created FFS instances wallet addresses.
This means:
- The Lotus node owned by Powergate must have a wallet address; the _master address_. 
- _master address_ should have funds.
- This address will be used to automatically send funds to the new FFS instance that will run the benchmark.
- Powergate should be started with two flags:
  - `--lotusmasteraddr`: with the above mentioned address.
  - `--walletinitialfund`: an reasonable amount of _attoFIL_ that will be transferred from the master address to the created FFS instance. It should be enough to fund _all_ deals that will be ran in the scenario.
  - _Note: env variables can be used instead of command line flags, i.e: `POWD_LOTUSMASTERADDR` and `POWD_WALLETINITIALFUND` respectively._

