// Package tui implements the terminal user interface
package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/benmatselby/wibble/pkg/client"
	"github.com/benmatselby/wibble/pkg/config"
	"github.com/benmatselby/wibble/pkg/dao"
	"github.com/benmatselby/wibble/pkg/feed"
	"github.com/benmatselby/wibble/pkg/theme"
)

// ── Pane ──────────────────────────────────────────────────────────────────────

type pane int

const (
	paneFeeds pane = iota
	paneArticles
	paneArticle
)

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	db               dao.DaoClient
	feedsList        *list.Model
	articlesList     *list.Model
	articlesTitle    string
	readableDelegate readableDelegate
	focusedPane      pane
	currentFeedID    int64
	currentArticleID int64
	isDark           bool
	theme            theme.Theme
	styles           styles
	keys             KeyMap
	width            int
	height           int
	ready            bool
	showHelp         bool
	status           *statusMsg
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.styles = newStyles(m.theme, m.isDark)
		m.readableDelegate.Styles = m.styles.listItem
		m.readableDelegate.readItemTitleColor = m.styles.readItemTitle
		m.feedsList.SetDelegate(m.readableDelegate)
		m.articlesList.SetDelegate(m.readableDelegate)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resize()
		return m, nil

	case articlesLoadedMsg:
		return handleArticlesLoaded(msg, m)

	case feedsLoadedMsg:
		return handleFeedsLoaded(msg, m)

	case statusMsg:
		return handleStatusMessage(m, msg)

	case tea.KeyPressMsg:
		m1, c, shouldReturn := handleKeypress(msg, m)
		if shouldReturn {
			return m1, c
		}
	}

	// Route list updates to the focused pane
	var cmd tea.Cmd
	switch m.focusedPane {
	case paneFeeds:
		newList, c := m.feedsList.Update(msg)
		*m.feedsList = newList
		cmd = c

	case paneArticles:
		newList, c := m.articlesList.Update(msg)
		*m.articlesList = newList
		cmd = c
	case paneArticle:
		// noop
	}

	return m, cmd
}

func (m *model) resize() {
	frameHeight := 3                          // top margin (1) + border top/bottom (2)
	availHeight := m.height - frameHeight - 3 // subtract help line + status bar line

	feedsWidth, articlesWidth := m.panelWidths()

	m.feedsList.SetSize(feedsWidth-4, availHeight)
	m.articlesList.SetSize(articlesWidth-4, availHeight)
}

// panelWidths returns (feedsWidth, articlesWidth) based on terminal width.
func (m model) panelWidths() (int, int) {
	if m.width == 0 {
		return 40, 60
	}

	// Leave a little breathing room for borders
	total := m.width - 2
	feedWidth := total * 35 / 100
	feedWidth = max(feedWidth, 20)
	articleWidth := total - feedWidth
	return feedWidth, articleWidth
}

// ── Entry point ───────────────────────────────────────────────────────────────

// Run is the entry point to the TUI.
func Run(db dao.DaoClient, rssClient client.API, t theme.Theme) error {
	feeds, err := db.GetFeeds()
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	feedItems := make([]list.Item, len(feeds))
	for i, f := range feeds {
		feedItems[i] = feedItem{feed: f}
	}

	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	s := newStyles(t, isDark)

	listDelegate := list.NewDefaultDelegate()
	listDelegate.ShowDescription = false
	listDelegate.SetSpacing(0)
	listDelegate.SetHeight(1)

	rd := readableDelegate{
		DefaultDelegate:    listDelegate,
		readItemTitleColor: s.readItemTitle,
	}
	rd.Styles = s.listItem

	feedsList := list.New(feedItems, rd, 0, 0)
	feedsList.SetShowHelp(false)
	feedsList.SetShowTitle(false)
	feedsList.SetFilteringEnabled(true)
	feedsList.Styles.Title = s.focusedTitle
	feedsFilterStyles := feedsList.FilterInput.Styles()
	feedsList.FilterInput.SetStyles(configureFilterStyles(feedsFilterStyles, s))

	articlesList := list.New([]list.Item{}, rd, 0, 0)
	articlesList.SetShowHelp(false)
	articlesList.SetShowTitle(false)
	articlesList.SetFilteringEnabled(true)
	articlesList.Styles.Title = s.focusedTitle
	articlesFilterStyles := articlesList.FilterInput.Styles()
	articlesList.FilterInput.SetStyles(configureFilterStyles(articlesFilterStyles, s))

	m := model{
		db:               db,
		feedsList:        &feedsList,
		articlesList:     &articlesList,
		articlesTitle:    "Select a feed →",
		readableDelegate: rd,
		focusedPane:      paneFeeds,
		isDark:           isDark,
		theme:            t,
		styles:           s,
		keys:             DefaultKeyMap,
	}

	p := tea.NewProgram(m)

	// syncAndReport runs a full sync cycle, reporting per-feed progress via the
	// status bar. Errors auto-dismiss after 20 seconds (handled in Update).
	syncAndReport := func() {
		p.Send(statusMsg{text: "Syncing feeds...", level: statusInfo})
		err := feed.Sync(db, rssClient, func(title string, syncErr error) {
			if syncErr != nil {
				p.Send(statusMsg{
					text:  fmt.Sprintf("Error syncing %q: %v", title, syncErr),
					level: statusError,
				})
			} else {
				p.Send(statusMsg{text: fmt.Sprintf("Syncing: %s", title), level: statusInfo})
				p.Send(fetchFeeds(db)())
			}
		})
		if err != nil {
			p.Send(statusMsg{text: fmt.Sprintf("Sync failed: %v", err), level: statusError})
		} else {
			when := fmt.Sprintf("Sync completed at %s", time.Now().Format("15:04:05"))
			p.Send(statusMsg{text: when, level: statusInfo})
		}
	}

	// Sync immediately on startup, then every N minutes.
	ctx, cancel := context.WithCancel(context.Background())
	go syncAndReport()
	go func() {
		ticker := time.NewTicker(time.Duration(config.GetRefreshInterval()) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				syncAndReport()
			case <-ctx.Done():
				return
			}
		}
	}()

	_, err = p.Run()
	cancel()
	return err
}
