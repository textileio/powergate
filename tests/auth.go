package tests

import (
	"os"
	"os/exec"
	"path/filepath"
)

const (
	DaemonAddr  = "127.0.0.1:1234"
	LotusFolder = "~/.lotus"
)

func GetLotusToken() (string, error) {
	if _, err := os.Stat(filepath.Join(LotusFolder, "token")); err != nil {
		if os.IsNotExist(err) {
			return createAdminToken()
		}
		return "", err
	}
	cmd := exec.Command("cat", "~/.lotus/token")
	if err := cmd.Run(); err != nil {
		return "", err
	}
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
