package tests

import (
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

// LaunchIPFSDocker runs a fresh go-ipfs docker image and returns the resource for
// container metadata.
func LaunchIPFSDocker(t require.TestingT) (*dockertest.Resource, func()) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	ipfsDocker, err := pool.Run("ipfs/go-ipfs", "v0.6.0", []string{"IPFS_PROFILE=test"})
	require.NoError(t, err)

	err = ipfsDocker.Expire(180)
	require.NoError(t, err)

	time.Sleep(time.Second * 3)
	return ipfsDocker, func() {
		err = pool.Purge(ipfsDocker)
		require.NoError(t, err)
	}
}
