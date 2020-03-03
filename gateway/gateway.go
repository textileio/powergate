package gateway

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	logger "github.com/ipfs/go-log"
	assets "github.com/jessevdk/go-assets"
	"github.com/rs/cors"
	gincors "github.com/rs/cors/wrapper/gin"
	"github.com/textileio/fil-tools/index/ask"
	"github.com/textileio/fil-tools/index/miner"
	"github.com/textileio/fil-tools/index/slashing"
	"github.com/textileio/fil-tools/reputation"
	"github.com/textileio/go-threads/broadcast"
)

const handlerTimeout = time.Second * 10

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
	askIndex         *ask.AskIndex
	minerIndex       *miner.MinerIndex
	slashingIndex    *slashing.SlashingIndex
	reputationModule *reputation.Module
	sessionBus       *broadcast.Broadcaster
}

// NewGateway returns a new gateway.
func NewGateway(
	addr string,
	askIndex *ask.AskIndex,
	minerIndex *miner.MinerIndex,
	slashingIndex *slashing.SlashingIndex,
	reputationModule *reputation.Module,
) *Gateway {
	return &Gateway{
		addr:             addr,
		askIndex:         askIndex,
		minerIndex:       minerIndex,
		slashingIndex:    slashingIndex,
		reputationModule: reputationModule,
	}
}

// Start the gateway.
func (g *Gateway) Start() {
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

	router.Use(static.Serve("", &fileSystem{Assets}))

	router.GET("/health", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusNoContent)
	})

	router.GET("/asks", g.asksHandler)
	router.GET("/miners", g.minersHandler)
	router.GET("/slashing", g.slashingHandler)
	router.GET("/reputation", g.reputationHandler)

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
		for {
			select {
			case err, ok := <-errc:
				if err != nil {
					if err == http.ErrServerClosed {
						return
					}
					log.Errorf("gateway error: %s", err)
				}
				if !ok {
					log.Info("gateway was shutdown")
					return
				}
			}
		}
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

	headers := []string{"Miner", "Price", "Min Piece Size", "Timestamp", "Expiry"}

	rows := make([][]interface{}, len(index.Storage))
	i := 0
	for _, ask := range index.Storage {
		rows[i] = []interface{}{
			ask.Miner,
			ask.Price,
			ask.MinPieceSize,
			timeToString(uint64ToTime(ask.Timestamp)),
			timeToString(uint64ToTime(ask.Expiry)),
		}
		i++
	}

	foo := [][]string{
		[]string{
			"asdfsad",
			"1234",
			"253",
			"Timestamp",
			"Expiry",
		},
		[]string{
			"asdfsad",
			"1234",
			"253",
			"Timestamp",
			"Expiry",
		},
		[]string{
			"asdfsad",
			"1234",
			"253",
			"Timestamp",
			"Expiry",
		},
		[]string{
			"asdfsad",
			"1234",
			"253",
			"Timestamp",
			"Expiry",
		},
		[]string{
			"asdfsad",
			"1234",
			"253",
			"Timestamp",
			"Expiry",
		},
		[]string{
			"asdfsad",
			"1234",
			"253",
			"Timestamp",
			"Expiry",
		},
	}

	c.HTML(http.StatusOK, "/public/html/asks.gohtml", gin.H{
		"MenuItems": menuItems,
		"Title":     "Available Asks",
		"Subtitle":  subtitle,
		"Headers":   headers,
		"Rows":      foo,
	})
}

