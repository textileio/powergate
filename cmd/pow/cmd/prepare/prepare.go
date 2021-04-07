package prepare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb/v3"
	bsrv "github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-car"
	"github.com/ipfs/go-cid"
	badger "github.com/ipfs/go-ds-badger"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core/coreunix"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
)

func init() {
	Cmd.AddCommand(genCar)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "offline",
	Short: "Provides commands to prepare data for Filecoin onbarding",
	Long:  `Provides commands to prepare data for Filecoin onbarding`,
}

var genCar = &cobra.Command{
	Use:   "prepare [cid | path] [output file path]",
	Short: "prepare generates a CAR file for data",
	Long:  `prepare generates a CAR file for data`,
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// TTODO: Accept Cids, ask for ipfs node api
		// TTODO: if output path not provided, spit to stdout
		// TTODO: prety mode
		// TTODO: quiet mode, no events
		// TTODO: tests
		// TTODO: define final command name and help text
		c.FmtOutput = os.Stderr

		dagService, cls, err := createTmpDAGService()
		if err != nil {
			c.Fatal(fmt.Errorf("creating temporal dag-service: %s", err))
		}
		defer cls()

		ctx := context.Background()
		path := args[0]

		c.Message("Creating data DAG...")
		dataCid, err := dagify(ctx, dagService, path)
		if err != nil {
			c.Fatal(fmt.Errorf("creating dag for data: %s", err))
		}
		c.Message("DAG created.")

		outputFile := os.Stdout
		if len(args) > 1 {
			outputFile, err = os.OpenFile(args[1], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
			if err != nil {
				c.Fatal(fmt.Errorf("creating output file: %s", err))
			}
			defer func() {
				if err := outputFile.Close(); err != nil {
					c.Fatal(fmt.Errorf("closing output file: %s", err))
				}
			}()
		}

		pr, pw := io.Pipe()
		var writeCarErr error
		go func() {
			defer pw.Close()
			start := time.Now()
			c.Message("Creating CAR file...")
			if err := car.WriteCar(ctx, dagService, []cid.Cid{dataCid}, pw); err != nil {
				writeCarErr = err
				return
			}
			c.Message("CAR file created in %.02f seconds.", time.Since(start).Seconds())
		}()
		if _, err := io.Copy(outputFile, pr); err != nil {
			c.Fatal(fmt.Errorf("writing CAR file to output: %s", err))
		}
		if writeCarErr != nil {
			c.Fatal(fmt.Errorf("generating CAR file: %s", err))
		}

	},
}

type CloseFunc func() error

func createTmpDAGService() (ipld.DAGService, CloseFunc, error) {
	badgerFolder, err := ioutil.TempDir("", "powprepare-*")
	if err != nil {
		return nil, nil, fmt.Errorf("creating temporary badger folder: %s", err)
	}

	ds, err := badger.NewDatastore(badgerFolder, &badger.DefaultOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("creating temporal badger datastore: %s", err)
	}
	bstore := blockstore.NewBlockstore(ds)

	return dag.NewDAGService(bsrv.New(bstore, offline.Exchange(bstore))),
		func() error {
			if err := ds.Close(); err != nil {
				return fmt.Errorf("closing datastore: %s", err)
			}
			os.RemoveAll(badgerFolder)

			return nil
		}, nil
}

func dagify(ctx context.Context, dagService ipld.DAGService, path string) (cid.Cid, error) {
	events := make(chan interface{}, 10)

	fileAdder, err := coreunix.NewAdder(ctx, nil, nil, dagService)
	if err != nil {
		return cid.Undef, fmt.Errorf("creating unixfs adder: %s", err)
	}
	fileAdder.Pin = false
	fileAdder.Progress = true
	fileAdder.Out = events

	f, err := os.Open(path)
	if err != nil {
		return cid.Undef, fmt.Errorf("opening path: %s", err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return cid.Undef, fmt.Errorf("getting stat of data: %s", err)
	}
	fs, err := files.NewSerialFile(path, false, stat)
	if err != nil {
		return cid.Undef, fmt.Errorf("creating serial file: %s", err)
	}
	defer fs.Close()

	dataSize := int(stat.Size())
	if stat.IsDir() {
		dataSize = 0
		err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				dataSize += int(info.Size())
			}
			return err
		})
		if err != nil {
			return cid.Undef, fmt.Errorf("walking path: %s", err)
		}
	}
	bar := pb.StartNew(dataSize)
	bar.Set(pb.Bytes, true)

	var (
		dagifyErr error
		dataCid   cid.Cid
	)
	go func() {
		defer close(events)

		nd, err := fileAdder.AddAllAndPin(fs)
		if err != nil {
			dagifyErr = err
			return
		}
		dataCid = nd.Cid()
	}()
	currentName := ""
	var previousSize int64
	for event := range events {
		output, ok := event.(*coreiface.AddEvent)
		if !ok {
			c.CheckErr(errors.New("unknown event type"))
		}
		if output.Name == "" {
			continue
		}
		panic(1)
		if currentName != output.Name {
			currentName = output.Name
			previousSize = 0
		}
		if output.Bytes > 0 {
			c.Message("LA")
			bar.Add64(-previousSize + output.Bytes)
		}
		previousSize = output.Bytes
	}
	bar.Finish()
	if dagifyErr != nil {
		return cid.Undef, fmt.Errorf("creating dag for data: %s", dagifyErr)
	}

	return dataCid, nil
}
