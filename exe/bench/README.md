# Powergaet bench

Powergate `bench` is an utility tool that allows to run benchmark scenarios against a properly configured Powergate server.

```bash
> go build -o powbench .
> ./powbench -h
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
- The Lotus node owned by Powergate should have a defined default wallet address.
- This wallet address should have enough funds.
- Powergate should be started with two flags:
  - `--lotusmasteraddr`: with the above mentioned default Lotus address.
  - `--walletinitialfund`: an reasonable amount of _attoFIL_ that will be transferred from the master address to newly created addresses of FFS instances. It should be enought to fund _all_ deals that will be ran in the scenario.
  - _Note: env variables can be used instead of command line flags, i.e: `TEXPOWERGATE_LOTUSMASTERADDR` and `TEXPOWERGATE_WALLETINITIALFUND` respectively._

