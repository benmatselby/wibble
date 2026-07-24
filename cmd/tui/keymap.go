package tui

import "charm.land/bubbles/v2/key"

// KeyMap defines all keybindings for the TUI. Embed it in the model to allow
// per-instance customisation and to drive the help bar from binding metadata.
type KeyMap struct {
	// Global
	Quit       key.Binding
	ListDown   key.Binding
	ListUp     key.Binding
	ListFilter key.Binding

	// Feeds pane
	OpenFeed       key.Binding
	MarkAllAsRead  key.Binding
	ToggleTagsPane key.Binding

	// Tags pane
	OpenTag key.Binding

	// Articles pane
	MarkAsRead  key.Binding
	ViewArticle key.Binding
	OpenArticle key.Binding
	AddTag      key.Binding
	RemoveTag   key.Binding
	Back        key.Binding
}

// DefaultKeyMap is the out-of-the-box keybinding configuration.
var DefaultKeyMap = KeyMap{
	// Global
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	ListDown: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "down"),
	),
	ListUp: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "up"),
	),
	ListFilter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),

	// Feeds pane
	OpenFeed: key.NewBinding(
		key.WithKeys("enter", "right", "l"),
		key.WithHelp("enter/→/l", "open feed"),
	),
	MarkAllAsRead: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "mark feed as read"),
	),
	ToggleTagsPane: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "toggle tags"),
	),

	// Tags pane
	OpenTag: key.NewBinding(
		key.WithKeys("enter", "right", "l"),
		key.WithHelp("enter/→/l", "open tag"),
	),

	// Articles pane
	MarkAsRead: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "mark article as read"),
	),
	ViewArticle: key.NewBinding(
		key.WithKeys("enter", "right", "l"),
		key.WithHelp("enter/→/l", "view article"),
	),
	OpenArticle: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open link"),
	),
	AddTag: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "add tag"),
	),
	RemoveTag: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "remove tag"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "left", "h", "backspace"),
		key.WithHelp("esc/←/h", "back"),
	),
}
