package summary

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/olekukonko/tablewriter"
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	headerHeight = 0
	footerHeight = 2
)

var (
	term = termenv.ColorProfile()
)

type model struct {
	cids           []string
	summary        []*userPb.CidSummary
	summaryYOffset int
	cidInfo        *userPb.CidInfo
	err            error
	cursor         int
	viewport       viewport.Model
	ready          bool
}

func (m model) Init() tea.Cmd {
	return getSummaryCmd(m.cids)
}

func getSummaryCmd(cids []string) tea.Cmd {
	return func() tea.Msg {
		res, err := getSummary(cids)
		if err != nil {
			return errMsg{err}
		}
		return res
	}
}

func getInfo(cid string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Data.CidInfo(c.MustAuthCtx(ctx), cid)
		if err != nil {
			return errMsg{err}
		}
		return res.CidInfo
	}
}

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.viewport, _ = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case *userPb.CidSummaryResponse:
		m.summary = msg.CidSummary
		m.viewport.SetContent(m.getViewportContent())
		return m, nil

	case *userPb.CidInfo:
		m.cidInfo = msg
		m.viewport.SetContent(m.getViewportContent())
		m.viewport.YOffset = 0
		return m, nil

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.WindowSizeMsg:
		verticalMargins := headerHeight + footerHeight
		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - verticalMargins}
			m.viewport.SetContent(m.getViewportContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargins
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyUp:
			if m.cidInfo == nil && m.cursor > 0 {
				m.cursor--
				m.summaryYOffset = m.viewport.YOffset
				m.viewport.SetContent(m.getViewportContent())
			}

		case tea.KeyDown:
			if m.cidInfo == nil && m.cursor < len(m.summary)-1 {
				m.cursor++
				m.summaryYOffset = m.viewport.YOffset
				m.viewport.SetContent(m.getViewportContent())
			}

		case tea.KeyLeft:
			m.cidInfo = nil
			m.viewport.SetContent(m.getViewportContent())
			m.viewport.YOffset = m.summaryYOffset

		case tea.KeyEnter, tea.KeySpace:
			if len(m.summary) > 0 {
				return m, getInfo(m.summary[m.cursor].Cid)
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	footerTop := termenv.String(strings.Repeat("─", m.viewport.Width)).Foreground(term.Color("241")).String()
	footerBottomLeft := m.getFooterContent()
	fits := m.viewport.AtTop() && (m.viewport.AtBottom() || m.viewport.PastBottom())
	footerBottomRight := fmt.Sprintf("%3.f%% ", m.viewport.ScrollPercent()*100)
	if fits {
		footerBottomRight = ""
	}
	gapSize := m.viewport.Width - (runewidth.StringWidth(footerBottomLeft) + runewidth.StringWidth(footerBottomRight))
	footerBottom := termenv.String(footerBottomLeft + strings.Repeat(" ", gapSize) + footerBottomRight).Foreground(term.Color("241")).String()

	footer := fmt.Sprintf("%s\n%s", footerTop, footerBottom)

	return fmt.Sprintf("%s\n%s", m.viewport.View(), footer)
}

func (m model) getViewportContent() string {
	if m.cidInfo != nil {
		return renderCidInfo(m.cidInfo)
	} else if m.summary != nil {
		return renderSummary(m.summary, m.cursor)
	}
	return "Loading..."
}

func (m model) getFooterContent() string {
	if m.cidInfo != nil {
		return " ←: Back • Ctrl+C: Quit"
	} else if m.summary != nil {
		return " ↑/↓: Select • Enter: View Seleted • Ctrl+C: Quit"
	}
	return "Ctrl+C: Quit"
}

func renderCidInfo(cidInfo *userPb.CidInfo) string {
	json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(cidInfo)
	if err != nil {
		return fmt.Sprintf("error marshaling json: %v", err)
	}
	return fmt.Sprintf("%s\n", string(json))
}

func renderSummary(summary []*userPb.CidSummary, cursor int) string {
	s := &strings.Builder{}

	// Set table header
	table := tablewriter.NewWriter(s)
	table.SetHeader([]string{"Cid", "Stored", "Executing Job", "Queued Jobs"})
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})

	// Add data to table
	for i, v := range summary {
		data := []string{
			v.Cid,
			fmt.Sprintf("%v", v.Stored),
			fmt.Sprintf("%v", len(v.ExecutingJob) > 0),
			fmt.Sprintf("%v", len(v.QueuedJobs)),
		}
		if cursor == i {
			// Color active item
			c := tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.BgWhiteColor}
			var colors []tablewriter.Colors
			for range data {
				colors = append(colors, c)
			}
			table.Rich(data, colors)
		} else {
			table.Append(data)
		}
	}

	table.Render()
	return s.String()
}
