package prepare

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/dustin/go-humanize"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commP "github.com/filecoin-project/go-fil-commp-hashhash"
	bsrv "github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-car"
	"github.com/ipfs/go-cid"
	badger "github.com/ipfs/go-ds-badger"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	files "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/go-ipfs/core/coreunix"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
)

func init() {
	Cmd.AddCommand(prepare, genCar, commp)

	prepare.Flags().String("tmpdir", os.TempDir(), "path of folder where a temporal blockstore is created for processing data")

	commp.Flags().Bool("json", false, "print output in json format")
	commp.Flags().Bool("skip-car-validation", false, "skips CAR validation when processing a path")

	genCar.Flags().String("tmpdir", os.TempDir(), "path of folder where a temporal blockstore is created for processing data")
	genCar.Flags().String("ipfs-api", "", "IPFS HTTP API multiaddress that stores the cid (only for Cid processing instead of file/folder path)")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "offline",
	Short: "Provides commands to prepare data for Filecoin onbarding",
	Long:  `Provides commands to prepare data for Filecoin onbarding`,
}

var genCar = &cobra.Command{
	Use:   "car [path | cid] [output path]",
	Short: "car generates a CAR file from the data",
	Long:  `car generates a CAR file from the data`,
	Args:  cobra.RangeArgs(0, 2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var dataCid cid.Cid
		var err error
		ctx := context.Background()

		w := os.Stdout
		if len(args) == 2 {
			w, err = os.OpenFile(args[1], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
			if err != nil {
				c.Fatal(fmt.Errorf("creating output file: %s", err))
			}
			defer func() {
				if err := w.Close(); err != nil {
					c.Fatal(fmt.Errorf("closing output file: %s", err))
				}
			}()
		}

		dataCid, dagService, cls, err := prepareDAGService(cmd, args)
		if err != nil {
			c.Fatal(fmt.Errorf("creating dag-service: %s", err))
		}
		defer cls()

		err = car.WriteCar(ctx, dagService, []cid.Cid{dataCid}, w)
		c.CheckErr(err)
	},
}

var commp = &cobra.Command{
	Use:     "commp [path]",
	Aliases: []string{"commP"},
	Short:   "commP calculates the piece size and cid for a CAR file",
	Long:    `commP claculates the piece size and cid for a CAR file`,
	Args:    cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var (
			r   io.Reader
			err error
		)
		r = os.Stdin
		if len(args) > 0 && args[0] != "-" {
			r, err := os.Open(args[0])
			if err != nil {
				c.Fatal(fmt.Errorf("opening the file %s: %s", args[0], err))
			}
			defer r.Close()

			skipCARValidation, err := cmd.Flags().GetBool("skip-car-validation")
			if err != nil {
				c.Fatal(fmt.Errorf("getting skip-car-validation flag: %s", err))
			}
			if !skipCARValidation {
				_, err = car.ReadHeader(bufio.NewReader(r))
				if err != nil {
					c.Fatal(fmt.Errorf("wrong car file format: %s", err))
				}

				if _, err := r.Seek(0, io.SeekStart); err != nil {
					c.Fatal(fmt.Errorf("rewind file to start: %s", err))
				}
			}
		}

		pieceCID, pieceSize, err := calcCommP(r)
		if err != nil {
			c.Fatal(fmt.Errorf("calculating commP: %s", err))
		}

		jsonFlag, err := cmd.Flags().GetBool("json")
		if err != nil {
			c.Fatal(fmt.Errorf("parsing json flag: %s", err))
		}
		if jsonFlag {
			outData := struct {
				PieceSize uint64 `json:"piece_size"`
				PieceCid  string `json:"piece_cid"`
			}{
				PieceSize: pieceSize,
				PieceCid:  pieceCID.String(),
			}
			out, err := json.Marshal(outData)
			c.CheckErr(err)
			fmt.Println(string(out))

			return
		}
		c.Message("Piece-size: %s", humanize.IBytes(pieceSize))
		c.Message("PieceCID: %s", pieceCID)
	},
}

