package flog

import (
	"github.com/hunterhug/golog"
	"path/filepath"
)

// InitLog Init the log
func InitLog(logFile string, debugLevel bool) {
	golog.SetOutputFile(filepath.Split(logFile))
	if debugLevel {
		golog.SetLevel(golog.DebugLevel)
	}
	golog.InitLogger()
}
