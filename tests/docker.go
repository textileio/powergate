package tests

import (
	"fmt"
	"time"

	"github.com/ory/dockertest"
)

func LaunchDocker() (*dockertest.Resource, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("couldn't create docker pool: %s", err))
	}
	ipfsDocker, err := pool.Run("ipfs/go-ipfs", "latest", []string{"IPFS_PROFILE=test"})
	if err != nil {
		panic(fmt.Sprintf("couldn't run ipfs docker container: %s", err))
	}
	ipfsDocker.Expire(180)

	time.Sleep(time.Second * 3)

	return ipfsDocker, func() {
		if err := pool.Purge(ipfsDocker); err != nil {
			panic(fmt.Sprintf("couldn't purge ipfs from docker pool: %s", err))
		}
	}
}
