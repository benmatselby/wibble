package tui

import "charm.land/bubbles/v2/key"

// KeyMap defines all keybindings for the TUI. Embed it in the model to allow
// per-instance customisation and to drive the help bar from binding metadata.
type KeyMap struct {
	// Global
	Quit key.Binding
	Help key.Binding

	// Feeds pane
	OpenFeed      key.Binding
	MarkAllAsRead key.Binding

	// Articles pane
	MarkAsRead  key.Binding
	OpenArticle key.Binding
	Back        key.Binding
}

// DefaultKeyMap is the out-of-the-box keybinding configuration.
var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "show keymaps"),
	),
	OpenFeed: key.NewBinding(
		key.WithKeys("enter", "right", "l"),
		key.WithHelp("enter/→", "open feed"),
	),
	MarkAllAsRead: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "mark feed as read"),
	),
	MarkAsRead: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "mark article as read"),
	),
	OpenArticle: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open link"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "left", "h", "backspace"),
		key.WithHelp("esc/←", "back"),
	),
}
