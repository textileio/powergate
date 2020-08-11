package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

const (
	lotusHost = "127.0.0.1"
	lotusPort = 1234
)

// GetLotusToken returns the lotus token from a Lotus repo path.
func GetLotusToken(lotusFolderPath string) (string, error) {
	tokenFullPath := filepath.Join(lotusFolderPath, "token")
	if _, err := os.Stat(tokenFullPath); err != nil {
		if os.IsNotExist(err) {
			return createAdminToken()
		}
		return "", err
	}
	cmd := exec.Command("cat", tokenFullPath)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func createAdminToken() (string, error) {
	cmd := exec.Command("lotus", "auth", "create-token", "--perm", "admin")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), err
}

// ClientConfigMA returns the prepared multiaddress and Lotus token,
// to connect to a Lotus node.
func ClientConfigMA(t *testing.T) (ma.Multiaddr, string) {
	addr := fmt.Sprintf("/ip4/%v/tcp/%v", lotusHost, lotusPort)
	multi, err := ma.NewMultiaddr(addr)
	require.NoError(t, err)
	authToken, ok := os.LookupEnv("TEXTILE_LOTUS_TOKEN")
	if !ok {
		home, err := os.UserHomeDir()
		require.NoError(t, err)
		path := filepath.Join(home, ".lotus")
		authToken, err = GetLotusToken(path)
		require.NoError(t, err)
	}

	return multi, authToken
}
