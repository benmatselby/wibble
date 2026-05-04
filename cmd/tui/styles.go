package tui

import (
	"io"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/benmatselby/wibble/pkg/theme"
)

// styles holds all lipgloss styles derived from a Theme.
type styles struct {
	focusedTitle    lipgloss.Style
	unfocusedTitle  lipgloss.Style
	focusedBorder   lipgloss.Style
	unfocusedBorder lipgloss.Style
	help            lipgloss.Style
	statusInfo      lipgloss.Style
	statusError     lipgloss.Style
	listItem        list.DefaultItemStyles
	filterPrompt    lipgloss.Style
	filterCursor    lipgloss.Style
	readItemTitle   lipgloss.Style
}

// newStyles builds a styles set from the provided Theme, adapting to the
// terminal background (dark/light).
func newStyles(t theme.Theme, hasDarkBackground bool) styles {
	ld := lipgloss.LightDark(hasDarkBackground)

	focusedColor := ld(lipgloss.Color(t.Light.FocusedColor), lipgloss.Color(t.Dark.FocusedColor))
	unfocusedColor := ld(lipgloss.Color(t.Light.UnfocusedColor), lipgloss.Color(t.Dark.UnfocusedColor))
	focusedTitleTextColor := ld(lipgloss.Color(t.Light.FocusedTitleTextColor), lipgloss.Color(t.Dark.FocusedTitleTextColor))
	unfocusedTitleTextColor := ld(lipgloss.Color(t.Light.UnfocusedTitleTextColor), lipgloss.Color(t.Dark.UnfocusedTitleTextColor))
	helpColor := ld(lipgloss.Color(t.Light.HelpColor), lipgloss.Color(t.Dark.HelpColor))

	normalTitle := ld(lipgloss.Color(t.Light.NormalTitleColor), lipgloss.Color(t.Dark.NormalTitleColor))
	selectedTitle := ld(lipgloss.Color(t.Light.SelectedTitleColor), lipgloss.Color(t.Dark.SelectedTitleColor))
	readTitle := ld(lipgloss.Color(t.Light.ReadTitleColor), lipgloss.Color(t.Dark.ReadTitleColor))

	itemStyles := list.NewDefaultItemStyles(hasDarkBackground)
	itemStyles.NormalTitle = itemStyles.NormalTitle.Foreground(normalTitle)
	itemStyles.SelectedTitle = itemStyles.SelectedTitle.
		Foreground(selectedTitle).
		BorderLeftForeground(selectedTitle)

	return styles{
		focusedTitle: lipgloss.NewStyle().
			Foreground(focusedTitleTextColor).
			Background(focusedColor).
			Padding(0, 1),

		unfocusedTitle: lipgloss.NewStyle().
			Foreground(unfocusedTitleTextColor).
			Background(unfocusedColor).
			Padding(0, 1),

		focusedBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(focusedColor).
			Padding(0, 1),

		unfocusedBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(unfocusedColor).
			Padding(0, 1),

		help: lipgloss.NewStyle().Foreground(helpColor).Margin(0, 1),

		statusInfo: lipgloss.NewStyle().
			Foreground(helpColor).
			Margin(0, 1),

		statusError: lipgloss.NewStyle().
			Foreground(ld(lipgloss.Color("#cc0000"), lipgloss.Color("#ff6666"))).
			Margin(0, 1),

		listItem: itemStyles,

		filterPrompt: lipgloss.NewStyle().
			Foreground(ld(lipgloss.Color(t.Light.FocusedColor), lipgloss.Color(t.Dark.FocusedColor))),

		filterCursor: lipgloss.NewStyle().
			Foreground(ld(lipgloss.Color(t.Light.FocusedColor), lipgloss.Color(t.Dark.FocusedColor))),

		readItemTitle: lipgloss.NewStyle().Foreground(readTitle),
	}
}

// readableItem is implemented by any list item that has a read/unread state.
type readableItem interface {
	IsRead() bool
}

// readableDelegate is a list delegate that applies ReadTitleColor to items
// that have already been read.
type readableDelegate struct {
	list.DefaultDelegate
	readItemTitleColor lipgloss.Style
}

// Render prints an item, styling read items with readTitleStyle unless the
// item is currently selected (in which case the selection style takes priority).
func (d readableDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ri, ok := item.(readableItem)
	if !ok || !ri.IsRead() || index == m.Index() {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	// Read, unselected item: override NormalTitle with readTitleStyle,
	// preserving the padding so layout is unchanged.
	d.Styles.NormalTitle = d.readItemTitleColor.
		Padding(d.Styles.NormalTitle.GetPadding())
	d.DefaultDelegate.Render(w, m, index, item)
}

func configureFilterStyles(filterStyles textinput.Styles, s styles) textinput.Styles {
	filterStyles.Focused.Prompt = s.unfocusedTitle.UnsetBackground()
	filterStyles.Focused.Text = s.unfocusedTitle.UnsetBackground()
	filterStyles.Blurred.Prompt = s.focusedTitle
	filterStyles.Blurred.Text = s.focusedTitle.UnsetBackground()
	filterStyles.Cursor.Color = s.filterCursor.GetForeground()

	return filterStyles
}
