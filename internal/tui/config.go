package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
)

var (
	SelectedColor = tcell.ColorYellow
	NormalColor   = tcell.ColorWhite
	DeclinedColor = tcell.ColorRed
	MergedColor   = tcell.ColorYellow
)

var (
	StatusColumnId   = 4
	CommentsColumnId = 5
)

func initIconsMap(config *viper.Viper) map[string]string {
	iconsMap := map[string]string{
		"Title":            "TITLE",
		"ID":               "#",
		"User":             "AUTHOR",
		"Status":           "📖",
		"ChangesRequested": "✋",
		"Comment":          "💬",
		"Approval":         "✅",
		"Branch":           "🛫",
		"Merge":            "🛬",
	}

	if config.GetBool("general.useNerdFontIcons") {
		nerdIconsMaps := map[string]string{
			"Title":            "TITLE",
			"ID":               "",
			"User":             "",
			"Status":           "",
			"ChangesRequested": "",
			"Comment":          "󰆈",
			"Approval":         "󰱒",
			"Branch":           "",
			"Merge":            "",
		}

		for k := range nerdIconsMaps {
			iconsMap[k] = nerdIconsMaps[k]
		}
	}

	for k := range iconsMap {
		p := fmt.Sprintf("icons.%s", k)
		if icon := config.GetString(p); icon != "" {
			iconsMap[k] = icon
		}
	}

	return iconsMap
}
