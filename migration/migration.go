package migration

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
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

// Migration runs a vA->v(A+1) migration.
type Migration func(datastore.Txn) error

func New(ds datastore.TxnDatastore, migrations map[int]Migration) *Migrator {
	m := &Migrator{
		ds:         ds,
		migrations: migrations,
	}
	return m
}

func (m *Migrator) Ensure() error {
	currentVersion, emptyDS, err := m.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("getting current version: %s", err)
	}

	targetVersion := m.getTargetVersion()

	// If the database is empty, we can assume Powergate started fresh
	// and requires no migration. Set current version to latest version
	// in migrations script.
	if emptyDS {
		if err := m.bootstrapEmptyDatastore(targetVersion); err != nil {
			return fmt.Errorf("bootstrapping empty database: %s", err)
		}

		return nil
	}

	log.Infof("Current datastore version is %d, target version %d", currentVersion, targetVersion)

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

// getCurrentVersion returns the current database version. If it isn't one
// defined wil be 0. If the datastore is considered to be completely empty,
// it returns true in the second return parameter.
func (m *Migrator) getCurrentVersion() (int, bool, error) {
	isDSEmpty, err := m.isDSEmpty()
	if err != nil {
		return 0, false, fmt.Errorf("detecting if datastore is empty: %s", err)
	}
	if isDSEmpty {
		return 0, true, nil
	}

	var current currentVersion
	buf, err := m.ds.Get(keyCurrentVersion)
	if err == datastore.ErrNotFound {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("getting version from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &current); err != nil {
		return 0, false, fmt.Errorf("unmarshaling current version: %s", err)
	}

	return current.Version, false, nil
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
	if err := txn.Put(keyCurrentVersion, newVerBuf); err != nil {
		return fmt.Errorf("saving new version: %s", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %s", err)
	}

	return nil
}

func (m *Migrator) isDSEmpty() (bool, error) {
	q := query.Query{Limit: 1}
	res, err := m.ds.Query(q)
	if err != nil {
		return false, fmt.Errorf("executing query: %s", err)
	}
	defer func() { _ = res.Close() }()

	all, err := res.Rest()
	if err != nil {
		return false, fmt.Errorf("getting query results: %s", err)
	}

	return len(all) == 0, nil

}

func (m *Migrator) getTargetVersion() int {
	var maxVersion int
	for ver := range m.migrations {
		if ver > maxVersion {
			maxVersion = ver
		}
	}
	return maxVersion
}

func (m *Migrator) bootstrapEmptyDatastore(version int) error {
	newVer := currentVersion{Version: version}
	newVerBuf, err := json.Marshal(newVer)
	if err != nil {
		return fmt.Errorf("marshaling new version: %s", err)
	}
	if err := m.ds.Put(keyCurrentVersion, newVerBuf); err != nil {
		return fmt.Errorf("saving new version: %s", err)
	}

	return nil
}
