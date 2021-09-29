package assets

import (
	"embed"
)

//go:embed *.png *.ico *.webmanifest *.css
var FS embed.FS
