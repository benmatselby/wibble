// Package theme provides functionality for loading and managing color themes
// for the TUI.
package theme

import (
	"github.com/spf13/viper"
)

// Default color values (Rose Pine).
const (
	// (Light Mode)
	DefaultLightFocusedColor            = "#b4637a"
	DefaultLightUnfocusedColor          = "#dfdad9"
	DefaultLightFocusedTitleTextColor   = "#dfdad9"
	DefaultLightUnfocusedTitleTextColor = "#b4637a"
	DefaultLightHelpColor               = "#797593"
	DefaultLightNormalTitleColor        = "#575279"
	DefaultLightSelectedTitleColor      = "#ea9d34"
	DefaultLightReadTitleColor          = "#cecacd"

	// (Dark Mode)
	DefaultDarkFocusedColor            = "#eb6f92"
	DefaultDarkUnfocusedColor          = "#44415a"
	DefaultDarkFocusedTitleTextColor   = "#44415a"
	DefaultDarkUnfocusedTitleTextColor = "#eb6f92"
	DefaultDarkHelpColor               = "#908caa"
	DefaultDarkNormalTitleColor        = "#e0def4"
	DefaultDarkSelectedTitleColor      = "#f6c177"
	DefaultDarkReadTitleColor          = "#56526e"
)

// ThemeColors holds the color hex values for a single mode (light or dark).
type ThemeColors struct {
	// FocusedColor is used for focused panel borders and title backgrounds.
	FocusedColor string
	// UnfocusedColor is used for unfocused panel borders and title backgrounds.
	UnfocusedColor string
	// FocusedTitleTextColor is used for the text on panel title bars.
	FocusedTitleTextColor string
	// UnfocusedTitleTextColor is used for the text on panel title bars.
	UnfocusedTitleTextColor string
	// HelpColor is used for the help bar at the bottom of the screen.
	HelpColor string
	// NormalTitleColor is used for the title text of unselected list items.
	NormalTitleColor string
	// SelectedTitleColor is used for the title text of the selected list item.
	SelectedTitleColor string
	// ReadTitleColor is used for the title text of read articles in the articles list.
	ReadTitleColor string
}

// Theme holds the light and dark color sets. lipgloss.AdaptiveColor will
// select between them automatically based on the terminal background.
type Theme struct {
	Light ThemeColors
	Dark  ThemeColors
}

// Load reads theme configuration from Viper and returns a Theme, falling back
// to built-in defaults for any values not present in the config file.
func Load() Theme {
	setDefaults()

	theme := Theme{
		Light: ThemeColors{
			FocusedColor:            viper.GetString("theme.light.focused_color"),
			UnfocusedColor:          viper.GetString("theme.light.unfocused_color"),
			FocusedTitleTextColor:   viper.GetString("theme.light.focused_title_text_color"),
			UnfocusedTitleTextColor: viper.GetString("theme.light.unfocused_title_text_color"),
			HelpColor:               viper.GetString("theme.light.help_color"),
			NormalTitleColor:        viper.GetString("theme.light.normal_title_color"),
			SelectedTitleColor:      viper.GetString("theme.light.selected_title_color"),
			ReadTitleColor:          viper.GetString("theme.light.read_title_color"),
		},
		Dark: ThemeColors{
			FocusedColor:            viper.GetString("theme.dark.focused_color"),
			UnfocusedColor:          viper.GetString("theme.dark.unfocused_color"),
			FocusedTitleTextColor:   viper.GetString("theme.dark.focused_title_text_color"),
			UnfocusedTitleTextColor: viper.GetString("theme.dark.unfocused_title_text_color"),
			HelpColor:               viper.GetString("theme.dark.help_color"),
			NormalTitleColor:        viper.GetString("theme.dark.normal_title_color"),
			SelectedTitleColor:      viper.GetString("theme.dark.selected_title_color"),
			ReadTitleColor:          viper.GetString("theme.dark.read_title_color"),
		},
	}

	return theme
}

// setDefaults registers the built-in color defaults with Viper so that any
// key omitted from the config file still returns a sensible value.
func setDefaults() {
	viper.SetDefault("theme.light.focused_color", DefaultLightFocusedColor)
	viper.SetDefault("theme.light.unfocused_color", DefaultLightUnfocusedColor)
	viper.SetDefault("theme.light.focused_title_text_color", DefaultLightFocusedTitleTextColor)
	viper.SetDefault("theme.light.unfocused_title_text_color", DefaultLightUnfocusedTitleTextColor)
	viper.SetDefault("theme.light.help_color", DefaultLightHelpColor)
	viper.SetDefault("theme.light.normal_title_color", DefaultLightNormalTitleColor)
	viper.SetDefault("theme.light.selected_title_color", DefaultLightSelectedTitleColor)
	viper.SetDefault("theme.light.read_title_color", DefaultLightReadTitleColor)

	viper.SetDefault("theme.dark.focused_color", DefaultDarkFocusedColor)
	viper.SetDefault("theme.dark.unfocused_color", DefaultDarkUnfocusedColor)
	viper.SetDefault("theme.dark.focused_title_text_color", DefaultDarkFocusedTitleTextColor)
	viper.SetDefault("theme.dark.unfocused_title_text_color", DefaultDarkUnfocusedTitleTextColor)
	viper.SetDefault("theme.dark.help_color", DefaultDarkHelpColor)
	viper.SetDefault("theme.dark.normal_title_color", DefaultDarkNormalTitleColor)
	viper.SetDefault("theme.dark.selected_title_color", DefaultDarkSelectedTitleColor)
	viper.SetDefault("theme.dark.read_title_color", DefaultDarkReadTitleColor)
}
