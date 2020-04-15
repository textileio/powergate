package tests

import "testing"

// CheckErr is a helper for checking an error and failing a test
func CheckErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
