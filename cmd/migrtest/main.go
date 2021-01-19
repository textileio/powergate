package main

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/ipfs/go-datastore"
	logger "github.com/ipfs/go-log/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/textileio/dsutils/clone"
	"github.com/textileio/powergate/v2/api/server"
	"github.com/textileio/powergate/v2/migration"
)

const (
	tmpRemoteMongoDB = "tmp_pow_migrtest"
)

var (
	log = logger.Logger("migrtest")
)

func main() {
	config := viper.New()
	logger.SetAllLoggers(logger.LevelInfo)

	if err := wireFlagsAndEnvs(config); err != nil {
		log.Fatalf("wiring flags/envs: %s", err)
	}

	var (
		err      error
		copiedDS datastore.TxnDatastore

		runName        = fmt.Sprintf("run_%s", time.Now().Format("2006-01-02-15:04:05"))
		originRemote   = config.GetString("origin-remote")
		runRemote      = config.GetBool("run-remote")
		verbose        = config.GetBool("verbose")
		skipMigrations = config.GetBool("skip-migrations")
	)
	defer log.Infof("Inspect folder %s to see migration assets.", runName)

	switch {
	case len(originRemote) > 0:
		mongoSplits := strings.Split(originRemote, ";")
		mongoURI := mongoSplits[0]
		mongoDatabase := mongoSplits[1]

		if runRemote {
			copiedDS, err = clone.CloneFromRemoteToRemote(mongoURI, mongoDatabase, "kvstore", mongoURI, tmpRemoteMongoDB, runName, 5000, verbose)
			if err != nil {
				log.Fatalf("cloning remote to local: %s", err)
			}
		} else {
			badgerPath := path.Join(runName, "badger-migrated")
			copiedDS, err = clone.CloneFromRemoteToLocal(mongoURI, mongoDatabase, "kvstore", badgerPath, 1000, verbose)
			if err != nil {
				log.Fatalf("cloning remote to local: %s", err)
			}
		}

	default:
		log.Fatalf("unsupported flag combination")
	}

	if skipMigrations {
		log.Warnf("Skipping migrations, your detination datastore is a clean copy of the origin")
		return
	}

	m := migration.New(copiedDS, server.Migrations)
	if err = m.Ensure(); err != nil {
		log.Fatalf("running migrations: %s", err)
	}
}

func wireFlagsAndEnvs(config *viper.Viper) error {
	pflag.String("origin-remote", "", "(MongoDBURI;database-name)")
	pflag.Bool("run-remote", false, "Copies the origin to the remote to test the migration remotely. If not set, runs locally.")
	pflag.Bool("verbose", false, "Verbose output")
	pflag.Bool("skip-migrations", false, "Skips running migrations, the destination datastore would remain a clean copy of the origin")
	pflag.Parse()

	config.SetEnvPrefix("MIGRTEST")
	config.AutomaticEnv()
	if err := config.BindPFlags(pflag.CommandLine); err != nil {
		return fmt.Errorf("binding flags: %s", err)
	}

	return nil
}
