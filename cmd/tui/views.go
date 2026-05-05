package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

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
	case paneArticle:
		// noop
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

	// -- Help view -------------------------------------------------------------
	if m.showHelp {
		v = tea.NewView(lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.renderHelpModal(),
		))
		v.AltScreen = true
	}

	// -- View article view -----------------------------------------------------
	if m.focusedPane == paneArticle {
		v = tea.NewView(lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.renderArticleModal(),
		))
		v.AltScreen = true
	}

	return v
}

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

// renderArticleModal renders the current article for viewing in the app.
func (m model) renderArticleModal() string {
	article, err := m.db.GetArticleByID(m.currentArticleID)
	if err != nil {
		return err.Error()
	}

	theme := "light"
	if m.isDark {
		theme = "dark"
	}

	converter := md.NewConverter("", true, nil)
	markdown, _ := converter.ConvertString(article.Summary)

	out, _ := glamour.Render(markdown, theme)

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.styles.focusedTitle.Render(article.Title),
		out,
	)

	return m.styles.focusedBorder.Padding(1, 3).Render(content)
}
