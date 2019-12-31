package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	defaultAddr = "127.0.0.1:1234"
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

func ClientConfig(t *testing.T) (string, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(home, ".lotus")
	authToken, err := GetLotusToken(path)
	checkErr(t, err)

	return defaultAddr, authToken
}

func getDefaultAddr() string {
	return defaultAddr
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
