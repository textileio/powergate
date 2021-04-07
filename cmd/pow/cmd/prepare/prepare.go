package prepare

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ipfs/go-car"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-cidutil"
	chunker "github.com/ipfs/go-ipfs-chunker"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/ipfs/go-unixfs/importer/balanced"
	ihelper "github.com/ipfs/go-unixfs/importer/helpers"
	mh "github.com/multiformats/go-multihash"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
)

func init() {
	Cmd.AddCommand(genCar)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "prepare",
	Short: "Provides commands to prepare data for Filecoin onbarding",
	Long:  `Provides commands to prepare data for Filecoin onbarding`,
}

var genCar = &cobra.Command{
	Use:   "gen-car [cid | file | folder] [output file path]",
	Short: "gen-car generates a CAR file for data",
	Long:  `gen-car generates a CAR file for data`,
	Args:  cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		//TODO parse different args
		f, err := os.Open(args[0])
		c.CheckErr(err)

		z := files.NewReaderFile(f)

		prefix, err := merkledag.PrefixForCidVersion(0)
		c.CheckErr(err)
		prefix.MhType = uint64(mh.SHA2_256)

		dserv := dstest.Mock()
		params := ihelper.DagBuilderParams{
			Maxlinks:  174,
			RawLeaves: false,
			CidBuilder: cidutil.InlineBuilder{
				Builder: prefix,
				Limit:   32,
			},
			Dagserv: dserv,
		}

		db, err := params.New(chunker.NewSizeSplitter(z, 256*1024))
		c.CheckErr(err)
		nd, err := balanced.Layout(db)
		c.CheckErr(err)
		fmt.Printf("CID: %s\n", nd.Cid())
		rr, rw := io.Pipe()
		go func() {
			defer rw.Close()
			if err := car.WriteCar(context.Background(), dserv, []cid.Cid{nd.Cid()}, rw); err != nil {
				panic(err) // TODO: fix this
			}
		}()

		_, err = io.Copy(io.Discard, rr)
		c.CheckErr(err)
	},
}
