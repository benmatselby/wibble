package tui

import (
	"fmt"
	"os"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

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
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

	theme := "light"
	if isDark {
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
