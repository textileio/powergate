package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger2"
	logger "github.com/ipfs/go-log/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	mongods "github.com/textileio/go-ds-mongo"
	"github.com/textileio/powergate/buildinfo"
)

var (
	log    = logger.Logger("powcfg")
	config = viper.New()
)

func main() {
	logger.SetAllLoggers(logger.LevelInfo)
	log.Infof("starting powcfg:\n%s", buildinfo.Summary())

	if err := wireFlagsAndEnvs(); err != nil {
		log.Fatalf("wiring flags/envs: %s", err)
	}

	mongoURI := config.GetString("mongouri")
	mongoDB := config.GetString("mongodb")
	mongoCollection := config.GetString("mongocollection")
	badgerrepo := config.GetString("badgerrepo")
	dryrun := config.GetBool("dryrun")
	ds, err := createDatastore(mongoURI, mongoDB, mongoCollection, badgerrepo)
	if err != nil {
		log.Fatalf("opening datastore: %s", err)
	}

	count, err := applyTransform(ds, dryrun, bumpIpfsAddTimeout(480))
	if err != nil {
		log.Fatalf("applying transformation: %s", err)
	}

	log.Infof("Transformation modified %d storage configs", count)
	if dryrun {
		log.Warnf("Dryrun: No changes applied")
	}
}

func wireFlagsAndEnvs() error {
	pflag.String("mongouri", "", "MongoDB URI")
	pflag.String("mongodb", "", "MongoDB database name")
	pflag.String("mongocollection", "", "MongoDB collection name")
	pflag.String("badgerrepo", "", "Badger Repo")
	pflag.Bool("dryrun", false, "Avoid any write to the datastore")
	config.SetEnvPrefix("POWCFG")
	config.AutomaticEnv()
	pflag.Parse()
	if err := config.BindPFlags(pflag.CommandLine); err != nil {
		return fmt.Errorf("binding flags: %s", err)
	}
	return nil
}

func createDatastore(mongoURI, mongoDB, mongoCollection, badgerrepo string) (datastore.TxnDatastore, error) {
	if mongoURI != "" {
		log.Info("Opening Mongo database...")
		mongoCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if mongoDB == "" {
			return nil, fmt.Errorf("mongo database name is empty")
		}
		if mongoCollection == "" {
			return nil, fmt.Errorf("mongo collection name is empty")
		}
		opts := []mongods.Option{
			mongods.WithCollName(mongoCollection),
			mongods.WithOpTimeout(time.Hour),
			mongods.WithTxnTimeout(time.Hour),
		}
		ds, err := mongods.New(mongoCtx, mongoURI, mongoDB, opts...)
		if err != nil {
			return nil, fmt.Errorf("opening mongo datastore: %s", err)
		}
		return ds, nil
	}

	log.Info("Opening badger database...")
	opts := &badger.DefaultOptions
	ds, err := badger.NewDatastore(badgerrepo, opts)
	if err != nil {
		return nil, fmt.Errorf("opening badger datastore: %s", err)
	}
	return ds, nil
}
