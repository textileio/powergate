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

	files "github.com/ipfs/go-ipfs-files"
	unixfile "github.com/ipfs/go-unixfs/file"

	"github.com/cheggaaa/pb/v3"
	"github.com/dustin/go-humanize"
	aggregator "github.com/filecoin-project/go-dagaggregator-unixfs"
	bsrv "github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-car"
	"github.com/ipfs/go-cid"
	badger "github.com/ipfs/go-ds-badger"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"github.com/textileio/powergate/v2/dataprep"
)

func init() {
	Cmd.AddCommand(prepare, genCar, commp)

	prepare.Flags().String("tmpdir", os.TempDir(), "path of folder where a temporal blockstore is created for processing data")
	prepare.Flags().String("ipfs-api", "", "IPFS HTTP API multiaddress that stores the cid (only for Cid processing instead of file/folder path)")
	prepare.Flags().Bool("json", false, "avoid pretty output and use json formatting")
	prepare.Flags().Bool("aggregate", false, "aggregates a folder of files")

	commp.Flags().Bool("json", false, "avoid pretty output and use json formatting")
	commp.Flags().Bool("skip-car-validation", false, "skips CAR validation when processing a path")

	genCar.Flags().String("tmpdir", os.TempDir(), "path of folder where a temporal blockstore is created for processing data")
	genCar.Flags().String("ipfs-api", "", "IPFS HTTP API multiaddress that stores the cid (only for Cid processing instead of file/folder path)")
	genCar.Flags().Bool("quiet", false, "avoid pretty output")
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
	Long: `Generates a CAR file from the data source. This data-source can be a file/folder path or a Cid.

If a file/folder path is provided, this command will DAGify the data and generate the CAR file.
If a Cid is provided, an extra --ipfs-api flag should be provided to connect to the IPFS node that contains this Cid data.`,
	Args: cobra.RangeArgs(1, 2),
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

		quiet, err := cmd.Flags().GetBool("quiet")
		if err != nil {
			c.Fatal(fmt.Errorf("parsing json flag: %s", err))
		}
		dataCid, dagService, _, cls, err := prepareDAGService(cmd, args, quiet)
		if err != nil {
			c.Fatal(fmt.Errorf("creating dag-service: %s", err))
		}
		defer func() { _ = cls() }()

		if err = car.WriteCar(ctx, dagService, []cid.Cid{dataCid}, w); err != nil {
			c.Fatal(fmt.Errorf("generating car file: %s", err))
		}
	},
}

var commp = &cobra.Command{
	Use:     "commp [path]",
	Aliases: []string{"commP"},
	Short:   "commP calculates the piece size and cid for a CAR file",
	Long: `commP calculates the piece-size and PieceCID for a CAR file.

This command calculates the piece-size and piece-cid (CommP) from a CAR file.
This command only makes sense to run for a CAR file, so it does some quick check if the input file *seems* to be well-formated. 
You can use the --skip-car-validation, but usually shouldn't be done unless you know what you're doing (e.g.: benchmarks, or other tests)`,
	Args: cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		r := io.Reader(os.Stdin)
		if len(args) > 0 && args[0] != "-" {
			f, err := os.Open(args[0])
			if err != nil {
				c.Fatal(fmt.Errorf("opening the file %s: %s", args[0], err))
			}
			defer func() { _ = f.Close() }()

			skipCARValidation, err := cmd.Flags().GetBool("skip-car-validation")
			if err != nil {
				c.Fatal(fmt.Errorf("getting skip-car-validation flag: %s", err))
			}
			if !skipCARValidation {
				_, err = car.ReadHeader(bufio.NewReader(f))
				if err != nil {
					c.Fatal(fmt.Errorf("wrong car file format: %s", err))
				}
				if _, err := f.Seek(0, io.SeekStart); err != nil {
					c.Fatal(fmt.Errorf("rewind file to start: %s", err))
				}
			}
			r = f
		}

		pieceCID, pieceSize, err := dataprep.CommP(r)
		if err != nil {
			c.Fatal(fmt.Errorf("calculating commP: %s", err))
		}

		jsonFlag, err := cmd.Flags().GetBool("json")
		if err != nil {
			c.Fatal(fmt.Errorf("parsing json flag: %s", err))
		}
		if jsonFlag {
			printJSONResult(pieceSize, cid.Undef, pieceCID, nil)
			return
		}
		c.Message("Piece-size: %d (%s)", pieceSize, humanize.IBytes(pieceSize))
		c.Message("PieceCID: %s", pieceCID)
	},
}

