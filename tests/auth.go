package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	ma "github.com/multiformats/go-multiaddr"
)

const (
	lotusHost = "127.0.0.1"
	lotusPort = 1234
)

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

func ClientConfigMA() (ma.Multiaddr, string) {
	addr := fmt.Sprintf("/ip4/%v/tcp/%v", lotusHost, lotusPort)
	multi, err := ma.NewMultiaddr(addr)
	if err != nil {
		panic(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(home, ".lotus")
	authToken, err := GetLotusToken(path)
	if err != nil {
		panic(err)
	}

	return multi, authToken
}
