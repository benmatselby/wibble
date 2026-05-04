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
	"github.com/benmatselby/wibble/pkg/models"
	"github.com/benmatselby/wibble/pkg/theme"
)

// ── Pane ──────────────────────────────────────────────────────────────────────

type pane int

const (
	paneFeeds pane = iota
	paneArticles
)

// ── Messages ──────────────────────────────────────────────────────────────────

type articlesLoadedMsg struct {
	articles []models.Article
	err      error
}

type feedsLoadedMsg struct {
	feeds []models.Feed
	err   error
}

type statusLevel int

const (
	statusInfo  statusLevel = iota
	statusError statusLevel = iota
)

// statusMsg sets the text shown in the status bar.
type statusMsg struct {
	text  string
	level statusLevel
}

// clearStatusMsg clears the status bar. A version field guards against stale
// timers clearing a newer message.
type clearStatusMsg struct {
	version int
}

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	db               dao.DaoClient
	feedsList        *list.Model
	articlesList     *list.Model
	articlesTitle    string
	readableDelegate readableDelegate
	focusedPane      pane
	currentFeedID    int64
	theme            theme.Theme
	styles           styles
	keys             KeyMap
	width            int
	height           int
	ready            bool
	showHelp         bool
	status           *statusMsg
	statusVersion    int
}

func (m model) Init() tea.Cmd {
	return nil
}

// fetchArticles is a Bubbletea command that loads articles for a given feed.
func fetchArticles(db dao.DaoClient, feedID int64) tea.Cmd {
	return func() tea.Msg {
		articles, err := db.GetArticlesByFeedID(feedID)
		return articlesLoadedMsg{articles: articles, err: err}
	}
}

// fetchFeeds is a Bubbletea command that reloads all feeds from the database.
func fetchFeeds(db dao.DaoClient) tea.Cmd {
	return func() tea.Msg {
		feeds, err := db.GetFeeds()
		return feedsLoadedMsg{feeds: feeds, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.styles = newStyles(m.theme, lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
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

	case clearStatusMsg:
		return handleClearStatusMessage(msg, m)

	case tea.KeyPressMsg:
		// Global quit
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

func (m model) View() tea.View {
	if !m.ready {
		return tea.NewView("\n  Loading...")
	}

	feedsW, articlesW := m.panelWidths()

	// ── Feeds panel ───────────────────────────────────────────────────────────
	feedsTitle := m.styles.focusedTitle.Width(feedsW - 4).Render("Feeds")
	feedsBorder := m.styles.focusedBorder
	if m.focusedPane != paneFeeds {
		feedsTitle = m.styles.unfocusedTitle.Width(feedsW - 4).Render("Feeds")
		feedsBorder = m.styles.unfocusedBorder
	}
	feedsContent := feedsBorder.
		Width(feedsW).
		Render(lipgloss.JoinVertical(lipgloss.Left, feedsTitle, m.feedsList.View()))

	// ── Articles panel ────────────────────────────────────────────────────────
	articlesTitle := m.styles.unfocusedTitle.Width(articlesW - 4).Render(m.articlesTitle)
	articlesBorder := m.styles.unfocusedBorder
	if m.focusedPane == paneArticles {
		articlesTitle = m.styles.focusedTitle.Width(articlesW - 4).Render(m.articlesTitle)
		articlesBorder = m.styles.focusedBorder
	}
	articlesContent := articlesBorder.
		Width(articlesW).
		Render(lipgloss.JoinVertical(lipgloss.Left, articlesTitle, m.articlesList.View()))

	// ── Layout ────────────────────────────────────────────────────────────────
	panels := lipgloss.JoinHorizontal(lipgloss.Top, feedsContent, articlesContent)

	// ── Help bar ──────────────────────────────────────────────────────────────
	var help string
	switch m.focusedPane {
	case paneFeeds:
		help = m.styles.help.Render(fmt.Sprintf(
			"j/k navigate • %s %s • / filter • %s %s • %s %s",
			m.keys.OpenFeed.Help().Key, m.keys.OpenFeed.Help().Desc,
			m.keys.Help.Help().Key, m.keys.Help.Help().Desc,
			m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
		))
	case paneArticles:
		help = m.styles.help.Render(fmt.Sprintf(
			"j/k navigate • %s %s • %s %s • / filter • %s %s • %s %s",
			m.keys.OpenArticle.Help().Key, m.keys.OpenArticle.Help().Desc,
			m.keys.Back.Help().Key, m.keys.Back.Help().Desc,
			m.keys.Help.Help().Key, m.keys.Help.Help().Desc,
			m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
		))
	}

	// ── Status bar ────────────────────────────────────────────────────────────
	var statusBar string
	if m.status != nil {
		style := m.styles.statusInfo
		if m.status.level == statusError {
			style = m.styles.statusError
		}
		statusBar = style.Render(m.status.text)
	}

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, panels, help, statusBar))
	v.AltScreen = true

	if m.showHelp {
		v = tea.NewView(lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.renderHelpModal(),
		))
		v.AltScreen = true
	}

	return v
}

// ── Entry point ───────────────────────────────────────────────────────────────

// renderHelpModal returns a styled modal box listing all keybindings.
func (m model) renderHelpModal() string {
	k := m.keys

	row := func(key, desc string) string {
		return fmt.Sprintf("  %-14s %s", key, desc)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.styles.focusedTitle.Render("Keymaps"),
		"",
		m.styles.help.Render("Global"),
		row(k.Quit.Help().Key, k.Quit.Help().Desc),
		row(k.Help.Help().Key, k.Help.Help().Desc),
		"",
		m.styles.help.Render("Feeds pane"),
		row("j/k/↑/↓", "navigate"),
		row("/", "filter"),
		row(k.OpenFeed.Help().Key, k.OpenFeed.Help().Desc),
		row(k.MarkAllAsRead.Help().Key, k.MarkAllAsRead.Help().Desc),
		"",
		m.styles.help.Render("Articles pane"),
		row("j/k/↑/↓", "navigate"),
		row("/", "filter"),
		row(k.OpenArticle.Help().Key, k.OpenArticle.Help().Desc),
		row(k.MarkAsRead.Help().Key, k.MarkAsRead.Help().Desc),
		row(k.Back.Help().Key, k.Back.Help().Desc),
		"",
		m.styles.help.Render("Press ? or esc to close"),
	)

	return m.styles.focusedBorder.Padding(1, 3).Render(content)
}

func Run(db dao.DaoClient, rssClient client.API, t theme.Theme) error {
	feeds, err := db.GetFeeds()
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	feedItems := make([]list.Item, len(feeds))
	for i, f := range feeds {
		feedItems[i] = feedItem{feed: f}
	}

	s := newStyles(t, lipgloss.HasDarkBackground(os.Stdin, os.Stdout))

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
			p.Send(statusMsg{text: "Sync complete", level: statusInfo})
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
