package gateway

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	logger "github.com/ipfs/go-log/v2"
	assets "github.com/jessevdk/go-assets"
	"github.com/rs/cors"
	gincors "github.com/rs/cors/wrapper/gin"
	askRunner "github.com/textileio/powergate/v2/index/ask/runner"
	faultsModule "github.com/textileio/powergate/v2/index/faults/module"
	minerModule "github.com/textileio/powergate/v2/index/miner/lotusidx"
	"github.com/textileio/powergate/v2/reputation"
)

const numTopMiners = 100

var log = logger.Logger("gateway")

// fileSystem extends the binary asset file system with Exists,
// enabling its use with the static middleware.
type fileSystem struct {
	*assets.FileSystem
}

// Exists returns whether or not the path exists in the binary assets.
func (f *fileSystem) Exists(prefix, path string) bool {
	pth := strings.TrimPrefix(path, prefix)
	if pth == "/" {
		return false
	}
	_, ok := f.Files[pth]
	return ok
}

// Gateway provides HTTP-based access to Textile.
type Gateway struct {
	addr             string
	server           *http.Server
	askIndex         *askRunner.Runner
	minerIndex       *minerModule.Index
	faultsIndex      *faultsModule.Index
	reputationModule *reputation.Module
}

// NewGateway returns a new gateway.
func NewGateway(
	addr string,
	askIndex *askRunner.Runner,
	minerIndex *minerModule.Index,
	faultsIndex *faultsModule.Index,
	reputationModule *reputation.Module,
) *Gateway {
	return &Gateway{
		addr:             addr,
		askIndex:         askIndex,
		minerIndex:       minerIndex,
		faultsIndex:      faultsIndex,
		reputationModule: reputationModule,
	}
}

// Start the gateway.
func (g *Gateway) Start(basePath string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(location.Default())

	// @todo: Config based headers
	options := cors.Options{}
	router.Use(gincors.New(options))

	temp, err := loadTemplate()
	if err != nil {
		log.Fatal(err)
	}
	router.SetHTMLTemplate(temp)

	if basePath == "/" {
		basePath = ""
	}
	router.Use(static.Serve(basePath, &fileSystem{Assets}))
	rg := router.Group(basePath)
	rg.GET("/asks", g.asksHandler)
	rg.GET("/miners", g.minersHandler)
	rg.GET("/faults", g.faultsHandler)
	rg.GET("/reputation", g.reputationHandler)

	rg.GET("/", func(c *gin.Context) {
		c.Request.URL.Path = basePath + "/asks"
		router.HandleContext(c)
	})

	router.NoRoute(func(c *gin.Context) {
		g.render404(c)
	})

	g.server = &http.Server{
		Addr:    g.addr,
		Handler: router,
	}

	errc := make(chan error)
	go func() {
		errc <- g.server.ListenAndServe()
		close(errc)
	}()
	go func() {
		for err := range errc {
			if err != nil {
				if err != http.ErrServerClosed {
					log.Errorf("gateway error: %s", err)
				}
				return
			}
		}
		log.Info("gateway was shutdown")
	}()
	log.Infof("gateway listening at %s", g.server.Addr)
}

// Addr returns the gateway's address.
func (g *Gateway) Addr() string {
	return g.server.Addr
}

// Stop the gateway.
func (g *Gateway) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := g.server.Shutdown(ctx); err != nil {
		log.Errorf("error shutting down gateway: %s", err)
		return err
	}
	return nil
}

func (g *Gateway) asksHandler(c *gin.Context) {
	menuItems := makeMenuItems(0)

	index := g.askIndex.Get()

	subtitle := fmt.Sprintf("Last updated: %v, storage median price: %v", timeToString(index.LastUpdated), index.StorageMedianPrice)
	headers := []string{"Miner", "Price", "Verified Price", "Min Piece Size", "Max Piece Size", "Timestamp", "Expiry"}

	rows := make([][]interface{}, len(index.Storage))
	i := 0
	for _, ask := range index.Storage {
		rows[i] = []interface{}{
			ask.Miner,
			ask.Price,
			ask.VerifiedPrice,
			ask.MinPieceSize,
			ask.MaxPieceSize,
			ask.Timestamp,
			ask.Expiry,
		}
		i++
	}

	c.HTML(http.StatusOK, "/public/html/asks.gohtml", gin.H{
		"MenuItems": menuItems,
		"Title":     "Available Asks",
		"Subtitle":  subtitle,
		"Headers":   headers,
		"Rows":      rows,
	})
}

