package prepare

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// NOTE: Testing only the `prepare` subcommand will indirectly test
// the `car` and `commp` subcommands. This test simply prepares
// some data and compares the final piece-size and piece-cid to a
// known correct value. If anything in the process (DAGification, CARing)
// misbehaves, it will result in a different PieceCID since, at the end of
// the day, PieceCID is a fingerprint of the prepared data.
func TestOfflinePreparation(t *testing.T) {
	testCases := []struct {
		size int
		json string
	}{
		{size: 10000, json: `{"payload_cid":"QmRP8TCp9bthhzLACAao6mh8cfLypqXncdNbzuPtuqFYP7","piece_size":16384,"piece_cid":"baga6ea4seaqjuk4uh5g7cu5znbvrr7wvfsn2l3xj47rbymvi63uiiroya44lkiy"}`},
		{size: 1000, json: `{"payload_cid":"QmQRAjpSLWziADGz8p5ezxTguFNn18yYbSnduKqMrbJ93c","piece_size":2048,"piece_cid":"baga6ea4seaqadahcx4ct54tlbvgkqlhmif7kxxkvxz3yf3vr2e4puhvsxdbrgka"}`},
		{size: 100, json: `{"payload_cid":"QmY6zHswPvyZkAyxwM9uup1DDPb67hZqChv8hnyu4MrFWk","piece_size":256,"piece_cid":"baga6ea4seaqd4hgfl6texpf377k7igx2ga2mfwn3lb4c4kdpaq3g3oao2yftuki"}`},
	}

	for _, test := range testCases {
		test := test
		t.Run(strconv.Itoa(test.size), func(t *testing.T) {
			out, err := run(t, test.size)
			require.NoError(t, err)
			require.Equal(t, test.json, out)
		})
	}
}

func run(t *testing.T, size int) (string, error) {
	f, err := os.CreateTemp("", "preparetest")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %s", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	z := sha256.New()
	_, err = z.Write([]byte("A")) //bytes.Repeat([]byte{byte('A')}, size))
	require.NoError(t, err)
	fmt.Printf("asjkdlas: %x\n", z.Sum(nil))

	if _, err := f.Write(bytes.Repeat([]byte{byte('A')}, size)); err != nil {
		return "", fmt.Errorf("generating temp file: %s", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing file: %s", err)
	}

	stdbuf := bytes.NewBuffer(nil)
	jsonOutput = stdbuf
	Cmd.SetArgs([]string{"prepare", "--json", f.Name()})

	if _, err := Cmd.ExecuteC(); err != nil {
		return "", fmt.Errorf("executing command: %s", err)
	}

	return stdbuf.String(), nil
}
