package migration

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	logger "github.com/ipfs/go-log/v2"
)

var (
	log = logger.Logger("migrations")

	keyCurrentVersion = datastore.NewKey("version")
)

type Migrator struct {
	ds         datastore.TxnDatastore
	migrations map[int]Migration
}

type Migration func(datastore.Txn) error

func New(ds datastore.TxnDatastore, migrations map[int]Migration) *Migrator {
	m := &Migrator{
		ds:         ds,
		migrations: migrations,
	}
	return m
}

func (m *Migrator) Ensure(targetVersion int) error {
	currentVersion, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("getting current version: %s", err)
	}
	log.Infof("Current datastore version is %d, target version %d", currentVersion, targetVersion)

	if currentVersion > targetVersion {
		return fmt.Errorf("current datastore version %d greater than target version %d", currentVersion, targetVersion)
	}

	if currentVersion == targetVersion {
		return nil
	}

	for i := currentVersion + 1; i <= targetVersion; i++ {
		log.Infof("Running %d migration...")
		if err := m.run(i); err != nil {
			return fmt.Errorf("running migration %d: %s", i, err)
		}
		log.Infof("Migration %d ran successfully")
	}

	return nil
}

type currentVersion struct {
	Version int
}

func (m *Migrator) getCurrentVersion() (int, error) {
	var current currentVersion
	buf, err := m.ds.Get(keyCurrentVersion)
	if err == datastore.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("getting version from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &current); err != nil {
		return 0, fmt.Errorf("unmarshaling current version: %s", err)
	}
	return current.Version, nil
}

func (m *Migrator) run(version int) error {
	txn, err := m.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("creating txn for migration: %s", err)
	}
	defer txn.Discard()

	script, ok := m.migrations[version]
	if !ok {
		return fmt.Errorf("migration script not found")
	}

	if err := script(txn); err != nil {
		return fmt.Errorf("running migration script: %s", err)
	}

	newVer := currentVersion{Version: version}
	newVerBuf, err := json.Marshal(newVer)
	if err != nil {
		return fmt.Errorf("marshaling new version: %s", err)
	}
	if err := m.ds.Put(keyCurrentVersion, newVerBuf); err != nil {
		return fmt.Errorf("saving new version: %s", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("commiting transaction: %s", err)
	}

	return nil

}
