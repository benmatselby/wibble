package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

// View is called after every update and its return value is written to the
// terminal. It's a pure function — it should only read from the model's state
// and produce a string representation of the UI, with no side effects.
func (m model) View() tea.View {
	if !m.ready {
		return tea.NewView("\n  Loading...")
	}

	leftPanelWidth, rightPanelWidth := m.panelWidths()

	leftTitle, leftContent := m.leftPanelTitleAndContent()

	var help string
	var panels string
	switch m.focusedPane {
	case paneFeeds:
		panels = lipgloss.JoinHorizontal(
			lipgloss.Top,
			renderPanel(m.styles.focusedTitle, m.styles.focusedBorder, leftPanelWidth, leftTitle, leftContent),
			renderPanel(m.styles.unfocusedTitle, m.styles.unfocusedBorder, rightPanelWidth, m.articlesTitle, m.articlesList.View()),
		)
		if m.leftPaneMode == leftPaneTags {
			help = m.styles.help.Render(fmt.Sprintf(
				"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
				fmt.Sprintf("%s/%s", m.keys.ListDown.Help().Key, m.keys.ListUp.Help().Key), "navigation",
				m.keys.ListFilter.Help().Key, m.keys.ListFilter.Help().Desc,
				m.keys.OpenTag.Help().Key, m.keys.OpenTag.Help().Desc,
				m.keys.DeleteTag.Help().Key, m.keys.DeleteTag.Help().Desc,
				m.keys.ToggleTagsPane.Help().Key, "show feeds",
				m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
			))
		} else {
			help = m.styles.help.Render(fmt.Sprintf(
				"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
				fmt.Sprintf("%s/%s", m.keys.ListDown.Help().Key, m.keys.ListUp.Help().Key), "navigation",
				m.keys.ListFilter.Help().Key, m.keys.ListFilter.Help().Desc,
				m.keys.OpenFeed.Help().Key, m.keys.OpenFeed.Help().Desc,
				m.keys.MarkAllAsRead.Help().Key, m.keys.MarkAllAsRead.Help().Desc,
				m.keys.ToggleTagsPane.Help().Key, "show tags",
				m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
			))
		}
	case paneArticles:
		panels = lipgloss.JoinHorizontal(
			lipgloss.Top,
			renderPanel(m.styles.unfocusedTitle, m.styles.unfocusedBorder, leftPanelWidth, leftTitle, leftContent),
			renderPanel(m.styles.focusedTitle, m.styles.focusedBorder, rightPanelWidth, m.articlesTitle, m.articlesList.View()),
		)

		if m.addingTag {
			help = m.styles.help.Render(fmt.Sprintf("tag name: %s (enter to confirm, esc to cancel)", m.tagInput.View()))
		} else {
			help = m.styles.help.Render(fmt.Sprintf(
				"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
				fmt.Sprintf("%s/%s", m.keys.ListDown.Help().Key, m.keys.ListUp.Help().Key), "navigation",
				m.keys.ListFilter.Help().Key, m.keys.ListFilter.Help().Desc,
				m.keys.Back.Help().Key, m.keys.Back.Help().Desc,
				m.keys.OpenArticle.Help().Key, m.keys.OpenArticle.Help().Desc,
				m.keys.MarkAsRead.Help().Key, m.keys.MarkAsRead.Help().Desc,
				m.keys.AddTag.Help().Key, m.keys.AddTag.Help().Desc,
				m.keys.RemoveTag.Help().Key, m.keys.RemoveTag.Help().Desc,
				m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
			))
		}
	case paneArticle:
		panels = lipgloss.JoinHorizontal(
			lipgloss.Top,
			renderPanel(m.styles.unfocusedTitle, m.styles.unfocusedBorder, leftPanelWidth, leftTitle, leftContent),
			renderPanel(m.styles.focusedTitle, m.styles.focusedBorder, rightPanelWidth, m.articlesTitle, m.articleViewport.View()),
		)

		if m.addingTag {
			help = m.styles.help.Render(fmt.Sprintf("tag name: %s (enter to confirm, esc to cancel)", m.tagInput.View()))
		} else {
			help = m.styles.help.Render(fmt.Sprintf(
				"%s %s • %s %s • %s %s • %s %s • %s %s • %s %s",
				fmt.Sprintf("%s/%s", m.keys.ListDown.Help().Key, m.keys.ListUp.Help().Key), "scroll",
				m.keys.Back.Help().Key, m.keys.Back.Help().Desc,
				m.keys.OpenArticle.Help().Key, m.keys.OpenArticle.Help().Desc,
				m.keys.AddTag.Help().Key, m.keys.AddTag.Help().Desc,
				m.keys.RemoveTag.Help().Key, m.keys.RemoveTag.Help().Desc,
				m.keys.Quit.Help().Key, m.keys.Quit.Help().Desc,
			))
		}
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
	return v
}

// leftPanelTitleAndContent returns the title and rendered list content for
// the left-hand panel, depending on whether it's currently showing Feeds or
// Tags.
func (m model) leftPanelTitleAndContent() (string, string) {
	if m.leftPaneMode == leftPaneTags {
		return "Tags", m.tagsList.View()
	}
	return "Feeds", m.feedsList.View()
}

// renderPanel aims to essentially ensure all panels look the same, and make
// the code cleaner
func renderPanel(titleStyle, borderStyle lipgloss.Style, width int, title, content string) string {
	panel := borderStyle.
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, titleStyle.Width(width-4).Render(title), content))

	return panel
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
	markdown, err := converter.ConvertString(article.Summary)
	if err != nil {
		return fmt.Sprintf("Error converting article content: %v", err)
	}

	out, err := glamour.Render(fmt.Sprintf("# %s\n%s", article.Title, markdown), theme)
	if err != nil {
		return fmt.Sprintf("Error rendering article content: %v", err)
	}

	return out
}
