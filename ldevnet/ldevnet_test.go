package ldevnet

import (
	"context"
	"fmt"
	"testing"
)

func TestSetup(t *testing.T) {
	num := 1
	ld, err := New(t, num)
	checkErr(t, err)
	defer ld.Close()

	maddrs, err := ld.Client.StateListMiners(context.Background(), nil)
	checkErr(t, err)
	if len(maddrs) != num {
		t.Fatalf("# of miners should be %d", num)
	}
	fmt.Println(maddrs)

	g, err := ld.Client.ChainGetGenesis(context.Background())
	checkErr(t, err)
	h, err := ld.Client.ChainHead(context.Background())
	checkErr(t, err)
	path, err := ld.Client.ChainGetPath(context.Background(), g.Key(), h.Key())
	checkErr(t, err)
	for _, l := range path {
		fmt.Printf("height: %d, \n", l.Val.Height())
		for _, b := range l.Val.Blocks() {
			fmt.Printf("\tminer: %s", b.Miner)
		}
		fmt.Println()
	}
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
