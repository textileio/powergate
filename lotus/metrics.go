package lotus

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	mLotusHeight = stats.Int64("lotus/height", "Height of Lotus node", "By")
	vHeight      = &view.View{
		Name:        "lotus/height_count",
		Measure:     mLotusHeight,
		Description: "Current height of Lotus node",
		Aggregation: view.LastValue(),
	}
)
