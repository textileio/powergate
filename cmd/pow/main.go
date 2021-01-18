package main

import (
	"fmt"
	"os"

	"github.com/textileio/powergate/v2/cmd/pow/cmd"
)

func main() {
	if err := cmd.Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
