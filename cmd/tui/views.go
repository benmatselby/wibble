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

	leftPanelWidth, rightPanelWidth := m.panelWidths()

	var help string
	var panels string
	switch m.focusedPane {
	case paneFeeds:
		panels = lipgloss.JoinHorizontal(
			lipgloss.Top,
			renderPanel(m.styles.focusedTitle, m.styles.focusedBorder, leftPanelWidth, "Feeds", m.feedsList.View()),
			renderPanel(m.styles.unfocusedTitle, m.styles.unfocusedBorder, rightPanelWidth, "Articles", m.articlesList.View()),
		)
		help = m.styles.help.Render(fmt.Sprintf(
			"%s %s • %s %s • %s %s • %s %s • %s %s",
			fmt.Sprintf("%s/%s", m.keys.ListDown.Help().Key, m.keys.ListUp.Help().Key), "navigation",
			m.keys.ListFilter.Help().Key, m.keys.ListFilter.Help().Desc,
			m.keys.OpenFeed.Help().Key, m.keys.OpenFeed.Help().Desc,
			m.keys.Help.Help().Key, m.keys.Help.Help().Desc,
			m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
		))
	case paneArticles:
		panels = lipgloss.JoinHorizontal(
			lipgloss.Top,
			renderPanel(m.styles.unfocusedTitle, m.styles.unfocusedBorder, leftPanelWidth, "Feeds", m.feedsList.View()),
			renderPanel(m.styles.focusedTitle, m.styles.focusedBorder, rightPanelWidth, "Articles", m.articlesList.View()),
		)

		help = m.styles.help.Render(fmt.Sprintf(
			"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
			fmt.Sprintf("%s/%s", m.keys.ListDown.Help().Key, m.keys.ListUp.Help().Key), "navigation",
			m.keys.ListFilter.Help().Key, m.keys.ListFilter.Help().Desc,
			m.keys.Back.Help().Key, m.keys.Back.Help().Desc,
			m.keys.OpenArticle.Help().Key, m.keys.OpenArticle.Help().Desc,
			m.keys.Help.Help().Key, m.keys.Help.Help().Desc,
			m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
		))
	case paneArticle:
		title, content := m.renderArticleModal()
		panels = lipgloss.JoinHorizontal(
			lipgloss.Top,
			renderPanel(m.styles.unfocusedTitle, m.styles.unfocusedBorder, leftPanelWidth, "Feeds", m.feedsList.View()),
			renderPanel(m.styles.focusedTitle, m.styles.focusedBorder, rightPanelWidth, title, content),
		)

		help = m.styles.help.Render(fmt.Sprintf(
			"%s %s • %s %s •  %s %s • %s %s",
			m.keys.Back.Help().Key, m.keys.Back.Help().Desc,
			m.keys.OpenArticle.Help().Key, m.keys.OpenArticle.Help().Desc,
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

	// -- Help view -------------------------------------------------------------
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

// renderPanel aims to essentially ensure all panels look the same, and make
// the code cleaner
func renderPanel(titleStyle, borderStyle lipgloss.Style, width int, title, content string) string {
	panel := borderStyle.
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, titleStyle.Width(width-4).Render(title), content))

	return panel
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
func (m model) renderArticleModal() (string, string) {
	article, err := m.db.GetArticleByID(m.currentArticleID)
	if err != nil {
		return "Error", err.Error()
	}

	theme := "light"
	if m.isDark {
		theme = "dark"
	}

	converter := md.NewConverter("", true, nil)
	markdown, _ := converter.ConvertString(article.Summary)

	out, _ := glamour.Render(markdown, theme)

	return article.Title, out
}
