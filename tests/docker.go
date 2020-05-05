package tests

import (
	"fmt"
	"time"

	"github.com/ory/dockertest"
)

// LaunchIPFSDocker runs a fresh go-ipfs docker image and returns the resource for
// container metadata.
func LaunchIPFSDocker() (*dockertest.Resource, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("couldn't create docker pool: %s", err))
	}
	ipfsDocker, err := pool.Run("ipfs/go-ipfs", "v0.5.0", []string{"IPFS_PROFILE=test"})
	if err != nil {
		panic(fmt.Sprintf("couldn't run ipfs docker container: %s", err))
	}
	if err := ipfsDocker.Expire(180); err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 3)

	return ipfsDocker, func() {
		if err := pool.Purge(ipfsDocker); err != nil {
			panic(fmt.Sprintf("couldn't purge ipfs from docker pool: %s", err))
		}
	}
}