func (g *Gateway) minersHandler(c *gin.Context) {
	menuItems := makeMenuItems(1)

	index := g.minerIndex.Get()

	metaHeaders := []string{"Miner", "Location", "User Agent", "Updated"}
	metaRows := make([][]interface{}, len(index.Meta.Info))
	i := 0
	for id, meta := range index.Meta.Info {
		metaRows[i] = []interface{}{
			id,
			meta.Location.Country,
			meta.UserAgent,
			timeToString(meta.LastUpdated),
		}
		i++
	}

	chainSubtitle := fmt.Sprintf("Last updated %v", timeToString(epochToTime(index.OnChain.LastUpdated)))
	chainHeaders := []string{"Miner", "Power", "RelativePower", "SectorSize", "SectorsActive", "SectorsLive", "SectorsFaulty"}
	var chainRows [][]interface{}
	i = 0
	for id, onchainData := range index.OnChain.Miners {
		if onchainData.Power == 0 {
			continue
		}
		chainRows = append(chainRows, []interface{}{
			id,
			onchainData.Power,
			onchainData.RelativePower,
			onchainData.SectorSize,
			onchainData.SectorsActive,
			onchainData.SectorsLive,
			onchainData.SectorsFaulty,
		})
		i++
	}

	sort.Slice(chainRows, func(i, j int) bool {
		l := chainRows[i][0].(string)
		r := chainRows[j][0].(string)
		return index.OnChain.Miners[l].RelativePower >= index.OnChain.Miners[r].RelativePower
	})

	c.HTML(http.StatusOK, "/public/html/miners.gohtml", gin.H{
		"MenuItems": menuItems,
		"MetaData": gin.H{
			"Title":   "Miner Metadata",
			"Headers": metaHeaders,
			"Rows":    metaRows,
		},
		"ChainData": gin.H{
			"Title":    "Miner On-Chain Data",
			"Subtitle": chainSubtitle,
			"Headers":  chainHeaders,
			"Rows":     chainRows,
		},
	})
}

func (g *Gateway) faultsHandler(c *gin.Context) {
	menuItems := makeMenuItems(2)

	index := g.faultsIndex.Get()

	subtitle := fmt.Sprintf("Current tip set key: %v", index.TipSetKey)

	headers := []string{"Miner", "Faults Epochs"}

	rows := make([][]interface{}, len(index.Miners))
	i := 0
	for id, faults := range index.Miners {
		epochs := make([]string, len(faults.Epochs))
		for j, epoch := range faults.Epochs {
			epochs[j] = strconv.FormatInt(epoch, 10)
		}
		rows[i] = []interface{}{
			id,
			strings.Join(epochs, ", "),
		}
		i++
	}

	sort.Slice(rows, func(i, j int) bool {
		l := rows[i][0].(string)
		r := rows[j][0].(string)
		return len(index.Miners[l].Epochs) >= len(index.Miners[r].Epochs)
	})

	c.HTML(http.StatusOK, "/public/html/faults.gohtml", gin.H{
		"MenuItems": menuItems,
		"Title":     "Miner Faults",
		"Subtitle":  subtitle,
		"Headers":   headers,
		"Rows":      rows,
	})
}

func (g *Gateway) reputationHandler(c *gin.Context) {
	menuItems := makeMenuItems(3)

	topMiners, err := g.reputationModule.GetTopMiners(numTopMiners)
	if err != nil {
		g.renderError(c, http.StatusInternalServerError, err)
		return
	}

	headers := []string{"Miner", "Score"}

	rows := make([][]interface{}, len(topMiners))
	for i, minerScore := range topMiners {
		rows[i] = []interface{}{
			minerScore.Addr,
			minerScore.Score,
		}
	}

	c.HTML(http.StatusOK, "/public/html/reputation.gohtml", gin.H{
		"MenuItems": menuItems,
		"Title":     fmt.Sprintf("Top %v Miners", numTopMiners),
		"Headers":   headers,
		"Rows":      rows,
	})
}

func epochToTime(value int64) time.Time {
	genesisEpochTime := int64(1598295600)
	return time.Unix(genesisEpochTime+value*30, 0)
}

func timeToString(t time.Time) string {
	return t.Format("01/02/06 3:04 PM")
}

type menuItem struct {
	Name     string
	Path     string
	Selected bool
}

func makeMenuItems(selectedIndex int) []menuItem {
	menuItems := []menuItem{
		{
			Name:     "Asks",
			Path:     "asks",
			Selected: false,
		},
		{
			Name:     "Miners",
			Path:     "miners",
			Selected: false,
		},
		{
			Name:     "Faults",
			Path:     "faults",
			Selected: false,
		},
		{
			Name:     "Reputation",
			Path:     "reputation",
			Selected: false,
		},
	}
	menuItems[selectedIndex].Selected = true
	return menuItems
}

// render404 renders the 404 template.
func (g *Gateway) render404(c *gin.Context) {
	c.HTML(http.StatusNotFound, "/public/html/404.gohtml", nil)
}

// renderError renders the error template.
func (g *Gateway) renderError(c *gin.Context, code int, err error) {
	c.HTML(code, "/public/html/error.gohtml", gin.H{
		"Code":  code,
		"Error": formatError(err),
	})
}

// loadTemplate loads HTML templates.
func loadTemplate() (*template.Template, error) {
	t := template.New("")
	for name, file := range Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".gohtml") {
			continue
		}
		h, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		t, err = t.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// formatError formats a go error for browser display.
func formatError(err error) string {
	words := strings.SplitN(err.Error(), " ", 2)
	words[0] = strings.Title(words[0])
	return strings.Join(words, " ") + "."
}