var prepare = &cobra.Command{
	Use:     "prepare [cid | path] [output CAR file path]",
	Aliases: []string{"prep"},
	Short:   "prepare generates a CAR file for data",
	Long: `Prepares a data source generating all needed to execute an offline deal.
The data source can be a file/folder path or a Cid.

If a file/folder path is provided, this command will DAGify the data and generate the CAR file.
If a Cid is provided, an extra --ipfs-api flag should be provided to connect to the IPFS node that contains this Cid data.

This command prepares data in a more efficiently than running car+commp subcommands, since it already starts calculating CommP at the same time that the CAR file is being generated.

By default prints to stdout the generated CAR file. You can provide a second argument to
specify the output file path, or simply pipe the stdout result.

The piece-size and piece-cid are printed to stderr. For scripting usage, its recommended to use the --json flag.`,
	Args: cobra.RangeArgs(1, 2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		c.FmtOutput = os.Stderr

		json, err := cmd.Flags().GetBool("json")
		if err != nil {
			c.Fatal(fmt.Errorf("parsing json flag: %s", err))
		}
		aggregate, err := cmd.Flags().GetBool("aggregate")
		if err != nil {
			c.Fatal(fmt.Errorf("aggregate flag: %s", err))
		}
		ipfsAPI, err := cmd.Flags().GetString("ipfs-api")
		if err != nil {
			c.Fatal(fmt.Errorf("get api addr flag: %s", err))
		}
		if ipfsAPI != "" && aggregate {
			c.Fatal(errors.New("the --aggregate flag can't be used with a remote go-ipfs node"))
		}

		payloadCid, dagService, aggrFiles, cls, err := prepareDAGService(cmd, args, json)
		if err != nil {
			c.Fatal(fmt.Errorf("creating temporal dag-service: %s", err))
		}
		defer func() { _ = cls() }()

		ctx := context.Background()

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

		if !json {
			c.Message("Creating CAR and calculating piece-size and PieceCID...")
		}
		start := time.Now()
		prCAR, pwCAR := io.Pipe()
		var writeCarErr error
		go func() {
			defer func() {
				if err := pwCAR.Close(); err != nil {
					c.Fatal(fmt.Errorf("closing car writer: %s", err))
				}
			}()
			if err := car.WriteCar(ctx, dagService, []cid.Cid{payloadCid}, pwCAR); err != nil {
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
			pieceCid, pieceSize, errCommP = dataprep.CommP(prCommP)
		}()
		if _, err := io.Copy(outputFile, teeCAR); err != nil {
			c.Fatal(fmt.Errorf("writing CAR file to output: %s", err))
		}
		if writeCarErr != nil {
			c.Fatal(fmt.Errorf("generating CAR file: %s", err))
		}
		if err := pwCommP.Close(); err != nil {
			c.Fatal(fmt.Errorf("closing tee-reader writer: %s", err))
		}
		wg.Wait()
		if errCommP != nil {
			c.Fatal(fmt.Errorf("calculating piece-size and PieceCID: %s", err))
		}

		if aggregate {
			rootNd, err := dagService.Get(ctx, payloadCid)
			if err != nil {
				c.Fatal(fmt.Errorf("get root node: %s", err))
			}
			manifestLnk := rootNd.Links()[0]
			manifestNd, err := dagService.Get(ctx, manifestLnk.Cid)
			if err != nil {
				c.Fatal(fmt.Errorf("get manifest node: %s", err))
			}
			manifestF, err := unixfile.NewUnixfsFile(context.Background(), dagService, manifestNd)
			if err != nil {
				c.Fatal(fmt.Errorf("get unifxfs manifest file: %s", err))
			}
			manifest := manifestF.(files.File)
			fManifest, err := os.Create(args[1] + ".manifest")
			if err != nil {
				c.Fatal(fmt.Errorf("creating manifest file: %s", err))

			}
			defer func() {
				if err := fManifest.Close(); err != nil {
					c.Fatal(fmt.Errorf("closing manifest file: %s", err))
				}
			}()
			if _, err := io.Copy(fManifest, manifest); err != nil {
				c.Fatal(fmt.Errorf("writing manifest file: %s", err))
			}
		}

		if json {
			printJSONResult(pieceSize, payloadCid, pieceCid, aggrFiles)
			return
		}
		c.Message("Created CAR file, and piece digest in %.02fs.", time.Since(start).Seconds())
		c.Message("Payload cid: %s", payloadCid)
		c.Message("Piece size: %d (%s)", pieceSize, humanize.IBytes(pieceSize))
		c.Message("Piece CID: %s\n", pieceCid)

		c.Message("Lotus offline-deal command:")
		c.Message("lotus client deal --manual-piece-cid=%s --manual-piece-size=%d %s <miner> <price> <duration>", pieceCid, pieceSize, payloadCid)
	},
}

type closeFunc func() error

func prepareDAGService(cmd *cobra.Command, args []string, quiet bool) (cid.Cid, ipld.DAGService, []aggregatedFile, closeFunc, error) {
	ipfsAPI, err := cmd.Flags().GetString("ipfs-api")
	if err != nil {
		return cid.Undef, nil, nil, nil, fmt.Errorf("getting ipfs api flag: %s", err)
	}
	aggregate, err := cmd.Flags().GetBool("aggregate")
	if err != nil {
		return cid.Undef, nil, nil, nil, fmt.Errorf("aggregate flag: %s", err)
	}
	if ipfsAPI == "" {
		path := args[0]
		tmpDir, err := cmd.Flags().GetString("tmpdir")
		if err != nil {
			return cid.Undef, nil, nil, nil, fmt.Errorf("getting tmpdir directory: %s", err)
		}

		dagService, cls, err := createTmpDAGService(tmpDir)
		if err != nil {
			return cid.Undef, nil, nil, nil, fmt.Errorf("creating temporary dag-service: %s", err)
		}
		dataCid, aggregatedFiles, err := dagify(context.Background(), dagService, path, quiet, aggregate)
		if err != nil {
			return cid.Undef, nil, nil, nil, fmt.Errorf("creating dag for data: %s", err)
		}

		return dataCid, dagService, aggregatedFiles, cls, nil
	}

	if len(args) == 0 {
		return cid.Undef, nil, nil, nil, fmt.Errorf("cid argument is empty")
	}
	dataCid, err := cid.Decode(args[0])
	if err != nil {
		return cid.Undef, nil, nil, nil, fmt.Errorf("parsing cid: %s", err)
	}

	ipfsAPIMA, err := multiaddr.NewMultiaddr(ipfsAPI)
	if err != nil {
		return cid.Undef, nil, nil, nil, fmt.Errorf("parsing ipfs-api multiaddress: %s", err)
	}
	ipfs, err := httpapi.NewApi(ipfsAPIMA)
	if err != nil {
		return cid.Undef, nil, nil, nil, fmt.Errorf("creating ipfs client: %s", err)
	}

	return dataCid, ipfs.Dag(), nil, closeFunc(func() error { return nil }), nil
}

func createTmpDAGService(tmpDir string) (ipld.DAGService, closeFunc, error) {
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
			_ = os.RemoveAll(badgerFolder)

			return nil
		}, nil
}