func (g *Gateway) minersHandler(c *gin.Context) {
	menuItems := makeMenuItems(1)

	index := g.minerIndex.Get()

	metaSubtitle := fmt.Sprintf("%v miners online, %v miners offline", index.Meta.Online, index.Meta.Offline)
	metaHeaders := []string{"Miner", "Location", "Online", "User Agent", "Updated"}
	metaRows := make([][]interface{}, len(index.Meta.Info))
	i := 0
	for id, meta := range index.Meta.Info {
		metaRows[i] = []interface{}{
			id,
			meta.Location.Country,
			meta.Online,
			meta.UserAgent,
			timeToString(meta.LastUpdated),
		}
		i++
	}

	chainSubtitle := fmt.Sprintf("Last updated %v", timeToString(uint64ToTime(index.Chain.LastUpdated)))
	chainHeaders := []string{"Miner", "Power", "Relative"}
	chainRows := make([][]interface{}, len(index.Chain.Power))
	i = 0
	for id, power := range index.Chain.Power {
		chainRows[i] = []interface{}{
			id,
			power.Power,
			power.Relative,
		}
		i++
	}

	c.HTML(http.StatusOK, "/public/html/miners.gohtml", gin.H{
		"MenuItems": menuItems,
		"MetaData": gin.H{
			"Title":    "Miner Metadata",
			"Subtitle": metaSubtitle,
			"Headers":  metaHeaders,
			"Rows":     metaRows,
		},
		"ChainData": gin.H{
			"Title":    "Miner on chain data",
			"Subtitle": chainSubtitle,
			"Headers":  chainHeaders,
			"Rows":     chainRows,
		},
	})
}

func (g *Gateway) slashingHandler(c *gin.Context) {
	menuItems := makeMenuItems(2)

	index := g.slashingIndex.Get()

	subtitle := fmt.Sprintf("Current tip set key: %v", index.TipSetKey)

	headers := []string{"Miner", "Slashed Epochs"}

	rows := make([][]interface{}, len(index.Miners))
	i := 0
	for id, slashes := range index.Miners {
		epochs := make([]string, len(slashes.Epochs))
		for j, epoch := range slashes.Epochs {
			epochs[j] = string(epoch)
		}
		rows[i] = []interface{}{
			id,
			strings.Join(epochs, ", "),
		}
		i++
	}

	c.HTML(http.StatusOK, "/public/html/slashing.gohtml", gin.H{
		"MenuItems": menuItems,
		"Title":     "Miner Slashes",
		"Subtitle":  subtitle,
		"Headers":   headers,
		"Rows":      rows,
	})
}

func (g *Gateway) reputationHandler(c *gin.Context) {
	menuItems := makeMenuItems(3)

	topMiners, err := g.reputationModule.GetTopMiners(30)
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
		"Title":     "Top 30 Miners",
		"Headers":   headers,
		"Rows":      rows,
	})
}