var prepare = &cobra.Command{
	Use:     "prepare [cid | path] [output file path]",
	Aliases: []string{"prep"},
	Short:   "prepare generates a CAR file for data",
	Long:    `prepare generates a CAR file for data`,
	Args:    cobra.RangeArgs(0, 2),
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
		// TTODO: print lotus and powergate commands to fire the offline deal
		c.FmtOutput = os.Stderr

		tmpDir, err := cmd.Flags().GetString("tmpdir")
		c.CheckErr(err)

		dagService, cls, err := createTmpDAGService(tmpDir)
		if err != nil {
			c.Fatal(fmt.Errorf("creating temporal dag-service: %s", err))
		}
		defer cls()

		ctx := context.Background()
		path := args[0]

		c.Message("Creating data DAG...")
		start := time.Now()
		dataCid, err := dagify(ctx, dagService, path)
		if err != nil {
			c.Fatal(fmt.Errorf("creating dag for data: %s", err))
		}
		c.Message("DAG created in %.02fs.", time.Since(start).Seconds())

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

		c.Message("Creating CAR and calculating piece-size and PieceCID...")
		start = time.Now()
		prCAR, pwCAR := io.Pipe()
		var writeCarErr error
		go func() {
			defer pwCAR.Close()
			if err := car.WriteCar(ctx, dagService, []cid.Cid{dataCid}, pwCAR); err != nil {
				writeCarErr = err
				return
			}
		}()

		prCommP, pwCommP := io.Pipe()
		teeCAR := io.TeeReader(prCAR, pwCommP)
		var (
			errCommP  error
			wg        sync.WaitGroup
			pieceCid  cid.Cid
			pieceSize uint64
		)
		wg.Add(1)
		go func() {
			defer wg.Done()
			pieceCid, pieceSize, errCommP = calcCommP(prCommP)
		}()
		if _, err := io.Copy(outputFile, teeCAR); err != nil {
			c.Fatal(fmt.Errorf("writing CAR file to output: %s", err))
		}
		if writeCarErr != nil {
			c.Fatal(fmt.Errorf("generating CAR file: %s", err))
		}
		pwCommP.Close()
		wg.Wait()
		if errCommP != nil {
			c.Fatal(fmt.Errorf("calculating piece-size and PieceCID: %s", err))
		}
		c.Message("Created CAR file, and piece digest in %.02fs.", time.Since(start).Seconds())

		c.Message("Piece size: %s", humanize.IBytes(pieceSize))
		c.Message("Piece CID: %s", pieceCid)
	},
}

type CloseFunc func() error

func createTmpDAGService(tmpDir string) (ipld.DAGService, CloseFunc, error) {
	badgerFolder, err := ioutil.TempDir(tmpDir, "powprepare-*")
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
		if stat.IsDir() && output.Name == "" {
			continue
		}
		if currentName != output.Name {
			currentName = output.Name
			previousSize = 0
		}
		if output.Bytes > 0 {
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

func calcCommP(r io.Reader) (cid.Cid, uint64, error) {
	cp := &commP.Calc{}
	_, err := io.Copy(cp, r)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("copying data to aggregator: %s", err)
	}

	rawCommP, pieceSize, err := cp.Digest()
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("calculating final digest: %s", err)
	}
	pieceCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("converting commP to cid: %s", err)
	}

	return pieceCid, pieceSize, nil

}

// TTODO: move lib related things to... libs. Only leave here CLI stuff.

func prepareDAGService(cmd *cobra.Command, args []string) (cid.Cid, ipld.DAGService, CloseFunc, error) {
	ipfsAPI, err := cmd.Flags().GetString("ipfs-api")
	if err != nil {
		return cid.Undef, nil, nil, fmt.Errorf("getting ipfs api flag: %s", err)
	}

	if ipfsAPI == "" {
		path := "/dev/stdin"
		if len(args) > 0 && args[0] != "-" {
			path = args[0]
		}

		tmpDir, err := cmd.Flags().GetString("tmpdir")
		if err != nil {
			return cid.Undef, nil, nil, fmt.Errorf("getting tmpdir directory: %s", err)
		}

		dagService, cls, err := createTmpDAGService(tmpDir)
		if err != nil {
			return cid.Undef, nil, nil, fmt.Errorf("creating temporary dag-service: %s", err)
		}
		ctx := context.Background()
		dataCid, err := dagify(ctx, dagService, path)
		if err != nil {
			return cid.Undef, nil, nil, fmt.Errorf("creating dag for data: %s", err)
		}
		return dataCid, dagService, cls, nil
	}

	if len(args) == 0 {
		return cid.Undef, nil, nil, fmt.Errorf("cid argument is empty")
	}
	dataCid, err := cid.Decode(args[0])
	if err != nil {
		return cid.Undef, nil, nil, fmt.Errorf("parsing cid: %s", err)
	}

	ipfsAPIMA, err := multiaddr.NewMultiaddr(ipfsAPI)
	if err != nil {
		return cid.Undef, nil, nil, fmt.Errorf("parsing ipfs-api multiaddress: %s", err)
	}
	ipfs, err := httpapi.NewApi(ipfsAPIMA)
	if err != nil {
		return cid.Undef, nil, nil, fmt.Errorf("creating ipfs client: %s", err)
	}

	return dataCid, ipfs.Dag(), CloseFunc(func() error { return nil }), nil
}
