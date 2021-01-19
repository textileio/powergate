# migrtest

`migrtest` is a tool for testing migrations on real data. This tools allows to run migrations on databases in a safe way since they're run in a copy of the target database. This tool is useful while developing new migrations, or to Powergate operations which might want to test a new release which has migrations on real data to see if something bad would happen in the real update.

## Usage

Run `migrtest -h` to know about the available flags.


### Target a remote `go-datastore` and run locally
```bash
migrtool --origin-remote "$MONGO_URI;$MONGO_DBNAME"
```
This command will:
- Create a folder `run_<TIMESTAMP>` which will have the run assets.
- Make a copy of the remote `go-datastore` to `run_<TIMESTAMP>/badger-migrated`.
- Run all detected/needed migrations in the copied datastore.
- Print to stdout Powergate output while doing the migrations.
- `run_<TIMESTAMP>/badger-migrated` will be the migrated `go-datastore` that might be useful for inspection.

###  Target a remote `go-datastore` and run remotely
```bash
migrtool --origin-remote "$MONGO_URI;$MONGO_DBNAME" --run-remote
```
This command will:
- Create a folder `run_<TIMESTAMP>` which will have the run assets.
- Make a copy of the remote `go-datastore` to the same MongoDB cluster, in a fixed `tmp_powergate_migrtest` database, with a collection name `run_<TIMESTAMP>`.
- Run all detected/needed migrations in the copied datastore.
- Print to stdout Powergate output while doing the migrations.
- The remote collection is kept that might be useful for inspection.

This setup is useful to double-check that running using `go-ds-mongo` client works correctly. This may be slower since it's targeting a remote datastore, but as close as simulating the real migration as possible.

### Target a remote `go-datastore` and skip the migration
This setup might be useful to just have a local copy of the remote `go-datastore`.
```bash
migrtool --origin-remote "$MONGO_URI;$MONGO_DBNAME" --skip-migrations
```
This command will:
- Create a folder `run_<TIMESTAMP>` which will have the run assets.
- Make a copy of the remote `go-datastore` to the same MongoDB cluster, in a fixed `tmp_powergate_migrtest` database, with a collection name `run_<TIMESTAMP>`.