func uint64ToTime(value uint64) time.Time {
	return time.Unix(int64(value), 0)
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
			Name:     "Slashing",
			Path:     "slashing",
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

// // bucketHandler renders a bucket as a website.
// func (g *Gateway) bucketHandler(c *gin.Context) {
// 	log.Debugf("host: %s", c.Request.Host)

// 	buckName, err := bucketFromHost(c.Request.Host, g.bucketDomain)
// 	if err != nil {
// 		abort(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
// 	defer cancel()
// 	rep, err := g.client.ListBucketPath(ctx, "", buckName, g.clientAuth)
// 	if err != nil {
// 		abort(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	for _, item := range rep.Item.Items {
// 		if item.Name == "index.html" {
// 			c.Writer.WriteHeader(http.StatusOK)
// 			c.Writer.Header().Set("Content-Type", "text/html")
// 			pth := strings.Replace(item.Path, rep.Root.Path, rep.Root.Name, 1)
// 			if err := g.client.PullBucketPath(ctx, pth, c.Writer, g.clientAuth); err != nil {
// 				abort(c, http.StatusInternalServerError, err)
// 			}
// 			return
// 		}
// 	}

// 	// No index was found, use the default (404 for now)
// 	g.render404(c)
// }

// type link struct {
// 	Name  string
// 	Path  string
// 	Size  string
// 	Links string
// }

// // dashHandler renders a project dashboard.
// // Currently, this just shows bucket files and directories.
// func (g *Gateway) dashHandler(c *gin.Context) {
// 	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
// 	defer cancel()

// 	project := c.Param("project")
// 	rep, err := g.client.ListBucketPath(ctx, project, c.Param("path"), g.clientAuth)
// 	if err != nil {
// 		abort(c, http.StatusNotFound, fmt.Errorf("project not found"))
// 		return
// 	}

// 	if !rep.Item.IsDir {
// 		if err := g.client.PullBucketPath(ctx, c.Param("path"), c.Writer, g.clientAuth); err != nil {
// 			abort(c, http.StatusInternalServerError, err)
// 		}
// 	} else {
// 		projectPath := path.Join("dashboard", project)

// 		links := make([]link, len(rep.Item.Items))
// 		for i, item := range rep.Item.Items {
// 			var pth string
// 			if rep.Root != nil {
// 				pth = strings.Replace(item.Path, rep.Root.Path, rep.Root.Name, 1)
// 			} else {
// 				pth = item.Name
// 			}

// 			links[i] = link{
// 				Name:  item.Name,
// 				Path:  path.Join(projectPath, pth),
// 				Size:  byteCountDecimal(item.Size),
// 				Links: strconv.Itoa(len(item.Items)),
// 			}
// 		}

// 		var root, back string
// 		if rep.Root != nil {
// 			root = strings.Replace(rep.Item.Path, rep.Root.Path, rep.Root.Name, 1)
// 		} else {
// 			root = ""
// 		}
// 		if root == "" {
// 			back = projectPath
// 		} else {
// 			back = path.Dir(path.Join(projectPath, root))
// 		}
// 		c.HTML(http.StatusOK, "/public/html/buckets.gohtml", gin.H{
// 			"Path":  rep.Item.Path,
// 			"Root":  root,
// 			"Back":  back,
// 			"Links": links,
// 		})
// 	}
// }

// // confirmEmail verifies an emailed secret.
// func (g *Gateway) confirmEmail(c *gin.Context) {
// 	if err := g.sessionBus.Send(g.parseUUID(c, c.Param("secret"))); err != nil {
// 		g.renderError(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	c.HTML(http.StatusOK, "/public/html/confirm.gohtml", nil)
// }

// // consentInvite adds a user to a team.
// func (g *Gateway) consentInvite(c *gin.Context) {
// 	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
// 	defer cancel()

// 	invite, err := g.collections.Invites.Get(ctx, g.parseUUID(c, c.Param("invite")))
// 	if err != nil {
// 		g.render404(c)
// 		return
// 	}
// 	if invite.Expiry < int(time.Now().Unix()) {
// 		g.renderError(c, http.StatusPreconditionFailed, fmt.Errorf("this invitation has expired"))
// 		return
// 	}

// 	dev, err := g.collections.Developers.GetOrCreateByEmail(ctx, invite.ToEmail)
// 	if err != nil {
// 		g.renderError(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	team, err := g.collections.Teams.Get(ctx, invite.TeamID)
// 	if err != nil {
// 		g.render404(c)
// 		return
// 	}
// 	if err = g.collections.Developers.JoinTeam(ctx, dev, team.ID); err != nil {
// 		g.renderError(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	c.HTML(http.StatusOK, "/public/html/consent.gohtml", gin.H{
// 		"Team": team.Name,
// 	})
// }

// type registrationParams struct {
// 	Token    string `json:"token" binding:"required"`
// 	DeviceID string `json:"device_id" binding:"required"`
// }

// // registerUser adds a user to a team.
// func (g *Gateway) registerUser(c *gin.Context) {
// 	var params registrationParams
// 	err := c.BindJSON(&params)
// 	if err != nil {
// 		abort(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
// 	defer cancel()

// 	token, err := g.collections.Tokens.Get(ctx, params.Token)
// 	if err != nil {
// 		abort(c, http.StatusNotFound, fmt.Errorf("token not found"))
// 		return
// 	}
// 	proj, err := g.collections.Projects.Get(ctx, token.ProjectID)
// 	if err != nil {
// 		abort(c, http.StatusNotFound, fmt.Errorf("project not found"))
// 		return
// 	}
// 	user, err := g.collections.Users.GetOrCreate(ctx, proj.ID, params.DeviceID)
// 	if err != nil {
// 		abort(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	session, err := g.collections.Sessions.Create(ctx, user.ID, user.ID)
// 	if err != nil {
// 		abort(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"id":         user.ID,
// 		"session_id": session.ID,
// 	})
// }

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

// abort the request with code and error.
func abort(c *gin.Context, code int, err error) {
	c.AbortWithStatusJSON(code, gin.H{
		"error": err.Error(),
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

// byteCountDecimal formats bytes
func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// parseUUID parses a string as a UUID, adding back hyphens.
func (g *Gateway) parseUUID(c *gin.Context, param string) (parsed string) {
	id, err := uuid.Parse(param)
	if err != nil {
		g.render404(c)
		return
	}
	return id.String()
}
