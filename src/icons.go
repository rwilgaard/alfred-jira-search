package main

import (
	"fmt"
	"os"
	"strings"

	aw "github.com/deanishe/awgo"
)

func getIcon(query string) *aw.Icon {
    name := strings.ReplaceAll(query, " ", "_")
    if cfg.AltIcons {
        name += "-alt"
    }
    iconPath := fmt.Sprintf("icons/%s.png", name)
    icon := aw.IconWorkflow

    if _, err := os.Stat(iconPath); err == nil {
        icon = &aw.Icon{Value: iconPath}
    }

    return icon
}
