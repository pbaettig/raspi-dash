package config

import (
	"image/color"
	"time"

	gofont "golang.org/x/image/font"
)

const (
	PlotSizeWidth       = 800
	PlotSizeHeight      = 240
	PlotTitleFontWeight = gofont.WeightBold
	PlotTitleFontSize   = 14
	PlotTitleFontStyle  = gofont.StyleNormal

	PlotUpdateInterval   = 1 * time.Second
	BackupUpdateInterval = 1 * time.Minute
	PlotDatapoints       = 12 * 3600
	PlotDataRange        = time.Duration(PlotDatapoints) * PlotUpdateInterval
)

var (
	PlotTitleFontColor = color.RGBA{44, 44, 144, 255}
)