var jsonOutput = io.Writer(os.Stderr)

func printJSONResult(pieceSize uint64, payloadCid, pieceCID cid.Cid, aggrFiles []aggregatedFile) {
	outData := struct {
		PayloadCid      string           `json:"payload_cid,omitempty"`
		PieceSize       uint64           `json:"piece_size"`
		PieceCid        string           `json:"piece_cid"`
		AggregatedFiles []aggregatedFile `json:"files,omitempty"`
	}{
		PieceSize:       pieceSize,
		PieceCid:        pieceCID.String(),
		AggregatedFiles: aggrFiles,
	}
	if payloadCid.Defined() {
		outData.PayloadCid = payloadCid.String()
	}
	out, err := json.Marshal(outData)
	c.CheckErr(err)
	fmt.Fprint(jsonOutput, string(out))
}

func dagify(ctx context.Context, dagService ipld.DAGService, path string, quiet bool, aggregate bool) (cid.Cid, []aggregatedFile, error) {
	var progressChan chan int64
	if !quiet {
		f, err := os.Open(path)
		if err != nil {
			return cid.Undef, nil, fmt.Errorf("opening path: %s", err)
		}
		stat, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return cid.Undef, nil, fmt.Errorf("getting stat of data: %s", err)
		}
		_ = f.Close()

		c.Message("Creating data DAG...")
		start := time.Now()
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
				return cid.Undef, nil, fmt.Errorf("walking path: %s", err)
			}
		}
		bar := pb.StartNew(dataSize)
		bar.Set(pb.Bytes, true)

		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Wait()
		progressChan = make(chan int64)
		defer close(progressChan)
		go func() {
			defer wg.Done()
			for bytesProgress := range progressChan {
				bar.Add64(bytesProgress)
			}
			bar.Finish()
			c.Message("DAG created in %.02fs.", time.Since(start).Seconds())
		}()
	}

	var aggregatedFiles []aggregatedFile
	var dataCid cid.Cid
	var err error
	if !aggregate {
		dataCid, err = dataprep.Dagify(ctx, dagService, path, progressChan)
		if err != nil {
			return cid.Undef, nil, fmt.Errorf("creating dag for data: %s", err)
		}
	} else {
		var lst []aggregator.AggregateDagEntry
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			dcid, err := dataprep.Dagify(ctx, dagService, p, progressChan)
			if err != nil {
				return err
			}
			lst = append(lst, aggregator.AggregateDagEntry{
				RootCid: dcid,
			})
			aggregatedFiles = append(aggregatedFiles, aggregatedFile{Name: p, Cid: dcid.String()})
			return err
		})
		if err != nil {
			return cid.Undef, nil, fmt.Errorf("walking the target folder: %s", err)
		}

		dataCid, err = aggregator.Aggregate(ctx, dagService, lst)
		if err != nil {
			return cid.Undef, nil, fmt.Errorf("aggregating: %s", err)
		}
	}

	return dataCid, aggregatedFiles, nil
}

type aggregatedFile struct {
	Name string `json:"name"`
	Cid  string `jsong:"cid"`
}
